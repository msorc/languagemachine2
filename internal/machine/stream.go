package machine

import (
	"fmt"
	"languagemachine2/internal/utils"

	"github.com/liyue201/gostl/ds/list/bidlist"
)

type Stream struct {
	mode      GenMode // stream mode
	codeIndex uint    // code index

	currentSymbol Element // current symbol
	currentValue  Element // current value

	operands  bidlist.List[Element]
	variables VarElement // list of all variables
	engine    *Engine    // the engine

	start    Element
	eof      Element
	nil      Element
	zlm      Element
	put      Element
	mark     Element
	dropFn   Element
	getFn    Element
	strFn    Element
	actFn    Element
	bindFn   Element
	takeFn   Element
	doneFn   Element
	injFn    Element
	appendFn Element
	repeatFn Element
	optionFn Element
	repeatFx Element
	optionFx Element

	qualifier  string    // for tracing
	codeVector []Element // code vector

	compiledRulesCodeIndex uint // code index from compiled rules
}

func NewStream() *Stream {
	s := &Stream{}
	return s
}

func NewStreamFromString(q string) *Stream {
	s := NewStream()
	s.qualifier = q
	return s
}

func NewStreamFromEngine(e *Engine, q string, i uint) *Stream {
	s := NewStream()
	s.engine = e
	s.qualifier = q
	s.codeIndex = i
	return s
}

func (s *Stream) Act(st *Stream, m GenMode) GenMode {
	if s.codeIndex < uint(len(s.codeVector)) {
		s.codeIndex++
		mode := s.codeVector[s.codeIndex-1].Act(st, m)
		return mode
	} else {
		return m.Ret()
	}
}

func (s *Stream) Rep(st *Stream, m GenMode) GenMode {
	if s.codeIndex < uint(len(s.codeVector)) {
		s.codeIndex++
		mode := s.codeVector[s.codeIndex-1].Act(st, m)
		return mode
	} else {
		s.codeIndex = 0
		return m.Ret()
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

func (s *Stream) Countx() uint {
	return uint(s.operands.Len())
}

func (s *Stream) CountXBefore(k Element) uint {
	var n uint

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

func (s *Stream) TerminalSymbols() *Dict {
	return s.engine.terminalSymbols
}

func (s *Stream) FunctionSymbols() *Dict {
	return s.engine.functionSymbols
}

func (s *Stream) NonTerminalSymbols() *Dict {
	return s.engine.nonTerminalSymbols
}

func (s *Stream) UserSymbols() *Dict {
	return s.engine.userSymbols
}

func (s *Stream) PredefinedSymbols() *Predef { return s.engine.predefinedSymbols }

func (s *Stream) TheRef(sMode GenMode, k Element, x LMScope) GenMode {
	return s.engine.TheRef(sMode, k, x)
}

func (s *Stream) EachRef(sMode GenMode, k Element, x LMScope) GenMode {
	return s.engine.EachRef(sMode, k, x)
}

func (s *Stream) AllRef(sMode GenMode, k Element, x LMScope) GenMode {
	return s.engine.AllRef(sMode, k, x)
}

func (s *Stream) BindCvar(l, r Element) bool {
	return s.engine.BindCvar(l, r)
}

func (s *Stream) ExternalSystem() *LMExternal {
	return s.engine.externalSystem
}

func (s *Stream) Initialise(sStream *Stream) {}

func (s *Stream) MakeNt(x int) Element {
	return NewNumber(LMNumber(x))
}

func (s *Stream) MakeMt(x string) Element {
	if x == "null" {
		return s.engine.predefinedSymbols.nil
	}
	return s.engine.nonTerminalSymbols.UniqueE(NewSym(x))
}

func (s *Stream) MakeDt(x string) Element {
	return NewQuote(s.engine.nonTerminalSymbols.UniqueE(NewSym(x)))
}

func (s *Stream) MakeTt(x string) Element {
	return s.engine.terminalSymbols.UniqueR(rune(utils.Unescape(utils.Decode(x))[0]))
}

func (s *Stream) MakeVt(x string) Element {
	return s.engine.varSymbols.UniqueE(NewVarSym(x))
}

func (s *Stream) Makext(x string) Element {
	return s.engine.nonTerminalSymbols.UniqueE(NewLexFromEngine(x, s.engine))
}

func (s *Stream) SetSymbols(x *Predef) {
	s.start = x.start
	s.eof = x.eof
	s.nil = x.nil
	s.zlm = x.zlm
	s.put = x.put
	s.mark = x.mark
	s.dropFn = x.dropFn
	s.getFn = x.getFn
	s.strFn = x.strFn
	s.actFn = x.actFn
	s.bindFn = x.bindFn
	s.takeFn = x.takeFn
	s.doneFn = x.doneFn
	s.injFn = x.injFn
	s.appendFn = x.appendFn
	s.repeatFn = x.repeatFn
	s.optionFn = x.optionFn
	s.repeatFx = x.repeatFx
	s.optionFx = x.optionFx
}

func (s *Stream) M(p uint) uint {
	return LPri(p)
}

func (s *Stream) L(p uint) uint {
	return LPri(p)
}

func (s *Stream) R(p uint) uint {
	return RPri(p)
}

func (s *Stream) B(p uint) uint {
	return BPri(p)
}

func (s *Stream) Ztr(str string, x any) {
	if (s.engine.tracer != nil) && (s.engine.tracer.Flags&DEBUG == DEBUG) {
		fmt.Printf("\t%s %5s %4d %4d %8x\n", s.qualifier, str, s.compiledRulesCodeIndex, s.codeIndex, x)
	}
}
