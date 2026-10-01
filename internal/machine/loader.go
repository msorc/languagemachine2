package machine

import (
	"github.com/msorc/languagemachine2/internal/utils"
	"regexp"
	"strconv"
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
}

func NewPredef() *Predef {
	return &Predef{}
}

// Loader builds rules from bytecode (see docs/bytecode.md) into its engine's
// grammars, using the engine's symbol dictionaries.
type Loader struct {
	engine     *Engine
	operands   bidlist.List[Element]
	count      int
	ruleText   string
	ruleNumber int
}

func NewLoader(e *Engine) *Loader {
	return &Loader{engine: e, ruleText: "rule"}
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
	l.count = l.Pop().ToInt()
}

func (l *Loader) Take(n int) []Element {
	v := make([]Element, n)
	for i := len(v); i > 0; i-- {
		v[i-1] = l.Pop()
	}
	return v
}

func (l *Loader) L(x int) {
	l.Push(NewNumber(LMNumber(x * 2)))
}

func (l *Loader) R(x int) {
	l.Push(NewNumber(LMNumber(x*2 + 1)))
}

func (l *Loader) B(x int) {
	l.Push(NewNumber(LMNumber(x*2 | BRACKET)))
}

// M encodes maximal priority: the rule can always start (BRACKET bit), and
// its context priority is PRIMASK, at which ResolveE refuses further nesting.
// The level x is not significant.
func (l *Loader) M(x int) {
	l.Push(NewNumber(LMNumber(PRIMASK | BRACKET)))
}

func (l *Loader) n(x float64) {
	l.Push(NewNumber(LMNumber(x)))
}

func (l *Loader) c(x string) {
	for _, ch := range x {
		//+
		l.Push(l.engine.terminalSymbols.UniqueR(rune(ch)))
	}
}

func (l *Loader) d(x string) {
	l.Push(NewQuote(l.engine.nonTerminalSymbols.UniqueE(NewSym(x))))
}

func (l *Loader) m(x string) {
	l.Push(l.engine.nonTerminalSymbols.UniqueE(NewSym(x)))
}

func (l *Loader) f(x string) {
	l.Push(l.engine.functionSymbols.UniqueE(NewSym(x)))
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
	l.engine.AddRule(l.Take(5), l.ruleText, l.ruleNumber)
	l.ruleNumber++
}

func (l *Loader) A() {
	l.Push(NewAllRef(l.Pop()))
}

func (l *Loader) e() {
	l.Push(NewEachRef(l.Pop()))
}

// p (and P, which is the same) binds: v:X p is get X, then bind.
func (l *Loader) p() {
	v := l.Pop()
	l.Push(NewGetXF(v))
	l.Push(l.engine.predefinedSymbols.bindFn)
}

func (l *Loader) t() {
	l.Push(l.engine.predefinedSymbols.takeFn)
}

func (l *Loader) b() {
	l.Push(l.engine.predefinedSymbols.bindFn)
}

func (l *Loader) g() {
	l.Push(l.engine.predefinedSymbols.getFn)
}

func (l *Loader) X() {
	l.Push(l.engine.predefinedSymbols.dropFn)
}

func (l *Loader) G() {
	l.Push(NewGetXF(l.Pop()))
}

func (l *Loader) V() {
	l.Push(NewGetVF(l.Pop()))
}

func (l *Loader) s() {
	l.Push(l.engine.predefinedSymbols.strFn)
}

func (l *Loader) a() {
	l.Push(l.engine.predefinedSymbols.actFn)
}

func (l *Loader) z() {
	l.Push(l.engine.predefinedSymbols.nil)
}

func (l *Loader) w() {
	l.Push(NewNewVar())
}

func (l *Loader) l(x string) {
	l.Push(l.engine.nonTerminalSymbols.UniqueE(NewLexFromEngine(x, l.engine)))
}

func (l *Loader) v(x string) {
	l.Push(l.engine.varSymbols.UniqueE(NewVarSym(x)))
}

// MStr decodes the text of an X:value token: URL decoding, then C escapes.
func (l *Loader) MStr(s string) string {
	d, err := utils.Decode(s)
	if err == nil {
		d, err = utils.Unescape(d)
	}
	if err != nil {
		fail("bad rule text `%s`: %v", s, err)
	}
	return d
}

// level decodes the level of a priority token such as L:20.
func (l *Loader) level(i int, st string) int {
	n, err := utils.Strtoi(st[2:])
	if err != nil {
		fail("bad priority: %d `%s`", i, st)
	}
	return n
}

// tokenRE splits bytecode into tokens. The final (\S) catches
// single-character opcodes the loader does not know, so they reach the
// bad-format error instead of being skipped.
var tokenRE = regexp.MustCompile(`([().reAtpbPgGVsawz])|(.:\S*)|#[^\n]*\n|(\S)|\s*`)

// Load defines the rules in tt, bytecode as described in docs/bytecode.md.
func (l *Loader) Load(tt string) (err error) {
	defer catch(&err)
	sa := tokenRE.FindAllString(tt, -1)
	for i, st := range sa {
		// comment lines (incl. a #! shebang) are separators in the original's
		// RegExp.split, so they never reach the dispatch
		if len(strings.TrimSpace(st)) == 0 || st[0] == '#' {
			continue
		}
		if t := l.engine.tracer; t != nil && t.Tracing(LOAD) != 0 {
			l.engine.printf("load: %s\n", st)
		}
		switch st {
		case "E":
			l.Push(NewEachX("each"))
			continue
		case "B":
			l.Push(NewAllX("all"))
			continue
		case "T":
			// lmn2mbe has a rule for top, but lmn2xfe never produces it
			fail("unsupported opcode: %d `%s` (top is not implemented)", i, st)
		}
		switch st[0] {
		case 'M':
			l.M(l.level(i, st))
		case 'L':
			l.L(l.level(i, st))
		case 'R':
			l.R(l.level(i, st))
		case 'B':
			l.B(l.level(i, st))
		case 'n':
			// numeric literals may be real (n:2.5), not just the rule offset
			x, err := strconv.ParseFloat(st[2:], 64)
			if err != nil {
				fail("bad number: %d `%s`", i, st)
			}
			l.n(x)
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
			l.e()
		case 't':
			l.t()
		case 'p', 'P':
			l.p()
		case 'b':
			l.b()
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
			fail("bad load format: %d `%s`", i, st)
		}
	}
	return nil
}
