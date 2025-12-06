package machine

import (
	"fmt"

	"github.com/liyue201/gostl/ds/list/bidlist"
)

type Stream struct {
	mode      GenMode // stream mode
	codeIndex int     // code index

	currentSymbol Element // current symbol
	currentValue  Element // current value

	operands  bidlist.List[Element]
	variables VarElement // list of all variables
	Engine    *Engine    // the engine

	qualifier  string    // for tracing
	codeVector []Element // code vector
}

func NewStream(e *Engine, q string, i int) *Stream {
	return &Stream{Engine: e, qualifier: q, codeIndex: i}
}

func (s *Stream) Act(st *Stream, m GenMode) GenMode {
	if s.codeIndex < len(s.codeVector) {
		s.codeIndex++
		mode := s.codeVector[s.codeIndex-1].Act(st, m)
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

func (s *Stream) Rep(st *Stream, m GenMode) GenMode {
	if s.codeIndex < len(s.codeVector) {
		s.codeIndex++
		mode := s.codeVector[s.codeIndex-1].Act(st, m)
		return mode
	} else {
		s.codeIndex = 0
		return m.Return()
	}
}

func (s *Stream) Operands() bidlist.List[Element] {
	return s.operands
}

func (s *Stream) Operand() Element {
	return s.operands.Front()
}

func (s *Stream) RestoreOperands(operands bidlist.List[Element]) {
	s.operands = operands
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
	s.operands.PushFront(x)
	return x
}

func (s *Stream) Popx() Element {
	return s.operands.PopFront()
}

func (s *Stream) PushNum(x int) int {
	s.operands.PushFront(NewNumber(LMNumber(x)))
	return x
}

func (s *Stream) PushDbl(x float64) float64 {
	s.operands.PushFront(NewNumber(LMNumber(x)))
	return x
}

func (s *Stream) PushBool(x bool) bool {
	s.operands.PushFront(NewBoolean(x))
	return x
}

func (s *Stream) Countx() int {
	return s.operands.Len()
}

func (s *Stream) CountXBefore(k Element) int {
	var n int

	for i := s.operands.FrontNode(); i != nil && i.Value != k; i = i.Next() {
		n++
	}

	return n
}

func (s *Stream) DumpXPlain() {
	s.operands.Traversal(func(e Element) bool {
		fmt.Printf("\tx: %s\n", e.ToString())
		return true
	})
}

func (s *Stream) Dumpx() {
	s.DumpXPlain()
	fmt.Println("------")
}

func (s *Stream) DumpxWithString(str string) {
	fmt.Printf("\tstack: %s\n", str)
	s.Dumpx()
}

func (s *Stream) ToRow() []Element {
	v := make([]Element, s.Countx())
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	return v
}

func (s *Stream) ToArgv(k Element) []Element {
	v := make([]Element, s.CountXBefore(k)+1)
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	v[0] = s.Popx() // TODO: why?
	return v
}
