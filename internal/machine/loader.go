package machine

import (
	"languagemachine2/internal/utils"
	"fmt"
	"regexp"
	"strings"
)

// --- symbols
func defineSymbols(e *Engine) {
	// e.nonTerminalSymbols.UniqueE(NewZzz("_voidv"))
	// e.nonTerminalSymbols.UniqueE(NewSym("__"))

	e.nonTerminalSymbols.UniqueE(theNull())
	e.predefinedSymbols.zlm = e.varSymbols.UniqueE(theNull())

	e.nonTerminalSymbols.UniqueE(NewSym("start"))
	e.nonTerminalSymbols.UniqueE(NewSym("eof"))
	e.nonTerminalSymbols.UniqueE(NewSpSym("sp"))
	e.nonTerminalSymbols.UniqueE(NewNlSym("nl"))
	e.nonTerminalSymbols.UniqueE(NewRepnSym("repeatN"))
	e.nonTerminalSymbols.UniqueE(NewAnything("anything"))
	e.nonTerminalSymbols.UniqueE(NewAnySym("nonTerminal"))
	e.nonTerminalSymbols.UniqueE(NewAnyChr("terminal"))
	e.nonTerminalSymbols.UniqueE(NewUriSym("uri"))
	e.nonTerminalSymbols.UniqueE(NewUrdSym("urd"))
	e.nonTerminalSymbols.UniqueE(NewOutSym("out"))
	e.nonTerminalSymbols.UniqueE(NewErrSym("err"))
	e.nonTerminalSymbols.UniqueE(NewLnoSym("lineNo"))
	e.nonTerminalSymbols.UniqueE(NewIfnSym("fileName"))
	e.nonTerminalSymbols.UniqueE(NewFlagSym("flagError"))
	e.nonTerminalSymbols.UniqueE(NewWarnSym("warnError"))

	e.functionSymbols.UniqueE(NewSym("mark"))

	e.predefinedSymbols.appendFn = e.functionSymbols.UniqueE(NewAppendXSym("append"))
	e.predefinedSymbols.repeatFn = e.nonTerminalSymbols.UniqueE(NewRepSym("repeat"))
	e.predefinedSymbols.optionFn = e.nonTerminalSymbols.UniqueE(NewOptSym("option"))
	e.predefinedSymbols.repeatFx = NewRepxSym("repeat")
	e.predefinedSymbols.optionFx = NewOptxSym("option")

	e.predefinedSymbols.nil = NewZzz("-")
	e.predefinedSymbols.getFn = NewGetF("g")
	e.predefinedSymbols.strFn = NewStrF("s")
	e.predefinedSymbols.actFn = NewActF("a")
	e.predefinedSymbols.bindFn = NewBindF(":")
	e.predefinedSymbols.takeFn = NewTakeF("%")
	e.predefinedSymbols.start = e.nonTerminalSymbols.GetByString("start")
	e.predefinedSymbols.eof = e.nonTerminalSymbols.GetByString("eof")
	e.predefinedSymbols.put = e.nonTerminalSymbols.GetByString("out")
	e.predefinedSymbols.mark = e.functionSymbols.GetByString("mark")

	e.predefinedSymbols.dropFn = e.functionSymbols.UniqueE(NewDropF("drop"))
	e.predefinedSymbols.doneFn = e.functionSymbols.UniqueE(NewDoneF("done"))
	e.predefinedSymbols.injFn = e.functionSymbols.UniqueE(NewInjF("inj"))

	e.nonTerminalSymbols.UniqueE(NewTrueSym("true"))
	e.nonTerminalSymbols.UniqueE(NewFalseSym("false"))
	e.functionSymbols.UniqueE(NewTrueF("true"))
	e.functionSymbols.UniqueE(NewFalseF("false"))
	e.functionSymbols.UniqueE(NewApplyF("apply"))

	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toStr", NewToStrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLstr", NewToLstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUstr", NewToUstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toQuote", NewToQuoteFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toSym", NewToSymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLsym", NewToLsymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUsym", NewToUsymFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toSys", NewToSysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toLsys", NewToLsysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUsys", NewToUsysFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toVar", NewToVarFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toNum", NewToNumFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toOct", NewToOctFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toHex", NewToHexFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toBin", NewToBinFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUrn", NewToUrNstrFromEngine(e)))
	e.nonTerminalSymbols.UniqueE(NewIOSymbol("toUrd", NewToUrDstrFromEngine(e)))

	e.functionSymbols.UniqueE(NewTestf("test"))
	e.functionSymbols.UniqueE(NewIff("if"))
	e.functionSymbols.UniqueE(NewLoopf("loop"))
	e.functionSymbols.UniqueE(NewForeachf("foreach"))
	e.functionSymbols.UniqueE(NewRetf("ret"))
	e.functionSymbols.UniqueE(NewLamdaf("lamda"))
	e.functionSymbols.UniqueE(NewSpecf("spec"))
	e.functionSymbols.UniqueE(NewArgsf("args"))
	e.functionSymbols.UniqueE(NewCellf("cell"))
	e.functionSymbols.UniqueE(NewArrayf("array"))
	e.functionSymbols.UniqueE(NewFunf("fun"))
	e.functionSymbols.UniqueE(NewIdxf("idx"))
	e.functionSymbols.UniqueE(NewIdtf("idt"))
	e.functionSymbols.UniqueE(NewSelF("sel"))
	e.functionSymbols.UniqueE(NewStoValf("stoVal"))
	e.functionSymbols.UniqueE(NewStoValf("="))
	e.functionSymbols.UniqueE(NewStoAddf("stoAdd"))
	e.functionSymbols.UniqueE(NewStoAddf("+="))
	e.functionSymbols.UniqueE(NewStoSubf("stoSub"))
	e.functionSymbols.UniqueE(NewStoSubf("-="))
	e.functionSymbols.UniqueE(NewStoMulf("stoMul"))
	e.functionSymbols.UniqueE(NewStoMulf("*="))
	e.functionSymbols.UniqueE(NewStoDivf("stoDiv"))
	e.functionSymbols.UniqueE(NewStoDivf("/="))
	e.functionSymbols.UniqueE(NewStoModf("stoMod"))
	e.functionSymbols.UniqueE(NewStoModf("%="))

	e.functionSymbols.UniqueE(NewEqf("eeq"))
	e.functionSymbols.UniqueE(NewEeqf("==="))
	e.functionSymbols.UniqueE(NewNef("nee"))
	e.functionSymbols.UniqueE(NewNeef("!=="))

	e.functionSymbols.UniqueE(NewInf("in"))
	e.functionSymbols.UniqueE(NewEqf("eq"))
	e.functionSymbols.UniqueE(NewEqf("=="))
	e.functionSymbols.UniqueE(NewNef("ne"))
	e.functionSymbols.UniqueE(NewNef("!="))
	e.functionSymbols.UniqueE(NewLtf("lt"))
	e.functionSymbols.UniqueE(NewLtf("<"))
	e.functionSymbols.UniqueE(NewGtf("gt"))
	e.functionSymbols.UniqueE(NewGtf(">"))
	e.functionSymbols.UniqueE(NewLef("le"))
	e.functionSymbols.UniqueE(NewLef("<="))
	e.functionSymbols.UniqueE(NewGef("ge"))
	e.functionSymbols.UniqueE(NewGef(">="))
	e.functionSymbols.UniqueE(NewOrOrf("orOr"))
	e.functionSymbols.UniqueE(NewOrOrf("||"))
	e.functionSymbols.UniqueE(NewAndAndf("andAnd"))
	e.functionSymbols.UniqueE(NewAndAndf("&&"))
	e.functionSymbols.UniqueE(NewBitOrf("bitOr"))
	e.functionSymbols.UniqueE(NewBitOrf("|"))
	e.functionSymbols.UniqueE(NewBitXorf("bitXor"))
	e.functionSymbols.UniqueE(NewBitXorf("^"))
	e.functionSymbols.UniqueE(NewBitAndf("bitAnd"))
	e.functionSymbols.UniqueE(NewBitAndf("&"))
	e.functionSymbols.UniqueE(NewAddf("add"))
	e.functionSymbols.UniqueE(NewAddf("+"))
	e.functionSymbols.UniqueE(NewSubf("sub"))
	e.functionSymbols.UniqueE(NewSubf("-"))
	e.functionSymbols.UniqueE(NewMulf("mul"))
	e.functionSymbols.UniqueE(NewMulf("*"))
	e.functionSymbols.UniqueE(NewDivf("div"))
	e.functionSymbols.UniqueE(NewDivf("/"))
	e.functionSymbols.UniqueE(NewModf("mod"))
	e.functionSymbols.UniqueE(NewModf("%"))
	e.functionSymbols.UniqueE(NewPreincf("preinc"))
	e.functionSymbols.UniqueE(NewPredecf("predec"))
	e.functionSymbols.UniqueE(NewPostincf("postinc"))
	e.functionSymbols.UniqueE(NewPostdecf("postdec"))
	e.functionSymbols.UniqueE(NewInvf("inv"))
	e.functionSymbols.UniqueE(NewNotf("not"))
	e.functionSymbols.UniqueE(NewNegf("neg"))
}

