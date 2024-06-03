package machine

import (
	"fmt"
	"strings"
)

type Stream struct {
	SM GenMode        // stream mode
	CI uint           // code index
	SY GrammarElement // current symbol
	SV GrammarElement // current value
	RV GrammarElement // return value from machine
	XS *Opnd          // operand stack
	AV *Var           // list of all variables
	LM *Engine        // the engine

	TT []GrammarElement
	MT []GrammarElement
	DT []GrammarElement
	VT []GrammarElement
	XT []GrammarElement
	NT []GrammarElement
	ST []GrammarElement
	FT []GrammarElement

	Start    GrammarElement
	EOF      GrammarElement
	Nil      GrammarElement
	Zlm      GrammarElement
	Put      GrammarElement
	Mark     GrammarElement
	DropFn   GrammarElement
	GetFn    GrammarElement
	StrFn    GrammarElement
	ActFn    GrammarElement
	BindFn   GrammarElement
	TakeFn   GrammarElement
	DoneFn   GrammarElement
	InjFn    GrammarElement
	AppendFn GrammarElement
	RepeatFn GrammarElement
	OptionFn GrammarElement
	RepeatFx GrammarElement
	OptionFx GrammarElement

	LK any // jump address to current point in string
	LX any // jump address to exit from string

	Ntv []GrammarElement
	MtV []GrammarElement
	DtV []GrammarElement
	VtV []GrammarElement
	XtV []GrammarElement
	TtV []GrammarElement
	StV []GrammarElement
	FtV []GrammarElement

	QU string           // for tracing
	CV []GrammarElement // code vector

	CZ uint // code index from compiled rules
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

func (s *Stream) Pushx(x GrammarElement) GrammarElement {
	s.XS = NewOpnd(s.XS, x)
	return x
}

func (s *Stream) Popx() GrammarElement {
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

func (s *Stream) CountxWithElement(k GrammarElement) uint {
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

func (s *Stream) ToRow() []GrammarElement {
	v := make([]GrammarElement, s.Countx())
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	return v
}

func (s *Stream) ToArgv(k GrammarElement) []GrammarElement {
	v := make([]GrammarElement, s.CountxWithElement(k)+1)
	for i := len(v); i > 0; i-- {
		v[i-1] = s.Popx()
	}
	v[0] = s.Popx()
	return v
}

func (s *Stream) Tsy() *Dict {
	return s.LM.Tsy
}

func (s *Stream) Fsy() *Dict {
	return s.LM.Fsy
}

func (s *Stream) Nsy() *Dict {
	return s.LM.Nsy
}

func (s *Stream) Usy() *Dict {
	return s.LM.Usy
}

func (s *Stream) Ssy() *Predef { return s.LM.Ssy }

func (s *Stream) TheRef(sMode GenMode, k GrammarElement, x LMScope) GenMode {
	return s.LM.TheRef(sMode, k, x)
}

func (s *Stream) EachRef(sMode GenMode, k GrammarElement, x LMScope) GenMode {
	return s.LM.EachRef(sMode, k, x)
}

func (s *Stream) AllRef(sMode GenMode, k GrammarElement, x LMScope) GenMode {
	return s.LM.AllRef(sMode, k, x)
}

func (s *Stream) BindCvar(l, r GrammarElement) bool {
	return s.LM.BindCvar(l, r)
}

func (s *Stream) System() *LMExternal {
	return s.LM.System
}

func (s *Stream) Initialise(sStream *Stream) {}

func (s *Stream) MakeNt(x int) GrammarElement {
	return NewNumber(LMNumber(x))
}

func (s *Stream) MakeMt(x string) GrammarElement {
	if x == "null" {
		return s.LM.Ssy.Nil
	}
	return s.LM.Nsy.UniqueE(NewSym(x))
}

func (s *Stream) MakeDt(x string) GrammarElement {
	return NewQuote(s.LM.Nsy.UniqueE(NewSym(x)))
}

func (s *Stream) MakeTt(x string) GrammarElement {
	return s.LM.Tsy.UniqueR(rune(Unescape(UrlUnescape(x))[0]))
}

func (s *Stream) MakeVt(x string) GrammarElement {
	return s.LM.Vsy.UniqueE(NewVarSym(x))
}

func (s *Stream) Makext(x string) GrammarElement {
	return s.LM.Nsy.UniqueE(NewLexFromEngine(x, s.LM))
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
	if (s.LM.Trace != nil) && (s.LM.Trace.Flags&DEBUG == DEBUG) {
		fmt.Printf("\t%s %5s %8x %8x %4d %4d %8x\n", s.QU, str, s.LK, s.LX, s.CZ, s.CI, x)
	}
}

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
