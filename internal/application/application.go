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
	"slices"
	"strconv"

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

func strtoui(s string) uint {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic("failed to convert string to uint")
	}
	return uint(value)
}

type OptionCallbacks map[string]func() error

type Application struct {
	args   []string
	engine *machine.Engine
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

func NewApplicationFromEngine(args []string, e *machine.Engine) *Application {
	app := &Application{
		args:   args,
		engine: e,
	}
	app.engine.Load()
	return app
}

func NewApplicationFromLMEString(args []string, s machine.LMEString) *Application {
	app := &Application{
		args:   args,
		engine: machine.NewEngine(),
	}
	app.engine.LoadFromLMEString(s)
	return app
}

func NewApplicationFromLMExternal(args []string, ext *machine.LMExternal) *Application {
	app := &Application{
		args:   args,
		engine: machine.NewEngine(),
	}
	app.engine.SetExternal(ext)
	return app
}

func (a *Application) Start() uint {
	err := a.ProcessOptions()
	if err != nil {
		log.Fatal(err)
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
		_, err := io.WriteString(os.Stdout, fmt.Sprintf(shebang, shebang))
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
	callbacks["stdin"] = func() error {
		a.engine.AddInput(machine.NewGramInputBuffer(a.engine, iOpt))
		return nil
	}

	var siOpt bool
	fs.BoolVar(&siOpt, "stdin", false, "stdin as input file")
	callbacks["stdin"] = func() error {
		a.engine.AddInput(machine.NewGramStdioFromEngine(a.engine))
		return nil
	}

	var lxOpt uint
	fs.UintVar(&lxOpt, "lexpri", 1, "lexical priority")
	callbacks["lexpri"] = func() error {
		a.engine.SetLexicalMismatchPriority(lxOpt)
		return nil
	}

	var bOpt uint
	fs.UintVar(&bOpt, "buffer", 1, "buffer length")
	callbacks["buffer"] = func() error {
		a.engine.SetBuffer(bOpt)
		return nil
	}

	var mrOpt uint
	fs.UintVar(&mrOpt, "max-repeat", 1, "max repeats")
	callbacks["max-repeat"] = func() error {
		a.engine.SetMaxRepeat(mrOpt)
		return nil
	}

	var mdOpt uint
	fs.UintVar(&mdOpt, "max-depth", 1, "max depth")
	callbacks["max-depth"] = func() error {
		a.engine.SetMaxDepth(mdOpt)
		return nil
	}

	var dwOpt uint
	fs.UintVar(&dwOpt, "dwidth", 1, "width for diagram (use before -t D)")
	callbacks["dwidth"] = func() error {
		a.engine.SetDisplayW(dwOpt)
		return nil
	}

	traceMap := map[string]uint{
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
  v  ARITHMETIC
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
Multiple options can be combined, e.g. -t m,s or -t m -t s`
	fs.Var(&tOpt, "trace", traceHelp)
	callbacks["trace"] = func() error {
		for _, option := range tOpt.Values {
			if flag, exists := traceMap[option]; exists {
				a.engine.SetTraceFlag(flag)
			} else {
				return errors.New("invalid trace option: " + option)
			}
		}
		return nil
	}

	err := fs.Parse(a.args[1:])
	if err != nil {
		return nil, err
	}

	// Handle positional arguments (input files)
	callbacks["files"] = func() error {
		for _, file := range fs.Args() {
			a.engine.AddInput(machine.NewGramInputFile(a.engine, file))
		}
		return nil
	}

	return callbacks, nil
}

func (a *Application) ApplyOptions(fs *flag.FlagSet, callbacks OptionCallbacks) error {
	var err error

	fs.Visit(func(f *flag.Flag) {
		if callback, exists := callbacks[f.Name]; exists {
			err = callback()
			if err != nil {
				return
			}
		} else {
			err = errors.New("Invalid trace option (no callback): " + f.Name)
			return
		}
	})

	if err != nil {
		return err
	}

	if callback, exists := callbacks["files"]; exists {
		err = callback()
	} else {
		return errors.New("No callback for files")
	}

	return err
}