// --- predefined elements
type Predef struct {
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
}

func NewPredef() *Predef {
	return &Predef{}
}

type Loader struct {
	engine        *Engine
	tracer        *Tracer
	operandsStack *Opnd
	count         uint
	ruleText      string
	ruleNumber    uint

	functionSymbols    *Dict // operator symbols
	terminalSymbols    *Dict // terminal symbols
	nonTerminalSymbols *Dict // system non-terminal symbols
	varSymbols         *Dict // variables
	userSymbols        *Dict // user non-terminal symbols
	predefinedSymbols  *Predef
}

func NewLoader(e *Engine) *Loader {
	l := &Loader{
		engine:             e,
		tracer:             e.tracer,
		functionSymbols:    e.functionSymbols,
		terminalSymbols:    e.terminalSymbols,
		nonTerminalSymbols: e.nonTerminalSymbols,
		varSymbols:         e.varSymbols,
		userSymbols:        e.userSymbols,
		predefinedSymbols:  e.predefinedSymbols,
	}
	defineSymbols(l.engine)
	return l
}

func (l *Loader) SetTrace(t *Tracer) {
	l.tracer = t
}

func (l *Loader) Push(x Element) {
	l.operandsStack = NewOpnd(l.operandsStack, x)
	l.count++
}

func (l *Loader) Pop() Element {
	v := l.operandsStack.V
	l.operandsStack = l.operandsStack.S
	l.count--
	return v
}

