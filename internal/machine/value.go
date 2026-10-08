package machine

import (
	"math"
	"strconv"
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
)

type Number struct {
	GenericElement
	V LMNumber
}

func NewNumber(x LMNumber) *Number {
	n := MakeSelf[Number]()
	n.V = x
	return n
}

func (n *Number) Weight() int {
	return 1
}

func (n *Number) Compare(e *Engine, r Element) bool {
	return r.IsNumber() && r.ToNumber() == n.V
}

// ToString formats with %g and 6 significant digits, which is what the
// original machine printed: 720, 33.3333.
func (n *Number) ToString() string {
	switch f := float64(n.V); {
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

func (n *Number) ToEncode() string {
	return n.ToString()
}

func (n *Number) ToNumber() LMNumber {
	return n.V
}

func (n *Number) IsNumber() bool {
	return true
}

func (n *Number) ToBool() bool {
	return n.V != 0.0
}

func (n *Number) ToInt() int {
	return int(n.V)
}

func (n *Number) Negf(sr *Stream) Element {
	return NewNumber(-n.V)
}

func (n *Number) Invf(sr *Stream) Element {
	return NewNumber(LMNumber(^n.Self().ToInt()))
}

func (n *Number) BitXorf(sr *Stream, y Element) Element {
	return NewNumber(LMNumber(n.Self().ToInt() ^ y.Self().ToInt()))
}

func (n *Number) BitOrf(sr *Stream, y Element) Element {
	return NewNumber(LMNumber(n.Self().ToInt() | y.Self().ToInt()))
}

func (n *Number) BitAndf(sr *Stream, y Element) Element {
	return NewNumber(LMNumber(n.Self().ToInt() & y.Self().ToInt()))
}

func (n *Number) Addf(sr *Stream, y Element) Element {
	return NewNumber(n.V + y.ToNumber())
}

func (n *Number) Subf(sr *Stream, y Element) Element {
	return NewNumber(n.V - y.ToNumber())
}

func (n *Number) Mulf(sr *Stream, y Element) Element {
	return NewNumber(n.V * y.ToNumber())
}

func (n *Number) Divf(sr *Stream, y Element) Element {
	return NewNumber(n.V / y.ToNumber())
}

// Modf is % on doubles (fmod), as in the original: no panic on a zero divisor.
func (n *Number) Modf(sr *Stream, y Element) Element {
	return NewNumber(LMNumber(math.Mod(float64(n.V), float64(y.ToNumber()))))
}

func (n *Number) Eqf(sr *Stream, y Element) Element {
	return NewBoolean(n.V == y.ToNumber())
}

func (n *Number) Nef(sr *Stream, y Element) Element {
	return NewBoolean(n.V != y.ToNumber())
}

func (n *Number) Ltf(sr *Stream, y Element) Element {
	return NewBoolean(n.V < y.ToNumber())
}

func (n *Number) Gtf(sr *Stream, y Element) Element {
	return NewBoolean(n.V > y.ToNumber())
}

func (n *Number) Lef(sr *Stream, y Element) Element {
	return NewBoolean(n.V <= y.ToNumber())
}

func (n *Number) Gef(sr *Stream, y Element) Element {
	return NewBoolean(n.V >= y.ToNumber())
}

type Boolean struct {
	GenericElement
	V bool
}

func NewBoolean(x bool) *Boolean {
	b := MakeSelf[Boolean]()
	b.V = x
	return b
}

func (b *Boolean) ToBool() bool {
	return b.V
}

func (b *Boolean) ToInt() int {
	if b.V {
		return 1
	}
	return 0
}

func (b *Boolean) Compare(e *Engine, r Element) bool {
	return r.ToBool() == b.V
}

func (b *Boolean) ToString() string {
	if b.V {
		return "true"
	}
	return "false"
}

func (b *Boolean) ToEncode() string {
	return b.ToString()
}

func (b *Boolean) Notf(sr *Stream) Element {
	return NewBoolean(!b.V)
}

type Symbol struct {
	GenericElement
	V string
}

func NewSymbol(x string) *Symbol {
	el := MakeSelf[Symbol]()
	el.V = x
	return el
}

func (s *Symbol) ToDump() string {
	return "m:" + conv.Encode(s.V)
}

func (s *Symbol) ToString() string {
	return s.V
}

func (s *Symbol) Weight() int {
	return 1
}

func (s *Symbol) Match(e *Engine, r Element) bool {
	if r.Token() == s.Self() {
		return e.Matched3E(s.Self(), r, r)
	}
	return e.ResolveE(s.Self(), r)
}

func (s *Symbol) Eqf(sr *Stream, y Element) Element {
	return NewBoolean(y.Token() == s.Self())
}

func (s *Symbol) Nef(sr *Stream, y Element) Element {
	return NewBoolean(y.Token() != s.Self())
}

type Quote struct {
	GenericElement
	V Element
}

func NewQuote(x Element) *Quote {
	el := MakeSelf[Quote]()
	el.V = x
	return el
}

func (q *Quote) Token() Element {
	return q.V.Token()
}

func (q *Quote) ToDump() string {
	return "d:" + conv.Encode(q.V.ToString())
}

func (q *Quote) ToEncode() string {
	return conv.Encode(q.ToString())
}

func (q *Quote) ToString() string {
	return q.V.ToString()
}

func (q *Quote) Weight() int {
	return 1
}

func (q *Quote) Match(e *Engine, r Element) bool {
	if r.Token() == q.V {
		return e.Matched3E(q.Self(), r, r)
	}
	return e.ResolveE(q.Self(), r)
}

func (q *Quote) ToBool() bool {
	return q.Self().Token().ToBool()
}

func (q *Quote) Eqf(sr *Stream, y Element) Element {
	return NewBoolean(y.Token() == q.Self().Token())
}

func (q *Quote) Nef(sr *Stream, y Element) Element {
	return NewBoolean(y.Token() != q.Self().Token())
}

type Chr struct {
	Symbol
	V rune
}

func NewChr(x rune) *Chr {
	el := MakeSelf[Chr]()
	el.V = x
	return el
}

func (c *Chr) ToString() string {
	return string(c.V)
}

func (c *Chr) ToTrace() string {
	return "'" + c.escaped(c.ToString()) + "'"
}

func (c *Chr) ToEncode() string {
	return conv.Encode(c.escaped(c.ToString()))
}

func (c *Chr) ToDump() string {
	return "d:" + conv.Encode(c.ToString())
}

func (c *Chr) escaped(x string) string {
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

type ZLM struct {
	Symbol
}

func NewZLM(x string) *ZLM {
	el := MakeSelf[ZLM]()
	el.V = x
	return el
}

func (z *ZLM) ToBool() bool {
	return false
}

func (z *ZLM) Append(sr *Stream, x Element) Element {
	return NewLMBufferFromElement(x)
}

type Zzz struct {
	Symbol
}

func NewZzz(x string) *Zzz {
	el := MakeSelf[Zzz]()
	el.V = x
	return el
}

// NewSym is NewSymbol.
func NewSym(x string) *Symbol {
	return NewSymbol(x)
}

type Str struct {
	GenericElement
	V []Element
}

func NewStr(x []Element) *Str {
	el := MakeSelf[Str]()
	el.V = x
	return el
}

func (s *Str) Token() Element {
	return nil
}

func (s *Str) ToBody() []Element {
	return s.V
}

func (s *Str) ToString() string {
	var r strings.Builder
	for _, x := range s.V {
		r.WriteString(x.ToString())
	}
	return r.String()
}

func (s *Str) ToTrace() string {
	var r strings.Builder
	r.WriteString("{ ")
	for _, x := range s.V {
		r.WriteString(x.ToTrace() + " ")
	}
	r.WriteString("}")
	return r.String()
}

func (s *Str) NewRHX(m GenMode, c ContextHolder, x ScopeHolder) GenMode {
	return NewRHModeFromParamsAndScope(m, s.V, 0, c, x)
}

func (s *Str) Act(sr *Stream, m GenMode) GenMode {
	return NewSTModeFromElements(m, s.V, m)
}

func (s *Str) Reference(sr *Stream, m GenMode, x ScopeHolder) GenMode {
	return NewSTModeFromElements(m, s.V, x)
}

type ChrStr struct {
	Str
}

func NewChrStr(x []Element) *ChrStr {
	return ReSelf(&ChrStr{Str: *NewStr(x)})
}

type LMBuffer struct {
	GenericElement
	V string
}

func NewLMBuffer() *LMBuffer {
	el := MakeSelf[LMBuffer]()
	el.V = ""
	return el
}

func NewLMBufferFromElement(x Element) *LMBuffer {
	el := MakeSelf[LMBuffer]()
	el.V = x.ToString()
	return el
}

func (lb *LMBuffer) Append(sr *Stream, x Element) Element {
	lb.V += x.ToString()
	return lb.Self()
}

func (lb *LMBuffer) ToString() string {
	return lb.V
}
