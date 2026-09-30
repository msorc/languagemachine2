package application

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"languagemachine2/internal/machine"
	"languagemachine2/internal/summary"
	"log"
	"maps"
	"os"
	"path/filepath"
	"runtime/trace"
	"slices"

	"github.com/sgreben/flagvar"
)

const shebang = `#! %s -r 
# Language Machine (C) 2005 Peri Hankey (mpah@users.sourceforge.net). Redistribution permitted subject to GNU GPLv2.
# The Language Machine is free software as defined by the Gnu GPL and comes with ABSOLUTELY NO WARRANTY.`

type OptionCallbacks map[string]func() error

type Application struct {
	args      []string
	engine    *machine.Engine
	traceStop func()
	order     []string    // flag names in command-line order
	out       io.Writer   // standard output, or the -output file
	closers   []io.Closer // files to close when the run ends
}

// orderedValue records the order in which flags appear on the command line,
// so their callbacks can be run in that order (flag.Visit is alphabetical).
type orderedValue struct {
	flag.Value
	name  string
	order *[]string
}

func (v *orderedValue) Set(s string) error {
	if err := v.Value.Set(s); err != nil {
		return err
	}
	if !slices.Contains(*v.order, v.name) {
		*v.order = append(*v.order, v.name)
	}
	return nil
}

// String is also called by flag on a zero orderedValue, to find defaults.
func (v *orderedValue) String() string {
	if v == nil || v.Value == nil {
		return ""
	}
	return v.Value.String()
}

func (v *orderedValue) IsBoolFlag() bool {
	b, ok := v.Value.(interface{ IsBoolFlag() bool })
	return ok && b.IsBoolFlag()
}

func NewApplication(args []string) *Application {
	app := &Application{
		args:   args,
		engine: machine.NewEngine(),
		out:    os.Stdout,
	}
	return app
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

	if err := a.ProcessOptions(); err != nil {
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
	fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(a.args[0]), err)
}

func (a *Application) close() {
	_ = a.engine.Flush()
	for _, c := range a.closers {
		if err := c.Close(); err != nil {
			log.Print(err)
		}
	}
}

