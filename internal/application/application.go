package application

import (
	"errors"
	"fmt"
	"io"
	"languagemachine2/internal/machine"
	"languagemachine2/internal/options"
	"languagemachine2/internal/summary"
	"log"
	"os"
	"strconv"
	"flag"
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

type Application struct {
	options options.OptArgs
	args    []string
	engine  *machine.Engine
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

func (a *Application) Arguments(args []string, e *machine.Engine) error {
	a.options = *options.NewOptArgs()
	a.options.Add("-v", NewHelpOpt(args, e, 0, "--version", " ", "display version information"))
	a.options.Add("-h", NewHelpOpt(args, e, 0, "--help", " ", "usage summary"))
	a.options.Add("-H", NewHelpOpt(args, e, 0, "--detail", " ", "more detailed usage"))
	a.options.Add("-L", NewHelpOpt(args, e, 0, "--license", " ", "display license information"))
	a.options.Add("-s", NewMainOpt(args, e, 1, "--shebang", "path", "output shebang script header with PATH"))
	a.options.Add("-g", NewMainOpt(args, e, 0, "--gomain", " ", "output Go language main program"))
	a.options.Add("-r", NewRuleSOpt(args, e, 1, "--rules", "file", "file of rules in .lmr format"))
	a.options.Add("-a", NewRuleXOpt(args, e, 1, "--add", "file", "additional rules in .lmr format"))
	a.options.Add("-o", NewOutOpt(args, e, 1, "--output", "file", "output file"))
	a.options.Add("-e", NewErrOpt(args, e, 1, "--errout", "file", "error output"))
	a.options.Add("-i", NewInputOpt(args, e, 1, "--input", "string", "string to process as input"))
	a.options.Add("-", NewFileOpt(args, e, 0, "", "", "stdin as input file"))
	a.options.Add("-l", NewLexPriOpt(args, e, 1, "--lexpri", "number", "lexical priority"))
	a.options.Add("-b", NewBufferOpt(args, e, 1, "--buffer", "number", "buffer length"))
	a.options.Add("-N", NewMRTraceOpt(args, e, 1, "--max-repeat", "number", "max repeats"))
	a.options.Add("-D", NewMDTraceOpt(args, e, 1, "--max-depth", "number", "max depth"))
	a.options.Add("-W", NewDWidthOpt(args, e, 1, "--dwidth", "number", "width for diagram (use before -t D)"))
	a.options.Add("...", NewFileOpt(args, e, 0, "", "files", "input files"))
	// a.Options.Add("-E",  NewEngOpt(args, e, 1, "--eng", "nz", "engine options"));
	return a.options.Arguments(args)
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
	callbacks := a.ConfigureOptions(fs)
	err := a.ApplyOptions(fs, callbacks)
	if err != nil {
		return err
	}
	
	return nil
}

func (a *Application) ConfigureOptions(fs *flag.FlagSet) OptionCallbacks {
	var callbacks = make(OptionCallbacks)

	trace := flagvar.EnumsCSV{Choices: []string{"m", "s", "x", "c", "U", "r", "R", "X", "v", "V", "w", "e", "E", "f", "y", "A", "q", "v", "l", "S", "I", "L", "b", "d", "D", "G", "a", "z"}, CaseSensitive: true, Accumulate: true}
	traceHelp := trace.Help() +
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
Multiple options can be combined, e.g. -t m,s or -t m -t s
`
	fs.Var(&trace, "trace", traceHelp)
	callbacks["trace"] = func() error {
		return a.applyTraceOptions(trace.Values)
	}

	fs.Parse(a.args[1:])

	return callbacks
}

func (a *Application) ApplyOptions(fs *flag.FlagSet, callbacks OptionCallbacks) error {
	var err error
	
	fs.Visit(func(f *flag.Flag) {
		if callback, exists := callbacks[f.Name]; exists {
			err = callback()
			if err != nil {
				return
			}
		}
	})
	
	return err
}

func (a *Application) applyTraceOptions(options []string) error {
	for _, option := range options {
		switch option {
		case "m":
			a.engine.SetTraceFlag(machine.MISMATCH)
		case "s":
			a.engine.SetTraceFlag(machine.SYMBOLS)
		case "x":
			a.engine.SetTraceFlag(machine.CXSCOPE)
		case "c":
			a.engine.SetTraceFlag(machine.CVAR)
		case "U":
			a.engine.SetTraceFlag(machine.LVAR)
		case "r":
			a.engine.SetTraceFlag(machine.RVAR)
		case "R":
			a.engine.SetTraceFlag(machine.RVAR_VAR)
		case "X":
			a.engine.SetTraceFlag(machine.RVARSCOPE)
		case "v":
			a.engine.SetTraceFlag(machine.REF)
		case "V":
			a.engine.SetTraceFlag(machine.REFSCOPE)
		case "w":
			a.engine.SetTraceFlag(machine.REFVAR)
		case "e":
			a.engine.SetTraceFlag(machine.EACH)
		case "E":
			a.engine.SetTraceFlag(machine.EACHSCOPE)
		case "f":
			a.engine.SetTraceFlag(machine.EACHREFVAR)
		case "y":
			a.engine.SetTraceFlag(machine.DEBUG)
		case "A":
			a.engine.SetTraceFlag(machine.ACT)
		case "q":
			a.engine.SetTraceFlag(machine.APPLY)
		case "l":
			a.engine.SetTraceFlag(machine.RELATION)
		case "S":
			a.engine.SetTraceFlag(machine.ASSIGN)
		case "I":
			a.engine.SetTraceFlag(machine.INDEX)
		case "L":
			a.engine.SetTraceFlag(machine.LOOP)
		case "b":
			a.engine.SetTraceFlag(machine.LOAD)
		case "d":
			a.engine.SetTraceFlag(machine.DIAGRAMT)
		case "D":
			a.engine.SetTraceFlag(machine.DIAGRAM)
		case "G":
			a.engine.SetTraceFlag(machine.GRAMMAR)
		case "a":
			a.engine.SetTraceFlag(^(machine.DIAGRAMT | machine.DIAGRAM))
		case "z":
			a.engine.SetTraceFlag(0)
		default:
			return errors.New("invalid trace option: " + string(option))
		}
	}
	return nil
}

type OptionCallbacks map[string]func() (error)

type EngineOpt struct {
	options.OptArg
	engine *machine.Engine
}

func NewEngineOpt(e *machine.Engine, n uint, l, x, u string) *EngineOpt {
	return &EngineOpt{
		OptArg: *options.NewOptArg(n, l, x, u),
		engine: e,
	}
}

type HelpOpt struct {
	EngineOpt
	args []string
}

func NewHelpOpt(argv []string, e *machine.Engine, n uint, l, x, u string) *HelpOpt {
	return &HelpOpt{
		EngineOpt: *NewEngineOpt(e, n, l, x, u),
		args:      argv,
	}
}

func (h *HelpOpt) OptionAction(a, x string) (noActon bool, err error) {
	switch h.S {
	case "-v":
		fmt.Printf("%s: language machine version %s\n%s\n", h.args[0], summary.VersionString, summary.Summary)
	case "-h":
		fmt.Printf("%s\n", summary.Summary)
		h.Options.Usage(0)
	case "-H":
		fmt.Printf("%s\n", summary.Summary)
		h.Options.Usage(1)
	case "-L":
		fmt.Printf("%s\n", summary.Copyright)
	default:
		noActon = true
	}

	return
}


type EngOpt struct {
	EngineOpt
	flag map[rune]uint
}

func NewEngOpt(args []string, e *machine.Engine, n uint, l, a, h string) *EngOpt {
	opt := &EngOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h),
		flag:      make(map[rune]uint),
	}
	// algorithm options -          //     default: try l:r r:z z:z z:l
	opt.flag['n'] = machine.ZLONLY  // = 0x0000001; try l:r r:z z:l (no z:z resolution)
	opt.flag['z'] = machine.ZZFINAL // = 0x0000002; try l:r r:z z:l z:z in that order
	return opt
}

func (eo *EngOpt) OptionAction(a, x string) {
	for _, c := range x {
		if val, exists := eo.flag[c]; exists {
			eo.engine.SetOption(val)
		}
	}
}

type RuleXOpt struct {
	EngineOpt
}

func NewRuleXOpt(args []string, e *machine.Engine, n uint, l, a, h string) *RuleXOpt {
	return &RuleXOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (rxo *RuleXOpt) OptionAction(a, x string) (noAction bool, err error) {
	data, err := os.ReadFile(x)
	if err != nil {
		return
	}
	rxo.engine.LoadFromStringReset(string(data), false)
	return
}

type RuleSOpt struct {
	EngineOpt
}

func NewRuleSOpt(args []string, e *machine.Engine, n uint, l, a, h string) *RuleSOpt {
	return &RuleSOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (rso *RuleSOpt) OptionAction(a, x string) (noAction bool, err error) {
	data, err := os.ReadFile(x)
	if err != nil {
		return
	}
	rso.engine.LoadFromString(string(data))
	return
}

type GoModOpt struct {
	EngineOpt
}

func NewGoModOpt(args []string, e *machine.Engine, n uint, l, a, h string) *GoModOpt {
	return &GoModOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (gomo *GoModOpt) OptionAction(a, x string) (noAction bool, err error) {
	noAction = true
	return
}

type OutOpt struct {
	EngineOpt
}

func NewOutOpt(args []string, e *machine.Engine, n uint, l, a, h string) *OutOpt {
	return &OutOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (oo *OutOpt) OptionAction(a, x string) (noAction bool, err error) {
	file, err := os.OpenFile(x, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to open file: %s\n", err)
		return
	}
	//defer file.Close() // Ensure the file is closed when the program exits
	os.Stdout = file
	return
}

type ErrOpt struct {
	EngineOpt
}

func NewErrOpt(args []string, e *machine.Engine, n uint, l, a, h string) *ErrOpt {
	return &ErrOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (eo *ErrOpt) OptionAction(a, x string) (noAction bool, err error) {
	fmt.Println("e")
	return
}

type FileOpt struct {
	EngineOpt
}

func NewFileOpt(args []string, e *machine.Engine, n uint, l, a, h string) *FileOpt {
	return &FileOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (fo *FileOpt) OptionAction(a, x string) (noAction bool, err error) {
	if x == "-" {
		fo.engine.AddInput(machine.NewGramInputFromEngine(fo.engine))
	} else {
		fo.engine.AddInput(machine.NewGramInputFile(fo.engine, x))
	}
	return
}

type InputOpt struct {
	EngineOpt
}

func NewInputOpt(args []string, e *machine.Engine, n uint, l, a, h string) *InputOpt {
	return &InputOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (io *InputOpt) OptionAction(a, x string) (noAction bool, err error) {
	io.engine.AddInput(machine.NewGramInputBuffer(io.engine, x))
	return
}

type BufferOpt struct {
	EngineOpt
}

func NewBufferOpt(args []string, e *machine.Engine, n uint, l, a, h string) *BufferOpt {
	return &BufferOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (bo *BufferOpt) OptionAction(a, x string) (noAction bool, err error) {
	bo.engine.SetBuffer(strtoui(x))
	return
}

type NTraceOpt struct {
	EngineOpt
}

func NewNTraceOpt(args []string, e *machine.Engine, n uint, l, a, h string) *NTraceOpt {
	return &NTraceOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (nto *NTraceOpt) OptionAction(a, x string) (noAction bool, err error) {
	nto.engine.SetTraceFlag(strtoui(x))
	return
}

type DWidthOpt struct {
	EngineOpt
}

func NewDWidthOpt(args []string, e *machine.Engine, n uint, l, a, h string) *DWidthOpt {
	return &DWidthOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (dwo *DWidthOpt) OptionAction(a, x string) (noAction bool, err error) {
	dwo.engine.SetDisplayW(strtoui(x))
	return
}

type MRTraceOpt struct {
	EngineOpt
}

func NewMRTraceOpt(args []string, e *machine.Engine, n uint, l, a, h string) *MRTraceOpt {
	return &MRTraceOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (mrto *MRTraceOpt) OptionAction(a, x string) (noAction bool, err error) {
	mrto.engine.SetMaxRepeat(strtoui(x))
	return
}

type MDTraceOpt struct {
	EngineOpt
}

func NewMDTraceOpt(args []string, e *machine.Engine, n uint, l, a, h string) *MDTraceOpt {
	return &MDTraceOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (mdto *MDTraceOpt) OptionAction(a, x string) (noAction bool, err error) {
	mdto.engine.SetMaxDepth(strtoui(x))
	return
}

type LexPriOpt struct {
	EngineOpt
}

func NewLexPriOpt(args []string, e *machine.Engine, n uint, l, a, h string) *LexPriOpt {
	return &LexPriOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h)}
}

func (lpo *LexPriOpt) OptionAction(a, x string) (noAction bool, err error) {
	lpo.engine.SetLexicalMismatchPriority(strtoui(x))
	return
}

type MainOpt struct {
	EngineOpt
	args []string
}

func NewMainOpt(argv []string, e *machine.Engine, n uint, l, a, h string) *MainOpt {
	return &MainOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h),
		args:      argv,
	}
}

func (mo *MainOpt) OptionAction(a, x string) (noAction bool, err error) {
	switch mo.S {
	case "-s":
		if _, err = io.WriteString(os.Stdout, shebang); err != nil {
			return
		}
	case "-g":
		if _, err = io.WriteString(os.Stdout, goMain); err != nil {
			return
		}
	}
	return
}
