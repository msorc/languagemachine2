package machine

import (
// "errors"
)

type LMScope interface {
	ScopeVariables() VarElement             // variable reference LMScope
	ScopeContextLimitVariables() VarElement // limit of context
	ScopeContextMode() EngineStateContext   // variable context
	ScopeReferenceContext() LMScope         // variable LMScope
	MakeVar(Element, Element, LMScope, VarElement) VarElement
	RfScope() LMScope
}

type GenMode interface {
	LMScope
	SelfPointer[GenMode]
	Stream() *Stream
	Variables() VarElement
	ReferenceContext() LMScope
	ContextMode() EngineStateContext
	CodeIndex() uint
	CodeVector() []Element
	LK() any
	What() uint
	Ret() GenMode
	Restore() GenMode
	Advance(*Stream) GenMode
	Save() GenMode
	More() GenMode
	Ends() GenMode
	Cont() GenMode
	EndRep(GenMode) GenMode
	Trace(Element)
	TraceRet(*Stream, *Tracer)
}

// LMScope
// element generator modes produce symbols for the engine to match
type Mode struct {
	SelfPointing[GenMode]
	stream           *Stream   // stream registers
	currentSymbol    Element   // current symbol
	currentValue     Element   // current value
	codeVector       []Element // code vector
	codeIndex        uint      // code index
	lk               any
	operandsStack    *Opnd              // operand stack
	variables        VarElement         // variables visible in this level
	referenceContext LMScope            // reference context
	contextMode      EngineStateContext // mode context
	stackMode        GenMode            // mode stack link
}

func NewMode() *Mode {
	return MakeSelf[Mode]()
}

func NewModeFromVar(s GenMode, v VarElement) *Mode {
	mode := NewMode()

	mode.stackMode = s
	mode.stream = s.Stream()
	mode.currentSymbol = s.Stream().currentSymbol
	mode.currentValue = s.Stream().currentValue
	mode.codeVector = s.Stream().codeVector
	mode.codeIndex = s.Stream().codeIndex
	mode.operandsStack = s.Stream().operandsStack
	mode.lk = s.Stream().LK
	mode.contextMode = s.ContextMode()
	mode.referenceContext = v
	mode.variables = v

	mode.stream.currentSymbol = nil
	mode.stream.codeVector = make([]Element, 0)
	mode.stream.codeIndex = 0
	mode.stream.LK = nil

	return mode
}

func NewModeFromElements(s GenMode, v []Element, i uint, c EngineStateContext, x LMScope) *Mode {
	mode := NewMode()

	mode.stackMode = s
	mode.stream = s.Stream()
	mode.currentSymbol = s.Stream().currentSymbol
	mode.currentValue = s.Stream().currentValue
	mode.codeVector = s.Stream().codeVector
	mode.codeIndex = s.Stream().codeIndex
	mode.operandsStack = s.Stream().operandsStack
	mode.lk = s.Stream().LK
	mode.contextMode = c
	mode.referenceContext = x
	mode.variables = x.ScopeVariables()

	mode.stream.currentSymbol = nil
	// mode.sr.LK = nil
	mode.stream.codeVector = v
	mode.stream.codeIndex = i

	return mode
}

func NewModeFromMode(s GenMode) *Mode {
	mode := NewMode()

	mode.stackMode = s
	mode.stream = s.Stream()
	mode.currentSymbol = s.Stream().currentSymbol
	mode.currentValue = s.Stream().currentValue
	mode.codeVector = s.Stream().codeVector
	mode.codeIndex = s.Stream().codeIndex
	mode.operandsStack = s.Stream().operandsStack
	mode.lk = s.Stream().LK
	mode.variables = s.Variables()
	mode.referenceContext = s.ReferenceContext()
	mode.contextMode = s.ContextMode()

	return mode
}

func (m *Mode) Stream() *Stream                 { return m.stream }
func (m *Mode) Variables() VarElement           { return m.variables }
func (m *Mode) ReferenceContext() LMScope       { return m.referenceContext }
func (m *Mode) ContextMode() EngineStateContext { return m.contextMode }
func (m *Mode) CodeIndex() uint                 { return m.codeIndex }
func (m *Mode) CodeVector() []Element           { return m.codeVector }
func (m *Mode) LK() any                         { return m.lk }

