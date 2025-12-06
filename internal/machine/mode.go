package machine

import "github.com/liyue201/gostl/ds/list/bidlist"

type GenMode interface {
	ScopeHolder
	SelfPointer[GenMode]
	Stream() *Stream
	Variables() VarElement
	ReferenceContext() ScopeHolder
	ContextMode() ContextHolder
	CodeIndex() uint
	CodeVector() []Element
	What() uint
	Return() GenMode
	Restore() GenMode
	Advance() GenMode
	Save() GenMode
	More() GenMode
	Ends() GenMode
	Cont() GenMode
	Trace(Element)
	TraceRet(*Tracer)
}

// ScopeHolder
// element generator modes produce symbols for the engine to match
type Mode struct {
	SelfPointing[GenMode]
	stream           *Stream   // stream registers
	currentSymbol    Element   // current symbol
	currentValue     Element   // current value
	codeVector       []Element // code vector
	codeIndex        uint      // code index
	operands         bidlist.List[Element]
	variables        VarElement    // variables visible in this level
	referenceContext ScopeHolder   // reference context
	contextMode      ContextHolder // mode context
	stackMode        GenMode       // mode stack link
}

func newMode(s GenMode) *Mode {
	mode := MakeSelf[Mode]()

	mode.stackMode = s
	mode.stream = s.Stream()
	mode.currentSymbol = s.Stream().currentSymbol
	mode.currentValue = s.Stream().currentValue
	mode.codeVector = s.Stream().codeVector
	mode.codeIndex = s.Stream().codeIndex
	mode.operands = s.Stream().Operands()

	return mode
}

func NewModeFromVar(s GenMode, v VarElement) *Mode {
	mode := newMode(s)

	mode.contextMode = s.ContextMode()
	mode.referenceContext = v
	mode.variables = v

	mode.stream.currentSymbol = nil
	mode.stream.codeVector = make([]Element, 0)
	mode.stream.codeIndex = 0

	return mode
}

func NewModeFromElements(s GenMode, v []Element, i uint, c ContextHolder, x ScopeHolder) *Mode {
	mode := newMode(s)

	mode.contextMode = c
	mode.referenceContext = x
	mode.variables = x.ScopeVariables()

	mode.stream.currentSymbol = nil
	mode.stream.codeVector = v
	mode.stream.codeIndex = i

	return mode
}

func NewModeFromMode(s GenMode) *Mode {
	mode := newMode(s)

	mode.variables = s.Variables()
	mode.referenceContext = s.ReferenceContext()
	mode.contextMode = s.ContextMode()

	return mode
}

func (m *Mode) Stream() *Stream               { return m.stream }
func (m *Mode) Variables() VarElement         { return m.variables }
func (m *Mode) ReferenceContext() ScopeHolder { return m.referenceContext }
func (m *Mode) ContextMode() ContextHolder    { return m.contextMode }
func (m *Mode) CodeIndex() uint               { return m.codeIndex }
func (m *Mode) CodeVector() []Element         { return m.codeVector }

func (m *Mode) What() uint {
	return 1
}

func (m *Mode) Return() GenMode {
	m.stream.currentSymbol = m.currentSymbol
	m.stream.currentValue = m.currentValue
	m.stream.codeVector = m.codeVector
	m.stream.codeIndex = m.codeIndex
	return m.stackMode
}

func (m *Mode) Restore() GenMode {
	m.stream.RestoreOperands(m.operands)
	m.stream.currentSymbol = m.currentSymbol
	m.stream.currentValue = m.currentValue
	m.stream.codeVector = m.codeVector
	m.stream.codeIndex = m.codeIndex
	return m.stackMode
}

