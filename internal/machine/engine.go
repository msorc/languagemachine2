package machine

import (
	"fmt"
	"languagemachine2/internal/utils"

	"github.com/liyue201/gostl/ds/list/bidlist"
)

const (
	// Engine
	LEXPRI    uint = 1000 // default value for lexical priority
	ZLONLY    uint = 1    // (not used) resolution: try r:l r:z z:l
	ZZFINAL   uint = 2    // (not used) resolution: try r:l r:z z:l z:z
	MAXLENGTH uint = 64 * 1024
)

var theZlm *ZLM

func TxE(s string, x Element) Element {
	var xtrace string
	if x != nil {
		xtrace = x.ToTrace()
	} else {
		xtrace = "---"
	}
	fmt.Printf("\t%6s:     %p %24s\n", s, x, xtrace)
	return x
}

func TxV(r, s string, w VarElement) Element {
	var n string

	if w != nil {
		n = w.ToDump()
	} else {
		n = "---"
	}

	fmt.Printf("\t%6s:%4s %p %24s %p %p %p %p\n", r, s, w, n, w.Value(), w.Variables(), w.ScopeReferenceContext(), w.ScopeVariables())

	return w
}

func theNull() Element { return theZlm }

type LMEString func(*Stream) GenMode
type LMNFunc func(*Stream) Element
type LMDString func(*Stream) GenMode

// main parsing engine: everything needed to load and apply grammars
type Engine struct {
	state         *State             // state at start of a new context
	contextsCount uint               // count of new contexts used to give each a unique identity
	lhsContext    ContextHolder // lhs context stack for mismatch events being resolved
	rhsContext    ContextHolder // rhs context stack for rhs of rules that have matched

	lhsStream *Stream // lhs registers
	rhsStream *Stream // rhs registers
	lhsMode   GenMode // lhs registers, LZMode
	rhsMode   GenMode // rhs registers, RZMode
	// Lhs       GenMode           // lhs element mode generator
	// Rhs       GenMode           // rhs element mode generator
	rsLastMatchElement Element // element resulting from last match

	grammars   *Selector // table of grammars selected by symbol
	oneGrammar *Grammar  // initial grammar

	terminalSymbols    *Dict   // terminal symbols
	nonTerminalSymbols *Dict   // non-terminal symbols
	varSymbols         *Dict   // variables
	userSymbols        *Dict   // user symbols - guaranteed not to match system symbols
	functionSymbols    *Dict   // primitive operator symbols
	predefinedSymbols  *Predef // predefined symbols with special significance

	loader *Loader // rule loader

	inputs bidlist.List[GrammarIO] // stack of input sources
	input  GrammarIO               // current input

	rhsBuffer *RZBuffer // circular buffer at outermost level of rhs

	flagErrors uint // incremented by flagSym
	warnErrors uint // incremented by warnSym

	externalSystem *LMExternal // external interfaces

	tracer *Tracer // trace handler - null if no tracing required

	display      *Diagram // to display trace as diagram
	displayWidth uint     // width for diagram display

	options                 uint // engine control options
	maxDepth                uint // limit on analysis recursion depth - zero means no limit
	maxRepeat               uint // limit on repetition at repeat     - zero means no limit
	rhsOffset               uint // offset applied to input position
	bufferLength            uint // default size of rhz circular buffer
	maxLength               uint // default size of rhz circular buffer
	lexicalMismatchPriority uint // artificial priority of context at lexical mismatch
}

func NewEngine() *Engine {
	return NewEngineFromLength(MAXLENGTH)
}

func NewEngineFromLength(len uint) *Engine {
	theZlm = NewZLM("null")
	e := &Engine{
		maxLength:          len,
		displayWidth:       80,
		bufferLength:       1024,
		functionSymbols:    NewDict(),
		terminalSymbols:    NewDict(),
		nonTerminalSymbols: NewDict(),
		varSymbols:         NewDict(),
		userSymbols:        NewDict(),
		predefinedSymbols:  NewPredef(),
		externalSystem:     NewLMExternal(),
		grammars:           NewSelector(),
		input:              NewGramStdio(), //+ do we need it?
	}
	e.rhsBuffer = NewRZBuffer(make([]Element, e.bufferLength), e.maxLength)
	e.state = NewState(e, nil, nil, nil, e.input, 0, 0, 0, e.contextsCount)
	e.contextsCount++
	e.lhsContext = NewContextFromState(LHContext, e.state)
	e.rhsContext = NewContextFromState(RHContext, e.state)
	e.lhsStream = NewStream(e, "lh", 0)
	e.rhsStream = NewStream(e, "rh", 0)
	e.lhsMode = NewLZModeFromContext(e.lhsContext, e.lhsStream)
	e.lhsStream.mode = e.lhsMode
	e.rhsMode = NewRZModeFromContext(e.rhsContext, e.rhsStream)
	e.rhsStream.mode = e.rhsMode
	e.SetLexicalMismatchPriority(LEXPRI)
	return e
}

