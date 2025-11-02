package machine

import (
	"fmt"
	"languagemachine2/internal/utils"
	"regexp"
	"strings"

	"github.com/liyue201/gostl/ds/list/bidlist"
)

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
	engine     *Engine
	tracer     *Tracer
	operands   bidlist.List[Element]
	count      uint
	ruleText   string
	ruleNumber uint

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

	return l
}

func (l *Loader) SetTrace(t *Tracer) {
	l.tracer = t
}

func (l *Loader) Push(x Element) {
	l.operands.PushFront(x)
	l.count++
}

func (l *Loader) Pop() Element {
	l.count--
	return l.operands.PopFront()
}

func (l *Loader) BMark() {
	l.operands.PushFront(NewNumber(LMNumber(l.count)))
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
	r1 := regexp.MustCompile(`([().reAtpbPgGVsawz])|(.:\S*)|#[^\n]*\n|\s*`)
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
