package machine

type GenMode interface {
	scopeHolder
	Stream() *Stream
	Variables() varElement
	ReferenceContext() scopeHolder
	ContextMode() contextHolder
	exit() GenMode
	advance() GenMode
	more() GenMode
	ends() GenMode
	cont() GenMode
	StackMode() GenMode
	trace(Element)
	traceRet(*tracer)
}

// Mode is a generator of symbols for the engine to match: the left side of a
// rule, the right side, the input, a statement block, a loop or a variable
// reference. A mode saves the stream registers it replaces and restores them
// when it returns.
//
// The kinds of mode embed Mode and define Advance, the step that produces
// the next symbol; Mode itself is not a GenMode.
type mode struct {
	tag              string        // names the kind of mode in traces
	stream           *Stream       // stream registers
	currentSymbol    Element       // current symbol
	codeVector       []Element     // code vector
	codeIndex        int           // code index
	variables        varElement    // variables visible in this level
	referenceContext scopeHolder   // reference context
	contextMode      contextHolder // mode context
	stackMode        GenMode       // mode stack link
}

// The init methods fill in a mode that is already allocated, so the types
// that embed Mode are built in place.
func (mode *mode) init(s GenMode) {
	mode.stackMode = s
	mode.stream = s.Stream()
	mode.currentSymbol = s.Stream().currentSymbol
	mode.codeVector = s.Stream().codeVector
	mode.codeIndex = s.Stream().codeIndex
}

func (mode *mode) initFromVar(s GenMode, v varElement) {
	mode.init(s)

	mode.contextMode = s.ContextMode()
	mode.referenceContext = v
	mode.variables = v

	mode.stream.currentSymbol = nil
	mode.stream.codeVector = make([]Element, 0)
	mode.stream.codeIndex = 0
}

func (mode *mode) initFromElements(s GenMode, v []Element, i int, c contextHolder, x scopeHolder) {
	mode.init(s)

	mode.contextMode = c
	mode.referenceContext = x
	mode.variables = x.ScopeVariables()

	mode.stream.currentSymbol = nil
	mode.stream.codeVector = v
	mode.stream.codeIndex = i
}

func (mode *mode) initFromMode(s GenMode) {
	mode.init(s)

	mode.variables = s.Variables()
	mode.referenceContext = s.ReferenceContext()
	mode.contextMode = s.ContextMode()
}

func (m *mode) Stream() *Stream               { return m.stream }
func (m *mode) Variables() varElement         { return m.variables }
func (m *mode) ReferenceContext() scopeHolder { return m.referenceContext }
func (m *mode) ContextMode() contextHolder    { return m.contextMode }
func (m *mode) StackMode() GenMode            { return m.stackMode }

func (m *mode) exit() GenMode {
	m.stream.restoreFromMode(m)
	return m.stackMode
}

func (m *mode) ScopeVariables() varElement {
	return m.variables
}

func (m *mode) scopeContextLimitVariables() varElement {
	if m.referenceContext != nil {
		return m.referenceContext.scopeContextLimitVariables()
	}
	return nil
}

func (m *mode) scopeReferenceContext() scopeHolder {
	return m.referenceContext
}

func (m *mode) scopeContextMode() contextHolder {
	return m.contextMode
}

func (m *mode) makeVar(k, v Element, s scopeHolder, a varElement) varElement {
	return m.referenceContext.makeVar(k, v, s, a)
}

func (m *mode) more() GenMode {
	return m.stackMode.more()
}

func (m *mode) ends() GenMode {
	return m.stackMode.ends()
}

func (m *mode) cont() GenMode {
	return m.stackMode.cont()
}

func (m *mode) traceRet(t *tracer) {
}

// Trace writes the TraceDebug trace line for the element x about to act.
func (m *mode) trace(x Element) {
	if m.tag == "" {
		traceElement(m.stream.Engine.out, "mm", x)
		return
	}
	traceElement(m.stream.Engine.out, m.contextMode.trace(m.tag), x)
}

// LHS mode: symbols produced from LHS of rules that are being matched
type lhMode struct {
	mode
}

func (m *lhMode) advance() GenMode {
	return m.stream.act(m)
}

func newLHMode(s GenMode, v []Element, i int, c contextHolder) *lhMode {
	mode := &lhMode{}
	mode.tag = "lh"
	mode.initFromElements(s, v, i, c, c)
	return mode
}