// --- symbols
func (e *Engine) defineSymbols() {
	// e.nonTerminalSymbols.UniqueE(NewZzz("_voidv"))
	// e.nonTerminalSymbols.UniqueE(NewSym("__"))

	e.nonTerminalSymbols.UniqueE(theNull())
	e.predefinedSymbols.zlm = e.varSymbols.UniqueE(theNull())

	e.nonTerminalSymbols.UniqueE(NewSym("start"))
	e.nonTerminalSymbols.UniqueE(NewSym("eof"))
	e.nonTerminalSymbols.UniqueE(NewSpSym("sp"))
	e.nonTerminalSymbols.UniqueE(NewNlSym("nl"))
	e.nonTerminalSymbols.UniqueE(NewRepnSym("repeatN"))
	e.nonTerminalSymbols.UniqueE(NewAnything("anything"))
	e.nonTerminalSymbols.UniqueE(NewAnySym("nonTerminal"))
	e.nonTerminalSymbols.UniqueE(NewAnyChr("terminal"))
	e.nonTerminalSymbols.UniqueE(NewUriSym("uri"))
	e.nonTerminalSymbols.UniqueE(NewUrdSym("urd"))
	e.nonTerminalSymbols.UniqueE(NewOutSym("out"))
	e.nonTerminalSymbols.UniqueE(NewErrSym("err"))
	e.nonTerminalSymbols.UniqueE(NewLnoSym("lineNo"))
	e.nonTerminalSymbols.UniqueE(NewIfnSym("fileName"))
	e.nonTerminalSymbols.UniqueE(NewFlagSym("flagError"))
	e.nonTerminalSymbols.UniqueE(NewWarnSym("warnError"))

	e.functionSymbols.UniqueE(NewSym("mark"))

	e.predefinedSymbols.appendFn = e.functionSymbols.UniqueE(NewAppendXSym("append"))
	e.predefinedSymbols.repeatFn = e.nonTerminalSymbols.UniqueE(NewRepSym("repeat"))
	e.predefinedSymbols.optionFn = e.nonTerminalSymbols.UniqueE(NewOptSym("option"))
	e.predefinedSymbols.repeatFx = NewRepxSym("repeat")
	e.predefinedSymbols.optionFx = NewOptxSym("option")

	e.predefinedSymbols.nil = NewZzz("-")
	e.predefinedSymbols.getFn = NewGetF("g")
	e.predefinedSymbols.strFn = NewStrF("s")
	e.predefinedSymbols.actFn = NewActF("a")
	e.predefinedSymbols.bindFn = NewBindF(":")
	e.predefinedSymbols.takeFn = NewTakeF("%")
	e.predefinedSymbols.start = e.nonTerminalSymbols.GetByString("start")
	e.predefinedSymbols.eof = e.nonTerminalSymbols.GetByString("eof")
	e.predefinedSymbols.put = e.nonTerminalSymbols.GetByString("out")
	e.predefinedSymbols.mark = e.functionSymbols.GetByString("mark")

	e.predefinedSymbols.dropFn = e.functionSymbols.UniqueE(NewDropF("drop"))
	e.predefinedSymbols.doneFn = e.functionSymbols.UniqueE(NewDoneF("done"))
	e.predefinedSymbols.injFn = e.functionSymbols.UniqueE(NewInjF("inj"))

	e.nonTerminalSymbols.UniqueE(NewTrueSym("true"))
	e.nonTerminalSymbols.UniqueE(NewFalseSym("false"))
	e.functionSymbols.UniqueE(NewTrueF("true"))
	e.functionSymbols.UniqueE(NewFalseF("false"))
	e.functionSymbols.UniqueE(NewApplyF("apply"))

	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toStr", NewToStrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLstr", NewToLstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUstr", NewToUstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toQuote", NewToQuoteFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toSym", NewToSymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLsym", NewToLsymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUsym", NewToUsymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toSys", NewToSysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLsys", NewToLsysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUsys", NewToUsysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toVar", NewToVarFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toNum", NewToNumFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toOct", NewToOctFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toHex", NewToHexFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toBin", NewToBinFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUrn", NewToUrNstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUrd", NewToUrDstrFromEngine(e)))

	e.functionSymbols.UniqueE(NewTestf("test"))
	e.functionSymbols.UniqueE(NewIff("if"))
	e.functionSymbols.UniqueE(NewLoopf("loop"))
	e.functionSymbols.UniqueE(NewForeachf("foreach"))
	e.functionSymbols.UniqueE(NewRetf("ret"))
	e.functionSymbols.UniqueE(NewLamdaf("lamda"))
	e.functionSymbols.UniqueE(NewSpecf("spec"))
	e.functionSymbols.UniqueE(NewArgsf("args"))
	e.functionSymbols.UniqueE(NewCellf("cell"))
	e.functionSymbols.UniqueE(NewArrayf("array"))
	e.functionSymbols.UniqueE(NewFunf("fun"))
	e.functionSymbols.UniqueE(NewIdxf("idx"))
	e.functionSymbols.UniqueE(NewIdtf("idt"))
	e.functionSymbols.UniqueE(NewSelF("sel"))
	e.functionSymbols.UniqueE(NewStoValf("stoVal"))
	e.functionSymbols.UniqueE(NewStoValf("="))
	e.functionSymbols.UniqueE(NewStoAddf("stoAdd"))
	e.functionSymbols.UniqueE(NewStoAddf("+="))
	e.functionSymbols.UniqueE(NewStoSubf("stoSub"))
	e.functionSymbols.UniqueE(NewStoSubf("-="))
	e.functionSymbols.UniqueE(NewStoMulf("stoMul"))
	e.functionSymbols.UniqueE(NewStoMulf("*="))
	e.functionSymbols.UniqueE(NewStoDivf("stoDiv"))
	e.functionSymbols.UniqueE(NewStoDivf("/="))
	e.functionSymbols.UniqueE(NewStoModf("stoMod"))
	e.functionSymbols.UniqueE(NewStoModf("%="))

	e.functionSymbols.UniqueE(NewEqf("eeq"))
	e.functionSymbols.UniqueE(NewEeqf("==="))
	e.functionSymbols.UniqueE(NewNef("nee"))
	e.functionSymbols.UniqueE(NewNeef("!=="))

	e.functionSymbols.UniqueE(NewInf("in"))
	e.functionSymbols.UniqueE(NewEqf("eq"))
	e.functionSymbols.UniqueE(NewEqf("=="))
	e.functionSymbols.UniqueE(NewNef("ne"))
	e.functionSymbols.UniqueE(NewNef("!="))
	e.functionSymbols.UniqueE(NewLtf("lt"))
	e.functionSymbols.UniqueE(NewLtf("<"))
	e.functionSymbols.UniqueE(NewGtf("gt"))
	e.functionSymbols.UniqueE(NewGtf(">"))
	e.functionSymbols.UniqueE(NewLef("le"))
	e.functionSymbols.UniqueE(NewLef("<="))
	e.functionSymbols.UniqueE(NewGef("ge"))
	e.functionSymbols.UniqueE(NewGef(">="))
	e.functionSymbols.UniqueE(NewOrOrf("orOr"))
	e.functionSymbols.UniqueE(NewOrOrf("||"))
	e.functionSymbols.UniqueE(NewAndAndf("andAnd"))
	e.functionSymbols.UniqueE(NewAndAndf("&&"))
	e.functionSymbols.UniqueE(NewBitOrf("bitOr"))
	e.functionSymbols.UniqueE(NewBitOrf("|"))
	e.functionSymbols.UniqueE(NewBitXorf("bitXor"))
	e.functionSymbols.UniqueE(NewBitXorf("^"))
	e.functionSymbols.UniqueE(NewBitAndf("bitAnd"))
	e.functionSymbols.UniqueE(NewBitAndf("&"))
	e.functionSymbols.UniqueE(NewAddf("add"))
	e.functionSymbols.UniqueE(NewAddf("+"))
	e.functionSymbols.UniqueE(NewSubf("sub"))
	e.functionSymbols.UniqueE(NewSubf("-"))
	e.functionSymbols.UniqueE(NewMulf("mul"))
	e.functionSymbols.UniqueE(NewMulf("*"))
	e.functionSymbols.UniqueE(NewDivf("div"))
	e.functionSymbols.UniqueE(NewDivf("/"))
	e.functionSymbols.UniqueE(NewModf("mod"))
	e.functionSymbols.UniqueE(NewModf("%"))
	e.functionSymbols.UniqueE(NewPreincf("preinc"))
	e.functionSymbols.UniqueE(NewPredecf("predec"))
	e.functionSymbols.UniqueE(NewPostincf("postinc"))
	e.functionSymbols.UniqueE(NewPostdecf("postdec"))
	e.functionSymbols.UniqueE(NewInvf("inv"))
	e.functionSymbols.UniqueE(NewNotf("not"))
	e.functionSymbols.UniqueE(NewNegf("neg"))
}

