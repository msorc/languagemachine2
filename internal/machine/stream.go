package machine

import (
	"fmt"
	"strings"
)

func Unescape(s string) string {
	var r strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' {
			i++
			if i < len(s) {
				switch s[i] {
				case 'a':
					r.WriteByte('\a')
				case 'b':
					r.WriteByte('\b')
				case '"':
					r.WriteByte('"')
				case '\'':
					r.WriteByte('\'')
				case '\\':
					r.WriteByte('\\')
				case 'n':
					r.WriteByte('\n')
				case 'r':
					r.WriteByte('\r')
				case 't':
					r.WriteByte('\t')
				case 'f':
					r.WriteByte('\f')
				case 'v':
					r.WriteByte('\v')
				default:
					r.WriteByte(s[i])
				}
			} else {
				fmt.Println(s)
				panic("bad unescape")
			}
		} else {
			r.WriteByte(s[i])
		}
		i++
	}
	return r.String()
}

type Stream struct {
	SM GenMode        // stream mode
	CI uint           // code index
	SY MachineElement // current symbol
	SV MachineElement // current value
	RV MachineElement // return value from machine
	XS *Opnd          // operand stack
	AV *Var           // list of all variables
	LM *Engine        // the engine

	TT []MachineElement
	MT []MachineElement
	DT []MachineElement
	VT []MachineElement
	XT []MachineElement
	NT []MachineElement
	ST []MachineElement
	FT []MachineElement

	Start    MachineElement
	EOF      MachineElement
	Nil      MachineElement
	Zlm      MachineElement
	Put      MachineElement
	Mark     MachineElement
	DropFn   MachineElement
	GetFn    MachineElement
	StrFn    MachineElement
	ActFn    MachineElement
	BindFn   MachineElement
	TakeFn   MachineElement
	DoneFn   MachineElement
	InjFn    MachineElement
	AppendFn MachineElement
	RepeatFn MachineElement
	OptionFn MachineElement
	RepeatFx MachineElement
	OptionFx MachineElement

	LK any // jump address to current point in string
	LX any // jump address to exit from string

	Ntv []MachineElement
	MtV []MachineElement
	DtV []MachineElement
	VtV []MachineElement
	XtV []MachineElement
	TtV []MachineElement
	StV []MachineElement
	FtV []MachineElement

	QU string           // for tracing
	CV []MachineElement // code vector

	CZ uint // code index from compiled rules
}

func NewStream() *Stream {
	return &Stream{}
}

func NewStreamFromString(s string) *Stream {
	return &Stream{QU: s}
}

func NewStreamFromEngine(e *Engine, s string, i uint) *Stream {
	return &Stream{LM: e, QU: s, CI: i}
}

func (s *Stream) Act(st *Stream, m GenMode) GenMode {
	if s.CI < uint(len(s.CV)) {
		i := s.CI
		s.CI++
		return s.CV[i].Act(st, m)
	} else {
		return m.Ret()
	}
}

func (s *Stream) Rep(st *Stream, m GenMode) GenMode {
	if s.CI < uint(len(s.CV)) {
		i := s.CI
		s.CI++
		return s.CV[i].Act(st, m)
	} else {
		s.CI = 0
		return m.Ret()
	}
}

func (s *Stream) Getx(m GenMode) GenMode {
	s.Pushx(s.CV[s.CI])
	s.CI++
	return m
}

func (s *Stream) Pushx(x MachineElement) MachineElement {
	s.XS = NewOpnd(s.XS, x)
	return x
}

func (s *Stream) Popx() MachineElement {
	x := s.XS
	s.XS = x.S
	return x.V
}

func (s *Stream) PushNum(x int) int {
	s.XS = NewOpnd(s.XS, NewNumber(LMNumber(x)))
	return x
}

func (s *Stream) PushDbl(x float64) float64 {
	s.XS = NewOpnd(s.XS, NewNumber(LMNumber(x)))
	return x
}

func (s *Stream) PushBool(x bool) bool {
	s.XS = NewOpnd(s.XS, NewBoolean(x))
	return x
}

func (s *Stream) Countx() uint {
	var n uint
	for x := s.XS; x != nil; x = x.S {
		n++
	}
	return n
}