func (m *Mode) What() uint {
	return 1
}

func (m *Mode) Ret() GenMode {
	m.stream.currentSymbol = m.currentSymbol
	m.stream.currentValue = m.currentValue
	m.stream.codeVector = m.codeVector
	m.stream.codeIndex = m.codeIndex
	m.stream.LK = m.lk
	return m.stackMode
}

func (m *Mode) Restore() GenMode {
	m.stream.operandsStack = m.operandsStack
	m.stream.currentSymbol = m.currentSymbol
	m.stream.currentValue = m.currentValue
	m.stream.codeVector = m.codeVector
	m.stream.codeIndex = m.codeIndex
	m.stream.LK = m.lk
	return m.stackMode
}

func (m *Mode) Advance(s *Stream) GenMode {
	return s.Act(s, m.Self())
}

func (m *Mode) ScopeVariables() VarElement {
	return m.variables
}

func (m *Mode) ScopeContextLimitVariables() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *Mode) ScopeReferenceContext() LMScope {
	return m.referenceContext
}

func (m *Mode) ScopeContextMode() EngineStateContext {
	return m.contextMode
}

func (m *Mode) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	return m.referenceContext.MakeVar(k, v, s, a)
}

func (m *Mode) RfScope() LMScope {
	return m.referenceContext
}

func (m *Mode) Save() GenMode {
	return NewModeFromMode(m)
}

func (m *Mode) More() GenMode {
	return m.stackMode.More()
}

func (m *Mode) Ends() GenMode {
	return m.stackMode.Ends()
}

func (m *Mode) Cont() GenMode {
	return m.stackMode.Cont()
}

func (m *Mode) EndRep(mode GenMode) GenMode {
	return m.Self().Ret()
}

func (m *Mode) TraceRet(sr *Stream, t *Tracer) {
}

func (m *Mode) Trace(x Element) {
	TxE("mm", x)
}

// LHS mode: symbols produced from LHS of rules that are being matched
type LHMode struct {
	Mode
}

func NewLHMode() *LHMode {
	return MakeSelf[LHMode]()
}

func NewLHModeFromElement(s GenMode, v []Element, i uint, c EngineStateContext) *LHMode {
	return ReSelf(&LHMode{
		Mode: *NewModeFromElements(s, v, i, c, c),
	})
}

func NewLHModeFromMode(s GenMode) *LHMode {
	return ReSelf(&LHMode{
		Mode: *NewModeFromMode(s),
	})
}

func (m *LHMode) Ret() GenMode {
	m.stream.currentSymbol = m.currentSymbol
	m.stream.codeVector = m.codeVector
	m.stream.codeIndex = m.codeIndex
	m.stream.LK = m.lk
	return nil
}

func (m *LHMode) Save() GenMode {
	return NewLHModeFromMode(m)
}

// variable reference lmScope
func (m *LHMode) ScopeVariables() VarElement {
	return m.referenceContext.ScopeVariables()
}

// limit of context
func (m *LHMode) vvq() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *LHMode) ScopeReferenceContext() LMScope {
	return m.referenceContext
}

func (m *LHMode) ScopeContextMode() EngineStateContext {
	return m.contextMode
}

func (m *LHMode) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	m.stream.variables = m.referenceContext.MakeVar(k, v, s, a)
	return m.stream.variables
}

func (m *LHMode) RfScope() LMScope {
	return m.referenceContext
}

// func (m *LHMode) Advance(s *Stream) GenMode {
//     return s.Act(s, m.Self())
// }

func (m *LHMode) Trace(x Element) {
	TxE(m.contextMode.Trace("lh"), x)
}

func (m *LHMode) TraceRet(sr *Stream, t *Tracer) {
	if (t.Flags&DIAGRAM == DIAGRAM) && m.contextMode.Rule().offset >= m.contextMode.Rule().Rhlength() {
		sr.engine.display.EndLevel("lx", m.contextMode.State().stateIndex, sr.engine.rhsStream.mode.ContextMode().State().stateIndex, m.contextMode.NestingDepth(), sr.engine.rhsStream.mode.ContextMode().NestingDepth())
	}
}

