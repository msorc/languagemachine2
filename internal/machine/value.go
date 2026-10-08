package machine

import (
	"math"
	"strconv"
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
)

type number struct {
	genericElement
	v LMNumber
}

func newNumber(x LMNumber) *number {
	n := makeSelf[number]()
	n.v = x
	return n
}

func (n *number) weight() int {
	return 1
}

func (n *number) compare(e *Engine, r Element) bool {
	return r.IsNumber() && r.ToNumber() == n.v
}

// ToString formats with %g and 6 significant digits, which is what the
// original machine printed: 720, 33.3333.
func (n *number) ToString() string {
	switch f := float64(n.v); {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	default:
		return strconv.FormatFloat(f, 'g', 6, 64)
	}
}

func (n *number) toEncode() string {
	return n.ToString()
}

func (n *number) ToNumber() LMNumber {
	return n.v
}

func (n *number) IsNumber() bool {
	return true
}

func (n *number) ToBool() bool {
	return n.v != 0.0
}

func (n *number) toInt() int {
	return int(n.v)
}

func (n *number) negf(sr *Stream) Element {
	return newNumber(-n.v)
}

func (n *number) invf(sr *Stream) Element {
	return newNumber(LMNumber(^n.self().toInt()))
}

func (n *number) bitXorf(sr *Stream, y Element) Element {
	return newNumber(LMNumber(n.self().toInt() ^ y.self().toInt()))
}

func (n *number) bitOrf(sr *Stream, y Element) Element {
	return newNumber(LMNumber(n.self().toInt() | y.self().toInt()))
}

func (n *number) bitAndf(sr *Stream, y Element) Element {
	return newNumber(LMNumber(n.self().toInt() & y.self().toInt()))
}

func (n *number) addf(sr *Stream, y Element) Element {
	return newNumber(n.v + y.ToNumber())
}

func (n *number) subf(sr *Stream, y Element) Element {
	return newNumber(n.v - y.ToNumber())
}

func (n *number) mulf(sr *Stream, y Element) Element {
	return newNumber(n.v * y.ToNumber())
}

func (n *number) divf(sr *Stream, y Element) Element {
	return newNumber(n.v / y.ToNumber())
}

// Modf is % on doubles (fmod), as in the original: no panic on a zero divisor.
func (n *number) modf(sr *Stream, y Element) Element {
	return newNumber(LMNumber(math.Mod(float64(n.v), float64(y.ToNumber()))))
}

func (n *number) eqf(sr *Stream, y Element) Element {
	return newBoolean(n.v == y.ToNumber())
}

func (n *number) nef(sr *Stream, y Element) Element {
	return newBoolean(n.v != y.ToNumber())
}

func (n *number) ltf(sr *Stream, y Element) Element {
	return newBoolean(n.v < y.ToNumber())
}

func (n *number) gtf(sr *Stream, y Element) Element {
	return newBoolean(n.v > y.ToNumber())
}

func (n *number) lef(sr *Stream, y Element) Element {
	return newBoolean(n.v <= y.ToNumber())
}

func (n *number) gef(sr *Stream, y Element) Element {
	return newBoolean(n.v >= y.ToNumber())
}

type boolean struct {
	genericElement
	v bool
}

func newBoolean(x bool) *boolean {
	b := makeSelf[boolean]()
	b.v = x
	return b
}

func (b *boolean) ToBool() bool {
	return b.v
}

func (b *boolean) toInt() int {
	if b.v {
		return 1
	}
	return 0
}

func (b *boolean) compare(e *Engine, r Element) bool {
	return r.ToBool() == b.v
}

func (b *boolean) ToString() string {
	if b.v {
		return "true"
	}
	return "false"
}

func (b *boolean) toEncode() string {
	return b.ToString()
}

func (b *boolean) notf(sr *Stream) Element {
	return newBoolean(!b.v)
}

type symbol struct {
	genericElement
	v string
}

func newSymbol(x string) *symbol {
	el := makeSelf[symbol]()
	el.v = x
	return el
}

func (s *symbol) toDump() string {
	return "m:" + conv.Encode(s.v)
}

func (s *symbol) ToString() string {
	return s.v
}

func (s *symbol) weight() int {
	return 1
}

func (s *symbol) match(e *Engine, r Element) bool {
	if r.token() == s.self() {
		return e.matchedWith(s.self(), r, r)
	}
	return e.resolve(s.self(), r)
}

func (s *symbol) eqf(sr *Stream, y Element) Element {
	return newBoolean(y.token() == s.self())
}