func (l *Loader) BMark() {
	l.operandsStack = NewOpnd(l.operandsStack, NewNumber(LMNumber(l.count)))
	l.count = 0
}

func (l *Loader) EMark() {
	l.count = uint(l.Pop().ToLong())
}

func (l *Loader) Take(n uint) []Element {
	v := make([]Element, n)
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
		l.Push(l.terminalSymbols.UniqueR(rune(ch)))
	}
}

func (l *Loader) d(x string) {
	l.Push(NewQuote(l.nonTerminalSymbols.UniqueE(NewSym(x))))
}

func (l *Loader) m(x string) {
	l.Push(l.nonTerminalSymbols.UniqueE(NewSym(x)))
}

func (l *Loader) F(x uint) {
	l.Push(l.engine.lhsStream.FtV[x])
}

func (l *Loader) f(x string) {
	l.Push(l.functionSymbols.UniqueE(NewSym(x)))
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
	l.engine.DefineElements(l.Take(5), l.ruleText, l.ruleNumber)
	l.ruleNumber++
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
	l.Push(l.predefinedSymbols.bindFn)
}

func (l *Loader) P() {
	v := l.Pop()
	l.Push(NewGetXF(v))
	l.Push(l.predefinedSymbols.bindFn)
}

func (l *Loader) t() {
	l.Push(l.predefinedSymbols.takeFn)
}

func (l *Loader) b() {
	l.Push(l.predefinedSymbols.bindFn)
}

func (l *Loader) g() {
	l.Push(l.predefinedSymbols.getFn)
}

func (l *Loader) X() {
	l.Push(l.predefinedSymbols.dropFn)
}

func (l *Loader) G() {
	l.Push(NewGetXF(l.Pop()))
}

func (l *Loader) V() {
	l.Push(NewGetVF(l.Pop()))
}

func (l *Loader) s() {
	l.Push(l.predefinedSymbols.strFn)
}

func (l *Loader) a() {
	l.Push(l.predefinedSymbols.actFn)
}

func (l *Loader) z() {
	l.Push(l.predefinedSymbols.nil)
}

func (l *Loader) w() {
	l.Push(NewNewVar())
}

func (l *Loader) l(x string) {
	l.Push(l.nonTerminalSymbols.UniqueE(NewLexFromEngine(x, l.engine)))
}

func (l *Loader) v(x string) {
	l.Push(l.varSymbols.UniqueE(NewVarSym(x)))
}

func (l *Loader) MStr(s string) string {
	return utils.Unescape(utils.Decode(s))
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