// RHS mode: input symbols and symbols produced by RHS of rules that have matched
type RHMode struct {
	Mode
}

func NewRHMode() *RHMode {
	return MakeSelf[RHMode]()
}

func NewRHModeFromParams(s GenMode, v []Element, i uint, c EngineStateContext) *RHMode {
	return ReSelf(&RHMode{
		Mode: *NewModeFromElements(s, v, i, c, c),
	})
}

func NewRHModeFromParamsAndScope(s GenMode, v []Element, i uint, c EngineStateContext, x LMScope) *RHMode {
	return ReSelf(&RHMode{
		Mode: *NewModeFromElements(s, v, i, c, x),
	})
}

func NewRHModeFromMode(s GenMode) *RHMode {
	return ReSelf(&RHMode{
		Mode: *NewModeFromMode(s),
	})
}

func (m *RHMode) Save() GenMode {
	return NewRHModeFromMode(m)
}

func (m *RHMode) ScopeVariables() VarElement {
	return m.referenceContext.ScopeVariables()
}

func (m *RHMode) ScopeContextLimitVariables() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *RHMode) ScopeReferenceContext() LMScope {
	return m.referenceContext
}

func (m *RHMode) ScopeContextMode() EngineStateContext {
	return m.contextMode
}

func (m *RHMode) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	return m.referenceContext.MakeVar(k, v, s, a)
}

// func (m *RHMode) Advance(s *Stream) GenMode {
//     return s.Act(m.Self())
// }

func (m *RHMode) Trace(x Element) {
	TxE(m.contextMode.Trace("rh"), x)
}

func (m *RHMode) TraceRet(sr *Stream, t *Tracer) {
	if (t.Flags & DIAGRAM) == DIAGRAM {
		sr.engine.display.EndLevel("rx", sr.engine.lhsContext.State().stateIndex, m.contextMode.State().stateIndex, sr.engine.lhsContext.NestingDepth(), m.contextMode.NestingDepth())
	}
}

type LZMode struct {
	Mode
}

func NewLZModeFromContext(z EngineStateContext, s *Stream) *LZMode {
	mode := MakeSelf[LZMode]()

	mode.contextMode = z
	mode.stream = s

	return mode
}

func NewLZModeFromMode(s GenMode) *LZMode {
	return ReSelf(&LZMode{Mode: *NewModeFromMode(s)})
}

func (m *LZMode) Save() GenMode {
	return NewLZModeFromMode(m)
}

func (m *LZMode) ScopeVariables() VarElement {
	return m.variables
}

func (m *LZMode) ScopeContextLimitVariables() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *LZMode) ScopeReferenceContext() LMScope {
	return m.referenceContext
}

func (m *LZMode) ScopeContextMode() EngineStateContext {
	return m.contextMode
}

func (m *LZMode) Advance(s *Stream) GenMode {
	s.currentSymbol = s.PredefinedSymbols().eof
	ci := s.codeIndex
	s.codeIndex++
	if ci > 0 {
		return nil
	}
	return m.Self()
}

func (m *LZMode) Trace(x Element) {
	TxE(m.contextMode.Trace("lz"), x)
}

type RZMode struct {
	Mode
}

func NewRZModeFromContext(z EngineStateContext, s *Stream) *RZMode {
	mode := MakeSelf[RZMode]()

	mode.contextMode = z
	mode.stream = s

	return mode
}

func NewRZModeFromMode(s GenMode) *RZMode {
	return ReSelf(&RZMode{Mode: *NewModeFromMode(s)})
}

func (m *RZMode) Save() GenMode {
	return NewRZModeFromMode(m)
}

func (m *RZMode) ScopeVariables() VarElement {
	return m.variables
}

func (m *RZMode) ScopeContextLimitVariables() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *RZMode) Vvs() LMScope {
	return m.referenceContext
}

func (m *RZMode) Vvc() EngineStateContext {
	return m.contextMode
}

