package application

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/trace"
	"strconv"
	"strings"

	"github.com/msorc/languagemachine2/internal/machine"
	"github.com/msorc/languagemachine2/internal/version"
)

const shebang = `#! %s -rules
# Language Machine (C) 2005 Peri Hankey (mpah@users.sourceforge.net). Redistribution permitted subject to GNU GPLv2.
# The Language Machine is free software as defined by the Gnu GPL and comes with ABSOLUTELY NO WARRANTY.
`

// Application runs the lm command line: it parses the options, then applies
// them to its engine in command-line order and runs the engine.
type Application struct {
	args      []string
	engine    *machine.Engine
	traceStop func()
	options   []option    // the flags given, in command-line order
	out       io.Writer   // standard output, or the -output file
	errOut    io.Writer   // standard error, for usage and error messages
	closers   []io.Closer // files to close when the run ends
	program   *Program    // rules built into the binary, if any
}

// option is one occurrence of a flag on the command line.
type option struct {
	name, value string
}

// Program is a ruleset built into the binary (see cmd/lmn2go). Its functions
// are registered and its rules loaded before the options run, so -rules
// replaces the rules and -add adds to them.
type Program struct {
	Rules string
	Funcs map[string]machine.ExtFn
}

// optionDef describes a flag. Every occurrence is recorded with its value
// when the command line is parsed, and apply runs once per occurrence, in
// command-line order, after parsing.
type optionDef struct {
	name, usage string
	isBool      bool
	parse       func(string) error // validates the value when the flag is parsed
	apply       func(a *Application, value string) error
}

// traceCode is a letter of the -trace option.
type traceCode struct {
	code string
	flag int
	name string
}

// traceCodes maps the letters of -trace to the engine's trace flags; the help
// text is generated from it. v is REF here, where the original used v for
// ARITHMETIC, which is o in the Go port.
var traceCodes = []traceCode{
	{"m", machine.MISMATCH, "MISMATCH"},
	{"s", machine.SYMBOLS, "SYMBOLS"},
	{"x", machine.CXSCOPE, "CXSCOPE"},
	{"c", machine.CVAR, "CVAR"},
	{"U", machine.LVAR, "LVAR"},
	{"r", machine.RVAR, "RVAR"},
	{"R", machine.RVAR_VAR, "RVAR_VAR"},
	{"X", machine.RVARSCOPE, "RVARSCOPE"},
	{"v", machine.REF, "REF"},
	{"V", machine.REFSCOPE, "REFSCOPE"},
	{"w", machine.REFVAR, "REFVAR"},
	{"e", machine.EACH, "EACH"},
	{"E", machine.EACHSCOPE, "EACHSCOPE"},
	{"f", machine.EACHREFVAR, "EACHREFVAR"},
	{"y", machine.DEBUG, "DEBUG"},
	{"A", machine.ACT, "ACT"},
	{"q", machine.APPLY, "APPLY"},
	{"o", machine.ARITHMETIC, "ARITHMETIC"},
	{"l", machine.RELATION, "RELATION"},
	{"S", machine.ASSIGN, "ASSIGN"},
	{"I", machine.INDEX, "INDEX"},
	{"L", machine.LOOP, "LOOP"},
	{"b", machine.LOAD, "LOAD"},
	{"d", machine.DIAGRAMT, "DIAGRAM text"},
	{"D", machine.DIAGRAM, "DIAGRAM"},
	{"G", machine.GRAMMAR, "GRAMMAR"},
	{"a", ^(machine.DIAGRAMT | machine.DIAGRAM), "all"},
	{"z", 0, "none"},
}

func traceCodeByLetter(code string) (traceCode, bool) {
	for _, c := range traceCodes {
		if c.code == code {
			return c, true
		}
	}
	return traceCode{}, false
}

func traceUsage() string {
	var b strings.Builder
	b.WriteString("trace codes, comma-separated or repeated:\n")
	for _, c := range traceCodes {
		fmt.Fprintf(&b, "  %s  %s\n", c.code, c.name)
	}
	b.WriteString("Multiple options can be combined, e.g. -trace m,s or -trace m -trace s")
	return b.String()
}