func newLHModeFrom(s GenMode) *lhMode {
	mode := &lhMode{}
	mode.tag = "lh"
	mode.initFromMode(s)
	return mode
}

func (m *lhMode) exit() GenMode {
	m.stream.restoreFromMode(&m.mode)
	return nil
}

// variable reference scopeHolder
func (m *lhMode) ScopeVariables() varElement {
	return m.referenceContext.ScopeVariables()
}

func (m *lhMode) makeVar(k, v Element, s scopeHolder, a varElement) varElement {
	m.stream.variables = m.referenceContext.makeVar(k, v, s, a)
	return m.stream.variables
}

func (m *lhMode) traceRet(t *tracer) {
	sr := m.Stream()
	if (t.flags&TraceDiagram == TraceDiagram) && m.contextMode.Rule().offset >= m.contextMode.Rule().rhsLen() {
		sr.Engine.display.endLevel("lx", m.contextMode.State().stateIndex, sr.Engine.rhsStream.mode.ContextMode().State().stateIndex, m.contextMode.NestingDepth(), sr.Engine.rhsStream.mode.ContextMode().NestingDepth())
	}
}

// RHS mode: input symbols and symbols produced by RHS of rules that have matched
type rhMode struct {
	mode
}

func (m *rhMode) advance() GenMode {
	return m.stream.act(m)
}

func newRHMode(s GenMode, v []Element, i int, c contextHolder, x scopeHolder) *rhMode {
	mode := &rhMode{}
	mode.tag = "rh"
	mode.initFromElements(s, v, i, c, x)
	return mode
}

func (m *rhMode) ScopeVariables() varElement {
	return m.referenceContext.ScopeVariables()
}

func (m *rhMode) traceRet(t *tracer) {
	sr := m.Stream()
	if (t.flags & TraceDiagram) == TraceDiagram {
		sr.Engine.display.endLevel("rx", sr.Engine.lhsContext.State().stateIndex, m.contextMode.State().stateIndex, sr.Engine.lhsContext.NestingDepth(), m.contextMode.NestingDepth())
	}
}

type lzMode struct {
	mode
}

func newLZMode(z contextHolder, s *Stream) *lzMode {
	mode := &lzMode{}
	mode.tag = "lz"

	mode.contextMode = z
	mode.stream = s

	return mode
}

func (m *lzMode) advance() GenMode {
	s := m.Stream()
	s.currentSymbol = s.Engine.predefinedSymbols.eof
	s.codeIndex++
	if s.codeIndex > 1 {
		return nil
	}
	return m
}

type rzMode struct {
	mode
}

func newRZMode(z contextHolder, s *Stream) *rzMode {
	mode := &rzMode{}
	mode.tag = "rz"

	mode.contextMode = z
	mode.stream = s

	return mode
}

func (m *rzMode) advance() GenMode {
	s := m.Stream()
	s.currentSymbol = m.contextMode.State().getChr(s.codeIndex)
	s.codeIndex++
	return m
}

type stMode struct {
	mode
}

func (m *stMode) advance() GenMode {
	return m.stream.act(m)
}

func newSTMode(s GenMode, v []Element, x scopeHolder) *stMode {
	mode := &stMode{}
	mode.tag = "st"
	mode.initFromElements(s, v, 0, s.ContextMode(), x)
	return mode
}

// rpMode repeats a loop body until a test or a break ends it. The body of a
// for loop is followed by its step, which continue resumes at.
type rpMode struct {
	mode
	next int // where continue resumes in the body
}

func newRPMode(s GenMode, v []Element) *rpMode {
	mode := &rpMode{}
	mode.tag = "rp"
	mode.initFromElements(s, v, 0, s.ContextMode(), s.ReferenceContext())
	return mode
}

func (m *rpMode) more() GenMode {
	return m
}

func (m *rpMode) ends() GenMode {
	return m.exit()
}

func (m *rpMode) cont() GenMode {
	m.codeIndex = 0
	return m
}

func (m *rpMode) advance() GenMode {
	return m.stream.rep(m)
}

type rfMode struct {
	mode
}

func newRFMode(s GenMode, v varElement) *rfMode {
	mode := &rfMode{}
	mode.tag = "rf"
	mode.initFromVar(s, v)
	return mode
}

func (m *rfMode) advance() GenMode {
	s := m.Stream()
	return m.variables.Value().reference(s, m.exit(), m.variables.scopeReferenceContext())
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
	ops    opStack // persistent list: later pushes and pops leave it intact
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