func (e *Engine) SetrhsOffset(x uint) {
	e.rhsOffset = x
	e.rhsStream.codeIndex = x
}

func (e *Engine) GetRhInput(i uint) Element {
	return e.rhsStream.mode.ContextMode().State().GetChr(i)
}

func (e *Engine) SetOption(x uint) {
	e.options |= x
}

func (e *Engine) SetLoader(x *Loader) {
	e.loader = x
}

func (e *Engine) SetMachineElement(g Element) {
	e.lhsContext.State().grammar = e.grammars.Select(g)
}

func (e *Engine) SetMachineElements(args []Element) Element {
	if len(args) < 2 {
		return e.predefinedSymbols.zlm
	}
	k := args[1].ToVal()
	g := e.grammars.Get(k)
	if g != nil {
		e.lhsContext.State().grammar = g
	}
	return k
}

func (e *Engine) DefineElement(gs, lx Element, ru *Rule) {
	gr := e.grammars.Select(gs)
	if e.oneGrammar == nil {
		e.oneGrammar = gr
		e.lhsContext.State().grammar = e.oneGrammar
	}
	lx.AddRule(gr, ru)
}

func (e *Engine) DefineElements(v []Element, t string, i uint) {
	gr := e.grammars.Select(v[0])
	if e.oneGrammar == nil {
		e.oneGrammar = gr
		e.lhsContext.State().grammar = e.oneGrammar
	}
	gr.Define(v, t, i)
}

