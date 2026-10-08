package machine

// opStack is an immutable (persistent) operand stack. Copying the value is an
// O(1) snapshot that later pushes and pops cannot disturb, which is what the
// mode snapshots (modeSnap) need when the engine backtracks.
type opStack struct {
	top *opNode
	n   int
}

type opNode struct {
	value Element
	next  *opNode
}

func (o *opStack) push(x Element) {
	o.top = &opNode{value: x, next: o.top}
	o.n++
}

// Pop removes and returns the top element, or nil if the stack is empty.
func (o *opStack) pop() Element {
	if o.top == nil {
		return nil
	}
	x := o.top.value
	o.top = o.top.next
	o.n--
	return x
}

func (o opStack) len() int    { return o.n }
func (o opStack) empty() bool { return o.n == 0 }
func (o *opStack) clear()     { *o = opStack{} }

// Each visits the elements from the top (most recently pushed) down, until f
// returns false.
func (o opStack) each(f func(Element) bool) {
	for p := o.top; p != nil; p = p.next {
		if !f(p.value) {
			return
		}
	}
}

// toSlice returns the elements oldest first, without changing the stack.
func (o opStack) toSlice() []Element {
	v := make([]Element, o.n)
	i := o.n
	for p := o.top; p != nil; p = p.next {
		i--
		v[i] = p.value
	}
	return v
}

type Stream struct {
	mode      GenMode // stream mode
	codeIndex int     // code index

	currentSymbol Element // current symbol

	operands  opStack
	variables varElement // list of all variables
	Engine    *Engine    // the engine

	qualifier  string    // for tracing
	codeVector []Element // code vector
}

func newStream(e *Engine, q string, i int) *Stream {
	return &Stream{Engine: e, qualifier: q, codeIndex: i}
}

func (s *Stream) act(m GenMode) GenMode {
	if s.codeIndex < len(s.codeVector) {
		s.codeIndex++
		return s.codeVector[s.codeIndex-1].act(s, m)
	}
	return m.exit()
}

func (s *Stream) modeAdvance() {
	for s.mode != nil && s.currentSymbol == nil {
		s.Engine.tracer.traceShort(s.mode)
		s.mode = s.mode.advance()
	}
}

func (s *Stream) rep(m GenMode) GenMode {
	if s.codeIndex < len(s.codeVector) {
		s.codeIndex++
		mode := s.codeVector[s.codeIndex-1].act(s, m)
		return mode
	}
	// end of the loop body: start it again; the loop ends when a test fails
	s.codeIndex = 0
	return m
}

func (s *Stream) Operands() opStack {
	return s.operands
}

// restoreFromMode puts back the registers that mode saved when it started;
// the operands stay as they are, so a mode returns its results on the stack.
func (s *Stream) restoreFromMode(mode *mode) {
	s.currentSymbol = mode.currentSymbol
	s.codeVector = mode.codeVector
	s.codeIndex = mode.codeIndex
}

func (s *Stream) clearX() {
	s.operands.clear()
}

func (s *Stream) emptyX() bool {
	return s.operands.empty()
}

func (s *Stream) getX(m GenMode) GenMode {
	s.codeIndex++
	s.pushX(s.codeVector[s.codeIndex-1])
	return m
}

func (s *Stream) pushX(x Element) {
	s.operands.push(x)
}

// popX pops an operand; an empty stack is a fault in the rules, not a nil
// element.
func (s *Stream) popX() Element {
	if s.operands.empty() {
		fail("operand stack underflow")
	}
	return s.operands.pop()
}

func (s *Stream) countX() int {
	return s.operands.len()
}

func (s *Stream) countXBefore(k Element) int {
	var n int

	s.operands.each(func(x Element) bool {
		if x == k {
			return false
		}
		n++
		return true
	})

	return n
}

func (s *Stream) dumpXPlain() {
	s.operands.each(func(e Element) bool {
		s.Engine.printf("\tx: %s\n", e.ToString())
		return true
	})
}

func (s *Stream) dumpX() {
	s.dumpXPlain()
	s.Engine.printf("------\n")
}

func (s *Stream) toRow() []Element {
	v := make([]Element, s.countX())
	for i := len(v); i > 0; i-- {
		v[i-1] = s.popX()
	}
	return v
}

// toArgv pops a call's operands, which are the function, the mark k pushed
// by args, and the arguments. It returns the function followed by the
// arguments; the mark is dropped.
func (s *Stream) toArgv(k Element) []Element {
	v := make([]Element, s.countXBefore(k)+1)
	for i := len(v) - 1; i > 0; i-- {
		v[i] = s.popX()
	}
	s.popX() // the mark
	v[0] = s.popX()
	return v
}
