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
	"runtime/trace"
	"slices"

	"github.com/sgreben/flagvar"
)

const shebang = `#! %s -r 
# Language Machine (C) 2005 Peri Hankey (mpah@users.sourceforge.net). Redistribution permitted subject to GNU GPLv2.
# The Language Machine is free software as defined by the Gnu GPL and comes with ABSOLUTELY NO WARRANTY.`

const goMain = `package main

import (
    "os"
    "languagemachine2/application"
)

func main() {
    args := os.Args
    app := application.NewApplication(args, lmdInit)
    result := app.Start()
    os.Exit(result)
}`

type OptionCallbacks map[string]func() error

type Application struct {
	args      []string
	engine    *machine.Engine
	traceStop func()
	order     []string // flag names in command-line order
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

func (v *orderedValue) IsBoolFlag() bool {
	b, ok := v.Value.(interface{ IsBoolFlag() bool })
	return ok && b.IsBoolFlag()
}

func NewApplication(args []string) *Application {
	app := &Application{
		args:   args,
		engine: machine.NewEngine(),
	}
	return app
}

func NewApplicationFromString(args []string, r string) *Application {
	app := &Application{
		args:   args,
		engine: machine.NewEngine(),
	}
	app.engine.LoadFromString(r)
	return app
}

func (a *Application) Start() int {
	err := a.ProcessOptions()
	if err != nil {
		log.Fatal(err)
	}

	if a.traceStop != nil {
		defer a.traceStop()
	}

	return a.engine.Start()
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
		fmt.Printf("%s: language machine version %s\n%s\n", a.args[0], summary.VersionString, summary.Summary)
		return nil
	}

	var lOpt bool
	fs.BoolVar(&lOpt, "license", false, "display license information")
	callbacks["license"] = func() error {
		fmt.Printf("%s\n", summary.Copyright)
		return nil
	}

	var sOpt string
	fs.StringVar(&sOpt, "shebang", "", "output shebang script header with PATH")
	callbacks["shebang"] = func() error {
		_, err := io.WriteString(os.Stdout, fmt.Sprintf(shebang, sOpt))
		return err
	}

	var gOpt bool
	fs.BoolVar(&gOpt, "gomain", false, "output Go language main program")
	callbacks["gomain"] = func() error {
		_, err := io.WriteString(os.Stdout, goMain)
		return err
	}

	var rOpt string
	fs.StringVar(&rOpt, "rules", "", "file of rules in .lmr format")
	callbacks["rules"] = func() error {
		data, err := os.ReadFile(rOpt)
		if err != nil {
			return err
		}
		a.engine.LoadFromString(string(data))
		return nil
	}

	var aOpt string
	fs.StringVar(&aOpt, "add", "", "additional rules in .lmr format")
	callbacks["add"] = func() error {
		data, err := os.ReadFile(aOpt)
		if err != nil {
			return err
		}
		a.engine.LoadFromStringReset(string(data), false)
		return nil
	}

	var oOpt string
	fs.StringVar(&oOpt, "output", "", "output file")
	callbacks["output"] = func() error {
		file, err := os.OpenFile(oOpt, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		//defer file.Close() // Ensure the file is closed when the program exits
		os.Stdout = file
		return nil
	}

	var eOpt string
	fs.StringVar(&eOpt, "errout", "", "error output")
	callbacks["errout"] = func() error {
		fmt.Println("errorout")
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
			f.Close()
			return err
		}
		a.traceStop = func() {
			trace.Stop()
			f.Close()
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
			a.engine.AppendInput(machine.NewGramInputFile(a.engine, file))
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
		return errors.New("No callback for files")
	}

	return err
}