func (e *Engine) LoadFromStream(l *Stream) {
	e.defineSymbols()
}

func (e *Engine) LoadFromLMEString(init LMEString) {
	e.defineSymbols()
	init(e.lhsStream)
}

func (e *Engine) LoadFromLMDString(init LMDString) {
	e.defineSymbols()
	init(e.lhsStream)
}

func (e *Engine) Load() {
	e.defineSymbols()
}

func (e *Engine) LoadFromString(rules string) {
	e.LoadFromStringReset(rules, true)
}

func (e *Engine) LoadFromStringReset(rules string, reset bool) {
	e.defineSymbols()
	if reset {
		e.grammars = NewSelector()
	}
	if e.loader == nil {
		e.loader = NewLoader(e)
	}
	e.loader.Load(rules)
}

func (e *Engine) Start() uint {
	if e.oneGrammar != nil {
		if e.inputs.Empty() {
			e.inputs.PushFront(NewGramInputFromEngine(e))
		} else {
			input := e.inputs.Front()
			e.inputs.PushFront(input)
		}

		e.input = e.inputs.Front()

		e.lhsContext.State().grammar = e.oneGrammar
		if e.tracer != nil {
			e.tracer.Dumpg(e.oneGrammar)
		}
		if e.oneGrammar.Get(e.predefinedSymbols.start.Token(), e.predefinedSymbols.eof.Token()) != nil {
			e.rhsStream.currentSymbol = e.predefinedSymbols.start
		}
		if e.Match() && e.flagErrors == 0 {
			return 0
		}
		return 1
	} else {
		return 1
	}
}

func (e *Engine) Gra() *Grammar {
	return e.lhsContext.State().grammar
}

func (e *Engine) Filename() string {
	return e.input.Filename()
}

func (e *Engine) Lineno() uint {
	return e.input.LineNo()
}

func (e *Engine) Charno() uint {
	return e.input.CharNo()
}