func (s *symbol) nef(sr *Stream, y Element) Element {
	return newBoolean(y.token() != s.self())
}

type quote struct {
	genericElement
	v Element
}

func newQuote(x Element) *quote {
	el := makeSelf[quote]()
	el.v = x
	return el
}

func (q *quote) token() Element {
	return q.v.token()
}

func (q *quote) toDump() string {
	return "d:" + conv.Encode(q.v.ToString())
}

func (q *quote) toEncode() string {
	return conv.Encode(q.ToString())
}

func (q *quote) ToString() string {
	return q.v.ToString()
}

func (q *quote) weight() int {
	return 1
}

func (q *quote) match(e *Engine, r Element) bool {
	if r.token() == q.v {
		return e.matchedWith(q.self(), r, r)
	}
	return e.resolve(q.self(), r)
}

func (q *quote) ToBool() bool {
	return q.self().token().ToBool()
}

func (q *quote) eqf(sr *Stream, y Element) Element {
	return newBoolean(y.token() == q.self().token())
}

func (q *quote) nef(sr *Stream, y Element) Element {
	return newBoolean(y.token() != q.self().token())
}

type chr struct {
	symbol
	v rune
}

func newChr(x rune) *chr {
	el := makeSelf[chr]()
	el.v = x
	return el
}

func (c *chr) ToString() string {
	return string(c.v)
}

func (c *chr) toTrace() string {
	return "'" + c.escaped(c.ToString()) + "'"
}

func (c *chr) toEncode() string {
	return conv.Encode(c.escaped(c.ToString()))
}

func (c *chr) toDump() string {
	return "d:" + conv.Encode(c.ToString())
}

func (c *chr) escaped(x string) string {
	switch x {
	case "\a":
		return "\\a"
	case "\b":
		return "\\b"
	case "\n":
		return "\\n"
	case "\r":
		return "\\r"
	case "\t":
		return "\\t"
	case "\f":
		return "\\f"
	case "\v":
		return "\\v"
	case "'":
		return "\\'"
	case "\"":
		return "\\\""
	case "\\":
		return "\\\\"
	case " ":
		return " "
	default:
		return x
	}
}

type nullSym struct {
	symbol
}

func newNullSym(x string) *nullSym {
	el := makeSelf[nullSym]()
	el.v = x
	return el
}

func (z *nullSym) ToBool() bool {
	return false
}

func (z *nullSym) append(sr *Stream, x Element) Element {
	return newBufferValueOf(x)
}

type dontCare struct {
	symbol
}

func newDontCare(x string) *dontCare {
	el := makeSelf[dontCare]()
	el.v = x
	return el
}

// newSym is newSymbol.
func newSym(x string) *symbol {
	return newSymbol(x)
}

type str struct {
	genericElement
	v []Element
}

func newStr(x []Element) *str {
	el := makeSelf[str]()
	el.v = x
	return el
}

func (s *str) token() Element {
	return nil
}

func (s *str) toBody() []Element {
	return s.v
}

func (s *str) ToString() string {
	var r strings.Builder
	for _, x := range s.v {
		r.WriteString(x.ToString())
	}
	return r.String()
}

func (s *str) toTrace() string {
	var r strings.Builder
	r.WriteString("{ ")
	for _, x := range s.v {
		r.WriteString(x.toTrace() + " ")
	}
	r.WriteString("}")
	return r.String()
}

func (s *str) newRHX(m GenMode, c contextHolder, x scopeHolder) GenMode {
	return newRHMode(m, s.v, 0, c, x)
}

func (s *str) act(sr *Stream, m GenMode) GenMode {
	return newSTMode(m, s.v, m)
}

func (s *str) reference(sr *Stream, m GenMode, x scopeHolder) GenMode {
	return newSTMode(m, s.v, x)
}

type chrStr struct {
	str
}

func newChrStr(x []Element) *chrStr {
	return reSelf(&chrStr{str: *newStr(x)})
}

type bufferValue struct {
	genericElement
	v string
}

func newBufferValue() *bufferValue {
	el := makeSelf[bufferValue]()
	el.v = ""
	return el
}

func newBufferValueOf(x Element) *bufferValue {
	el := makeSelf[bufferValue]()
	el.v = x.ToString()
	return el
}

func (lb *bufferValue) append(sr *Stream, x Element) Element {
	lb.v += x.ToString()
	return lb.self()
}

func (lb *bufferValue) ToString() string {
	return lb.v
}
