package machine

// OpStack is an immutable (persistent) operand stack. Copying the value is an
// O(1) snapshot that later pushes and pops cannot disturb, which is what mode
// Save/Restore needs when the engine backtracks.
type OpStack struct {
	top *opNode
	n   int
}

type opNode struct {
	value Element
	next  *opNode
}

func (o *OpStack) Push(x Element) {
	o.top = &opNode{value: x, next: o.top}
	o.n++
}

// Pop removes and returns the top element, or nil if the stack is empty.
func (o *OpStack) Pop() Element {
	if o.top == nil {
		return nil
	}
	x := o.top.value
	o.top = o.top.next
	o.n--
	return x
}

// Front returns the top element, or nil if the stack is empty.
func (o OpStack) Front() Element {
	if o.top == nil {
		return nil
	}
	return o.top.value
}

func (o OpStack) Len() int    { return o.n }
func (o OpStack) Empty() bool { return o.n == 0 }
func (o *OpStack) Clear()     { *o = OpStack{} }

// Each visits the elements from the top (most recently pushed) down, until f
// returns false.
func (o OpStack) Each(f func(Element) bool) {
	for p := o.top; p != nil; p = p.next {
		if !f(p.value) {
			return
		}
	}
}

// ToSlice returns the elements oldest first, without changing the stack.
func (o OpStack) ToSlice() []Element {
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

	operands  OpStack
	variables VarElement // list of all variables
	Engine    *Engine    // the engine

	qualifier  string    // for tracing
	codeVector []Element // code vector
}

func NewStream(e *Engine, q string, i int) *Stream {
	return &Stream{Engine: e, qualifier: q, codeIndex: i}
}

func (s *Stream) Act(m GenMode) GenMode {
	if s.codeIndex < len(s.codeVector) {
		s.codeIndex++
		mode := s.codeVector[s.codeIndex-1].Act(s, m)
		return mode
	} else {
		return m.Return()
	}
}

func (s *Stream) ModeAdvance() {
	for s.mode != nil && s.currentSymbol == nil {
		if s.Engine.tracer != nil {
			s.Engine.tracer.TraceShort(s.mode)
		}
		s.mode = s.mode.Advance()
	}
}

func (s *Stream) Rep(m GenMode) GenMode {
	if s.codeIndex < len(s.codeVector) {
		s.codeIndex++
		mode := s.codeVector[s.codeIndex-1].Act(s, m)
		return mode
	}
	// end of the loop body: start it again; the loop ends when a test fails
	s.codeIndex = 0
	return m
}

func (s *Stream) Operands() OpStack {
	return s.operands
}

func (s *Stream) RestoreFromMode(mode GenMode, restoreOperands bool) {
	s.currentSymbol = mode.CurrentSymbol()
	s.codeVector = mode.CodeVector()
	s.codeIndex = mode.CodeIndex()
	if restoreOperands {
		// Mode Restore
		s.operands = mode.Operands()
	} // else Mode Return
}

func (s *Stream) ClearX() {
	s.operands.Clear()
}

func (s *Stream) EmptyX() bool {
	return s.operands.Empty()
}

func (s *Stream) Getx(m GenMode) GenMode {
	s.codeIndex++
	s.Pushx(s.codeVector[s.codeIndex-1])
	return m
}

func (s *Stream) Pushx(x Element) Element {
	s.operands.Push(x)
	return x
}

// Popx pops an operand; an empty stack is a fault in the rules, not a nil
// element.
func (s *Stream) Popx() Element {
	if s.operands.Empty() {
		fail("operand stack underflow")
	}
	return s.operands.Pop()
}

func (s *Stream) Countx() int {
	return s.operands.Len()
}

func (s *Stream) CountXBefore(k Element) int {
	var n int

	s.operands.Each(func(x Element) bool {
		if x == k {
			return false
		}
		n++
		return true
	})

	return n
}

func (s *Stream) DumpXPlain() {
	s.operands.Each(func(e Element) bool {
		s.Engine.printf("\tx: %s\n", e.ToString())
		return true
	})
}

func (s *Stream) Dumpx() {
	s.DumpXPlain()
	s.Engine.printf("------\n")
}

func (s *Stream) ToRow() []Element {
	v := make([]Element, s.Countx())
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	return v
}

// ToArgv pops a call's operands, which are the function, the mark k pushed
// by args, and the arguments. It returns the function followed by the
// arguments; the mark is dropped.
func (s *Stream) ToArgv(k Element) []Element {
	v := make([]Element, s.CountXBefore(k)+1)
	for i := len(v) - 1; i > 0; i-- {
		v[i] = s.Popx()
	}
	s.Popx() // the mark
	v[0] = s.Popx()
	return v
}
