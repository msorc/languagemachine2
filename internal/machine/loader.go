package machine

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
)

// Loader builds rules from bytecode (see docs/bytecode.md) into its engine's
// grammars, using the engine's symbol dictionaries.
type loader struct {
	engine     *Engine
	operands   []Element // stack, top last
	count      int
	ruleNumber int
	text       string // the bytecode being loaded, for error positions
	pos        int    // byte offset of the current token in text
}

// fail reports a fault in the bytecode as line:col: message, for the caller
// to prefix with the name of the rules file.
func (l *loader) fail(format string, args ...any) {
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

func newLoader(e *Engine) *loader {
	return &loader{engine: e}
}

func (l *loader) push(x Element) {
	l.operands = append(l.operands, x)
	l.count++
}

func (l *loader) pop() Element {
	if len(l.operands) == 0 {
		l.fail("operand stack underflow")
	}
	l.count--
	x := l.operands[len(l.operands)-1]
	l.operands = l.operands[:len(l.operands)-1]
	return x
}

func (l *loader) openMark() {
	l.operands = append(l.operands, newNumber(LMNumber(l.count)))
	l.count = 0
}

func (l *loader) closeMark() {
	n, ok := l.pop().(*number)
	if !ok {
		l.fail("unbalanced parentheses")
	}
	l.count = n.toInt()
}

func (l *loader) take(n int) []Element {
	v := make([]Element, n)
	for i := len(v); i > 0; i-- {
		v[i-1] = l.pop()
	}
	return v
}

func (l *loader) pushPriority(p priority) {
	l.push(newNumber(LMNumber(p)))
}

func (l *loader) pushNum(x float64) {
	l.push(newNumber(LMNumber(x)))
}

func (l *loader) pushChars(x string) {
	for _, ch := range x {
		l.push(l.engine.terminalSymbols.uniqueR(ch))
	}
}

func (l *loader) pushQuote(x string) {
	l.push(newQuote(l.engine.nonTerminalSymbols.uniqueE(newSym(x))))
}

func (l *loader) pushNonTerminal(x string) {
	l.push(l.engine.nonTerminalSymbols.uniqueE(newSym(x)))
}

func (l *loader) pushFunction(x string) {
	l.push(l.engine.functionSymbols.uniqueE(newSym(x)))
}

func (l *loader) openList() {
	l.openMark()
}

func (l *loader) closeList() {
	v := l.take(l.count)
	l.closeMark()
	l.push(newStr(v))
}

// r defines a rule from the operands grammar, priority, offset, lhs, rhs.
func (l *loader) defineRule() {
	if len(l.operands) < 5 {
		l.fail("a rule needs 5 operands, found %d", len(l.operands))
	}
	v := l.take(5)
	if _, ok := v[1].(*number); !ok {
		l.fail("the priority of a rule is not a number: %s", v[1].ToString())
	}
	if _, ok := v[2].(*number); !ok {
		l.fail("the offset of a rule is not a number: %s", v[2].ToString())
	}
	for i, side := range []string{"left", "right"} {
		x, ok := v[3+i].(*str)
		if !ok {
			l.fail("the %s side of a rule is not a list: %s", side, v[3+i].ToString())
		}
		if len(x.v) == 0 {
			l.fail("the %s side of a rule is empty", side)
		}
	}
	l.engine.addRule(v, l.ruleNumber)
	l.ruleNumber++
}

func (l *loader) pushAll() {
	l.push(newAllRef(l.pop()))
}

func (l *loader) pushEach() {
	l.push(newEachRef(l.pop()))
}

// p (and P, which is the same) binds: v:X p is get X, then bind.
func (l *loader) pushBind() {
	v := l.pop()
	l.push(newGetXF(v))
	l.push(l.engine.predefinedSymbols.bindFn)
}

func (l *loader) pushTake() {
	l.push(l.engine.predefinedSymbols.takeFn)
}

func (l *loader) pushBindFn() {
	l.push(l.engine.predefinedSymbols.bindFn)
}

func (l *loader) pushGet() {
	l.push(l.engine.predefinedSymbols.getFn)
}

func (l *loader) pushDrop() {
	l.push(l.engine.predefinedSymbols.dropFn)
}

func (l *loader) pushGetX() {
	l.push(newGetXF(l.pop()))
}

func (l *loader) pushGetV() {
	l.push(newGetVF(l.pop()))
}

func (l *loader) pushStr() {
	l.push(l.engine.predefinedSymbols.strFn)
}

func (l *loader) pushAct() {
	l.push(l.engine.predefinedSymbols.actFn)
}

func (l *loader) pushDontCare() {
	l.push(l.engine.predefinedSymbols.nil)
}

func (l *loader) pushDeclare() {
	l.push(newDeclareVar())
}

func (l *loader) pushLex(x string) {
	l.push(l.engine.nonTerminalSymbols.uniqueE(newLex(x, l.engine)))
}

func (l *loader) pushVar(x string) {
	l.push(l.engine.varSymbols.uniqueE(newVarSym(x)))
}

// decode decodes the text of an X:value token: URL decoding, then C escapes.
func (l *loader) decode(s string) string {
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
func (l *loader) level(st string) int {
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
func (l *loader) load(tt string) (err error) {
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
		if l.engine.tracer.tracing(TraceLoad) {
			l.engine.printf("load: %s\n", st)
		}
		switch st {
		case "E":
			l.push(newEachX("each"))
			continue
		case "B":
			l.push(newAllX("all"))
			continue
		case "T":
			// lm2n2mbe has a rule for top, but lm2n2xfe never produces it
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
			l.pushNum(x)
		case 'c':
			l.pushChars(l.decode(st[2:]))
		case 'd':
			l.pushQuote(l.decode(st[2:]))
		case 'm':
			l.pushNonTerminal(l.decode(st[2:]))
		case 'f':
			l.pushFunction(l.decode(st[2:]))
		case 'l':
			l.pushLex(l.decode(st[2:]))
		case 'v':
			l.pushVar(l.decode(st[2:]))
		case '(':
			l.openList()
		case ')':
			l.closeList()
		case '.':
			l.pushDrop()
		case 'r':
			l.defineRule()
		case 'A':
			l.pushAll()
		case 'e':
			l.pushEach()
		case 't':
			l.pushTake()
		case 'p', 'P':
			l.pushBind()
		case 'b':
			l.pushBindFn()
		case 'g':
			l.pushGet()
		case 'G':
			l.pushGetX()
		case 'V':
			l.pushGetV()
		case 's':
			l.pushStr()
		case 'a':
			l.pushAct()
		case 'w':
			l.pushDeclare()
		case 'z':
			l.pushDontCare()
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