func (m *RZMode) Advance(s *Stream) GenMode {
	s.currentSymbol = m.contextMode.State().GetChr(s.codeIndex)
	s.codeIndex++
	return m
}

func (m *RZMode) Trace(x Element) {
	TxE(m.contextMode.Trace("rz"), x)
}

type STMode struct {
	Mode
}

func NewSTMode() *STMode {
	return MakeSelf[STMode]()
}

func NewSTModeFromElements(s GenMode, v []Element, x LMScope) *STMode {
	return ReSelf(&STMode{Mode: *NewModeFromElements(s, v, 0, s.ContextMode(), x)})
}

func NewSTModeFromMode(s GenMode) *STMode {
	return ReSelf(&STMode{Mode: *NewModeFromMode(s)})
}

func (m *STMode) Save() GenMode {
	return NewSTModeFromMode(m)
}

func (m *STMode) ScopeVariables() VarElement {
	return m.variables
}

func (m *STMode) ScopeContextLimitVariables() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *STMode) ScopeReferenceContext() LMScope {
	return m.referenceContext
}

func (m *STMode) ScopeContextMode() EngineStateContext {
	return m.contextMode
}

func (m *STMode) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	return m.referenceContext.MakeVar(k, v, s, a)
}

func (m *STMode) Trace(x Element) {
	TxE(m.contextMode.Trace("st"), x)
}

type RPMode struct {
	Mode
}

func NewRPMode() *RPMode {
	return MakeSelf[RPMode]()
}

func NewRPModeFromElement(s GenMode, v []Element) *RPMode {
	return ReSelf(&RPMode{Mode: *NewModeFromElements(s, v, 0, s.ContextMode(), s.ContextMode())})
}

func NewRPModeFromMode(s GenMode) *RPMode {
	return ReSelf(&RPMode{Mode: *NewModeFromMode(s)})
}

func (m *RPMode) Save() GenMode {
	return NewRPModeFromMode(m)
}

func (m *RPMode) More() GenMode {
	return m
}

func (m *RPMode) Ends() GenMode {
	return m.Self().Ret()
}

func (m *RPMode) Cont() GenMode {
	m.codeIndex = 0
	return m
}

func (m *RPMode) Advance(s *Stream) GenMode {
	return s.Rep(s, m)
}

func (m *RPMode) Trace(x Element) {
	TxE(m.contextMode.Trace("rp"), x)
}

type RFMode struct {
	Mode
}

func NewRFMode() *RFMode {
	return MakeSelf[RFMode]()
}

func NewRFModeFromVar(s GenMode, v VarElement) *RFMode {
	return ReSelf(&RFMode{Mode: *NewModeFromVar(s, v)})
}

func NewRFModeFromMode(s GenMode) *RFMode {
	return ReSelf(&RFMode{Mode: *NewModeFromMode(s)})
}

func (m *RFMode) Save() GenMode {
	return NewRFModeFromMode(m)
}

func (m *RFMode) Advance(s *Stream) GenMode {
	return m.variables.Value().Reference(s, m.Self().Ret(), m.variables.ScopeReferenceContext())
}

func (m *RFMode) ScopeVariables() VarElement {
	return m.variables
}

func (m *RFMode) ScopeContextLimitVariables() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *RFMode) Vvs() LMScope {
	return m.referenceContext
}

func (m *RFMode) Vvc() EngineStateContext {
	return m.contextMode
}

func (m *RFMode) Trace(x Element) {
	TxE(m.contextMode.Trace("rf"), x)
}

type APMode struct {
	Mode
	v Element
}

func NewAPMode() *APMode {
	return MakeSelf[APMode]()
}

func NewAPModeFromElement(s GenMode, x Element) *APMode {
	mode := ReSelf(&APMode{Mode: *NewModeFromMode(s)})
	mode.v = x
	return mode
}

func NewAPModeFromMode(s GenMode) *APMode {
	return ReSelf(&APMode{Mode: *NewModeFromMode(s)})
}

func (m *APMode) Save() GenMode {
	return NewAPModeFromMode(m)
}

