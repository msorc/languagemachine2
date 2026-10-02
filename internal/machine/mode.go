package machine

type GenMode interface {
	ScopeHolder
	SelfPointer[GenMode]
	Stream() *Stream
	Variables() VarElement
	ReferenceContext() ScopeHolder
	CurrentSymbol() Element
	ContextMode() ContextHolder
	CodeIndex() int
	CodeVector() []Element
	Operands() OpStack
	Return() GenMode
	Advance() GenMode
	More() GenMode
	Ends() GenMode
	Cont() GenMode
	StackMode() GenMode
	Trace(Element)
	TraceRet(*Tracer)
}

// ScopeHolder
// element generator modes produce symbols for the engine to match
type Mode struct {
	SelfPointing[GenMode]
	stream           *Stream   // stream registers
	currentSymbol    Element   // current symbol
	codeVector       []Element // code vector
	codeIndex        int       // code index
	operands         OpStack
	variables        VarElement    // variables visible in this level
	referenceContext ScopeHolder   // reference context
	contextMode      ContextHolder // mode context
	stackMode        GenMode       // mode stack link
}

// The init methods fill in a mode that is already allocated, so the types
// that embed Mode are built in place.
func (mode *Mode) init(s GenMode) {
	mode.stackMode = s
	mode.stream = s.Stream()
	mode.currentSymbol = s.Stream().currentSymbol
	mode.codeVector = s.Stream().codeVector
	mode.codeIndex = s.Stream().codeIndex
	mode.operands = s.Stream().Operands()
}

func (mode *Mode) initFromVar(s GenMode, v VarElement) {
	mode.init(s)

	mode.contextMode = s.ContextMode()
	mode.referenceContext = v
	mode.variables = v

	mode.stream.currentSymbol = nil
	mode.stream.codeVector = make([]Element, 0)
	mode.stream.codeIndex = 0
}

func (mode *Mode) initFromElements(s GenMode, v []Element, i int, c ContextHolder, x ScopeHolder) {
	mode.init(s)

	mode.contextMode = c
	mode.referenceContext = x
	mode.variables = x.ScopeVariables()

	mode.stream.currentSymbol = nil
	mode.stream.codeVector = v
	mode.stream.codeIndex = i
}

func (mode *Mode) initFromMode(s GenMode) {
	mode.init(s)

	mode.variables = s.Variables()
	mode.referenceContext = s.ReferenceContext()
	mode.contextMode = s.ContextMode()
}

func (m *Mode) Stream() *Stream               { return m.stream }
func (m *Mode) Variables() VarElement         { return m.variables }
func (m *Mode) ReferenceContext() ScopeHolder { return m.referenceContext }
func (m *Mode) ContextMode() ContextHolder    { return m.contextMode }
func (m *Mode) CodeIndex() int                { return m.codeIndex }
func (m *Mode) CodeVector() []Element         { return m.codeVector }
func (m *Mode) CurrentSymbol() Element        { return m.currentSymbol }
func (m *Mode) Operands() OpStack             { return m.operands }
func (m *Mode) StackMode() GenMode            { return m.stackMode }

func (m *Mode) Return() GenMode {
	m.stream.RestoreFromMode(m, false)
	return m.stackMode
}

func (m *Mode) Advance() GenMode {
	return m.stream.Act(m.Self())
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
	TxE(m.stream.Engine.out, "mm", x)
}

// LHS mode: symbols produced from LHS of rules that are being matched
type LHMode struct {
	Mode
}

func NewLHModeFromElement(s GenMode, v []Element, i int, c ContextHolder) *LHMode {
	mode := MakeSelf[LHMode]()
	mode.initFromElements(s, v, i, c, c)
	return mode
}

func NewLHModeFromMode(s GenMode) *LHMode {
	mode := MakeSelf[LHMode]()
	mode.initFromMode(s)
	return mode
}

func (m *LHMode) Return() GenMode {
	m.stream.RestoreFromMode(m, false)
	return nil
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

func (m *LHMode) Trace(x Element) {
	TxE(m.stream.Engine.out, m.contextMode.Trace("lh"), x)
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

func NewRHModeFromParamsAndScope(s GenMode, v []Element, i int, c ContextHolder, x ScopeHolder) *RHMode {
	mode := MakeSelf[RHMode]()
	mode.initFromElements(s, v, i, c, x)
	return mode
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

func (m *RHMode) Trace(x Element) {
	TxE(m.stream.Engine.out, m.contextMode.Trace("rh"), x)
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
	TxE(m.stream.Engine.out, m.contextMode.Trace("lz"), x)
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

func (m *RZMode) ScopeVariables() VarElement {
	return m.variables
}

func (m *RZMode) ScopeContextLimitVariables() VarElement {
	if m.referenceContext != nil {
		return m.referenceContext.ScopeContextLimitVariables()
	}
	return nil
}

func (m *RZMode) Advance() GenMode {
	s := m.Stream()
	s.currentSymbol = m.contextMode.State().GetChr(s.codeIndex)
	s.codeIndex++
	return m.Self()
}

func (m *RZMode) Trace(x Element) {
	TxE(m.stream.Engine.out, m.contextMode.Trace("rz"), x)
}

type STMode struct {
	Mode
}

func NewSTModeFromElements(s GenMode, v []Element, x ScopeHolder) *STMode {
	mode := MakeSelf[STMode]()
	mode.initFromElements(s, v, 0, s.ContextMode(), x)
	return mode
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
	TxE(m.stream.Engine.out, m.contextMode.Trace("st"), x)
}

// RPMode repeats a loop body until a test or a break ends it. The body of a
// for loop is followed by its step, which continue resumes at.
type RPMode struct {
	Mode
	next int // where continue resumes in the body
}

func NewRPModeFromElement(s GenMode, v []Element) *RPMode {
	mode := MakeSelf[RPMode]()
	mode.initFromElements(s, v, 0, s.ContextMode(), s.ReferenceContext())
	return mode
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
	return m.stream.Rep(m.Self())
}

func (m *RPMode) Trace(x Element) {
	TxE(m.stream.Engine.out, m.contextMode.Trace("rp"), x)
}

type RFMode struct {
	Mode
}

func NewRFModeFromVar(s GenMode, v VarElement) *RFMode {
	mode := MakeSelf[RFMode]()
	mode.initFromVar(s, v)
	return mode
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

func (m *RFMode) Trace(x Element) {
	TxE(m.stream.Engine.out, m.contextMode.Trace("rf"), x)
}

// modeSnap is a value copy of the stream registers and the live mode that was
// current when it was taken. It shares no mutable state with the engine, so
// later changes to the live mode or the registers cannot alter it.
type modeSnap struct {
	stream *Stream
	live   GenMode
	sym    Element
	vec    []Element
	idx    int
	ops    OpStack // persistent list: later pushes and pops leave it intact
}

func snapshot(m GenMode) modeSnap {
	s := m.Stream()
	return modeSnap{s, m, s.currentSymbol, s.codeVector, s.codeIndex, s.operands}
}

// restore puts the registers back and returns the mode that was current.
func (z modeSnap) restore() GenMode {
	z.stream.currentSymbol = z.sym
	z.stream.codeVector = z.vec
	z.stream.codeIndex = z.idx
	z.stream.operands = z.ops
	return z.live
}
