package machine

type GenMode interface {
	ScopeHolder
	Stream() *Stream
	Variables() VarElement
	ReferenceContext() ScopeHolder
	ContextMode() ContextHolder
	Return() GenMode
	Advance() GenMode
	More() GenMode
	Ends() GenMode
	Cont() GenMode
	StackMode() GenMode
	Trace(Element)
	TraceRet(*Tracer)
}

// Mode is a generator of symbols for the engine to match: the left side of a
// rule, the right side, the input, a statement block, a loop or a variable
// reference. A mode saves the stream registers it replaces and restores them
// when it returns.
//
// The kinds of mode embed Mode and define Advance, the step that produces
// the next symbol; Mode itself is not a GenMode.
type Mode struct {
	tag              string        // names the kind of mode in traces
	stream           *Stream       // stream registers
	currentSymbol    Element       // current symbol
	codeVector       []Element     // code vector
	codeIndex        int           // code index
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
func (m *Mode) StackMode() GenMode            { return m.stackMode }

func (m *Mode) Return() GenMode {
	m.stream.RestoreFromMode(m)
	return m.stackMode
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

// Trace writes the DEBUG trace line for the element x about to act.
func (m *Mode) Trace(x Element) {
	if m.tag == "" {
		TxE(m.stream.Engine.out, "mm", x)
		return
	}
	TxE(m.stream.Engine.out, m.contextMode.Trace(m.tag), x)
}

// LHS mode: symbols produced from LHS of rules that are being matched
type LHMode struct {
	Mode
}

func (m *LHMode) Advance() GenMode {
	return m.stream.Act(m)
}

func NewLHModeFromElement(s GenMode, v []Element, i int, c ContextHolder) *LHMode {
	mode := &LHMode{}
	mode.tag = "lh"
	mode.initFromElements(s, v, i, c, c)
	return mode
}

func NewLHModeFromMode(s GenMode) *LHMode {
	mode := &LHMode{}
	mode.tag = "lh"
	mode.initFromMode(s)
	return mode
}

func (m *LHMode) Return() GenMode {
	m.stream.RestoreFromMode(&m.Mode)
	return nil
}

// variable reference ScopeHolder
func (m *LHMode) ScopeVariables() VarElement {
	return m.referenceContext.ScopeVariables()
}

func (m *LHMode) MakeVar(k, v Element, s ScopeHolder, a VarElement) VarElement {
	m.stream.variables = m.referenceContext.MakeVar(k, v, s, a)
	return m.stream.variables
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

func (m *RHMode) Advance() GenMode {
	return m.stream.Act(m)
}

func NewRHModeFromParamsAndScope(s GenMode, v []Element, i int, c ContextHolder, x ScopeHolder) *RHMode {
	mode := &RHMode{}
	mode.tag = "rh"
	mode.initFromElements(s, v, i, c, x)
	return mode
}

func (m *RHMode) ScopeVariables() VarElement {
	return m.referenceContext.ScopeVariables()
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
	mode := &LZMode{}
	mode.tag = "lz"

	mode.contextMode = z
	mode.stream = s

	return mode
}

func (m *LZMode) Advance() GenMode {
	s := m.Stream()
	s.currentSymbol = s.Engine.predefinedSymbols.eof
	s.codeIndex++
	if s.codeIndex > 1 {
		return nil
	}
	return m
}

type RZMode struct {
	Mode
}

func NewRZModeFromContext(z ContextHolder, s *Stream) *RZMode {
	mode := &RZMode{}
	mode.tag = "rz"

	mode.contextMode = z
	mode.stream = s

	return mode
}

func (m *RZMode) Advance() GenMode {
	s := m.Stream()
	s.currentSymbol = m.contextMode.State().GetChr(s.codeIndex)
	s.codeIndex++
	return m
}

type STMode struct {
	Mode
}

func (m *STMode) Advance() GenMode {
	return m.stream.Act(m)
}

func NewSTModeFromElements(s GenMode, v []Element, x ScopeHolder) *STMode {
	mode := &STMode{}
	mode.tag = "st"
	mode.initFromElements(s, v, 0, s.ContextMode(), x)
	return mode
}

// RPMode repeats a loop body until a test or a break ends it. The body of a
// for loop is followed by its step, which continue resumes at.
type RPMode struct {
	Mode
	next int // where continue resumes in the body
}

func NewRPModeFromElement(s GenMode, v []Element) *RPMode {
	mode := &RPMode{}
	mode.tag = "rp"
	mode.initFromElements(s, v, 0, s.ContextMode(), s.ReferenceContext())
	return mode
}

func (m *RPMode) More() GenMode {
	return m
}

func (m *RPMode) Ends() GenMode {
	return m.Return()
}

func (m *RPMode) Cont() GenMode {
	m.codeIndex = 0
	return m
}

func (m *RPMode) Advance() GenMode {
	return m.stream.Rep(m)
}

type RFMode struct {
	Mode
}

func NewRFModeFromVar(s GenMode, v VarElement) *RFMode {
	mode := &RFMode{}
	mode.tag = "rf"
	mode.initFromVar(s, v)
	return mode
}

func (m *RFMode) Advance() GenMode {
	s := m.Stream()
	return m.variables.Value().Reference(s, m.Return(), m.variables.ScopeReferenceContext())
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
