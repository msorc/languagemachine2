package machine

import (
	"fmt"
	"net/url"
)

const (
	// Engine
	LEXPRI    uint = 1000 // default value for lexical priority
	ZLONLY    uint = 1    // (not used) resolution: try r:l r:z z:l
	ZZFINAL   uint = 2    // (not used) resolution: try r:l r:z z:l z:z
	MAXLENGTH uint = 64 * 1024
)

var theZlm *ZLM

func TxE(s string, x MachineElement) MachineElement {
	var xtrace string
	if x != nil {
		xtrace = x.ToTrace()
	} else {
		xtrace = "---"
	}
	fmt.Printf("\t%6s:     %p %24s\n", s, x, xtrace)
	return x
}

func TxV(r, s string, w *Var) MachineElement {
	var n string

	if w != nil {
		n = w.ToDump()
	} else {
		n = "---"
	}

	fmt.Printf("\t%6s:%4s %p %24s %p %p %p %p\n", r, s, w, n, w.Vv, w.Vp, w.Vx, w.Vs)

	return w
}

func tz(s string) { fmt.Printf("\ttz: %s\n", s) }
func star()       { tz("*") }

func UrlEscape(u string) string {
	return url.QueryEscape(u)
}

func UrlUnescape(u string) string {
	decoded, err := url.QueryUnescape(u)
	if err != nil {
		panic("Error decoding")
	}
	return decoded
}

func priAssoc(pri uint) string {
	if pri == 0 {
		return "L"
	}
	if pri&BRACKET != 0 {
		return "B"
	}
	if pri&1 != 0 {
		return "R"
	}
	return "L"
}
func priValue(pri uint) uint { return (pri & PRIMASK) / 2 }

func nullStr(s *Stream) GenMode        { return s.mode }
func nullFun(s *Stream) MachineElement { return nil }
func theNull() MachineElement          { return theZlm }

type LMEString func(*Stream) GenMode
type LMNFunc func(*Stream) MachineElement
type LMDString func(*Stream) GenMode

// main parsing engine: everything needed to load and apply grammars
type Engine struct {
	state         *State             // state at start of a new context
	contextsCount uint               // count of new contexts used to give each a unique identity
	lhsContext    EngineStateContext // lhs context stack for mismatch events being resolved
	rhsContext    EngineStateContext // rhs context stack for rhs of rules that have matched

	lhsStream *Stream // lhs registers
	rhsStream *Stream // rhs registers
	lhsMode   GenMode // lhs registers, LZMode
	rhsMode   GenMode // rhs registers, RZMode
	// Lhs       GenMode           // lhs element mode generator
	// Rhs       GenMode           // rhs element mode generator
	rsLastMatchElement MachineElement // element resulting from last match

	grammars   *Selector // table of grammars selected by symbol
	oneGrammar *Grammar  // initial grammar

	terminalSymbols    *Dict   // terminal symbols
	nonTerminalSymbols *Dict   // non-terminal symbols
	varSymbols         *Dict   // variables
	userSymbols        *Dict   // user symbols - guaranteed not to match system symbols
	functionSymbols    *Dict   // primitive operator symbols
	predefinedSymbols  *Predef // predefined symbols with special significance

	loader *Loader // rule loader

	inputs *IStack      // stack of input sources
	input  GrammarStdio // current input

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
		functionSymbols:    NewDict(),
		terminalSymbols:    NewDict(),
		nonTerminalSymbols: NewDict(),
		varSymbols:         NewDict(),
		userSymbols:        NewDict(),
		predefinedSymbols:  NewPredef(),
		rhsBuffer:          NewRZBuffer(make([]MachineElement, 1024), len),
		externalSystem:     NewLMExternal(),
		grammars:           NewSelector(),
	}
	e.state = NewState(e, nil, nil, nil, nil, 0, 0, 0, 0)
	e.lhsContext = NewLHContextFromState(e.state)
	e.rhsContext = NewRHContextFromState(e.state)
	e.lhsStream = NewStreamFromEngine(e, "lh", 0)
	e.rhsStream = NewStreamFromEngine(e, "rh", 0)
	e.lhsMode = NewLZModeFromContext(e.lhsContext, e.lhsStream)
	e.lhsStream.mode = e.lhsMode
	e.rhsMode = NewRZModeFromContext(e.rhsContext, e.rhsStream)
	e.rhsStream.mode = e.rhsMode
	e.lexicalMismatchPriority = LEXPRI
	return e
}

func (e *Engine) SetrhsOffset(x uint) {
	e.rhsOffset = x
	e.rhsStream.codeIndex = x
}