func (m *Mode) Advance() GenMode {
	s := m.Stream()
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

func (m *Mode) ScopeReferenceContext() ScopeHolder {
	return m.referenceContext
}

func (m *Mode) ScopeContextMode() ContextHolder {
	return m.contextMode
}

func (m *Mode) MakeVar(k, v Element, s ScopeHolder, a VarElement) VarElement {
	return m.referenceContext.MakeVar(k, v, s, a)
}

func (m *Mode) RfScope() ScopeHolder {
	return m.referenceContext
}

func (m *Mode) Save() GenMode {
	return NewModeFromMode(m.Self())
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

func (m *Mode) TraceRet(t *Tracer) {
}

func (m *Mode) Trace(x Element) {
	TxE("mm", x)
}

// LHS mode: symbols produced from LHS of rules that are being matched
type LHMode struct {
	Mode
}

func NewLHModeFromElement(s GenMode, v []Element, i uint, c ContextHolder) *LHMode {
	return ReSelf(&LHMode{
		Mode: *NewModeFromElements(s, v, i, c, c),
	})
}

func NewLHModeFromMode(s GenMode) *LHMode {
	return ReSelf(&LHMode{
		Mode: *NewModeFromMode(s),
	})
}

func (m *LHMode) Return() GenMode {
	m.stream.currentSymbol = m.currentSymbol
	m.stream.codeVector = m.codeVector
	m.stream.codeIndex = m.codeIndex
	return nil
}

func (m *LHMode) Save() GenMode {
	return NewLHModeFromMode(m.Self())
}

// variable reference ScopeHolder
func (m *LHMode) ScopeVariables() VarElement {
	return m.referenceContext.ScopeVariables()
}

// limit of context
func (m *LHMode) ScopeReferenceContext() ScopeHolder {
	return m.referenceContext
}

func (m *LHMode) ScopeContextMode() ContextHolder {
	return m.contextMode
}

func (m *LHMode) MakeVar(k, v Element, s ScopeHolder, a VarElement) VarElement {
	m.stream.variables = m.referenceContext.MakeVar(k, v, s, a)
	return m.stream.variables
}

func (m *LHMode) RfScope() ScopeHolder {
	return m.referenceContext
}

// func (m *LHMode) Advance() GenMode {
//     s := m.stream
//     return s.Act(s, m.Self())
// }

func (m *LHMode) Trace(x Element) {
	TxE(m.contextMode.Trace("lh"), x)
}

func (m *LHMode) TraceRet(t *Tracer) {
	sr := m.Stream()
	if (t.Flags&DIAGRAM == DIAGRAM) && m.contextMode.Rule().offset >= m.contextMode.Rule().Rhlength() {
		sr.Engine.display.EndLevel("lx", m.contextMode.State().stateIndex, sr.Engine.rhsStream.mode.ContextMode().State().stateIndex, m.contextMode.NestingDepth(), sr.Engine.rhsStream.mode.ContextMode().NestingDepth())
	}
}

// RHS mode: input symbols and symbols produced by RHS of rules that have matched
type RHMode struct {
	Mode
}

func NewRHModeFromParams(s GenMode, v []Element, i uint, c ContextHolder) *RHMode {
	return ReSelf(&RHMode{
		Mode: *NewModeFromElements(s, v, i, c, c),
	})
}

func NewRHModeFromParamsAndScope(s GenMode, v []Element, i uint, c ContextHolder, x ScopeHolder) *RHMode {
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
	return NewRHModeFromMode(m.Self())
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

func (m *RHMode) ScopeReferenceContext() ScopeHolder {
	return m.referenceContext
}

func (m *RHMode) ScopeContextMode() ContextHolder {
	return m.contextMode
}

func (m *RHMode) MakeVar(k, v Element, s ScopeHolder, a VarElement) VarElement {
	return m.referenceContext.MakeVar(k, v, s, a)
}

// func (m *RHMode) Advance() GenMode {
// 	   s := m.Stream()
//     return s.Act(m.Self())
// }

func (m *RHMode) Trace(x Element) {
	TxE(m.contextMode.Trace("rh"), x)
}

func (m *RHMode) TraceRet(t *Tracer) {
	sr := m.Stream()
	if (t.Flags & DIAGRAM) == DIAGRAM {
		sr.Engine.display.EndLevel("rx", sr.Engine.lhsContext.State().stateIndex, m.contextMode.State().stateIndex, sr.Engine.lhsContext.NestingDepth(), m.contextMode.NestingDepth())
	}
}

type LZMode struct {
	Mode
}

func NewLZModeFromContext(z ContextHolder, s *Stream) *LZMode {
	mode := MakeSelf[LZMode]()

	mode.contextMode = z
	mode.stream = s

	return mode
}

func NewLZModeFromMode(s GenMode) *LZMode {
	return ReSelf(&LZMode{Mode: *NewModeFromMode(s)})
}

func (m *LZMode) Save() GenMode {
	return NewLZModeFromMode(m.Self())
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

func (m *LZMode) ScopeReferenceContext() ScopeHolder {
	return m.referenceContext
}

func (m *LZMode) ScopeContextMode() ContextHolder {
	return m.contextMode
}

func (m *LZMode) Advance() GenMode {
	s := m.Stream()
	s.currentSymbol = s.Engine.predefinedSymbols.eof
	s.codeIndex++
	if s.codeIndex-1 > 0 {
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

func NewRZModeFromContext(z ContextHolder, s *Stream) *RZMode {
	mode := MakeSelf[RZMode]()

	mode.contextMode = z
	mode.stream = s

	return mode
}

func NewRZModeFromMode(s GenMode) *RZMode {
	return ReSelf(&RZMode{Mode: *NewModeFromMode(s)})
}

func (m *RZMode) Save() GenMode {
	return NewRZModeFromMode(m.Self())
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

func (m *RZMode) Vvs() ScopeHolder {
	return m.referenceContext
}

func (m *RZMode) Vvc() ContextHolder {
	return m.contextMode
}

func (m *RZMode) Advance() GenMode {
	s := m.Stream()
	s.currentSymbol = m.contextMode.State().GetChr(s.codeIndex)
	s.codeIndex++
	return m.Self()
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

func NewSTModeFromElements(s GenMode, v []Element, x ScopeHolder) *STMode {
	return ReSelf(&STMode{Mode: *NewModeFromElements(s, v, 0, s.ContextMode(), x)})
}

func NewSTModeFromMode(s GenMode) *STMode {
	return ReSelf(&STMode{Mode: *NewModeFromMode(s)})
}

func (m *STMode) Save() GenMode {
	return NewSTModeFromMode(m.Self())
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

func (m *STMode) ScopeReferenceContext() ScopeHolder {
	return m.referenceContext
}

func (m *STMode) ScopeContextMode() ContextHolder {
	return m.contextMode
}

func (m *STMode) MakeVar(k, v Element, s ScopeHolder, a VarElement) VarElement {
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
	return m.Self()
}

func (m *RPMode) Ends() GenMode {
	return m.Self().Return()
}

func (m *RPMode) Cont() GenMode {
	m.codeIndex = 0
	return m.Self()
}

func (m *RPMode) Advance() GenMode {
	s := m.Stream()
	return s.Rep(s, m.Self())
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
	return NewRFModeFromMode(m.Self())
}

func (m *RFMode) Advance() GenMode {
	s := m.Stream()
	return m.variables.Value().Reference(s, m.Self().Return(), m.variables.ScopeReferenceContext())
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

func (m *RFMode) Vvs() ScopeHolder {
	return m.referenceContext
}

func (m *RFMode) Vvc() ContextHolder {
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

func (m *APMode) Advance() GenMode {
	s := m.Stream()
	if m.v != nil {
		return m.v.Act(s, m.Self().Return())
	}
	return m.Self().Return()
}

func (m *APMode) Trace(x Element) {
	TxE(m.contextMode.Trace("ap"), x)
}
