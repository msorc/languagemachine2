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

func TxE(s string, x GrammarElement) GrammarElement {
	var xtrace string
	if x != nil {
		xtrace = x.ToTrace()
	} else {
		xtrace = "---"
	}
	fmt.Printf("\t%6s:     %p %24s\n", s, x, xtrace)
	return x
}

func TxV(r, s string, w *Var) GrammarElement {
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

func nullStr(s *Stream) GenMode        { return s.SM }
func nullFun(s *Stream) GrammarElement { return nil }
func theNull() GrammarElement          { return theZlm }

type LMEString func(*Stream) GenMode
type LMNFunc func(*Stream) GrammarElement
type LMDString func(*Stream) GenMode

// main parsing engine: everything needed to load and apply grammars
type Engine struct {
	Sta *State             // state at start of a new context
	Cxn uint               // count of new contexts used to give each a unique identity
	Lhx EngineStateContext // lhs context stack for mismatch events being resolved
	Rhx EngineStateContext // rhs context stack for rhs of rules that have matched

	Lhr  *Stream // lhs registers
	Rhr  *Stream // rhs registers
	Lhzz GenMode // rhs registers
	Rhzz GenMode // rhs registers
	// Lhs       GenMode           // lhs element mode generator
	// Rhs       GenMode           // rhs element mode generator
	Rsy GrammarElement // element resulting from last match

	Grammars *Selector // table of grammars selected by symbol
	One      *Grammar  // initial grammar
	Tsy      *Dict     // terminal symbols
	Nsy      *Dict     // non-terminal symbols
	Vsy      *Dict     // variables
	Usy      *Dict     // user symbols - guaranteed not to match system symbols
	Fsy      *Dict     // primitive operator symbols
	Ssy      *Predef   // predefined symbols with special significance
	Ldr      *Loader   // rule loader

	Inputs *IStack      // stack of input sources
	Input  GrammarStdio // current input
	Rhz    *RZBuffer    // circular buffer at outermost level of rhs

	FlagErrors uint // incremented by flagSym
	WarnErrors uint // incremented by warnSym

	System   *LMExternal // external interfaces
	Trace    *Tracer     // trace handler - null if no tracing required
	Display  *Diagram    // to display trace as diagram
	Displayw uint        // width for diagram display

	Options   uint // engine control options
	Maxdepth  uint // limit on analysis recursion depth - zero means no limit
	Maxrepeat uint // limit on repetition at repeat     - zero means no limit
	RhOffset  uint // offset applied to input position
	Buflength uint // default size of rhz circular buffer
	Maxlength uint // default size of rhz circular buffer
	Lexpri    uint // artificial priority of context at lexical mismatch
}

func NewEngine() *Engine {
	return NewEngineFromLength(MAXLENGTH)
}

func NewEngineFromLength(len uint) *Engine {
	theZlm = NewZLM("null")
	e := &Engine{
		Maxlength: len,
		Fsy:       NewDict(),
		Tsy:       NewDict(),
		Nsy:       NewDict(),
		Vsy:       NewDict(),
		Usy:       NewDict(),
		Ssy:       NewPredef(),
		Rhz:       NewRZBuffer(make([]GrammarElement, 1024), len),
		System:    NewLMExternal(),
		Grammars:  NewSelector(),
	}
	e.Sta = NewState(e, nil, nil, nil, nil, 0, 0, 0, 0)
	e.Lhx = NewLHContextFromState(e.Sta)
	e.Rhx = NewRHContextFromState(e.Sta)
	e.Lhr = NewStreamFromEngine(e, "lh", 0)
	e.Rhr = NewStreamFromEngine(e, "rh", 0)
	e.Lhr.QU = "lh"
	e.Lhr.LM = e
	e.Rhr.QU = "rh"
	e.Rhr.LM = e
	e.Lhzz = NewLZModeFromContext(e.Lhx, e.Lhr)
	e.Lhr.SM = e.Lhzz
	e.Rhzz = NewRZModeFromContext(e.Rhx, e.Rhr)
	e.Rhr.SM = e.Lhzz
	e.SetLexpri(LEXPRI)
	return e
}

func (e *Engine) SetRhOffset(x uint) {
	e.RhOffset = x
	e.Rhr.CI = x
}

func (e *Engine) GetRhInput(i uint) GrammarElement {
	return e.Rhr.SM.CX().St().GetChr(i)
}

func (e *Engine) SetOption(x uint) {
	e.Options |= x
}

func (e *Engine) SetLoader(x *Loader) {
	e.Ldr = x
}

func (e *Engine) SetGrammarElement(g GrammarElement) {
	e.Lhx.St().Gr = e.Grammars.Select(g)
}

func (e *Engine) SetGrammarElements(args []GrammarElement) GrammarElement {
	if len(args) < 2 {
		return e.Ssy.ZLM
	}
	k := args[1].ToVal()
	g := e.Grammars.Get(k)
	if g != nil {
		e.Lhx.St().Gr = g
	}
	return k
}

func (e *Engine) DefineElement(gs, lx GrammarElement, ru *Rule) {
	gr := e.Grammars.Select(gs)
	if e.One == nil {
		e.One = gr
		e.Lhx.St().Gr = e.One
	}
	lx.AddRule(gr, ru)
}

func (e *Engine) DefineElements(v []GrammarElement, t string, i uint) {
	gr := e.Grammars.Select(v[0])
	if e.One == nil {
		e.One = gr
		e.Lhx.St().Gr = e.One
	}
	gr.Define(v, t, i)
}

func (e *Engine) LoadFromStream(l *Stream) {
	defineSymbols(e)
	e.Lhr.SetSymbols(e.Ssy)
	e.Lhr.Initialise(e.Lhr)
	e.Rhr.SetSymbols(e.Ssy)
	e.Rhr.CopyTables(e.Lhr)
}

func (e *Engine) LoadFromLMEString(init LMEString) {
	defineSymbols(e)
	e.Lhr.SetSymbols(e.Ssy)
	init(e.Lhr)
	e.Rhr.SetSymbols(e.Ssy)
	e.Rhr.CopyTables(e.Lhr)
}

func (e *Engine) LoadFromLMDString(init LMDString) {
	defineSymbols(e)
	e.Lhr.SetSymbols(e.Ssy)
	init(e.Lhr)
	e.Rhr.SetSymbols(e.Ssy)
	e.Rhr.CopyTables(e.Lhr)
}

func (e *Engine) Load() {
	defineSymbols(e)
	e.Lhr.SetSymbols(e.Ssy)
	e.InitialiseStream(e.Lhr)
	e.Rhr.SetSymbols(e.Ssy)
	e.Rhr.CopyTables(e.Lhr)
}

func (e *Engine) LoadFromString(rules string) {
	e.LoadFromStringReset(rules, true)
	e.Lhr.SetSymbols(e.Ssy)
	e.Rhr.SetSymbols(e.Ssy)
	e.Rhr.CopyTables(e.Lhr)
}

func (e *Engine) LoadFromStringReset(rules string, reset bool) {
	if reset {
		e.Grammars = NewSelector()
	}
	if e.Ldr == nil {
		e.Ldr = NewLoader(e)
	}
	e.Ldr.Load(rules)
}

func (e *Engine) Start() uint {
	if e.One != nil {
		var t, s = e.Inputs, e.Inputs
		for s != nil {
			t = NewIStack(t, s.Input)
			s = s.Next
		}
		if t == nil {
			t = NewIStack(s, NewGramInputFromEngine(e))
		}
		e.Inputs = t
		e.Input = t.Input
		e.Lhx.St().Gr = e.One
		if e.Trace != nil {
			e.Trace.Dumpg(e.One)
		}
		if e.One.Get(e.Ssy.Start.Token(), e.Ssy.EOF.Token()) != nil {
			e.Rhr.SY = e.Ssy.Start
		}
		if e.Match() && e.FlagErrors == 0 {
			return 0
		}
		return 1
	} else {
		return 1
	}
}

func (e *Engine) Gra() *Grammar {
	return e.Lhx.St().Gr
}

func (e *Engine) Filename() string {
	return e.Input.Filename()
}

func (e *Engine) Lineno() uint {
	return e.Input.LineNo()
}

func (e *Engine) Charno() uint {
	return e.Input.CharNo()
}

func (e *Engine) Charpos() uint {
	return e.Input.CharPos()
}

func (e *Engine) SetExternal(x *LMExternal) {
	e.System = x
}

func (e *Engine) SetLexpri(x uint) {
	e.Lexpri = x * 2
}

func (e *Engine) SetBuffer(x uint) uint {
	e.Maxlength = x
	e.Rhz.SetMax(x)
	return x
}

func (e *Engine) GetInput() GrammarElement {
	x := e.Input.Get()
	for x == e.Ssy.EOF && e.Inputs != nil {
		e.Inputs = e.Inputs.Next
		e.Input = e.Inputs.Input
		x = e.Input.Get()
	}
	return x
}

func (e *Engine) AddInput(x GrammarStdio) {
	e.Inputs = NewIStack(e.Inputs, x)
	e.Input = e.Inputs.Input
}

func (e *Engine) Include(a []GrammarElement) GrammarElement {
	if len(a) < 2 {
		return e.Ssy.ZLM
	}
	x := a[1].ToVal().ToString()
	if x == "-" {
		e.AddInput(NewGramInputFromEngine(e))
	} else {
		e.AddInput(NewGramInputFile(e, x))
	}
	e.Input = e.Inputs.Input
	return NewNumber(0)
}

func (e *Engine) SetTrace(a []GrammarElement) GrammarElement {
	if len(a) < 2 {
		return e.Ssy.ZLM
	}
	x, ok := a[1].ToVal().(*Number)
	if !ok {
		return e.Ssy.ZLM
	}
	return NewNumber(LMNumber(e.SetTraceFlag(x.ToUlong())))
}

func (e *Engine) UnsetTrace(a []GrammarElement) GrammarElement {
	if len(a) < 2 {
		return e.Ssy.ZLM
	}
	x, ok := a[1].ToVal().(*Number)
	if !ok {
		return e.Ssy.ZLM
	}
	return NewNumber(LMNumber(e.UnsetTraceFlag(x.ToUlong())))
}

func (e *Engine) SetMaxDepth(x uint) uint {
	e.Maxdepth = x
	return e.Maxdepth
}

func (e *Engine) SetMaxRepeat(x uint) uint {
	e.Maxrepeat = x
	return e.Maxrepeat
}

func (e *Engine) SetDisplayW(x uint) uint {
	e.Displayw = x
	return e.Displayw
}

func (e *Engine) SetTraceFlag(x uint) uint {
	if e.Trace == nil {
		e.Trace = NewTracer(e)
	}
	e.Trace.Flags |= x
	if x&DIAGRAMT != 0 || x&DIAGRAM != 0 {
		e.Display = NewDiagram(e, e.Displayw)
		e.Trace.Flags |= DIAGRAM
		e.Trace.Flags |= MISMATCH
		e.Trace.Flags |= SYMBOLS
		e.Trace.Flags |= CXSCOPE
	}
	return e.Trace.Flags
}

func (e *Engine) UnsetTraceFlag(x uint) uint {
	if e.Trace == nil {
		e.Trace = NewTracer(e)
	}
	e.Trace.Flags &= ^x
	return e.Trace.Flags
}

func (e *Engine) PushRhx0(s *State, x *Rule, l EngineStateContext, lx *Opnd) {
	if e.Trace != nil {
		e.Trace.RuleScope("z=", s, e.Lhx.Cp(), e.Lhx.Cq())
	}
	e.Rhx = NewRHContextFromStateContexts(s, e.Rhr.SM.CX(), e.Lhx)
	e.Rhr.SM = x.Newrhs(e.Rhr.SM, e.Rhx, e.Rhx)
}

func (e *Engine) PushRhx1(s *State, x *Rule, l EngineStateContext, lx *Opnd) {
	if lx != nil {
		e.Lhx.MakeVar(e.Ssy.TakeFn, NewStr(e.Lhr.ToRow()), e.Lhx, e.Lhr.AV)
	}
	if e.Trace != nil {
		e.Trace.RuleScope("==", s, e.Lhx.Cp(), e.Lhx.Cq())
	}
	e.Rhx = NewRHContextFromStateContexts(s, e.Rhr.SM.CX(), e.Lhx)
	e.Rhr.SM = x.Newrhs(e.Rhr.SM, e.Rhx, l)
}

func (e *Engine) PushRhx(x GrammarElement) {
	e.Rhr.SM = x.NewRHX(e.Rhr.SM, e.Rhr.SM.CX(), e.Lhx)
}

func (e *Engine) Matched3E(l, r, x GrammarElement) bool {
	if l != nil {
		e.Lhr.SY = nil
	}
	if r != nil {
		e.Rhr.SY = nil
	}
	e.Rsy = x
	return true
}

func (e *Engine) Matched2E(l, r GrammarElement) bool {
	if l != nil {
		e.Lhr.SY = nil
	}
	if r != nil {
		e.Rhr.SY = nil
	}
	e.Rsy = nil
	return true
}

func (e *Engine) Match() bool {
	for {
		if e.Trace != nil {
			for e.Lhr.SM != nil && e.Lhr.SY == nil {
				e.Trace.TraceShort(e.Lhr, e.Lhr.SM)
				e.Lhr.SM = e.Lhr.SM.Advance(e.Lhr)
			}
			for e.Rhr.SM != nil && e.Rhr.SY == nil {
				e.Trace.TraceShort(e.Rhr, e.Rhr.SM)
				e.Rhr.SM = e.Rhr.SM.Advance(e.Rhr)
			}
			if e.Lhr.SM == nil {
				return true // exit from lhr.sm
			}
			if e.Rhr.SM == nil {
				return true // no more input
			}
			if e.Lhr.SY == e.Ssy.Nil {
				e.Lhr.SY = nil
				continue
			}
			if e.Rhr.SY == e.Ssy.Nil {
				e.Rhr.SY = nil
				continue
			}
			e.Trace.MatchSymbols(e.Lhr.SY, e.Rhr.SY)
			if e.Lhr.SY.Match(e, e.Rhr.SY) {
				continue
			}
			e.Trace.Back(e.Lhr.SY, e.Rhr.SY)
			return false
		} else {
			for e.Lhr.SM != nil && e.Lhr.SY == nil {
				e.Lhr.SM = e.Lhr.SM.Advance(e.Lhr)
			}
			for e.Rhr.SM != nil && e.Rhr.SY == nil {
				e.Rhr.SM = e.Rhr.SM.Advance(e.Rhr)
			}
			if e.Lhr.SM == nil {
				return true // exit from lhr.sm
			}
			if e.Rhr.SM == nil {
				return true // no more input
			}
			if e.Lhr.SY == e.Ssy.Nil {
				e.Lhr.SY = nil
				continue
			}
			if e.Rhr.SY == e.Ssy.Nil {
				e.Rhr.SY = nil
				continue
			}
			if e.Lhr.SY.Match(e, e.Rhr.SY) {
				continue
			}
			return false
		}
	}
}

func (e *Engine) ResolveE(l, r GrammarElement) bool {
	var sta *State
	var x *Rule
	var zl, zr GenMode
	pri := l.Priority(e.Lhx.Pr())

	if e.Trace != nil {
		e.Trace.Resolve(l, r, pri)
	}
	if e.Lhx.Pr() == PRIMASK {
		return false
	}

	if x = e.Gra().Get(r.Token(), l.Token()); x != nil {
		sta = NewState(e, e.Gra(), l, r, e.Input, e.Charpos(), e.Lineno(), e.Charno(), e.Cxn)
		zl = e.Lhr.SM.Save()
		zr = e.Rhr.SM.Save()
		if e.ResolveState(sta, x, nil, r, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(r.Token(), e.Ssy.Nil); x != nil {
		if sta != nil {
			sta = NewState(e, e.Gra(), l, r, e.Input, e.Charpos(), e.Lineno(), e.Charno(), e.Cxn)
			zl = e.Lhr.SM.Save()
			zr = e.Rhr.SM.Save()
		}
		if e.ResolveState(sta, x, nil, r, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(e.Ssy.Nil, e.Ssy.Nil); x != nil {
		if sta != nil {
			sta = NewState(e, e.Gra(), l, r, e.Input, e.Charpos(), e.Lineno(), e.Charno(), e.Cxn)
			zl = e.Lhr.SM.Save()
			zr = e.Rhr.SM.Save()
		}
		if e.ResolveState(sta, x, r, nil, pri, zl, zr) {
			return true
		}
	}

	if x = e.Gra().Get(e.Ssy.Nil, l.Token()); x != nil {
		if sta != nil {
			sta = NewState(e, e.Gra(), l, r, e.Input, e.Charpos(), e.Lineno(), e.Charno(), e.Cxn)
			zl = e.Lhr.SM.Save()
			zr = e.Rhr.SM.Save()
		}
		if e.ResolveState(sta, x, r, nil, pri, zl, zr) {
			return true
		}
	}

	return false
}

func (e *Engine) ResolveState(sta *State, a *Rule, v, s GrammarElement, pri uint, zl, zr GenMode) bool {
	x := a
	y := e.Lhx

	for {
		if x == nil {
			return false
		}
		if x.Allow(pri) {
			if x.Lhlength() == 1 {
				e.Rhr.SY = v
				if x.Off < x.Rhlength() {
					e.PushRhx0(sta, x, e.Lhx, nil)
				}
				e.Lhx = y
				return true
			}
			e.Lhx = NewLHContextFromRule(sta, e.Lhx, x)
			if e.Maxdepth > 0 {
				e.Lhx.CheckDepth(e.Maxdepth)
			}
			e.Rhr.SY = v
			e.Rsy = s
			e.Lhr.SM = x.Newlhs(e.Lhr.SM, e.Lhx)
			e.Lhr.XS = nil
			if x.Match(e) {
				break
			}
			e.Lhr.AV = y.Cp()
			e.Lhx = y
			e.Lhr.SM = zl.Restore()
			e.Rhr.SM = zr.Restore()
		}
		x = x.Nxt
	}
	// writefln("A %4d %4d", x.off, x.rhlength());
	if x.Off < x.Rhlength() {
		e.PushRhx1(sta, x, e.Lhx, e.Lhr.XS)
	}

	e.Lhr.SM = zl.Restore()
	e.Lhx = y
	return true
}

func (e *Engine) Repeat(max uint) bool {
	var w, x GenMode

	e.Lhr.SY = nil
	w = e.Lhr.SM
	x = e.Rhr.SM.Save()

	for i := uint(0); max == 0 || i < max; i++ {
		if e.Maxrepeat == 0 || i < e.Maxrepeat {
			e.Lhr.SM = NewLHModeFromMode(w)
			if !e.Match() {
				break
			}
			x = e.Rhr.SM.Save()
			if e.Trace != nil {
				e.Trace.Repeat(i)
			}
		} else {
			panic("NewMaxRepeatError")
		}
	}

	e.Rhr.SM = x.Restore()
	e.Lhr.SM = w.Ret()
	return true
}

func (e *Engine) Repeatx(max uint) bool {
	var w, x, z GenMode

	b := e.Lhr.Popx()
	e.Lhr.SY = nil
	w = e.Lhr.SM.Save()
	x = e.Rhr.SM.Save()

	for i := uint(0); max == 0 || i < max; i++ {
		if e.Maxrepeat == 0 || i < e.Maxrepeat {
			z = b.NewLHS(w)
			e.Lhr.SM = z
			if !e.Lhx.Ru().Match(e) {
				break
			}
			x = e.Rhr.SM.Save()
			if e.Trace != nil {
				e.Trace.Repeat(i)
			}
		} else {
			panic("NewMaxRepeatError")
		}
	}

	e.Rhr.SM = x.Restore()
	e.Lhr.SY = nil
	e.Lhr.CI = w.CI()
	e.Lhr.CV = w.CV()
	e.Lhr.LK = w.LK()
	e.Lhr.SM = w
	return true
}

func (e *Engine) PushX() {
	e.Lhr.XS = NewOpnd(e.Lhr.XS, e.Rhr.Popx())
}

func (e *Engine) PushR(x GrammarElement) {
	e.Lhr.XS = NewOpnd(e.Lhr.XS, x)
}

func (e *Engine) PushXElem(x GrammarElement) {
	e.Lhr.XS = NewOpnd(e.Lhr.XS, x)
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

func (e *Engine) BindCvar(l, r GrammarElement) bool {
	e.Lhx.MakeVar(l, r, e.Lhx, e.Lhr.AV)
	if e.Trace != nil {
		e.Trace.BindCvar(l, r)
	}
	return true
}

func (e *Engine) BindLvar(l, r GrammarElement) bool {
	e.Lhx.MakeVar(l, r, e.Lhx, e.Lhr.AV)
	if e.Trace != nil {
		e.Trace.BindLvar(l, r)
	}
	return true
}

func (e *Engine) BindXvarE(r GrammarElement) bool {
	l := e.Lhr.Popx()
	if e.Trace != nil {
		e.Trace.BindRvar(l, r)
	}
	e.Lhx.MakeVar(l, r, e.Rhr.SM, e.Lhr.AV)
	if e.Trace != nil {
		e.Trace.BindRvarScope(l, r, e.Rhr.SM)
	}
	return true
}

func (e *Engine) BindXVar() bool {
	l := e.Lhr.Popx()
	r := e.Rhr.Popx()
	if e.Trace != nil {
		e.Trace.BindRvar(l, r)
	}
	e.Lhx.MakeVar(l, r, e.Rhr.SM, e.Lhr.AV)
	if e.Trace != nil {
		e.Trace.BindRvarScope(l, r, e.Rhr.SM)
	}
	return true
}

func (e *Engine) BindUvar(l, r GrammarElement) bool {
	if e.Trace != nil {
		e.Trace.BindRvar(l, r)
	}
	e.Lhx.MakeVar(l, r, e.Rhr.SM, e.Lhr.AV)
	if e.Trace != nil {
		e.Trace.BindRvarScope(l, r, e.Rhr.SM)
	}
	return true
}

func (e *Engine) TakeTvar() bool {
	v := e.Rhr.SM.VvP()
	if v != nil && v.Vk == e.Ssy.TakeFn {
		s := v.Vv.(*Str)
		for _, x := range s.V {
			e.Lhr.Pushx(x)
		}
	}
	return true
}

func (e *Engine) BindTvar() bool {
	l := e.Lhr.Popx()
	v := e.Rhr.SM.VvP()
	var r GrammarElement
	if v != nil && v.Vk == e.Ssy.TakeFn {
		r = v.Vv
	} else {
		r = NewStr([]GrammarElement{})
	}
	if e.Trace != nil {
		e.Trace.BindRvar(l, r)
	}
	e.Lhx.MakeVar(l, r, e.Rhr.SM, e.Lhr.AV)
	if e.Trace != nil {
		e.Trace.BindRvarScope(l, r, e.Rhr.SM)
	}
	return true
}

func (e *Engine) Deref(pk GrammarElement, x LMScope) *Var {
	pp := x.VvP()
	pq := x.VvQ()
	if e.Trace != nil {
		e.Trace.TheRefVars(pk, pp, pq)
	}
	for pp != nil && pk != pp.Vk {
		pp = pp.Vs
	}
	return pp
}

func (e *Engine) TheRef(s GenMode, k GrammarElement, x LMScope) GenMode {
	v := e.Deref(k, x)
	if e.Trace != nil {
		e.Trace.TheRefVar(v)
	}
	if v != nil {
		return NewRFModeFromVar(s, v)
	}
	return s
}

func (e *Engine) TheValue(s GenMode, k GrammarElement, x LMScope) GrammarElement {
	v := e.Deref(k, x)
	if e.Trace != nil {
		e.Trace.TheRefVar(v)
	}
	if v != nil {
		return v.ToVal()
	}
	return e.Ssy.ZLM
}

func (e *Engine) EachRef(s GenMode, k GrammarElement, x LMScope) GenMode {
	pp := x.VvP()
	pq := x.VvQ()
	if e.Trace != nil {
		e.Trace.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Vk {
			if e.Trace != nil {
				e.Trace.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.Vs
	}
	return s
}

func (e *Engine) Lookvars(s GenMode, k GrammarElement, x LMScope, last *Var) GenMode {
	if x == x.VvP() {
		return s
	}
	tz(">LOOK>")
	pp := x.VvP()
	pq := x.VvQ()
	if pq == nil {
		return s
	}
	if e.Trace != nil {
		e.Trace.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq && pp != last {
		if pp.VvS() != nil {
			e.Lookvars(s, k, pp.VvS(), pp)
		}
		fmt.Printf("\tsi: %8d ", pp.Si)
		TxE("----", pp)
		pp = pp.Vs
	}
	tz("<LOOK<")
	return s
}

func (e *Engine) AllRef(s GenMode, k GrammarElement, x LMScope) GenMode {
	pp := x.VvP()
	pq := x.VvQ()
	if pq == nil {
		return s
	}
	if e.Trace != nil {
		e.Trace.EachRefVars(k, pp, pq)
	}
	for pp != nil && pp != pq {
		if k == pp.Vk {
			if e.Trace != nil {
				e.Trace.EachRefVar(pp)
			}
			s = NewRFModeFromVar(s, pp)
		}
		pp = pp.Va
	}
	return s
}

func (e *Engine) Count(k GrammarElement, p, q *Var) uint {
	var i uint
	for i = 0; p != nil && p != q; p = p.Vs {
		if p.Vk == k {
			i++
		}
	}
	return i
}

func (e *Engine) ToElements(k GrammarElement, p, q *Var) []GrammarElement {
	n := e.Count(k, p, q)
	r := make([]GrammarElement, n)
	for i := int(n); i > 0; p = p.Vs {
		if p.Vk == k {
			i--
			r[i] = p.Vv
		}
	}
	return r
}

func (e *Engine) ToString(k GrammarElement, p, q *Var) string {
	var r string
	for p != nil && p != q {
		if p.Vk == k {
			r = p.Vv.ToString() + r
		}
		p = p.Vs
	}
	return r
}