func (s *Stream) CountxWithElement(k MachineElement) uint {
	var n uint
	for x := s.XS; x != nil && x.V != k; x = x.S {
		n++
	}
	return n
}

func (s *Stream) Dumpx() {
	for x := s.XS; x != nil; x = x.S {
		fmt.Printf("\tx: %s\n", x.ToString())
	}
	fmt.Println("------")
}

func (s *Stream) DumpxWithString(str string) {
	fmt.Printf("\tstack: %s\n", str)
	s.Dumpx()
}

func (s *Stream) ToRow() []MachineElement {
	v := make([]MachineElement, s.Countx())
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	return v
}

func (s *Stream) ToArgv(k MachineElement) []MachineElement {
	v := make([]MachineElement, s.CountxWithElement(k)+1)
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	v[0] = s.Popx()
	return v
}

func (s *Stream) TerminalSymbols() *Dict {
	return s.LM.terminalSymbols
}

func (s *Stream) FunctionSymbols() *Dict {
	return s.LM.functionSymbols
}

func (s *Stream) NonTerminalSymbols() *Dict {
	return s.LM.nonTerminalSymbols
}

func (s *Stream) UserSymbols() *Dict {
	return s.LM.userSymbols
}

func (s *Stream) PredefinedSymbols() *Predef { return s.LM.predefinedSymbols }

func (s *Stream) TheRef(sMode GenMode, k MachineElement, x LMScope) GenMode {
	return s.LM.TheRef(sMode, k, x)
}

func (s *Stream) EachRef(sMode GenMode, k MachineElement, x LMScope) GenMode {
	return s.LM.EachRef(sMode, k, x)
}

func (s *Stream) AllRef(sMode GenMode, k MachineElement, x LMScope) GenMode {
	return s.LM.AllRef(sMode, k, x)
}

func (s *Stream) BindCvar(l, r MachineElement) bool {
	return s.LM.BindCvar(l, r)
}

func (s *Stream) ExternalSystem() *LMExternal {
	return s.LM.externalSystem
}

func (s *Stream) Initialise(sStream *Stream) {}

func (s *Stream) MakeNt(x int) MachineElement {
	return NewNumber(LMNumber(x))
}

func (s *Stream) MakeMt(x string) MachineElement {
	if x == "null" {
		return s.LM.predefinedSymbols.Nil
	}
	return s.LM.nonTerminalSymbols.UniqueE(NewSym(x))
}

func (s *Stream) MakeDt(x string) MachineElement {
	return NewQuote(s.LM.nonTerminalSymbols.UniqueE(NewSym(x)))
}

func (s *Stream) MakeTt(x string) MachineElement {
	return s.LM.terminalSymbols.UniqueR(rune(Unescape(UrlUnescape(x))[0]))
}

func (s *Stream) MakeVt(x string) MachineElement {
	return s.LM.varSymbols.UniqueE(NewVarSym(x))
}

func (s *Stream) Makext(x string) MachineElement {
	return s.LM.nonTerminalSymbols.UniqueE(NewLexFromEngine(x, s.LM))
}

// + attention
func (s *Stream) CopyTables(x *Stream) {
	if x.Ntv != nil {
		s.Ntv = x.Ntv
		s.NT = x.Ntv[0:]
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
	s.Start = x.Start
	s.EOF = x.EOF
	s.Nil = x.Nil
	s.Zlm = x.ZLM
	s.Put = x.Put
	s.Mark = x.Mark
	s.DropFn = x.DropFn
	s.GetFn = x.GetFn
	s.StrFn = x.StrFn
	s.ActFn = x.ActFn
	s.BindFn = x.BindFn
	s.TakeFn = x.TakeFn
	s.DoneFn = x.DoneFn
	s.InjFn = x.InjFn
	s.AppendFn = x.AppendFn
	s.RepeatFn = x.RepeatFn
	s.OptionFn = x.OptionFn
	s.RepeatFx = x.RepeatFx
	s.OptionFx = x.OptionFx
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
	if (s.LM.tracer != nil) && (s.LM.tracer.Flags&DEBUG == DEBUG) {
		fmt.Printf("\t%s %5s %8x %8x %4d %4d %8x\n", s.QU, str, s.LK, s.LX, s.CZ, s.CI, x)
	}
}