func (m *APMode) Advance(s *Stream) GenMode {
	if m.v != nil {
		return m.v.Act(s, m.Ret())
	}
	return m.Self().Ret()
}

func (m *APMode) Trace(x Element) {
	TxE(m.contextMode.Trace("ap"), x)
}

// Operand stack - pushdown list of operand elements
type Opnd struct {
	S *Opnd   // operand stack link
	V Element // operand element
}

func NewOpnd(p *Opnd, x Element) *Opnd {
	return &Opnd{
		S: p,
		V: x,
	}
}

func (o *Opnd) ToString() string {
	if o.V != nil {
		return o.V.ToString()
	}
	return "---"
}

// state information that can be fixed at the start of a context, ie when a mismatch occurs
type State struct {
	engine       *Engine   // the engine - for access to global properties
	grammar      *Grammar  // current grammar
	lsy          Element   // lh symbol at mismatch
	rsy          Element   // rh symbol at mismatch
	input        GrammarIO // input source object
	charPosition uint      // absolute char position in file
	lineNumber   uint      // line number
	charNumber   uint      // char number in line
	stateIndex   uint      // state index or identity
}

func NewState(e *Engine, g *Grammar, l, r Element, i GrammarIO, p uint, n uint, c uint, x uint) *State {
	return &State{
		engine:       e,
		grammar:      g,
		lsy:          l,
		rsy:          r,
		input:        i,
		charPosition: p,
		lineNumber:   n,
		charNumber:   c,
		stateIndex:   x,
	}
}

// Method to get character element
func (s *State) GetChr(ci uint) Element {
	return s.engine.rhsBuffer.GetChr(s.engine, ci)
}

type EngineStateContext interface {
	LMScope
	SelfPointer[EngineStateContext]
	Rule() *Rule
	State() *State
	Priority() uint
	OperandsStack() *Opnd
	Variables() VarElement
	ContextLimitVariable() VarElement
	NestingDepth() uint
	ContextStack() EngineStateContext
	CheckDepth(uint) error
	Trace(string) string
}

// LMScope
// contexts: the state of the engine as rules are applied
type Context struct {
	SelfPointing[EngineStateContext]
	state                *State             // state at start of context
	rule                 *Rule              // rule
	priority             uint               // context priority
	operandsStack        *Opnd              // operand stack
	variables            VarElement         // variables
	contextLimitVariable VarElement         // limit of context
	nestingDepth         uint               // context nesting depth
	contextStack         EngineStateContext // context stack
}

func NewContextFromState(s *State) *Context {
	return ReSelf(&Context{
		state: s,
	})
}

func NewContextFromParams(s *State, c EngineStateContext, x *Rule, n uint, p, q VarElement) *Context {
	return ReSelf(&Context{
		state:                s,
		contextStack:         c,
		priority:             n,
		rule:                 x,
		variables:            p,
		contextLimitVariable: q,
		nestingDepth:         c.NestingDepth() + 1,
	})
}

func NewContextFromContext(x EngineStateContext) *Context {
	return ReSelf(&Context{
		state:                x.State(),
		rule:                 x.Rule(),
		priority:             x.Priority(),
		operandsStack:        x.OperandsStack(),
		variables:            x.Variables(),
		contextLimitVariable: x.ContextLimitVariable(),
		contextStack:         x.ContextStack(),
	})
}

func (c *Context) Copy(x EngineStateContext) EngineStateContext {
	c.state = x.State()
	c.rule = x.Rule()
	c.priority = x.Priority()
	c.operandsStack = x.OperandsStack()
	c.variables = x.Variables()
	c.contextLimitVariable = x.ContextLimitVariable()
	c.contextStack = x.ContextStack()
	return c.Self()
}

func (c *Context) Dup() EngineStateContext {
	return NewContextFromContext(c.Self())
}

func (c *Context) CheckDepth(max uint) error {
	if max == 0 || c.nestingDepth < max {
		return nil
	}
	panic("maxDepthError")
	//return errors.New("maxDepthError")
}

func (c *Context) ScopeVariables() VarElement {
	return c.variables
}

func (c *Context) ScopeContextLimitVariables() VarElement {
	return c.contextLimitVariable
}