func (e *Engine) Charpos() uint {
	return e.input.CharPos()
}

func (e *Engine) SetExternal(x *LMExternal) {
	e.externalSystem = x
}

func (e *Engine) SetLexicalMismatchPriority(x uint) {
	e.lexicalMismatchPriority = x * 2
}

func (e *Engine) SetBuffer(x uint) uint {
	e.maxLength = x
	e.rhsBuffer.SetMax(x)
	return x
}

func (e *Engine) GetInput() Element {
	x := e.input.Get()
	for x == e.predefinedSymbols.eof && !e.inputs.Empty() && x != nil {
		e.inputs.PopFront()
		if e.inputs.Empty() {
			break
		}
		e.input = e.inputs.Front()
		x = e.input.Get()
	}
	return x
}

func (e *Engine) AddInput(x GrammarIO) {
	e.inputs.PushFront(x)
	e.input = x
}

func (e *Engine) Include(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x := a[1].ToVal().ToString()
	if x == "-" {
		e.AddInput(NewGramInputFromEngine(e))
	} else {
		e.AddInput(NewGramInputFile(e, x))
	}
	e.input = e.inputs.Front()
	return NewNumber(0)
}

func (e *Engine) SetTrace(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x, ok := a[1].ToVal().(*Number)
	if !ok {
		return e.predefinedSymbols.zlm
	}
	return NewNumber(LMNumber(e.SetTraceFlag(x.ToUlong())))
}

func (e *Engine) UnsetTrace(a []Element) Element {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x, ok := a[1].ToVal().(*Number)
	if !ok {
		return e.predefinedSymbols.zlm
	}
	return NewNumber(LMNumber(e.UnsetTraceFlag(x.ToUlong())))
}

func (e *Engine) SetMaxDepth(x uint) uint {
	e.maxDepth = x
	return e.maxDepth
}

func (e *Engine) SetMaxRepeat(x uint) uint {
	e.maxRepeat = x
	return e.maxRepeat
}

func (e *Engine) SetDisplayW(x uint) uint {
	e.displayWidth = x
	return e.displayWidth
}

func (e *Engine) SetTraceFlag(x uint) uint {
	if e.tracer == nil {
		e.tracer = NewTracer(e)
	}
	e.tracer.Flags |= x
	if x&DIAGRAMT != 0 || x&DIAGRAM != 0 {
		e.display = NewDiagram(e, e.displayWidth)
		e.tracer.Flags |= DIAGRAM
		e.tracer.Flags |= MISMATCH
		e.tracer.Flags |= SYMBOLS
		e.tracer.Flags |= CXSCOPE
	}
	return e.tracer.Flags
}

func (e *Engine) UnsetTraceFlag(x uint) uint {
	if e.tracer == nil {
		e.tracer = NewTracer(e)
	}
	e.tracer.Flags &= ^x
	return e.tracer.Flags
}