func (a *Application) ProcessOptions() error {
	fs := flag.NewFlagSet("languagemachine2", flag.ContinueOnError)
	callbacks, err := a.ConfigureOptions(fs)
	if err != nil {
		return err
	}
	err = a.ApplyOptions(fs, callbacks)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) ConfigureOptions(fs *flag.FlagSet) (OptionCallbacks, error) {
	var callbacks = make(OptionCallbacks)

	var vOpt bool
	fs.BoolVar(&vOpt, "version", false, "display version information")
	callbacks["version"] = func() error {
		_, err := fmt.Fprintf(a.out, "%s: language machine version %s\n%s\n", a.args[0], summary.VersionString, summary.Summary)
		return err
	}

	var lOpt bool
	fs.BoolVar(&lOpt, "license", false, "display license information")
	callbacks["license"] = func() error {
		_, err := fmt.Fprintf(a.out, "%s\n", summary.Copyright)
		return err
	}

	var sOpt string
	fs.StringVar(&sOpt, "shebang", "", "output shebang script header with PATH")
	callbacks["shebang"] = func() error {
		_, err := fmt.Fprintf(a.out, shebang, sOpt)
		return err
	}

	var rOpt string
	fs.StringVar(&rOpt, "rules", "", "file of rules in .lmr format")
	callbacks["rules"] = func() error {
		data, err := os.ReadFile(rOpt)
		if err != nil {
			return err
		}
		return a.engine.LoadFromString(string(data))
	}

	var aOpt string
	fs.StringVar(&aOpt, "add", "", "additional rules in .lmr format")
	callbacks["add"] = func() error {
		data, err := os.ReadFile(aOpt)
		if err != nil {
			return err
		}
		return a.engine.LoadFromStringReset(string(data), false)
	}

	var oOpt string
	fs.StringVar(&oOpt, "output", "", "output file")
	callbacks["output"] = func() error {
		file, err := os.OpenFile(oOpt, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		a.closers = append(a.closers, file)
		a.out = file
		a.engine.SetOutput(file)
		return nil
	}

	var eOpt string
	fs.StringVar(&eOpt, "errout", "", "error output file (for err)")
	callbacks["errout"] = func() error {
		file, err := os.OpenFile(eOpt, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		a.closers = append(a.closers, file)
		a.engine.SetErrOutput(file)
		return nil
	}

	var iOpt string
	fs.StringVar(&iOpt, "input", "", "string to process as input")
	callbacks["input"] = func() error {
		a.engine.AppendInput(machine.NewGramInputBuffer(a.engine, iOpt))
		return nil
	}

	var siOpt bool
	fs.BoolVar(&siOpt, "stdin", false, "stdin as input file")
	callbacks["stdin"] = func() error {
		a.engine.AppendInput(machine.NewGramStdioFromEngine(a.engine))
		return nil
	}

	var lxOpt int
	fs.IntVar(&lxOpt, "lexpri", machine.LEXPRI, "lexical priority")
	callbacks["lexpri"] = func() error {
		a.engine.SetLexicalMismatchPriority(lxOpt)
		return nil
	}

	var bOpt int
	fs.IntVar(&bOpt, "buffer", machine.MAXLENGTH, "maximum length of the backtracking input buffer")
	callbacks["buffer"] = func() error {
		a.engine.SetBuffer(bOpt)
		return nil
	}

	var mrOpt int
	fs.IntVar(&mrOpt, "max-repeat", 0, "max repeats (0 = no limit)")
	callbacks["max-repeat"] = func() error {
		a.engine.SetMaxRepeat(mrOpt)
		return nil
	}

	var mdOpt int
	fs.IntVar(&mdOpt, "max-depth", 0, "max depth (0 = no limit)")
	callbacks["max-depth"] = func() error {
		a.engine.SetMaxDepth(mdOpt)
		return nil
	}

	var dwOpt int
	fs.IntVar(&dwOpt, "dwidth", 80, "width for diagram (use before -trace D)")
	callbacks["dwidth"] = func() error {
		a.engine.SetDisplayW(dwOpt)
		return nil
	}

	traceMap := map[string]int{
		"m": machine.MISMATCH, "s": machine.SYMBOLS, "x": machine.CXSCOPE,
		"c": machine.CVAR, "U": machine.LVAR, "r": machine.RVAR, "R": machine.RVAR_VAR, "X": machine.RVARSCOPE,
		"v": machine.REF, "V": machine.REFSCOPE, "w": machine.REFVAR,
		"e": machine.EACH, "E": machine.EACHSCOPE, "f": machine.EACHREFVAR, "y": machine.DEBUG, "A": machine.ACT,
		"q": machine.APPLY, "l": machine.RELATION, "S": machine.ASSIGN, "I": machine.INDEX, "L": machine.LOOP, "b": machine.LOAD,
		"d": machine.DIAGRAMT, "D": machine.DIAGRAM, "G": machine.GRAMMAR, "a": ^(machine.DIAGRAMT | machine.DIAGRAM), "z": 0,
	}
	tOpt := flagvar.EnumsCSV{Choices: slices.Collect(maps.Keys(traceMap)), CaseSensitive: true, Accumulate: true}
	traceHelp := tOpt.Help() +
		`
Trace options:
  m  MISMATCH
  s  SYMBOLS
  x  CXSCOPE
  c  CVAR
  U  LVAR
  r  RVAR
  R  RVAR_VAR
  X  RVARSCOPE
  v  REF
  V  REFSCOPE
  w  REFVAR
  e  EACH
  E  EACHSCOPE
  f  EACHREFVAR
  y  DEBUG
  A  ACT
  q  APPLY
  l  RELATION
  S  ASSIGN
  I  INDEX
  L  LOOP
  b  LOAD
  d  DIAGRAM text
  D  DIAGRAM
  G  GRAMMAR
  a  all
  z  none
Multiple options can be combined, e.g. -trace m,s or -trace m -trace s`
	fs.Var(&tOpt, "trace", traceHelp)
	callbacks["trace"] = func() error {
		for _, option := range tOpt.Values {
			if option == "z" {
				a.engine.UnsetTraceFlag(^0)
			} else if flag, exists := traceMap[option]; exists {
				a.engine.SetTraceFlag(flag)
			} else {
				return errors.New("invalid trace option: " + option)
			}
		}
		return nil
	}

	var traceOut string
	fs.StringVar(&traceOut, "trace-out", "", "write execution trace to file")
	callbacks["trace-out"] = func() error {
		f, err := os.Create(traceOut)
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
				log.Print(err)
			}
		}
		return nil
	}

	fs.VisitAll(func(f *flag.Flag) {
		f.Value = &orderedValue{Value: f.Value, name: f.Name, order: &a.order}
	})

	err := fs.Parse(a.args[1:])
	if err != nil {
		return nil, err
	}

	// Handle positional arguments (input files)
	callbacks["files"] = func() error {
		for _, file := range fs.Args() {
			g, err := machine.NewGramInputFile(a.engine, file)
			if err != nil {
				return err
			}
			a.engine.AppendInput(g)
		}
		return nil
	}

	return callbacks, nil
}

func (a *Application) ApplyOptions(fs *flag.FlagSet, callbacks OptionCallbacks) error {
	var err error

	for _, name := range a.order {
		callback, exists := callbacks[name]
		if !exists {
			return errors.New("invalid option (no callback): " + name)
		}
		if err = callback(); err != nil {
			return err
		}
	}

	if callback, exists := callbacks["files"]; exists {
		err = callback()
	} else {
		return errors.New("no callback for files")
	}

	return err
}
