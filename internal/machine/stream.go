package machine

import (
	"languagemachine2/internal/utils"
	"fmt"
)

type Stream struct {
	mode      GenMode // stream mode
	codeIndex uint    // code index

	currentSymbol Element // current symbol
	currentValue  Element // current value
	returnValue   Element // return value from machine

	operandsStack *Opnd      // operand stack
	variables     VarElement // list of all variables
	engine        *Engine    // the engine

	TT []Element
	MT []Element
	DT []Element
	VT []Element
	XT []Element
	NT []Element
	ST []Element
	FT []Element

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

	LK any // jump address to current point in string
	LX any // jump address to exit from string

	NtV []Element
	MtV []Element
	DtV []Element
	VtV []Element
	XtV []Element
	TtV []Element
	StV []Element
	FtV []Element

	qualifier  string    // for tracing
	codeVector []Element // code vector

	compiledRulesCodeIndex uint // code index from compiled rules
}

func NewStream() *Stream {
	return &Stream{}
}

func NewStreamFromString(s string) *Stream {
	return &Stream{qualifier: s}
}

func NewStreamFromEngine(e *Engine, s string, i uint) *Stream {
	return &Stream{engine: e, qualifier: s, codeIndex: i}
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

func (s *Stream) Getx(m GenMode) GenMode {
	s.codeIndex++
	s.Pushx(s.codeVector[s.codeIndex-1])
	return m
}

func (s *Stream) Pushx(x Element) Element {
	s.operandsStack = NewOpnd(s.operandsStack, x)
	return x
}

func (s *Stream) Popx() Element {
	x := s.operandsStack
	s.operandsStack = x.S
	return x.V
}

func (s *Stream) PushNum(x int) int {
	s.operandsStack = NewOpnd(s.operandsStack, NewNumber(LMNumber(x)))
	return x
}

func (s *Stream) PushDbl(x float64) float64 {
	s.operandsStack = NewOpnd(s.operandsStack, NewNumber(LMNumber(x)))
	return x
}

func (s *Stream) PushBool(x bool) bool {
	s.operandsStack = NewOpnd(s.operandsStack, NewBoolean(x))
	return x
}

func (s *Stream) Countx() uint {
	var n uint
	for x := s.operandsStack; x != nil; x = x.S {
		n++
	}
	return n
}

func (s *Stream) CountxWithElement(k Element) uint {
	var n uint
	for x := s.operandsStack; x != nil && x.V != k; x = x.S {
		n++
	}
	return n
}

func (s *Stream) Dumpx() {
	for x := s.operandsStack; x != nil; x = x.S {
		fmt.Printf("\tx: %s\n", x.ToString())
	}
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
	v := make([]Element, s.CountxWithElement(k)+1)
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	v[0] = s.Popx()
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

// + attention
func (s *Stream) CopyTables(x *Stream) {
	if x.NtV != nil {
		s.NtV = x.NtV
		s.NT = x.NtV[0:]
	}
	if x.MtV != nil {
		s.MtV = x.MtV
		s.MT = x.MtV[0:]
	}
	if x.DtV != nil {
		s.DtV = x.DtV
		s.DT = x.DtV[0:]
	}
	if x.VtV != nil {
		s.VtV = x.VtV
		s.VT = x.VtV[0:]
	}
	if x.XtV != nil {
		s.XtV = x.XtV
		s.XT = x.XtV[0:]
	}
	if x.TtV != nil {
		s.TtV = x.TtV
		s.TT = x.TtV[0:]
	}
	if x.StV != nil {
		s.StV = x.StV
		s.ST = x.StV[0:]
	}
	if x.FtV != nil {
		s.FtV = x.FtV
		s.FT = x.FtV[0:]
	}
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
		fmt.Printf("\t%s %5s %8x %8x %4d %4d %8x\n", s.qualifier, str, s.LK, s.LX, s.compiledRulesCodeIndex, s.codeIndex, x)
	}
}