func (e *Engine) GetRhInput(i uint) MachineElement {
	return e.rhsStream.mode.ContextMode().State().GetChr(i)
}

func (e *Engine) SetOption(x uint) {
	e.options |= x
}

func (e *Engine) SetLoader(x *Loader) {
	e.loader = x
}

func (e *Engine) SetMachineElement(g MachineElement) {
	e.lhsContext.State().grammar = e.grammars.Select(g)
}

func (e *Engine) SetMachineElements(args []MachineElement) MachineElement {
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

func (e *Engine) DefineElement(gs, lx MachineElement, ru *Rule) {
	gr := e.grammars.Select(gs)
	if e.oneGrammar == nil {
		e.oneGrammar = gr
		e.lhsContext.State().grammar = e.oneGrammar
	}
	lx.AddRule(gr, ru)
}

func (e *Engine) DefineElements(v []MachineElement, t string, i uint) {
	gr := e.grammars.Select(v[0])
	if e.oneGrammar == nil {
		e.oneGrammar = gr
		e.lhsContext.State().grammar = e.oneGrammar
	}
	gr.Define(v, t, i)
}

func (e *Engine) LoadFromStream(l *Stream) {
	defineSymbols(e)
	e.lhsStream.SetSymbols(e.predefinedSymbols)
	e.lhsStream.Initialise(e.lhsStream)
	e.rhsStream.SetSymbols(e.predefinedSymbols)
	e.rhsStream.CopyTables(e.lhsStream)
}

func (e *Engine) LoadFromLMEString(init LMEString) {
	defineSymbols(e)
	e.lhsStream.SetSymbols(e.predefinedSymbols)
	init(e.lhsStream)
	e.rhsStream.SetSymbols(e.predefinedSymbols)
	e.rhsStream.CopyTables(e.lhsStream)
}

func (e *Engine) LoadFromLMDString(init LMDString) {
	defineSymbols(e)
	e.lhsStream.SetSymbols(e.predefinedSymbols)
	init(e.lhsStream)
	e.rhsStream.SetSymbols(e.predefinedSymbols)
	e.rhsStream.CopyTables(e.lhsStream)
}

func (e *Engine) Load() {
	defineSymbols(e)
	e.lhsStream.SetSymbols(e.predefinedSymbols)
	e.InitialiseStream(e.lhsStream)
	e.rhsStream.SetSymbols(e.predefinedSymbols)
	e.rhsStream.CopyTables(e.lhsStream)
}

func (e *Engine) LoadFromString(rules string) {
	e.LoadFromStringReset(rules, true)
	e.lhsStream.SetSymbols(e.predefinedSymbols)
	e.rhsStream.SetSymbols(e.predefinedSymbols)
	e.rhsStream.CopyTables(e.lhsStream)
}

func (e *Engine) LoadFromStringReset(rules string, reset bool) {
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
		var t, s = e.inputs, e.inputs
		for s != nil {
			t = NewIStack(t, s.Input)
			s = s.Next
		}
		if t == nil {
			t = NewIStack(s, NewGramInputFromEngine(e))
		}
		e.inputs = t
		e.input = t.Input
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

func (e *Engine) GetInput() MachineElement {
	x := e.input.Get()
	for x == e.predefinedSymbols.eof && e.inputs != nil {
		e.inputs = e.inputs.Next
		e.input = e.inputs.Input
		x = e.input.Get()
	}
	return x
}

func (e *Engine) AddInput(x GrammarStdio) {
	e.inputs = NewIStack(e.inputs, x)
	e.input = e.inputs.Input
}

func (e *Engine) Include(a []MachineElement) MachineElement {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x := a[1].ToVal().ToString()
	if x == "-" {
		e.AddInput(NewGramInputFromEngine(e))
	} else {
		e.AddInput(NewGramInputFile(e, x))
	}
	e.input = e.inputs.Input
	return NewNumber(0)
}

func (e *Engine) SetTrace(a []MachineElement) MachineElement {
	if len(a) < 2 {
		return e.predefinedSymbols.zlm
	}
	x, ok := a[1].ToVal().(*Number)
	if !ok {
		return e.predefinedSymbols.zlm
	}
	return NewNumber(LMNumber(e.SetTraceFlag(x.ToUlong())))
}

func (e *Engine) UnsetTrace(a []MachineElement) MachineElement {
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

func (e *Engine) PushRhx0(s *State, x *Rule, l EngineStateContext, lx *Opnd) {
	if e.tracer != nil {
		e.tracer.RuleScope("z=", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	}
	e.rhsContext = NewRHContextFromStateContexts(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.Newrhs(e.rhsStream.mode, e.rhsContext, e.rhsContext)
}

func (e *Engine) PushRhx1(s *State, x *Rule, l EngineStateContext, lx *Opnd) {
	if lx != nil {
		e.lhsContext.MakeVar(e.predefinedSymbols.takeFn, NewStr(e.lhsStream.ToRow()), e.lhsContext, e.lhsStream.variables)
	}
	if e.tracer != nil {
		e.tracer.RuleScope("==", s, e.lhsContext.Variables(), e.lhsContext.ContextLimitVariable())
	}
	e.rhsContext = NewRHContextFromStateContexts(s, e.rhsStream.mode.ContextMode(), e.lhsContext)
	e.rhsStream.mode = x.Newrhs(e.rhsStream.mode, e.rhsContext, l)
}

func (e *Engine) PushRhx(x MachineElement) {
	e.rhsStream.mode = x.NewRHX(e.rhsStream.mode, e.rhsStream.mode.ContextMode(), e.lhsContext)
}

func (e *Engine) Matched3E(l, r, x MachineElement) bool {
	if l != nil {
		e.lhsStream.currentSymbol = nil
	}
	if r != nil {
		e.rhsStream.currentSymbol = nil
	}
	e.rsLastMatchElement = x
	return true
}

func (e *Engine) Matched2E(l, r MachineElement) bool {
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
			if e.lhsStream.currentSymbol.Match(e, e.rhsStream.currentSymbol) {
				continue
			}
			return false
		}
	}
}

func (e *Engine) ResolveE(l, r MachineElement) bool {
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
		zl = e.lhsStream.mode.Save()
		zr = e.rhsStream.mode.Save()
		if e.ResolveState(sta, x, nil, r, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(r.Token(), e.predefinedSymbols.nil); x != nil {
		if sta != nil {
			sta = NewState(e, e.Gra(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
			zl = e.lhsStream.mode.Save()
			zr = e.rhsStream.mode.Save()
		}
		if e.ResolveState(sta, x, nil, r, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(e.predefinedSymbols.nil, e.predefinedSymbols.nil); x != nil {
		if sta != nil {
			sta = NewState(e, e.Gra(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
			zl = e.lhsStream.mode.Save()
			zr = e.rhsStream.mode.Save()
		}
		if e.ResolveState(sta, x, r, nil, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(e.predefinedSymbols.nil, l.Token()); x != nil {
		if sta != nil {
			sta = NewState(e, e.Gra(), l, r, e.input, e.Charpos(), e.Lineno(), e.Charno(), e.contextsCount)
			zl = e.lhsStream.mode.Save()
			zr = e.rhsStream.mode.Save()
		}
		if e.ResolveState(sta, x, r, nil, pri, zl, zr) {
			return true
		}
	}

	return false
}

func (e *Engine) ResolveState(sta *State, a *Rule, v, s MachineElement, pri uint, zl, zr GenMode) bool {
	x := a
	y := e.lhsContext

	for {
		if x == nil {
			return false
		}
		if x.Allow(pri) {
			if x.Lhlength() == 1 {
				e.rhsStream.currentSymbol = v
				if x.Off < x.Rhlength() {
					e.PushRhx0(sta, x, e.lhsContext, nil)
				}
				e.lhsContext = y
				return true
			}
			e.lhsContext = NewLHContextFromRule(sta, e.lhsContext, x)
			if e.maxDepth > 0 {
				e.lhsContext.CheckDepth(e.maxDepth)
			}
			e.rhsStream.currentSymbol = v
			e.rsLastMatchElement = s
			e.lhsStream.mode = x.Newlhs(e.lhsStream.mode, e.lhsContext)
			e.lhsStream.operandsStack = nil
			if x.Match(e) {
				break
			}
			e.lhsStream.variables = y.Variables()
			e.lhsContext = y
			e.lhsStream.mode = zl.Restore()
			e.rhsStream.mode = zr.Restore()
		}
		x = x.Nxt
	}
	// writefln("A %4d %4d", x.off, x.rhlength());
	if x.Off < x.Rhlength() {
		e.PushRhx1(sta, x, e.lhsContext, e.lhsStream.operandsStack)
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
	e.lhsStream.LK = w.LK()
	e.lhsStream.mode = w
	return true
}

func (e *Engine) PushX() {
	e.lhsStream.operandsStack = NewOpnd(e.lhsStream.operandsStack, e.rhsStream.Popx())
}

func (e *Engine) PushR(x MachineElement) {
	e.lhsStream.operandsStack = NewOpnd(e.lhsStream.operandsStack, x)
}

func (e *Engine) PushXElem(x MachineElement) {
	e.lhsStream.operandsStack = NewOpnd(e.lhsStream.operandsStack, x)
}

func (e *Engine) Initialise(s *Stream, m GenMode) {
}

func (e *Engine) InitialiseStream(s *Stream) {
}

func (e *Engine) BadRhs(s *Stream, m GenMode, i uint) {
	tz("bad rhs")
	panic("bad rhs")
}

func (e *Engine) BadCode(s *Stream, m GenMode, i uint) {
	tz("bad code")
	panic("bad code")
}

func (e *Engine) BindCvar(l, r MachineElement) bool {
	e.lhsContext.MakeVar(l, r, e.lhsContext, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindCvar(l, r)
	}
	return true
}

func (e *Engine) BindLvar(l, r MachineElement) bool {
	e.lhsContext.MakeVar(l, r, e.lhsContext, e.lhsStream.variables)
	if e.tracer != nil {
		e.tracer.BindLvar(l, r)
	}
	return true
}

func (e *Engine) BindXvarE(r MachineElement) bool {
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

func (e *Engine) BindUvar(l, r MachineElement) bool {
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
	if v != nil && v.Vk == e.predefinedSymbols.takeFn {
		s := v.Vv.(*Str)
		for _, x := range s.V {
			e.lhsStream.Pushx(x)
		}
	}
	return true
}

func (e *Engine) BindTvar() bool {
	l := e.lhsStream.Popx()
	v := e.rhsStream.mode.ScopeVariables()
	var r MachineElement
	if v != nil && v.Vk == e.predefinedSymbols.takeFn {
		r = v.Vv
	} else {
		r = NewStr([]MachineElement{})
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

func (e *Engine) Deref(pk MachineElement, x LMScope) *Var {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if e.tracer != nil {
		e.tracer.TheRefVars(pk, pp, pq)
	}
	for pp != nil && pk != pp.Vk {
		pp = pp.Vs
	}
	return pp
}

func (e *Engine) TheRef(s GenMode, k MachineElement, x LMScope) GenMode {
	v := e.Deref(k, x)
	if e.tracer != nil {
		e.tracer.TheRefVar(v)
	}
	if v != nil {
		return NewRFModeFromVar(s, v)
	}
	return s
}

func (e *Engine) TheValue(s GenMode, k MachineElement, x LMScope) MachineElement {
	v := e.Deref(k, x)
	if e.tracer != nil {
		e.tracer.TheRefVar(v)
	}
	if v != nil {
		return v.ToVal()
	}
	return e.predefinedSymbols.zlm
}

func (e *Engine) EachRef(s GenMode, k MachineElement, x LMScope) GenMode {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if e.tracer != nil {
		e.tracer.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Vk {
			if e.tracer != nil {
				e.tracer.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.Vs
	}
	return s
}

func (e *Engine) Lookvars(s GenMode, k MachineElement, x LMScope, last *Var) GenMode {
	if x == x.ScopeVariables() {
		return s
	}
	tz(">LOOK>")
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
		fmt.Printf("\tsi: %8d ", pp.Si)
		TxE("----", pp)
		pp = pp.Vs
	}
	tz("<LOOK<")
	return s
}

func (e *Engine) AllRef(s GenMode, k MachineElement, x LMScope) GenMode {
	pp := x.ScopeVariables()
	pq := x.ScopeContextLimitVariables()
	if pq == nil {
		return s
	}
	if e.tracer != nil {
		e.tracer.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Vk {
			if e.tracer != nil {
				e.tracer.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.Va
	}
	return s
}

func (e *Engine) Count(k MachineElement, p, q *Var) uint {
	var i uint
	for i = 0; p != nil && p != q; p = p.Vs {
		if p.Vk == k {
			i++
		}
	}
	return i
}

func (e *Engine) ToElements(k MachineElement, p, q *Var) []MachineElement {
	n := e.Count(k, p, q)
	r := make([]MachineElement, n)
	for i := int(n); i > 0; p = p.Vs {
		if p.Vk == k {
			i--
			r[i] = p.Vv
		}
	}
	return r
}

func (e *Engine) ToString(k MachineElement, p, q *Var) string {
	var r string
	for p != nil && p != q {
		if p.Vk == k {
			r = p.Vv.ToString() + r
		}
		p = p.Vs
	}
	return r
}