// parseTrace checks the codes of a -trace value.
func parseTrace(value string) error {
	for code := range strings.SplitSeq(value, ",") {
		if _, ok := traceCodeByLetter(code); !ok {
			return fmt.Errorf("unknown trace code %q", code)
		}
	}
	return nil
}

// applyTrace sets the flags of a -trace value in order; z turns every flag
// off.
func (a *Application) applyTrace(value string) error {
	for code := range strings.SplitSeq(value, ",") {
		c, ok := traceCodeByLetter(code)
		if !ok {
			return fmt.Errorf("unknown trace code %q", code)
		}
		if c.code == "z" {
			a.engine.UnsetTraceFlag(^0)
		} else {
			a.engine.SetTraceFlag(c.flag)
		}
	}
	return nil
}

const minDiagramWidth = 20

func parseInt(value string) error {
	_, err := strconv.Atoi(value)
	return err
}

// atoi converts a value that parseInt has already accepted.
func atoi(value string) int {
	n, _ := strconv.Atoi(value)
	return n
}

func (a *Application) openOutput(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	a.closers = append(a.closers, file)
	return file, nil
}

func loadRules(a *Application, name string, reset bool) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	if err := a.engine.LoadFromStringReset(string(data), reset); err != nil {
		return fmt.Errorf("%s:%w", name, err)
	}
	return nil
}

var optionDefs = []optionDef{
	{name: "version", usage: "display version information", isBool: true,
		apply: func(a *Application, _ string) error {
			_, err := fmt.Fprintf(a.out, "%s: language machine version %s\n%s\n", a.name(), version.Version, version.Summary)
			return err
		}},
	{name: "license", usage: "display license information", isBool: true,
		apply: func(a *Application, _ string) error {
			_, err := fmt.Fprintf(a.out, "%s\n", version.Copyright)
			return err
		}},
	{name: "shebang", usage: "output shebang script header with PATH",
		apply: func(a *Application, v string) error {
			_, err := fmt.Fprintf(a.out, shebang, v)
			return err
		}},
	{name: "rules", usage: "file of rules in .lm format; replaces rules loaded before it",
		apply: func(a *Application, v string) error { return loadRules(a, v, true) }},
	{name: "add", usage: "file of rules in .lm format, added to the rules loaded before it",
		apply: func(a *Application, v string) error { return loadRules(a, v, false) }},
	{name: "output", usage: "output file",
		apply: func(a *Application, v string) error {
			file, err := a.openOutput(v)
			if err != nil {
				return err
			}
			a.out = file
			a.engine.SetOutput(file)
			return nil
		}},
	{name: "errout", usage: "error output file (for err)",
		apply: func(a *Application, v string) error {
			file, err := a.openOutput(v)
			if err != nil {
				return err
			}
			a.engine.SetErrOutput(file)
			return nil
		}},
	{name: "input", usage: "string to process as input; may be repeated",
		apply: func(a *Application, v string) error {
			a.engine.AppendInput(machine.NewGramInputBuffer(a.engine, v))
			return nil
		}},
	{name: "stdin", usage: "stdin as input file", isBool: true,
		apply: func(a *Application, _ string) error {
			a.engine.AppendInput(machine.NewGramStdioFromEngine(a.engine))
			return nil
		}},
	{name: "lexpri", usage: fmt.Sprintf("lexical priority (default %d)", machine.LEXPRI), parse: parseInt,
		apply: func(a *Application, v string) error {
			a.engine.SetLexicalMismatchPriority(atoi(v))
			return nil
		}},
	{name: "buffer", usage: fmt.Sprintf("maximum length of the backtracking input buffer (default %d)", machine.MAXLENGTH), parse: parseInt,
		apply: func(a *Application, v string) error {
			a.engine.SetBuffer(atoi(v))
			return nil
		}},
	{name: "max-repeat", usage: "max repeats (default 0 = no limit)", parse: parseInt,
		apply: func(a *Application, v string) error {
			a.engine.SetMaxRepeat(atoi(v))
			return nil
		}},
	{name: "max-depth", usage: "max depth (default 0 = no limit)", parse: parseInt,
		apply: func(a *Application, v string) error {
			a.engine.SetMaxDepth(atoi(v))
			return nil
		}},
	{name: "dwidth", usage: fmt.Sprintf("width for diagram, at least %d (default 80); use before -trace D", minDiagramWidth),
		parse: func(v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			if n < minDiagramWidth {
				return fmt.Errorf("must be at least %d", minDiagramWidth)
			}
			return nil
		},
		apply: func(a *Application, v string) error {
			a.engine.SetDisplayW(atoi(v))
			return nil
		}},
	{name: "trace", usage: traceUsage(), parse: parseTrace,
		apply: (*Application).applyTrace},
	{name: "trace-out", usage: "write a Go runtime execution trace to file",
		apply: func(a *Application, v string) error {
			f, err := os.Create(v)
			if err != nil {
				return err
			}
			if err := trace.Start(f); err != nil {
				_ = f.Close()
				return err
			}
			a.traceStop = func() {
				trace.Stop()
				if err := f.Close(); err != nil {
					a.report(err)
				}
			}
			return nil
		}},
}