func (e *Engine) PushRhx0(s *State, x *Rule, l ContextHolder, operandsEmpty bool) {
	if e.tracer != nil {
		e.tracer.RuleScope("z=", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	}
	e.rhsContext = NewRHContextFromStateContext(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.Newrhs(e.rhsStream.mode, e.rhsContext, e.rhsContext)
}

func (e *Engine) PushRhx1(s *State, x *Rule, l ContextHolder, operandsEmpty bool) {
	if !operandsEmpty {
		e.lhsContext.MakeVar(e.predefinedSymbols.takeFn, NewStr(e.lhsStream.ToRow()), e.lhsContext, e.lhsStream.variables)
	}
	if e.tracer != nil {
		e.tracer.RuleScope("==", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	}
	e.rhsContext = NewRHContextFromStateContext(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.Newrhs(e.rhsStream.mode, e.rhsContext, l)
}

func (e *Engine) PushRhx(x Element) {
	e.rhsStream.mode = x.NewRHX(e.rhsStream.mode, e.rhsStream.mode.ContextMode(), e.lhsContext)
}

func (e *Engine) Matched3E(l, r, x Element) bool {
	if l != nil {
		e.lhsStream.currentSymbol = nil
	}
	if r != nil {
		e.rhsStream.currentSymbol = nil
	}
	e.rsLastMatchElement = x
	return true
}

func (e *Engine) Matched2E(l, r Element) bool {
	if l != nil {
		e.lhsStream.currentSymbol = nil
	}
	if r != nil {
		e.rhsStream.currentSymbol = nil
	}
	e.rsLastMatchElement = nil
	return true
}

func (e *Engine) Match() bool {
	for {
		if e.tracer != nil {
			for e.lhsStream.mode != nil && e.lhsStream.currentSymbol == nil {
				e.tracer.TraceShort(e.lhsStream, e.lhsStream.mode)
				e.lhsStream.mode = e.lhsStream.mode.Advance(e.lhsStream)
			}
			for e.rhsStream.mode != nil && e.rhsStream.currentSymbol == nil {
				e.tracer.TraceShort(e.rhsStream, e.rhsStream.mode)
				e.rhsStream.mode = e.rhsStream.mode.Advance(e.rhsStream)
			}
			if e.lhsStream.mode == nil {
				return true // exit from lhr.sm
			}
			if e.rhsStream.mode == nil {
				return true // no more input
			}
			if e.lhsStream.currentSymbol == e.predefinedSymbols.nil {
				e.lhsStream.currentSymbol = nil
				continue
			}
			if e.rhsStream.currentSymbol == e.predefinedSymbols.nil {
				e.rhsStream.currentSymbol = nil
				continue
			}
			e.tracer.MatchSymbols(e.lhsStream.currentSymbol, e.rhsStream.currentSymbol)
			if e.lhsStream.currentSymbol.Match(e, e.rhsStream.currentSymbol) {
				continue
			}
			e.tracer.Back(e.lhsStream.currentSymbol, e.rhsStream.currentSymbol)
			return false
		} else {
			for e.lhsStream.mode != nil && e.lhsStream.currentSymbol == nil {
				e.lhsStream.mode = e.lhsStream.mode.Advance(e.lhsStream)
			}
			for e.rhsStream.mode != nil && e.rhsStream.currentSymbol == nil {
				e.rhsStream.mode = e.rhsStream.mode.Advance(e.rhsStream)
			}
			if e.lhsStream.mode == nil {
				return true // exit from e.lhsStream.mode
			}
			if e.rhsStream.mode == nil {
				return true // no more input
			}
			if e.lhsStream.currentSymbol == e.predefinedSymbols.nil {
				e.lhsStream.currentSymbol = nil
				continue
			}
			if e.rhsStream.currentSymbol == e.predefinedSymbols.nil {
				e.rhsStream.currentSymbol = nil
				continue
			}
			if e.lhsStream.currentSymbol.Match(e, e.rhsStream.currentSymbol) {
				continue
			}
			return false
		}
	}
}

func (e *Engine) ResolveE(l, r Element) bool {
	var sta *State
	var x *Rule
	var zl, zr GenMode
	pri := l.Priority(e.lhsContext.Priority())

	if e.tracer != nil {
		e.tracer.Resolve(l, r, pri)
	}
	if e.lhsContext.Priority() == PRIMASK {
		return false
	}

	if x = e.Gra().Get(r.Token(), l.Token()); x != nil {
		sta = NewState(e, e.Gra(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
		e.contextsCount++
		zl = e.lhsStream.mode.Save()
		zr = e.rhsStream.mode.Save()
		if e.ResolveState(sta, x, nil, r, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(r.Token(), e.predefinedSymbols.nil); x != nil {
		if sta == nil {
			sta = NewState(e, e.Gra(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
			e.contextsCount++
			zl = e.lhsStream.mode.Save()
			zr = e.rhsStream.mode.Save()
		}
		if e.ResolveState(sta, x, nil, r, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(e.predefinedSymbols.nil, e.predefinedSymbols.nil); x != nil {
		if sta == nil {
			sta = NewState(e, e.Gra(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
			e.contextsCount++
			zl = e.lhsStream.mode.Save()
			zr = e.rhsStream.mode.Save()
		}
		if e.ResolveState(sta, x, r, nil, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(e.predefinedSymbols.nil, l.Token()); x != nil {
		if sta == nil {
			sta = NewState(e, e.Gra(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
			e.contextsCount++
			zl = e.lhsStream.mode.Save()
			zr = e.rhsStream.mode.Save()
		}
		if e.ResolveState(sta, x, r, nil, pri, zl, zr) {
			return true
		}
	}

	return false
}

func (e *Engine) ResolveState(sta *State, a *Rule, v, s Element, pri uint, zl, zr GenMode) bool {
	x := a
	y := e.lhsContext

	for {
		if x == nil {
			return false
		}
		if x.Allow(pri) {
			if x.Lhlength() == 1 {
				e.rhsStream.currentSymbol = v
				if x.offset < x.Rhlength() {
					e.PushRhx0(sta, x, e.lhsContext, false)
				}
				e.lhsContext = y
				return true
			}
			e.lhsContext = NewLHContextFromRule(sta, e.lhsContext, x)
			if e.maxDepth > 0 {
				if err := e.lhsContext.CheckDepth(e.maxDepth); err != nil {
					panic(err)
				}
			}
			e.rhsStream.currentSymbol = v
			e.rsLastMatchElement = s
			e.lhsStream.mode = x.Newlhs(e.lhsStream.mode, e.lhsContext)
			e.lhsStream.ClearX()
			if x.Match(e) {
				break
			}
			e.lhsStream.variables = y.Variables()
			e.lhsContext = y
			e.lhsStream.mode = zl.Restore()
			e.rhsStream.mode = zr.Restore()
		}
		x = x.next
	}
	// writefln("A %4d %4d", x.off, x.rhlength());
	if x.offset < x.Rhlength() {
		e.PushRhx1(sta, x, e.lhsContext, e.lhsStream.EmptyX())
	}

	e.lhsStream.mode = zl.Restore()
	e.lhsContext = y
	return true
}

func (e *Engine) Repeat(max uint) bool {
	var w, x GenMode

	e.lhsStream.currentSymbol = nil
	w = e.lhsStream.mode
	x = e.rhsStream.mode.Save()

	for i := uint(0); max == 0 || i < max; i++ {
		if e.maxRepeat == 0 || i < e.maxRepeat {
			e.lhsStream.mode = NewLHModeFromMode(w)
			if !e.Match() {
				break
			}
			x = e.rhsStream.mode.Save()
			if e.tracer != nil {
				e.tracer.Repeat(i)
			}
		} else {
			panic("NewMaxRepeatError")
		}
	}

	e.rhsStream.mode = x.Restore()
	e.lhsStream.mode = w.Ret()
	return true
}

func (e *Engine) Repeatx(max uint) bool {
	var w, x, z GenMode

	b := e.lhsStream.Popx()
	e.lhsStream.currentSymbol = nil
	w = e.lhsStream.mode.Save()
	x = e.rhsStream.mode.Save()

	for i := uint(0); max == 0 || i < max; i++ {
		if e.maxRepeat == 0 || i < e.maxRepeat {
			z = b.NewLHS(w)
			e.lhsStream.mode = z
			if !e.lhsContext.Rule().Match(e) {
				break
			}
			x = e.rhsStream.mode.Save()
			if e.tracer != nil {
				e.tracer.Repeat(i)
			}
		} else {
			panic("NewMaxRepeatError")
		}
	}

	e.rhsStream.mode = x.Restore()
	e.lhsStream.currentSymbol = nil
	e.lhsStream.codeIndex = w.CodeIndex()
	e.lhsStream.codeVector = w.CodeVector()
	e.lhsStream.mode = w
	return true
}

func (e *Engine) PushX() {
	e.lhsStream.Pushx(e.rhsStream.Popx())
}

func (e *Engine) PushR(x Element) {
	e.lhsStream.Pushx(x)
}

func (e *Engine) PushXElem(x Element) {
	e.lhsStream.Pushx(x)
}

func (e *Engine) Initialise(s *Stream, m GenMode) {
}

func (e *Engine) BadRhs(s *Stream, m GenMode, i uint) {
	utils.Tz("bad rhs")
	panic("bad rhs")
}

func (e *Engine) BadCode(s *Stream, m GenMode, i uint) {
	utils.Tz("bad code")
	panic("bad code")
}

func (e *Engine) BindCvar(l, r Element) bool {
	e.lhsContext.MakeVar(l, r, e.lhsContext, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindCvar(l, r)
	}
	return true
}

func (e *Engine) BindLvar(l, r Element) bool {
	e.lhsContext.MakeVar(l, r, e.lhsContext, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindLvar(l, r)
	}
	return true
}

func (e *Engine) BindXvarE(r Element) bool {
	l := e.lhsStream.Popx()
	if e.tracer != nil {
		e.tracer.BindRvar(l, r)
	}
	e.lhsContext.MakeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindRvarScope(l, r, e.rhsStream.mode)
	}
	return true
}

func (e *Engine) BindXVar() bool {
	l := e.lhsStream.Popx()
	r := e.rhsStream.Popx()
	if e.tracer != nil {
		e.tracer.BindRvar(l, r)
	}
	e.lhsContext.MakeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindRvarScope(l, r, e.rhsStream.mode)
	}
	return true
}

func (e *Engine) BindUvar(l, r Element) bool {
	if e.tracer != nil {
		e.tracer.BindRvar(l, r)
	}
	e.lhsContext.MakeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindRvarScope(l, r, e.rhsStream.mode)
	}
	return true
}

func (e *Engine) TakeTvar() bool {
	v := e.rhsStream.mode.ScopeVariables()
	if v != nil && v.Key() == e.predefinedSymbols.takeFn {
		s := v.Value().(*Str)
		for _, x := range s.V {
			e.lhsStream.Pushx(x)
		}
	}
	return true
}

func (e *Engine) BindTvar() bool {
	l := e.lhsStream.Popx()
	v := e.rhsStream.mode.ScopeVariables()
	var r Element
	if v != nil && v.Key() == e.predefinedSymbols.takeFn {
		r = v.Value()
	} else {
		r = NewStr([]Element{})
	}
	if e.tracer != nil {
		e.tracer.BindRvar(l, r)
	}
	e.lhsContext.MakeVar(l, r, e.rhsStream.mode, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindRvarScope(l, r, e.rhsStream.mode)
	}
	return true
}

func (e *Engine) Deref(pk Element, x LMScope) VarElement {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if e.tracer != nil {
		e.tracer.TheRefVars(pk, pp, pq)
	}
	for pp != nil && pk != pp.Key() {
		pp = pp.ScopeVariables()
	}
	return pp
}

func (e *Engine) TheRef(s GenMode, k Element, x LMScope) GenMode {
	v := e.Deref(k, x)
	if e.tracer != nil {
		e.tracer.TheRefVar(v)
	}
	if v != nil {
		return NewRFModeFromVar(s, v)
	}
	return s
}

func (e *Engine) TheValue(s GenMode, k Element, x LMScope) Element {
	v := e.Deref(k, x)
	if e.tracer != nil {
		e.tracer.TheRefVar(v)
	}
	if v != nil {
		return v.ToVal()
	}
	return e.predefinedSymbols.zlm
}

func (e *Engine) EachRef(s GenMode, k Element, x LMScope) GenMode {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if e.tracer != nil {
		e.tracer.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Key() {
			if e.tracer != nil {
				e.tracer.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.ScopeVariables()
	}
	return s
}

func (e *Engine) Lookvars(s GenMode, k Element, x LMScope, last VarElement) GenMode {
	if x == x.ScopeVariables() {
		return s
	}
	utils.Tz(">LOOK>")
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if pq == nil {
		return s
	}
	if e.tracer != nil {
		e.tracer.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq && pp != last {
		if pp.ScopeReferenceContext() != nil {
			e.Lookvars(s, k, pp.ScopeReferenceContext(), pp)
		}
		fmt.Printf("\tsi: %8d ", pp.Si())
		TxE("----", pp)
		pp = pp.ScopeVariables()
	}
	utils.Tz("<LOOK<")
	return s
}

func (e *Engine) AllRef(s GenMode, k Element, x LMScope) GenMode {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if pq == nil {
		return s
	}
	if e.tracer != nil {
		e.tracer.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Key() {
			if e.tracer != nil {
				e.tracer.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.AllVariables()
	}
	return s
}

func (e *Engine) Count(k Element, p, q VarElement) uint {
	var i uint
	for i = 0; p != nil && p != q; p = p.ScopeVariables() {
		if p.Key() == k {
			i++
		}
	}
	return i
}

func (e *Engine) ToElements(k Element, p, q VarElement) []Element {
	n := e.Count(k, p, q)
	r := make([]Element, n)
	for i := int(n); i > 0; p = p.ScopeVariables() {
		if p.Key() == k {
			i--
			r[i] = p.Value()
		}
	}
	return r
}

func (e *Engine) ToString(k Element, p, q VarElement) string {
	var r string
	for p != nil && p != q {
		if p.Key() == k {
			r = p.Value().ToString() + r
		}
		p = p.ScopeVariables()
	}
	return r
}
