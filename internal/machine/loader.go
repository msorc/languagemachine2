package machine

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
)

// Loader builds rules from bytecode (see docs/bytecode.md) into its engine's
// grammars, using the engine's symbol dictionaries.
type Loader struct {
	engine     *Engine
	operands   []Element // stack, top last
	count      int
	ruleNumber int
	text       string // the bytecode being loaded, for error positions
	pos        int    // byte offset of the current token in text
}

// fail reports a fault in the bytecode as line:col: message, for the caller
// to prefix with the name of the rules file.
func (l *Loader) fail(format string, args ...any) {
	line, col := 1, 1
	for i, c := range l.text[:l.pos] {
		if c == '\n' {
			line++
			col = l.pos - i
		}
	}
	if line == 1 {
		col = l.pos + 1
	}
	fail("%d:%d: "+format, append([]any{line, col}, args...)...)
}

func NewLoader(e *Engine) *Loader {
	return &Loader{engine: e}
}

func (l *Loader) Push(x Element) {
	l.operands = append(l.operands, x)
	l.count++
}

func (l *Loader) Pop() Element {
	if len(l.operands) == 0 {
		l.fail("operand stack underflow")
	}
	l.count--
	x := l.operands[len(l.operands)-1]
	l.operands = l.operands[:len(l.operands)-1]
	return x
}

func (l *Loader) BMark() {
	l.operands = append(l.operands, NewNumber(LMNumber(l.count)))
	l.count = 0
}

func (l *Loader) EMark() {
	n, ok := l.Pop().(*Number)
	if !ok {
		l.fail("unbalanced parentheses")
	}
	l.count = n.ToInt()
}

func (l *Loader) Take(n int) []Element {
	v := make([]Element, n)
	for i := len(v); i > 0; i-- {
		v[i-1] = l.Pop()
	}
	return v
}

func (l *Loader) pushPriority(p priority) {
	l.Push(NewNumber(LMNumber(p)))
}

func (l *Loader) n(x float64) {
	l.Push(NewNumber(LMNumber(x)))
}

func (l *Loader) c(x string) {
	for _, ch := range x {
		l.Push(l.engine.terminalSymbols.UniqueR(ch))
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

// r defines a rule from the operands grammar, priority, offset, lhs, rhs.
func (l *Loader) r() {
	if len(l.operands) < 5 {
		l.fail("a rule needs 5 operands, found %d", len(l.operands))
	}
	v := l.Take(5)
	if _, ok := v[1].(*Number); !ok {
		l.fail("the priority of a rule is not a number: %s", v[1].ToString())
	}
	if _, ok := v[2].(*Number); !ok {
		l.fail("the offset of a rule is not a number: %s", v[2].ToString())
	}
	for i, side := range []string{"left", "right"} {
		x, ok := v[3+i].(*Str)
		if !ok {
			l.fail("the %s side of a rule is not a list: %s", side, v[3+i].ToString())
		}
		if len(x.V) == 0 {
			l.fail("the %s side of a rule is empty", side)
		}
	}
	l.engine.AddRule(v, l.ruleNumber)
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
	d, err := conv.Decode(s)
	if err == nil {
		d, err = conv.Unescape(d)
	}
	if err != nil {
		l.fail("bad rule text `%s`: %v", s, err)
	}
	return d
}

// level decodes the level of a priority token such as L:20.
func (l *Loader) level(st string) int {
	n, err := conv.Strtoi(st[2:])
	if err != nil {
		l.fail("bad priority `%s`", st)
	}
	return n
}

// tokenRE splits bytecode into tokens. The final (\S) catches
// single-character opcodes the loader does not know, so they reach the
// bad-format error instead of being skipped.
var tokenRE = regexp.MustCompile(`([().reAtpbPgGVsawz])|(.:\S*)|#[^\n]*\n|(\S)|\s*`)

// Load defines the rules in tt, bytecode as described in docs/bytecode.md.
// A fault in the bytecode is returned as an *Error naming its line and
// column; Engine.LoadFromStringReset then takes out the rules defined
// before it.
func (l *Loader) Load(tt string) (err error) {
	defer catch(&err)
	l.text, l.operands, l.count = tt, l.operands[:0], 0
	for _, m := range tokenRE.FindAllStringIndex(tt, -1) {
		st := tt[m[0]:m[1]]
		// comment lines (incl. a #! shebang) are separators in the original's
		// RegExp.split, so they never reach the dispatch
		if len(strings.TrimSpace(st)) == 0 || st[0] == '#' {
			continue
		}
		l.pos = m[0]
		if t := l.engine.tracer; t != nil && t.Tracing(LOAD) {
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
			l.fail("unsupported opcode `%s` (top is not implemented)", st)
		}
		// opcodes that carry a value are written X:value
		if strings.IndexByte("MLRBncdmflv", st[0]) >= 0 && (len(st) < 2 || st[1] != ':') {
			l.fail("opcode %c needs a value: `%s`", st[0], st)
		}
		switch st[0] {
		case 'M':
			l.level(st) // checked, but the level of M is not significant
			l.pushPriority(maximal)
		case 'L':
			l.pushPriority(leftPriority(l.level(st)))
		case 'R':
			l.pushPriority(rightPriority(l.level(st)))
		case 'B':
			l.pushPriority(bracketPriority(l.level(st)))
		case 'n':
			// numeric literals may be real (n:2.5), not just the rule offset
			x, err := strconv.ParseFloat(st[2:], 64)
			if err != nil {
				l.fail("bad number `%s`", st)
			}
			l.n(x)
		case 'c':
			l.c(l.MStr(st[2:]))
		case 'd':
			l.d(l.MStr(st[2:]))
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
			l.fail("bad load format `%s`", st)
		}
	}
	if len(l.operands) != 0 {
		l.pos = len(tt)
		l.fail("%d operands left over: an unclosed ( or a rule without r", len(l.operands))
	}
	return nil
}