func (c *Context) ScopeReferenceContext() LMScope {
	return c
}

func (c *Context) ScopeContextMode() EngineStateContext {
	return c
}

func (c *Context) Rule() *Rule                      { return c.rule }
func (c *Context) State() *State                    { return c.state }
func (c *Context) Priority() uint                   { return c.priority }
func (c *Context) OperandsStack() *Opnd             { return c.operandsStack }
func (c *Context) Variables() VarElement            { return c.variables }
func (c *Context) ContextLimitVariable() VarElement { return c.contextLimitVariable }
func (c *Context) NestingDepth() uint               { return c.nestingDepth }
func (c *Context) ContextStack() EngineStateContext { return c.contextStack }

func (c *Context) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	c.variables = NewVarFromParams(c.variables, k, v, s, a)
	return c.variables
}

func (c *Context) RfScope() LMScope {
	return c.Self()
}

func (c *Context) Trace(s string) string {
	return "C-" + s
}

// LHS context: the context in which a rule is being tried
type LHContext struct {
	Context
}

func NewLHContext() *LHContext {
	return ReSelf(&LHContext{})
}

func NewLHContextFromState(s *State) *LHContext {
	return ReSelf(&LHContext{
		Context: *NewContextFromState(s),
	})
}

func NewLHContextFromContext(c EngineStateContext) *LHContext {
	return ReSelf(&LHContext{
		Context: *NewContextFromContext(c),
	})
}

func NewLHContextFromRule(s *State, c EngineStateContext, x *Rule) *LHContext {
	return ReSelf(&LHContext{
		Context: *NewContextFromParams(s, c, x, x.Cxtpri(c.Priority()), c.Variables(), c.Variables()),
	})
}

func (lh *LHContext) Dup() EngineStateContext {
	return NewLHContextFromContext(lh.Self())
}

func (lh *LHContext) ScopeVariables() VarElement {
	return lh.variables
}

func (lh *LHContext) ScopeContextLimitVariables() VarElement {
	return lh.contextLimitVariable
}

func (lh *LHContext) ScopeReferenceContext() LMScope {
	return lh
}

func (lh *LHContext) ScopeContextMode() EngineStateContext {
	return lh.Self()
}

func (lh *LHContext) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	lh.variables = NewVarFromParams(lh.variables, k, v, s, a)
	return lh.variables
}

func (lh *LHContext) RfScope() LMScope {
	return lh.Self()
}

func (lh *LHContext) Trace(s string) string {
	return "L-" + s
}

// RHS context: used to provide information to RHS modes and to variables
type RHContext struct {
	Context
}

func NewRHContext() *RHContext {
	return ReSelf(&RHContext{})
}

func NewRHContextFromState(s *State) *RHContext {
	return ReSelf(&RHContext{
		Context: *NewContextFromState(s),
	})
}

func NewRHContextFromContext(c EngineStateContext) *RHContext {
	return ReSelf(&RHContext{
		Context: *NewContextFromContext(c),
	})
}

func NewRHContextFromStateContexts(s *State, c, l EngineStateContext) *RHContext {
	return ReSelf(&RHContext{
		Context: *NewContextFromParams(s, c, l.Rule(), l.Priority(), l.Variables(), l.ContextLimitVariable()),
	})
}

func (rh *RHContext) Dup() EngineStateContext {
	return NewRHContextFromContext(rh.Self())
}

func (rh *RHContext) ScopeVariables() VarElement {
	return rh.variables
}

func (rh *RHContext) ScopeContextLimitVariables() VarElement {
	return rh.contextLimitVariable
}

func (rh *RHContext) ScopeReferenceContext() LMScope {
	return rh.Self()
}

func (rh *RHContext) ScopeContextMode() EngineStateContext {
	return rh.Self()
}

func (rh *RHContext) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	rh.variables = NewVarFromParams(rh.variables, k, v, s, a)
	return rh.variables
}

func (rh *RHContext) RfScope() LMScope {
	return rh.Self()
}

func (rh *RHContext) Trace(s string) string {
	return "R-" + s
}
