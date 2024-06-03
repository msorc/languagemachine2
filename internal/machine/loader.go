package machine

import (
	"fmt"
	"regexp"
	"strings"
)

// --- symbols
func defineSymbols(e *Engine) {
	e.Nsy.UniqueE(NewZzz("_voidv"))
	e.Nsy.UniqueE(NewSym("__"))

	e.Nsy.UniqueE(theNull())
	e.Ssy.ZLM = e.Vsy.UniqueE(theNull())

	e.Nsy.UniqueE(NewSym("start"))
	e.Nsy.UniqueE(NewSym("eof"))
	e.Nsy.UniqueE(NewSpSym("sp"))
	e.Nsy.UniqueE(NewNlSym("nl"))
	e.Nsy.UniqueE(NewRepnSym("repeatN"))
	e.Nsy.UniqueE(NewAnything("anything"))
	e.Nsy.UniqueE(NewAnySym("nonTerminal"))
	e.Nsy.UniqueE(NewAnyChr("terminal"))
	e.Nsy.UniqueE(NewUriSym("uri"))
	e.Nsy.UniqueE(NewUrdSym("urd"))
	e.Nsy.UniqueE(NewOutSym("out"))
	e.Nsy.UniqueE(NewErrSym("err"))
	e.Nsy.UniqueE(NewLnoSym("lineNo"))
	e.Nsy.UniqueE(NewIfnSym("fileName"))
	e.Nsy.UniqueE(NewFlagSym("flagError"))
	e.Nsy.UniqueE(NewWarnSym("warnError"))

	e.Fsy.UniqueE(NewSym("mark"))

	e.Ssy.AppendFn = e.Fsy.UniqueE(NewAppendXSym("append"))
	e.Ssy.RepeatFn = e.Nsy.UniqueE(NewRepSym("repeat"))
	e.Ssy.OptionFn = e.Nsy.UniqueE(NewOptSym("option"))
	e.Ssy.RepeatFx = NewRepxSym("repeat")
	e.Ssy.OptionFx = NewOptxSym("option")

	e.Ssy.Nil = NewZzz("-")
	e.Ssy.GetFn = NewGetF("g")
	e.Ssy.StrFn = NewStrF("s")
	e.Ssy.ActFn = NewActF("a")
	e.Ssy.BindFn = NewBindF(":")
	e.Ssy.TakeFn = NewTakeF("%")
	e.Ssy.Start = e.Nsy.GetByString("start")
	e.Ssy.EOF = e.Nsy.GetByString("eof")
	e.Ssy.Put = e.Nsy.GetByString("out")
	e.Ssy.Mark = e.Fsy.GetByString("mark")

	e.Ssy.DropFn = e.Fsy.UniqueE(NewDropF("drop"))
	e.Ssy.DoneFn = e.Fsy.UniqueE(NewDoneF("done"))
	e.Ssy.InjFn = e.Fsy.UniqueE(NewInjF("inj"))

	e.Nsy.UniqueE(NewTrueSym("true"))
	e.Nsy.UniqueE(NewFalseSym("false"))
	e.Fsy.UniqueE(NewTrueF("true"))
	e.Fsy.UniqueE(NewFalseF("false"))
	e.Fsy.UniqueE(NewApplyF("apply"))

	e.Nsy.UniqueE(NewIOSymbol("toStr", NewToStrFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toLstr", NewToLstrFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toUstr", NewToUstrFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toQuote", NewToQuoteFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toSym", NewToSymFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toLsym", NewToLsymFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toUsym", NewToUsymFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toSys", NewToSysFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toLsys", NewToLsysFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toUsys", NewToUsysFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toVar", NewToVarFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toNum", NewToNumFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toOct", NewToOctFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toHex", NewToHexFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toBin", NewToBinFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toUrn", NewToUrNstrFromEngine(e)))
	e.Nsy.UniqueE(NewIOSymbol("toUrd", NewToUrDstrFromEngine(e)))

	e.Fsy.UniqueE(NewTestf("test"))
	e.Fsy.UniqueE(NewIff("if"))
	e.Fsy.UniqueE(NewLoopf("loop"))
	e.Fsy.UniqueE(NewForeachf("foreach"))
	e.Fsy.UniqueE(NewRetf("ret"))
	e.Fsy.UniqueE(NewLamdaf("lamda"))
	e.Fsy.UniqueE(NewSpecf("spec"))
	e.Fsy.UniqueE(NewArgsf("args"))
	e.Fsy.UniqueE(NewCellf("cell"))
	e.Fsy.UniqueE(NewArrayf("array"))
	e.Fsy.UniqueE(NewFunf("fun"))
	e.Fsy.UniqueE(NewIdxf("idx"))
	e.Fsy.UniqueE(NewIdtf("idt"))
	e.Fsy.UniqueE(NewSelf("sel"))
	e.Fsy.UniqueE(NewStoValf("stoVal"))
	e.Fsy.UniqueE(NewStoValf("="))
	e.Fsy.UniqueE(NewStoAddf("stoAdd"))
	e.Fsy.UniqueE(NewStoAddf("+="))
	e.Fsy.UniqueE(NewStoSubf("stoSub"))
	e.Fsy.UniqueE(NewStoSubf("-="))
	e.Fsy.UniqueE(NewStoMulf("stoMul"))
	e.Fsy.UniqueE(NewStoMulf("*="))
	e.Fsy.UniqueE(NewStoDivf("stoDiv"))
	e.Fsy.UniqueE(NewStoDivf("/="))
	e.Fsy.UniqueE(NewStoModf("stoMod"))
	e.Fsy.UniqueE(NewStoModf("%="))

	e.Fsy.UniqueE(NewEqf("eeq"))
	e.Fsy.UniqueE(NewEeqf("==="))
	e.Fsy.UniqueE(NewNef("nee"))
	e.Fsy.UniqueE(NewNeef("!=="))

	e.Fsy.UniqueE(NewInf("in"))
	e.Fsy.UniqueE(NewEqf("eq"))
	e.Fsy.UniqueE(NewEqf("=="))
	e.Fsy.UniqueE(NewNef("ne"))
	e.Fsy.UniqueE(NewNef("!="))
	e.Fsy.UniqueE(NewLtf("lt"))
	e.Fsy.UniqueE(NewLtf("<"))
	e.Fsy.UniqueE(NewGtf("gt"))
	e.Fsy.UniqueE(NewGtf(">"))
	e.Fsy.UniqueE(NewLef("le"))
	e.Fsy.UniqueE(NewLef("<="))
	e.Fsy.UniqueE(NewGef("ge"))
	e.Fsy.UniqueE(NewGef(">="))
	e.Fsy.UniqueE(NewOrOrf("orOr"))
	e.Fsy.UniqueE(NewOrOrf("||"))
	e.Fsy.UniqueE(NewAndAndf("andAnd"))
	e.Fsy.UniqueE(NewAndAndf("&&"))
	e.Fsy.UniqueE(NewBitOrf("bitOr"))
	e.Fsy.UniqueE(NewBitOrf("|"))
	e.Fsy.UniqueE(NewBitXorf("bitXor"))
	e.Fsy.UniqueE(NewBitXorf("^"))
	e.Fsy.UniqueE(NewBitAndf("bitAnd"))
	e.Fsy.UniqueE(NewBitAndf("&"))
	e.Fsy.UniqueE(NewAddf("add"))
	e.Fsy.UniqueE(NewAddf("+"))
	e.Fsy.UniqueE(NewSubf("sub"))
	e.Fsy.UniqueE(NewSubf("-"))
	e.Fsy.UniqueE(NewMulf("mul"))
	e.Fsy.UniqueE(NewMulf("*"))
	e.Fsy.UniqueE(NewDivf("div"))
	e.Fsy.UniqueE(NewDivf("/"))
	e.Fsy.UniqueE(NewModf("mod"))
	e.Fsy.UniqueE(NewModf("%"))
	e.Fsy.UniqueE(NewPreincf("preinc"))
	e.Fsy.UniqueE(NewPredecf("predec"))
	e.Fsy.UniqueE(NewPostincf("postinc"))
	e.Fsy.UniqueE(NewPostdecf("postdec"))
	e.Fsy.UniqueE(NewInvf("inv"))
	e.Fsy.UniqueE(NewNotf("not"))
	e.Fsy.UniqueE(NewNegf("neg"))
}

// --- predefined elements
type Predef struct {
	Start    GrammarElement
	EOF      GrammarElement
	Nil      GrammarElement
	ZLM      GrammarElement
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
}

func NewPredef() *Predef {
	return &Predef{}
}

type Loader struct {
	lm      *Engine
	tracer  *Tracer
	stk     *Opnd
	count   uint
	ruleTxt string
	ruleNum uint

	fsy *Dict // operator symbols
	tsy *Dict // terminal symbols
	nsy *Dict // system non-terminal symbols
	vsy *Dict // variables
	usy *Dict // user non-terminal symbols
	ssy *Predef
}

func NewLoader(e *Engine) *Loader {
	l := &Loader{
		lm:     e,
		tracer: e.Trace,
		fsy:    e.Fsy,
		tsy:    e.Tsy,
		nsy:    e.Nsy,
		vsy:    e.Vsy,
		usy:    e.Usy,
		ssy:    e.Ssy,
	}
	defineSymbols(l.lm)
	return l
}

func (l *Loader) SetTrace(t *Tracer) {
	l.tracer = t
}

func (l *Loader) Push(x GrammarElement) {
	l.stk = NewOpnd(l.stk, x)
	l.count++
}

func (l *Loader) Pop() GrammarElement {
	v := l.stk.V
	l.stk = l.stk.S
	l.count--
	return v
}

func (l *Loader) BMark() {
	l.stk = NewOpnd(l.stk, NewNumber(LMNumber(l.count)))
	l.count = 0
}

func (l *Loader) EMark() {
	l.count = uint(l.Pop().ToLong())
}

func (l *Loader) Take(n uint) []GrammarElement {
	v := make([]GrammarElement, n)
	for i := len(v); i > 0; i-- {
		v[i-1] = l.Pop()
	}
	return v
}

func (l *Loader) L(x uint) {
	l.Push(NewNumber(LMNumber(x * 2)))
}

func (l *Loader) R(x uint) {
	l.Push(NewNumber(LMNumber(x*2 + 1)))
}

func (l *Loader) B(x uint) {
	l.Push(NewNumber(LMNumber(x*2 | BRACKET)))
}

func (l *Loader) n(x uint) {
	l.Push(NewNumber(LMNumber(x)))
}

func (l *Loader) c(x string) {
	for _, ch := range x {
		//+
		l.Push(l.tsy.UniqueR(rune(ch)))
	}
}

func (l *Loader) d(x string) {
	l.Push(NewQuote(l.nsy.UniqueE(NewSym(x))))
}

func (l *Loader) m(x string) {
	l.Push(l.nsy.UniqueE(NewSym(x)))
}

func (l *Loader) F(x uint) {
	l.Push(l.lm.Lhr.FtV[x])
}

func (l *Loader) f(x string) {
	l.Push(l.fsy.UniqueE(NewSym(x)))
}

func (l *Loader) O() {
	l.BMark()
}

func (l *Loader) C() {
	v := l.Take(l.count)
	l.EMark()
	l.Push(NewStr(v))
}

func (l *Loader) r() {
	l.lm.DefineElements(l.Take(5), l.ruleTxt, l.ruleNum)
	l.ruleNum++
}

func (l *Loader) A() {
	l.Push(NewAllRef(l.Pop()))
}

func (l *Loader) e() {
	l.Push(NewEachRef(l.Pop()))
}

func (l *Loader) p() {
	v := l.Pop()
	l.Push(NewGetXF(v))
	l.Push(l.ssy.BindFn)
}

func (l *Loader) P() {
	v := l.Pop()
	l.Push(NewGetXF(v))
	l.Push(l.ssy.BindFn)
}

func (l *Loader) t() {
	l.Push(l.ssy.TakeFn)
}

func (l *Loader) b() {
	l.Push(l.ssy.BindFn)
}

func (l *Loader) g() {
	l.Push(l.ssy.GetFn)
}

func (l *Loader) X() {
	l.Push(l.ssy.DropFn)
}

func (l *Loader) G() {
	l.Push(NewGetXF(l.Pop()))
}

func (l *Loader) V() {
	l.Push(NewGetVF(l.Pop()))
}

func (l *Loader) s() {
	l.Push(l.ssy.StrFn)
}

func (l *Loader) a() {
	l.Push(l.ssy.ActFn)
}

func (l *Loader) z() {
	l.Push(l.ssy.Nil)
}

func (l *Loader) w() {
	l.Push(NewNewVar())
}

func (l *Loader) l(x string) {
	l.Push(l.nsy.UniqueE(NewLexFromEngine(x, l.lm)))
}

func (l *Loader) v(x string) {
	l.Push(l.vsy.UniqueE(NewVarSym(x)))
}

func (l *Loader) Unescape(s string) string {
	var r strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			if i < len(s) {
				switch s[i] {
				case 'a':
					r.WriteRune('\a')
				case 'b':
					r.WriteRune('\b')
				case '"':
					r.WriteRune('"')
				case '\'':
					r.WriteRune('\'')
				case '\\':
					r.WriteRune('\\')
				case 'n':
					r.WriteRune('\n')
				case 'r':
					r.WriteRune('\r')
				case 't':
					r.WriteRune('\t')
				case 'f':
					r.WriteRune('\f')
				case 'v':
					r.WriteRune('\v')
				default:
					r.WriteByte(s[i])
				}
			} else {
				fmt.Printf("%s\n", s)
				panic("bad unescape")
			}
		} else {
			r.WriteByte(s[i])
		}
	}
	return r.String()
}