// NewApplication runs the command line args (args[0] is the program name)
// on the process's standard output and error.
func NewApplication(args []string) *Application {
	return newApplication(args, os.Stdout, os.Stderr)
}

func newApplication(args []string, stdout, stderr io.Writer) *Application {
	app := &Application{
		args:   args,
		engine: machine.NewEngine(),
		out:    stdout,
		errOut: stderr,
	}
	app.engine.SetOutput(stdout)
	app.engine.SetErrOutput(stderr)
	return app
}

// NewProgramApplication runs the command line args with the program p
// built in, on the given standard output and error.
func NewProgramApplication(args []string, p *Program, stdout, stderr io.Writer) *Application {
	app := newApplication(args, stdout, stderr)
	app.program = p
	return app
}

func (a *Application) name() string {
	return filepath.Base(a.args[0])
}

// Start runs the command line and returns the exit status. Errors are
// reported on stderr.
func (a *Application) Start() int {
	defer func() {
		if a.traceStop != nil {
			a.traceStop()
		}
	}()
	defer a.close()

	if err := a.processOptions(); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		a.report(err)
		return 1
	}

	status, err := a.engine.Start()
	if err != nil {
		a.report(err)
	}
	return status
}

func (a *Application) report(err error) {
	_ = a.engine.Flush()
	_, _ = fmt.Fprintf(a.errOut, "%s: %v\n", a.name(), err)
}

func (a *Application) close() {
	_ = a.engine.Flush()
	for _, c := range a.closers {
		if err := c.Close(); err != nil {
			a.report(err)
		}
	}
}

// processOptions loads a built-in program, parses the command line and
// applies the options in the order they were given; the positional input
// files come last.
func (a *Application) processOptions() error {
	if p := a.program; p != nil {
		for name, fn := range p.Funcs {
			a.engine.External().Set(name, fn)
		}
		if err := a.engine.LoadFromString(p.Rules); err != nil {
			return err
		}
	}
	fs := flag.NewFlagSet(a.name(), flag.ContinueOnError)
	fs.SetOutput(a.errOut)
	defs := make(map[string]*optionDef, len(optionDefs))
	for i := range optionDefs {
		d := &optionDefs[i]
		defs[d.name] = d
		record := func(value string) error {
			if d.parse != nil {
				if err := d.parse(value); err != nil {
					return err
				}
			}
			a.options = append(a.options, option{d.name, value})
			return nil
		}
		if d.isBool {
			fs.BoolFunc(d.name, d.usage, record)
		} else {
			fs.Func(d.name, d.usage, record)
		}
	}
	if err := fs.Parse(a.args[1:]); err != nil {
		return err
	}
	for _, o := range a.options {
		if err := defs[o.name].apply(a, o.value); err != nil {
			return err
		}
	}
	for _, file := range fs.Args() {
		g, err := machine.NewGramInputFile(a.engine, file)
		if err != nil {
			return err
		}
		a.engine.AppendInput(g)
	}
	return nil
}
