package application

import (
	"errors"
	"fmt"
	"io/ioutil"
	"languagemachine2/internal/machine"
	"languagemachine2/internal/options"
	"languagemachine2/internal/summary"
	"log"
	"os"
	"strconv"
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
	a.options.Add("-t", NewTraceOpt(args, e, 1, "--trace", "(--detail)", "trace options"))
	// a.Options.Add("-T",  NewNTraceOpt(args, e, 1, "--trace", "number", "trace options"));
	// a.Options.Add("-E",  NewEngOpt(args, e, 1, "--eng", "nz", "engine options"));
	// a.Options.Add("-d",  NewDModOpt(args, e, 1, "--dmodule", "file", "file for rules output as d module"));
	return a.options.Arguments(args)
}

func (a *Application) Start() uint {
	err := a.Arguments(a.args, a.engine)
	if err != nil {
		log.Fatal(err)
	}
	return a.engine.Start()
}

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

type TraceOpt struct {
	EngineOpt
	flag map[rune]uint
	what map[rune]string
}

func NewTraceOpt(args []string, e *machine.Engine, n uint, l, a, h string) *TraceOpt {
	to := &TraceOpt{
		EngineOpt: *NewEngineOpt(e, n, l, a, h),
		flag:      make(map[rune]uint),
		what:      make(map[rune]string),
	}

	to.setFlag('m', machine.MISMATCH, "MISMATCH")
	to.setFlag('s', machine.SYMBOLS, "SYMBOLS")
	to.setFlag('x', machine.CXSCOPE, "CXSCOPE")
	to.setFlag('c', machine.CVAR, "CVAR")
	to.setFlag('U', machine.LVAR, "LVAR")
	to.setFlag('r', machine.RVAR, "RVAR")
	to.setFlag('R', machine.RVAR_VAR, "RVAR_VAR")
	to.setFlag('X', machine.RVARSCOPE, "RVARSCOPE")
	to.setFlag('v', machine.REF, "REF")
	to.setFlag('V', machine.REFSCOPE, "REFSCOPE")
	to.setFlag('w', machine.REFVAR, "REFVAR")
	to.setFlag('e', machine.EACH, "EACH")
	to.setFlag('E', machine.EACHSCOPE, "EACHSCOPE")
	to.setFlag('f', machine.EACHREFVAR, "EACHREFVAR")
	to.setFlag('y', machine.DEBUG, "DEBUG")
	to.setFlag('A', machine.ACT, "ACT")
	to.setFlag('q', machine.APPLY, "APPLY")
	to.setFlag('v', machine.ARITHMETIC, "ARITHMETIC")
	to.setFlag('l', machine.RELATION, "RELATION")
	to.setFlag('S', machine.ASSIGN, "ASSIGN")
	to.setFlag('I', machine.INDEX, "INDEX")
	to.setFlag('L', machine.LOOP, "LOOP")
	to.setFlag('b', machine.LOAD, "LOAD")
	to.setFlag('d', machine.DIAGRAMT, "DIAGRAM text")
	to.setFlag('D', machine.DIAGRAM, "DIAGRAM")
	to.setFlag('G', machine.GRAMMAR, "GRAMMAR")
	to.setFlag('a', ^(machine.DIAGRAMT | machine.DIAGRAM), "all")
	to.setFlag('z', 0, "none")

	return to
}

func (to *TraceOpt) setFlag(x rune, v uint, s string) {
	to.flag[x] = v
	to.what[x] = s
}

func (to *TraceOpt) Explain(detail int) {
	to.EngineOpt.ExplainOption(0)
	if detail > 0 {
		for k := range to.flag {
			fmt.Printf("  %c %s\n", k, to.what[k])
		}
	}
}

func (to *TraceOpt) OptionAction(a, x string) (noAction bool, err error) {
	var n int
	for _, c := range string(x) {
		if v, ok := to.flag[c]; ok {
			to.engine.SetTraceFlag(v)
			n++
		}
	}
	if n == 0 {
		err = errors.New("bad arguments")
		return
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
	data, err := ioutil.ReadFile(x)
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
	data, err := ioutil.ReadFile(x)
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
		fmt.Println(shebang)
	case "-g":
		fmt.Println(goMain)
	}
	return
}