func (l *Loader) MStr(s string) string {
	return l.Unescape(s)
}

func (l *Loader) Load(tt string) {
	r1 := regexp.MustCompile("([\\(\\)\\.reAtpbPgGVsawz])|(.:\\S*)|#[^\\n]*\\n|\\s*")
	sa := r1.FindAllString(tt, -1)
	for i, st := range sa {
		if len(strings.TrimSpace(st)) == 0 {
			continue
		}
		if l.tracer != nil && (l.tracer.Tracing(LOAD) == LOAD) {
			fmt.Printf("load: %s\n", st)
		}
		switch st[0] {
		case 'F':
			l.F(strtoui(st[2:]))
		case 'L':
			l.L(strtoui(st[2:]))
		case 'R':
			l.R(strtoui(st[2:]))
		case 'B':
			l.B(strtoui(st[2:]))
		case 'n':
			l.n(strtoui(st[2:]))
		case 'c':
			if len(st) == 2 {
				l.c("")
			} else {
				l.c(l.MStr(st[2:]))
			}
		case 'd':
			if len(st) == 2 {
				l.d("")
			} else {
				l.d(l.MStr(st[2:]))
			}
		case 'm':
			l.m(l.MStr(st[2:]))
		case 'f':
			l.f(l.MStr(st[2:]))
		case 'l':
			l.l(l.MStr(st[2:]))
		case 'v':
			l.v(l.MStr(st[2:]))
		case '(':
			l.O()
		case ')':
			l.C()
		case '.':
			l.X()
		case 'r':
			l.r()
		case 'A':
			l.A()
		case 'e':
			l.r()
		case 't':
			l.t()
		case 'p':
			l.p()
		case 'P':
			l.P()
		case 'b':
			l.b()
		// case 'k':
		// 	l.k()
		// case 'K':
		// 	l.K()
		case 'g':
			l.g()
		case 'G':
			l.G()
		case 'V':
			l.V()
		case 's':
			l.s()
		case 'a':
			l.a()
		case 'w':
			l.w()
		case 'z':
			l.z()
		default:
			panic(fmt.Sprintf("bad load format: %d `%s`", i, st))
		}
	}
}
