package machine

import (
	"fmt"
	"os"
	"strings"
)

const (
	IN = iota
	C1
	E1
	C2
	RN
)

type LMNumber float64

type GrammarElement interface {
	AddRule(*Grammar, *Rule)
	Match(*Engine, GrammarElement) bool
	NewLHS(GenMode) GenMode
	NewRHX(GenMode, EngineStateContext, LMScope) GenMode
	Act(*Stream, GenMode) GenMode
	Compare(*Engine, GrammarElement) bool
	ToNumber() LMNumber
	IsNumber() bool
	ToBool() bool
	ToVar() *Var
	ToDouble() float64
	ToLong() int64
	ToUlong() uint
	ToInt() int
	ToType() string
	Put(*[]any)
	Len() uint
	Trace(*Stream, *Tracer)
	ToString() string
	ToTrace() string
	ToEncode() string
	ToDecode() string
	ToDump() string
	Dump()
	ToBody() []GrammarElement
	Weight() uint
	Token() GrammarElement
	Priority(uint) uint
	Reference(*Stream, GenMode, LMScope) GenMode
	ToExplore() GrammarElement
	InvalidOp(string) GrammarElement
	NotFound() GrammarElement
	ToVal() GrammarElement
	ToDeref(*Var) *Var
	Append(GrammarElement) GrammarElement
	Inf(GrammarElement) GrammarElement
	Idxf(y GrammarElement) GrammarElement
	Idtf(y GrammarElement) GrammarElement
	StoValf(y GrammarElement) GrammarElement
	StoAddf(y GrammarElement) GrammarElement
	StoSubf(y GrammarElement) GrammarElement
	StoMulf(y GrammarElement) GrammarElement
	StoDivf(y GrammarElement) GrammarElement
	StoModf(y GrammarElement) GrammarElement
	StoAndf(y GrammarElement) GrammarElement
	StoOrf(y GrammarElement) GrammarElement
	StoXorf(y GrammarElement) GrammarElement
	StoShlf(y GrammarElement) GrammarElement
	StoShrf(y GrammarElement) GrammarElement
	Eeqf(y GrammarElement) GrammarElement
	Neef(y GrammarElement) GrammarElement
	Eqf(y GrammarElement) GrammarElement
	Nef(y GrammarElement) GrammarElement
	Ltf(y GrammarElement) GrammarElement
	Gtf(y GrammarElement) GrammarElement
	Lef(y GrammarElement) GrammarElement
	Gef(y GrammarElement) GrammarElement
	BitXorf(y GrammarElement) GrammarElement
	BitOrf(y GrammarElement) GrammarElement
	BitAndf(y GrammarElement) GrammarElement
	OrOrf(y GrammarElement) GrammarElement
	AndAndf(y GrammarElement) GrammarElement
	Addf(y GrammarElement) GrammarElement
	Subf(y GrammarElement) GrammarElement
	Mulf(y GrammarElement) GrammarElement
	Divf(y GrammarElement) GrammarElement
	Modf(y GrammarElement) GrammarElement
	Preincf() GrammarElement
	Predecf() GrammarElement
	Postincf() GrammarElement
	Postdecf() GrammarElement
	Posf() GrammarElement
	Negf() GrammarElement
	Notf() GrammarElement
	Invf() GrammarElement
	OpAdd(GrammarElement) GrammarElement
	OpShl(GrammarElement) GrammarElement
	OpShr(GrammarElement) GrammarElement
	OpUShr(GrammarElement) GrammarElement
	OpCat(GrammarElement) GrammarElement
	OpEquals(GrammarElement) bool
	OpCmp(GrammarElement) int
	OpAndAssign(GrammarElement) GrammarElement
	OpOrAssign(GrammarElement) GrammarElement
	OpXorAssign(GrammarElement) GrammarElement
	OpShlAssign(GrammarElement) GrammarElement
	OpShrAssign(GrammarElement) GrammarElement
	OpUShrAssign(GrammarElement) GrammarElement
	OpCatAssign(GrammarElement) GrammarElement
	OpCall() GrammarElement
	OpIndex() GrammarElement
	OpIndexElement(GrammarElement) GrammarElement
	OpIndexAssign() GrammarElement
	OpIndexAssignElement(GrammarElement, GrammarElement) GrammarElement
	OpSlice() GrammarElement
}

type Element struct {
}

func (e *Element) AddRule(g *Grammar, x *Rule) {
	g.Add(x)
}

func (e *Element) Match(engine *Engine, r GrammarElement) bool {
	if e.Compare(engine, r) {
		engine.Matched3E(e, r, r)
		return true
	} else {
		return engine.ResolveE(e, r)
	}
}

func (e *Element) NewLHS(m GenMode) GenMode {
	panic("not implemented")
}

func (e *Element) NewRHX(m GenMode, c EngineStateContext, x LMScope) GenMode {
	panic("not implemented")
}

func (e *Element) Act(sr *Stream, s GenMode) GenMode {
	sr.SY = e
	return s
}

func (e *Element) Compare(engine *Engine, r GrammarElement) bool {
	return false
}

func (e *Element) ToNumber() LMNumber {
	panic("not implemented")
}

func (e *Element) IsNumber() bool {
	return false
}

func (e *Element) ToBool() bool {
	return true
}

func (e *Element) ToVar() *Var {
	panic("not implemented")
}

func (e *Element) ToDouble() float64 {
	panic("not implemented")
}

func (e *Element) ToLong() int64 {
	panic("not implemented")
}

func (e *Element) ToUlong() uint {
	panic("not implemented")
}

func (e *Element) ToInt() int {
	panic("not implemented")
}

func (e *Element) ToType() string {
	return "Element"
}

func (e *Element) Put(argp *[]any) {
	*argp = append(*argp, e)
}

func (e *Element) Len() uint {
	return 0
}

func (e *Element) Trace(s *Stream, t *Tracer) {
}

func (e *Element) ToString() string {
	return "element"
}

func (e *Element) ToTrace() string {
	return e.ToString()
}

func (e *Element) ToEncode() string {
	return e.ToString()
}

func (e *Element) ToDecode() string {
	return e.ToString()
}

func (e *Element) ToDump() string {
	return e.ToEncode()
}

func (e *Element) Dump() {
	fmt.Printf("%s ", e.ToEncode())
}

func (e *Element) ToBody() []GrammarElement {
	return nil
}

func (e *Element) Weight() uint {
	return 0
}

func (e *Element) Token() GrammarElement {
	return e
}

func (e *Element) Priority(p uint) uint {
	return p
}

func (e *Element) Reference(sr *Stream, s GenMode, x LMScope) GenMode {
	return e.Act(sr, s)
}

func (e *Element) ToExplore() GrammarElement {
	return TxE("E", e)
}

func (e *Element) InvalidOp(f string) GrammarElement {
	TxE("BAD "+f, e)
	return theNull()
}

func (e *Element) NotFound() GrammarElement {
	return theNull()
}

func (e *Element) ToVal() GrammarElement {
	return e
}

func (e *Element) ToDeref(x *Var) *Var {
	return nil
}

func (e *Element) Append(y GrammarElement) GrammarElement {
	return e.InvalidOp("~=")
}

func (e *Element) Inf(y GrammarElement) GrammarElement {
	return theNull()
}

func (e *Element) Idxf(y GrammarElement) GrammarElement {
	return theNull()
}

func (e *Element) Idtf(y GrammarElement) GrammarElement {
	return theNull()
}

func (e *Element) StoValf(y GrammarElement) GrammarElement {
	return e.InvalidOp("=")
}

func (e *Element) StoAddf(y GrammarElement) GrammarElement {
	return e.InvalidOp("+=")
}

func (e *Element) StoSubf(y GrammarElement) GrammarElement {
	return e.InvalidOp("-=")
}

func (e *Element) StoMulf(y GrammarElement) GrammarElement {
	return e.InvalidOp("*=")
}

func (e *Element) StoDivf(y GrammarElement) GrammarElement {
	return e.InvalidOp("/=")
}

func (e *Element) StoModf(y GrammarElement) GrammarElement {
	return e.InvalidOp("%=")
}

func (e *Element) StoAndf(y GrammarElement) GrammarElement {
	return e.InvalidOp("&=")
}

func (e *Element) StoOrf(y GrammarElement) GrammarElement {
	return e.InvalidOp("|=")
}

func (e *Element) StoXorf(y GrammarElement) GrammarElement {
	return e.InvalidOp("^=")
}

func (e *Element) StoShlf(y GrammarElement) GrammarElement {
	return e.InvalidOp("<<=")
}

func (e *Element) StoShrf(y GrammarElement) GrammarElement {
	return e.InvalidOp(">>=")
}

func (e *Element) Eeqf(y GrammarElement) GrammarElement {
	return NewBoolean(y.Token() == e.Token())
}

func (e *Element) Neef(y GrammarElement) GrammarElement {
	return NewBoolean(y.Token() != e.Token())
}

func (e *Element) Eqf(y GrammarElement) GrammarElement {
	return e.InvalidOp("==")
}

func (e *Element) Nef(y GrammarElement) GrammarElement {
	return e.InvalidOp("!=")
}

func (e *Element) Ltf(y GrammarElement) GrammarElement {
	return e.InvalidOp("<")
}

func (e *Element) Gtf(y GrammarElement) GrammarElement {
	return e.InvalidOp(">")
}

func (e *Element) Lef(y GrammarElement) GrammarElement {
	return e.InvalidOp("<=")
}

func (e *Element) Gef(y GrammarElement) GrammarElement {
	return e.InvalidOp(">=")
}

func (e *Element) BitXorf(y GrammarElement) GrammarElement {
	return e.InvalidOp("^")
}

func (e *Element) BitOrf(y GrammarElement) GrammarElement {
	return e.InvalidOp("|")
}

func (e *Element) BitAndf(y GrammarElement) GrammarElement {
	return e.InvalidOp("&")
}

func (e *Element) OrOrf(y GrammarElement) GrammarElement {
	return e.InvalidOp("||")
}

func (e *Element) AndAndf(y GrammarElement) GrammarElement {
	return e.InvalidOp("&&")
}

func (e *Element) Addf(y GrammarElement) GrammarElement {
	return e.InvalidOp("+")
}

func (e *Element) Subf(y GrammarElement) GrammarElement {
	return e.InvalidOp("-")
}

func (e *Element) Mulf(y GrammarElement) GrammarElement {
	return e.InvalidOp("*")
}

func (e *Element) Divf(y GrammarElement) GrammarElement {
	return e.InvalidOp("/")
}

func (e *Element) Modf(y GrammarElement) GrammarElement {
	return e.InvalidOp("%")
}

func (e *Element) Preincf() GrammarElement {
	return e.InvalidOp("++X")
}

func (e *Element) Predecf() GrammarElement {
	return e.InvalidOp("--X")
}

func (e *Element) Postincf() GrammarElement {
	return e.InvalidOp("X++")
}

func (e *Element) Postdecf() GrammarElement {
	return e.InvalidOp("X--")
}

func (e *Element) Posf() GrammarElement {
	return e.InvalidOp("u+")
}

func (e *Element) Negf() GrammarElement {
	return e.InvalidOp("u-")
}

func (e *Element) Notf() GrammarElement {
	return NewBoolean(!e.ToBool())
}

func (e *Element) Invf() GrammarElement {
	return e.InvalidOp("~")
}

func (e *Element) OpAdd(y GrammarElement) GrammarElement {
	return e.Addf(y.ToVal())
}

func (e *Element) OpShl(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpShr(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpUShr(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpCat(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpEquals(y GrammarElement) bool {
	panic("badType")
}

func (e *Element) OpCmp(y GrammarElement) int {
	panic("badType")
}

func (e *Element) OpAndAssign(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpOrAssign(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpXorAssign(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpShlAssign(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpShrAssign(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpUShrAssign(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpCatAssign(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpCall() GrammarElement {
	panic("badType")
}

func (e *Element) OpIndex() GrammarElement {
	panic("badType")
}

func (e *Element) OpIndexElement(y GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpIndexAssign() GrammarElement {
	panic("badType")
}

func (e *Element) OpIndexAssignElement(y, z GrammarElement) GrammarElement {
	panic("badType")
}

func (e *Element) OpSlice() GrammarElement {
	panic("badType")
}

type Number struct {
	Element
	V LMNumber
}

func NewNumber(x LMNumber) *Number {
	return &Number{V: x}
}

func (n *Number) Weight() uint {
	return 1
}

func (n *Number) ToBody() []GrammarElement {
	return nil
}

func (n *Number) Token() GrammarElement {
	return n
}

func (n *Number) Compare(e *Engine, r GrammarElement) bool {
	return r.ToNumber() == n.V
}

func (n *Number) Dump() {
	fmt.Printf("n:%f ", n.V)
}

func (n *Number) ToString() string {
	return fmt.Sprintf("n:%f ", n.V)
}

func (n *Number) ToEncode() string {
	return fmt.Sprintf("n:%f ", n.V)
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

func (n *Number) ToDouble() float64 {
	return float64(n.V)
}

func (n *Number) ToLong() int64 {
	return int64(n.V)
}

func (n *Number) ToUlong() uint {
	return uint(n.V)
}

func (n *Number) ToInt() int {
	return int(n.V)
}

func (n *Number) ToType() string {
	return fmt.Sprintf("%T", n.V)
}

func (n *Number) Put(argp *[]any) {
	*argp = append(*argp, n)
}

func (n *Number) Len() uint {
	return 8 // Assuming LMNumber is 8 bytes
}

func (n *Number) Posf() GrammarElement {
	return n
}

func (n *Number) Negf() GrammarElement {
	return NewNumber(-n.V)
}

func (n *Number) Invf() GrammarElement {
	return NewNumber(LMNumber(^n.ToUlong()))
}

func (n *Number) BitXorf(y GrammarElement) GrammarElement {
	return NewNumber(LMNumber(n.ToUlong() ^ y.ToUlong()))
}

func (n *Number) BitOrf(y GrammarElement) GrammarElement {
	return NewNumber(LMNumber(n.ToUlong() | y.ToUlong()))
}

func (n *Number) BitAndf(y GrammarElement) GrammarElement {
	return NewNumber(LMNumber(n.ToUlong() & y.ToUlong()))
}

func (n *Number) Addf(y GrammarElement) GrammarElement {
	return NewNumber(n.V + y.ToNumber())
}

func (n *Number) Subf(y GrammarElement) GrammarElement {
	return NewNumber(n.V - y.ToNumber())
}

func (n *Number) Mulf(y GrammarElement) GrammarElement {
	return NewNumber(n.V * y.ToNumber())
}

func (n *Number) Divf(y GrammarElement) GrammarElement {
	return NewNumber(n.V / y.ToNumber())
}

func (n *Number) Modf(y GrammarElement) GrammarElement {
	return NewNumber(LMNumber(int64(n.V) % y.ToLong()))
}

func (n *Number) Eqf(y GrammarElement) GrammarElement {
	return NewBoolean(n.V == y.ToNumber())
}

func (n *Number) Nef(y GrammarElement) GrammarElement {
	return NewBoolean(n.V != y.ToNumber())
}

func (n *Number) Ltf(y GrammarElement) GrammarElement {
	return NewBoolean(n.V < y.ToNumber())
}

func (n *Number) Gtf(y GrammarElement) GrammarElement {
	return NewBoolean(n.V > y.ToNumber())
}

func (n *Number) Lef(y GrammarElement) GrammarElement {
	return NewBoolean(n.V <= y.ToNumber())
}

func (n *Number) Gef(y GrammarElement) GrammarElement {
	return NewBoolean(n.V >= y.ToNumber())
}

type Boolean struct {
	Element
	V bool
}

func NewBoolean(x bool) *Boolean {
	return &Boolean{V: x}
}

func (b *Boolean) ToBool() bool {
	return b.V
}

func (b *Boolean) ToDouble() float64 {
	if b.V {
		return 1.0
	}
	return 0.0
}

func (b *Boolean) ToLong() int64 {
	if b.V {
		return 1
	}
	return 0
}

func (b *Boolean) ToUlong() uint {
	if b.V {
		return 1
	}
	return 0
}

func (b *Boolean) ToInt() int {
	if b.V {
		return 1
	}
	return 0
}

func (b *Boolean) Compare(e *Engine, r GrammarElement) bool {
	return r.ToBool() == b.V
}

func (b *Boolean) Dump() {
	fmt.Printf("q:%d ", b.ToInt())
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

func (b *Boolean) ToType() string {
	return fmt.Sprintf("%T", b.V)
}

func (b *Boolean) Put(argp *[]any) {
	*argp = append(*argp, b)
}

func (b *Boolean) Len() uint {
	return 1
}

func (b *Boolean) Notf() GrammarElement {
	return NewBoolean(!b.V)
}

type Symbol struct {
	Element
	V string
}

func NewSymbol(x string) *Symbol {
	return &Symbol{V: x}
}

func (s *Symbol) Token() GrammarElement {
	return s
}

func (s *Symbol) ToDump() string {
	return "m:" + s.V
}

func (s *Symbol) Dump() {
	fmt.Printf("m:%s ", s.V)
}

func (s *Symbol) ToString() string {
	return s.V
}

func (s *Symbol) ToEncode() string {
	return s.ToString()
}

func (s *Symbol) ToBody() []GrammarElement {
	return nil
}

func (s *Symbol) Weight() uint {
	return 1
}

func (s *Symbol) Act(sr *Stream, m GenMode) GenMode {
	sr.SY = s
	return m
}

func (s *Symbol) Match(e *Engine, r GrammarElement) bool {
	if r.Token() == s {
		return e.Matched3E(s, r, r)
	}
	return e.ResolveE(s, r)
}

func (s *Symbol) Append(y GrammarElement) GrammarElement {
	return s.InvalidOp("~=")
}

func (s *Symbol) ToVal() GrammarElement {
	return s
}

func (s *Symbol) ToDeref(x *Var) *Var {
	return x.Deref(s)
}

func (s *Symbol) Eqf(y GrammarElement) GrammarElement {
	return NewBoolean(y.Token() == s)
}

func (s *Symbol) Nef(y GrammarElement) GrammarElement {
	return NewBoolean(y.Token() != s)
}

type Quote struct {
	Element
	V GrammarElement
}

func NewQuote(x GrammarElement) *Quote {
	return &Quote{V: x}
}

func (q *Quote) Token() GrammarElement {
	return q.V.Token()
}

func (q *Quote) ToDump() string {
	return "d:" + q.V.ToString()
}

func (q *Quote) Dump() {
	fmt.Printf("d:%s ", q.V.ToString())
}

func (q *Quote) ToEncode() string {
	return q.ToString()
}

func (q *Quote) ToString() string {
	return q.V.ToString()
}

func (q *Quote) Weight() uint {
	return 1
}

func (q *Quote) Act(sr *Stream, m GenMode) GenMode {
	sr.SY = q
	return m
}

func (q *Quote) Match(e *Engine, r GrammarElement) bool {
	if r.Token() == q.V {
		return e.Matched3E(q, r, r)
	}
	return e.ResolveE(q, r)
}

func (q *Quote) ToVal() GrammarElement {
	return q
}

func (q *Quote) ToDeref(x *Var) *Var {
	return nil
}

func (q *Quote) ToBool() bool {
	return q.Token().ToBool()
}

func (q *Quote) Eqf(y GrammarElement) GrammarElement {
	return NewBoolean(y.Token() == q.Token())
}

func (q *Quote) Nef(y GrammarElement) GrammarElement {
	return NewBoolean(y.Token() != q.Token())
}

type Chr struct {
	Symbol
	V rune
}

func NewChr(x rune) *Chr {
	return &Chr{V: x}
}

func (c *Chr) ToString() string {
	return string(c.V)
}

func (c *Chr) Dump() {
	fmt.Printf("c:%s ", c.escaped(c.ToString()))
}

func (c *Chr) ToTrace() string {
	return "'" + c.escaped(c.ToString()) + "'"
}

func (c *Chr) ToEncode() string {
	return c.escaped(c.ToString())
}

func (c *Chr) ToDump() string {
	return "d:" + c.ToString()
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
	return &ZLM{Symbol: Symbol{V: x}}
}

func (z *ZLM) Dump() {
	fmt.Print("null ")
}

func (z *ZLM) ToBool() bool {
	return false
}

func (z *ZLM) Append(x GrammarElement) GrammarElement {
	return NewLMBufferFromElement(x)
}

type Zzz struct {
	Symbol
}

func NewZzz(x string) *Zzz {
	return &Zzz{Symbol: Symbol{V: x}}
}

func (z *Zzz) Dump() {
	fmt.Print("z ")
}

type Sym struct {
	Symbol
}

func NewSym(x string) *Sym {
	return &Sym{Symbol: Symbol{V: x}}
}

type Usr struct {
	Symbol
}

func NewUsr(x string) *Usr {
	return &Usr{Symbol: Symbol{V: x}}
}

type Str struct {
	Element
	V []GrammarElement
}

func NewStr(x []GrammarElement) *Str {
	return &Str{V: x}
}

func (s *Str) Token() GrammarElement {
	return nil
}

func (s *Str) ToBody() []GrammarElement {
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

func (s *Str) Dump() {
	isChr := func(ge GrammarElement) bool { _, ok := ge.(*Chr); return ok }
	fmt.Print("( ")
	i := 0
	for i < len(s.V) {
		c := ""
		if isChr(s.V[i]) {
			for i < len(s.V) && isChr(s.V[i]) {
				c += s.V[i].ToEncode()
				i++
			}
			fmt.Printf("c:%s ", c)
		} else {
			s.V[i].Dump()
			i++
		}
	}
	fmt.Print(") ")
}

func (s *Str) Weight() uint {
	return 0
}

func (s *Str) NewLHS(m GenMode) GenMode {
	return newLHModeFromElement(m, s.V, 0, m.CX())
}

func (s *Str) NewRHX(m GenMode, c EngineStateContext, x LMScope) GenMode {
	return NewRHModeFromParamsAndScope(m, s.V, 0, c, x)
}

func (s *Str) Act(sr *Stream, m GenMode) GenMode {
	return NewSTModeFromElements(m, s.V, m)
}

func (s *Str) Reference(sr *Stream, m GenMode, x LMScope) GenMode {
	return NewSTModeFromElements(m, s.V, x)
}

type ChrStr struct {
	Str
}

func NewChrStr(x []GrammarElement) *ChrStr {
	return &ChrStr{Str: *NewStr(x)}
}

func (cs *ChrStr) Dump() {
	fmt.Print("c:")
	for _, x := range cs.V {
		fmt.Print(x.ToString())
	}
	fmt.Print(" ")
}

type LMBuffer struct {
	Element
	V string
}

func NewLMBuffer() *LMBuffer {
	return &LMBuffer{V: ""}
}

func NewLMBufferFromElement(x GrammarElement) *LMBuffer {
	return &LMBuffer{V: x.ToString()}
}

func (lb *LMBuffer) ToVal() GrammarElement {
	return lb
}

func (lb *LMBuffer) Append(x GrammarElement) GrammarElement {
	lb.V += x.ToString()
	return lb
}

func (lb *LMBuffer) ToString() string {
	return lb.V
}

// LMScope
type Var struct {
	Element
	Vs *Var
	Va *Var
	Vk GrammarElement
	Vv GrammarElement
	Vi uint
	Vp *Var
	Vx LMScope
}

func NewVarDefault() *Var {
	return &Var{}
}

func NewVarFromParams(s *Var, k, v GrammarElement, q LMScope, a *Var) *Var {
	if q == nil {
		panic("vx cannot be nil")
	}
	return &Var{
		Va: a,
		Vs: s,
		Vk: k,
		Vv: v,
		Vp: q.VvP(),
		Vx: q,
	}
}

func (v *Var) Act(sr *Stream, s GenMode) GenMode {
	if v.Vv != nil {
		return v.Vv.Reference(sr, s, v)
	}
	return s
}

func (v *Var) VvP() *Var {
	return v.Vp
}

func (v *Var) VvQ() *Var {
	if v.Vx != nil {
		return v.Vx.VvQ()
	}
	return nil
}

func (v *Var) VvS() LMScope {
	return v.Vx
}

func (v *Var) VvC() EngineStateContext {
	if v.Vx != nil {
		return v.Vx.VvC()
	}
	return nil
}

func (v *Var) MakeVar(k, ve GrammarElement, s LMScope, a *Var) *Var {
	v.Vp = NewVarFromParams(v.Vp, k, ve, s, a)
	return v.Vp
}

func (v *Var) RfScope() LMScope {
	return v
}

func (v *Var) Key() string {
	if v.Vk != nil {
		return v.Vk.ToString()
	}
	return "---"
}

func (v *Var) Value() string {
	if v.Vv != nil {
		return v.Vv.ToString()
	}
	return "---"
}

func (v *Var) ToString() string {
	return v.Key()
}

func (v *Var) DumpIt(s string) {
	fmt.Printf("var: %s %s\n", s, v.ToDebug())
}

func (v *Var) ToDebug() string {
	return "var " + v.Key() + ": " + v.Value()
}

func (v *Var) Deref(k GrammarElement) *Var {
	pp := v.Vp
	for (pp != nil) && !(k == pp.Vk) {
		pp = pp.Vs
	}
	return pp
}

func (v *Var) ToExplore() GrammarElement {
	TxV("V", " ", v)
	if v.Vv != nil {
		v.Vv.ToExplore()
	}
	return v
}

func (v *Var) ToDeref(x *Var) *Var {
	if x.Vv != nil {
		return x.Deref(x.Vv)
	}
	return nil
}

func (v *Var) ToVal() GrammarElement {
	x := v.Vv
	if x != nil {
		if _, ok := x.(*VarSym); ok {
			x = v.Deref(x)
		}
	}
	if x == nil {
		return v.NotFound()
	}
	return x.ToVal()
}

func (v *Var) ToVar() *Var {
	return v
}

func (v *Var) ToBool() bool {
	if v.Vv == nil {
		v.Vv = NewBoolean(false)
	}
	return v.Vv.ToBool()
}

func (v *Var) ToDouble() float64 {
	if v.Vv == nil {
		v.Vv = NewNumber(0)
	}
	return v.Vv.ToDouble()
}

func (v *Var) ToLong() int64 {
	if v.Vv == nil {
		v.Vv = NewNumber(0)
	}
	return v.Vv.ToLong()
}

func (v *Var) ToUlong() uint {
	if v.Vv == nil {
		v.Vv = NewNumber(0)
	}
	return v.Vv.ToUlong()
}

func (v *Var) ToInt() int {
	if v.Vv == nil {
		v.Vv = NewNumber(0)
	}
	return v.Vv.ToInt()
}

func (v *Var) Append(y GrammarElement) GrammarElement {
	if v.Vv == nil || v.Vv == theNull() {
		v.Vv = NewLMBuffer()
	}
	return v.Vv.Append(y)
}

func (v *Var) Idxf(y GrammarElement) GrammarElement {
	return v.Vv.Idxf(y.ToVal())
}

func (v *Var) Idtf(y GrammarElement) GrammarElement {
	return v.Vv.Idtf(y.ToVal())
}

func (v *Var) StoValf(y GrammarElement) GrammarElement {
	v.Vv = y.ToVal()
	return v.Vv
}

func (v *Var) StoAddf(y GrammarElement) GrammarElement {
	v.Vv = v.Vv.Addf(y.ToVal())
	return v.Vv
}

func (v *Var) StoSubf(y GrammarElement) GrammarElement {
	v.Vv = v.Vv.Subf(y.ToVal())
	return v.Vv
}

func (v *Var) StoMulf(y GrammarElement) GrammarElement {
	v.Vv = v.Vv.Mulf(y.ToVal())
	return v.Vv
}

func (v *Var) StoDivf(y GrammarElement) GrammarElement {
	v.Vv = v.Vv.Divf(y.ToVal())
	return v.Vv
}

func (v *Var) StoModf(y GrammarElement) GrammarElement {
	v.Vv = v.Vv.Modf(y.ToVal())
	return v.Vv
}

func (v *Var) Preincf() GrammarElement {
	return v.StoAddf(NewNumber(1))
}

func (v *Var) Predecf() GrammarElement {
	return v.StoSubf(NewNumber(1))
}

func (v *Var) Postincf() GrammarElement {
	r := v.Vv.ToVal()
	v.Preincf()
	return r
}

func (v *Var) Postdecf() GrammarElement {
	r := v.Vv.ToVal()
	v.Predecf()
	return r
}

func (v *Var) Ru() *Rule {
	return v.VvC().Ru()
}

func (v *Var) Si() uint {
	return v.VvC().St().Si
}

func (v *Var) Gr() *Grammar {
	return v.VvC().St().Gr
}

func (v *Var) Gsy() GrammarElement {
	return v.VvC().St().Gr.Sy
}

func (v *Var) Rsy() GrammarElement {
	return v.VvC().St().rsy
}

func (v *Var) Lsy() GrammarElement {
	return v.VvC().St().lsy
}

func (v *Var) Ifn() string {
	return v.VvC().St().input.Filename()
}

func (v *Var) Cp() uint {
	return v.VvC().St().cp
}

func (v *Var) Ln() uint {
	return v.VvC().St().ln
}

func (v *Var) Cn() uint {
	return v.VvC().St().cn
}

type LMRef struct {
	Var
}

func NewLMRef() *LMRef {
	return &LMRef{}
}

func NewLMRefFromElement(k GrammarElement, q LMScope) *LMRef {
	lm := &LMRef{}
	lm.Vk = k
	lm.Vp = q.VvP() // Assuming Vvp() returns Var
	lm.Vx = q
	if lm.Vx == nil {
		panic("vx is null")
	}
	lm.Vv = lm.Var.Deref(lm.Vk)
	return lm
}

func (lm *LMRef) ToString() string {
	return lm.Key()
}

func (lm *LMRef) ToDebug() string {
	return "LMRef " + lm.Key() + ": " + lm.Value()
}

func (lm *LMRef) ToExplore() GrammarElement {
	TxV("R", " ", &lm.Var)
	if lm.Vv != nil {
		lm.Vv.ToExplore()
	}
	return lm
}

func (lm *LMRef) ToVal() GrammarElement {
	return lm.Var.ToVal()
}

func (lm *LMRef) ToRef() *Var {
	return &lm.Var
}

func (lm *LMRef) ToDeref(v *Var) *Var {
	return nil
}

func (lm *LMRef) Append(y GrammarElement) GrammarElement {
	return lm.Vv.Append(y)
}

func (lm *LMRef) Funf(y GrammarElement) GrammarElement {
	return nil
}

func (lm *LMRef) Inf(y GrammarElement) GrammarElement {
	return lm.Vv.Inf(y.ToVal())
}

func (lm *LMRef) Idxf(y GrammarElement) GrammarElement {
	return lm.Vv.Idxf(y.ToVal())
}

func (lm *LMRef) Idtf(y GrammarElement) GrammarElement {
	return lm.Vv.Idtf(y.ToVal())
}

func (lm *LMRef) StoValf(y GrammarElement) GrammarElement {
	return lm.Vv.StoValf(y.ToVal())
}

func (lm *LMRef) StoAddf(y GrammarElement) GrammarElement {
	return lm.Vv.StoAddf(y.ToVal())
}

func (lm *LMRef) StoSubf(y GrammarElement) GrammarElement {
	return lm.Vv.StoSubf(y.ToVal())
}

func (lm *LMRef) StoMulf(y GrammarElement) GrammarElement {
	return lm.Vv.StoMulf(y.ToVal())
}

func (lm *LMRef) StoDivf(y GrammarElement) GrammarElement {
	return lm.Vv.StoDivf(y.ToVal())
}

func (lm *LMRef) StoModf(y GrammarElement) GrammarElement {
	return lm.Vv.StoModf(y.ToVal())
}

type ARef struct {
	Var
	A *AArray
	K GrammarElement
}

func NewARef(x *AArray, y GrammarElement, z LMScope) *ARef {
	ar := &ARef{}
	ar.A = x
	ar.K = y
	ar.Vv = z.VvP()
	ar.Vx = z
	if ar.Vx == nil {
		panic("vx is null")
	}
	return ar
}

func (ar *ARef) Act(sr *Stream, s GenMode) GenMode {
	if _, ok := ar.A.A[ar.K]; ok {
		ar.Vv = ar.A.A[ar.K]
		return ar.Vv.Reference(sr, s, ar)
	}
	return s
}

func (ar *ARef) Key() string {
	if ar.K != nil {
		return ar.K.ToString()
	}
	return "---"
}

func (ar *ARef) Value() string {
	if _, ok := ar.A.A[ar.K]; ok {
		return ar.A.A[ar.K].ToString()
	}
	return "---"
}

func (ar *ARef) ToString() string {
	return "aref " + ar.Key() + ": " + ar.Value()
}

func (ar *ARef) ToRef() *ARef {
	return ar
}

func (ar *ARef) ToVal() GrammarElement {
	if val, ok := ar.A.A[ar.K]; ok {
		return val
	}
	//+ return NotFound()
	return nil
}

func (ar *ARef) ToDeref(v *Var) *Var {
	return nil
}

func (ar *ARef) Append(y GrammarElement) GrammarElement {
	return ar.StoValf(ar.ToVal().Append(y))
}

func (ar *ARef) Inf(y GrammarElement) GrammarElement {
	return ar.ToVal().Inf(y)
}

func (ar *ARef) Idxf(y GrammarElement) GrammarElement {
	return ar.ToVal().Idxf(y)
}

func (ar *ARef) Idtf(y GrammarElement) GrammarElement {
	return ar.ToVal().Idtf(y)
}

func (ar *ARef) StoValf(y GrammarElement) GrammarElement {
	ar.A.A[ar.K] = y
	return y
}

func (ar *ARef) StoAddf(y GrammarElement) GrammarElement {
	r := ar.ToVal().Addf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoSubf(y GrammarElement) GrammarElement {
	r := ar.ToVal().Subf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoMulf(y GrammarElement) GrammarElement {
	r := ar.ToVal().Mulf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoDivf(y GrammarElement) GrammarElement {
	r := ar.ToVal().Divf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoModf(y GrammarElement) GrammarElement {
	r := ar.ToVal().Modf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) Preincf() GrammarElement {
	return ar.StoAddf(NewNumber(1))
}

func (ar *ARef) Predecf() GrammarElement {
	return ar.StoSubf(NewNumber(1))
}

func (ar *ARef) Postincf() GrammarElement {
	r := ar.ToVal()
	ar.Preincf()
	return r
}

func (ar *ARef) Postdecf() GrammarElement {
	r := ar.ToVal()
	ar.Predecf()
	return r
}

type AArray struct {
	A map[GrammarElement]GrammarElement
}

func NewAArray() *AArray {
	return &AArray{A: make(map[GrammarElement]GrammarElement)}
}

type LMArray struct {
	Element
	aa *AArray
	sx LMScope
}

func NewLMArray(sr *Stream, s GenMode, z LMScope) *LMArray {
	la := LMArray{aa: NewAArray(), sx: s}

	if la.sx == nil {
		panic("sx cannot be nil")
	}

	var x *Opnd
	var v GrammarElement
	var i uint

	for x, i = sr.XS, 0; x != nil && x.V != sr.Ssy().Mark; x = x.S {
		if la.Assign(sr, x.V) == nil {
			i++
		}
	}
	for v = sr.Popx(); sr.XS != nil && v != sr.Ssy().Mark; v = sr.Popx() {
		if sr.XS.V == nil {
			i -= 1
			la.AssignE(sr, i, v)
		}
	}

	return &la
}

func (la *LMArray) ToVal() GrammarElement {
	return la
}

func (la *LMArray) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(la)
	return s
}

// func (la *LMArray) Assign(e *Stream, c *LMCell) GrammarElement {
func (la *LMArray) Assign(e *Stream, c GrammarElement) GrammarElement {
	if c != nil {
		lm, ok := c.(*LMCell)
		if !ok {
			panic("not lmcell")
			return nil
		}
		la.aa.A[e.Usy().UniqueE(lm.K)] = lm.V
		return lm.V
	}
	return nil
}

func (la *LMArray) AssignE(e *Stream, i uint, v GrammarElement) GrammarElement {
	la.aa.A[e.Usy().UniqueE(NewNumber(LMNumber(i)))] = v
	return v
}

func (la *LMArray) Idxf(y GrammarElement) GrammarElement {
	return NewARef(la.aa, la.sx.VvC().St().lm.Usy.UniqueE(y.ToVal()), la.sx)
}

func (la *LMArray) Idtf(y GrammarElement) GrammarElement {
	return NewARef(la.aa, la.sx.VvC().St().lm.Usy.UniqueE(y.ToVal()), la.sx)
}

type LMCell struct {
	Element
	K GrammarElement
	V GrammarElement
}

func NewLMCell(y, z GrammarElement) *LMCell {
	return &LMCell{K: y, V: z}
}

func (lc *LMCell) ToString() string {
	return "LMCell:" + lc.K.ToString() + lc.V.ToString()
}

type NewVar struct {
	Element
}

func NewNewVar() *NewVar {
	return &NewVar{}
}

func (nv *NewVar) ToString() string {
	return "newvar"
}

func (nv *NewVar) Act(sr *Stream, s GenMode) GenMode {
	v := sr.Popx().ToVal()
	k := sr.Popx()
	s.MakeVar(k, v, s, sr.AV)
	return s
}

type EachRef struct {
	Element
	K GrammarElement
}

func NewEachRef(x GrammarElement) *EachRef {
	return &EachRef{K: x}
}

func (er *EachRef) Token() GrammarElement {
	return nil
}

func (er *EachRef) ToBody() []GrammarElement {
	return nil
}

func (er *EachRef) ToLong() int64 {
	return 0
}

func (er *EachRef) Dump() {
	fmt.Printf("v:%s e", er.K)
}

func (er *EachRef) ToString() string {
	return "each " + er.K.ToString()
}

func (er *EachRef) Weight() uint {
	return 0
}

func (er *EachRef) Act(sr *Stream, s GenMode) GenMode {
	return sr.EachRef(s, er.K, s)
}

func (er *EachRef) Match(e *Engine, r GrammarElement) bool {
	return false
}

type AllRef struct {
	Element
	K GrammarElement
}

func NewAllRef(x GrammarElement) *AllRef {
	return &AllRef{K: x}
}

func (ar *AllRef) Token() GrammarElement {
	return nil
}

func (ar *AllRef) ToBody() []GrammarElement {
	return nil
}

func (ar *AllRef) ToLong() int64 {
	return 0
}

func (ar *AllRef) Dump() {
	fmt.Printf("v:%s A", ar.K)
}

func (ar *AllRef) ToString() string {
	return "all " + ar.K.ToString()
}

func (ar *AllRef) Weight() uint {
	return 0
}

func (ar *AllRef) Act(sr *Stream, s GenMode) GenMode {
	return sr.AllRef(s, ar.K, s)
}

func (ar *AllRef) Match(e *Engine, r GrammarElement) bool {
	return false
}

type VarSym struct {
	Symbol
}

func NewVarSym(x string) *VarSym {
	return &VarSym{Symbol{V: x}}
}

func (vs *VarSym) Token() GrammarElement {
	return vs
}

func (vs *VarSym) ToDump() string {
	return "v:" + vs.V
}

func (vs *VarSym) Dump() {
	fmt.Printf("v:%s ", vs.V)
}

func (vs *VarSym) Act(sr *Stream, s GenMode) GenMode {
	return vs.Reference(sr, s, s)
}

func (vs *VarSym) Match(e *Engine, r GrammarElement) bool {
	return false
}

func (vs *VarSym) Reference(sr *Stream, s GenMode, x LMScope) GenMode {
	return sr.TheRef(s, vs, x)
}

func (vs *VarSym) ToDeref(x *Var) *Var {
	TxE("var: ", vs)
	return x.Deref(vs)
}

type DoneF struct {
	Symbol
}

func NewDoneF(x string) *DoneF {
	return &DoneF{Symbol{V: x}}
}

func (df *DoneF) Match(e *Engine, r GrammarElement) bool {
	e.Lhr.AV = e.Lhx.Cp()
	return e.Matched3E(df, nil, nil)
}

type TakeF struct {
	Symbol
}

func NewTakeF(x string) *TakeF {
	return &TakeF{Symbol{V: x}}
}

func (tf *TakeF) Dump() {
	fmt.Printf("t ")
}

func (tf *TakeF) Match(e *Engine, r GrammarElement) bool {
	if r.Token() == tf { // %  %
		e.TakeTvar()
		return e.Matched3E(tf, r, nil)
	}
	if _, ok := r.(*BindF); ok { // %  :
		e.PushX()
		return e.Matched3E(tf, r, nil)
	}
	if e.Rsy != nil { // %  [matched]
		e.PushR(e.Rsy)
		return e.Matched3E(tf, nil, nil)
	}
	return false
}

type BindF struct {
	Symbol
}

func NewBindF(x string) *BindF {
	return &BindF{Symbol{V: x}}
}

func (b *BindF) Dump() {
	fmt.Print("p ")
}

func (b *BindF) Match(e *Engine, r GrammarElement) bool {
	if r.Token() == b {
		a := e.Lhr.Popx()
		bElem := e.Rhr.Popx().ToVal()
		// tx("l", a); tx("r", bElem);
		if a, ok := a.(*VarSym); ok {
			e.BindUvar(a, bElem)
			return e.Matched3E(b, r, nil)
		}
		e.Matched3E(b, r, nil)
		lh := []GrammarElement{a}
		e.Lhr.SM = NewSTModeFromElements(e.Lhr.SM, lh, e.Lhr.SM)
		rh := []GrammarElement{bElem}
		e.Rhr.SM = NewSTModeFromElements(e.Rhr.SM, rh, e.Rhr.SM)
		return true
	}
	if _, ok := r.(*TakeF); ok {
		e.BindTvar()
		return e.Matched3E(b, r, nil)
	}
	if e.Rsy != nil {
		e.BindXvarE(e.Rsy)
		return e.Matched3E(b, nil, r)
	}
	return false
}

type AppendSym struct {
	Symbol
}

func NewAppendSym(x string) *AppendSym {
	return &AppendSym{Symbol{V: x}}
}

func (a *AppendSym) Match(e *Engine, r GrammarElement) bool {
	e.Lhr.Popx().Append(r)
	return e.Matched3E(a, r, r)
}

type AppendXSym struct {
	Symbol
}

func NewAppendXSym(x string) *AppendXSym {
	return &AppendXSym{Symbol{V: x}}
}

// if there is captured material, append it
// otherwise match one symbol and append that
func (a *AppendXSym) Match(e *Engine, r GrammarElement) bool {
	b := e.Lhr.Popx()
	if e.Lhr.XS != nil {
		v := e.Lhr.ToRow()
		for _, x := range v {
			b.Append(x)
		}
		e.Matched3E(a, nil, nil)
	} else {
		b.Append(r)
		e.Matched3E(a, r, r)
	}
	return true
}

type ErrSym struct {
	Symbol
}

func NewErrSym(x string) *ErrSym {
	return &ErrSym{Symbol{V: x}}
}

func (e *ErrSym) Append(y GrammarElement) GrammarElement {
	fmt.Fprintf(os.Stderr, "%s", y)
	return e
}

func (e *ErrSym) Match(engine *Engine, r GrammarElement) bool {
	fmt.Fprintf(os.Stderr, "%s", r)
	return engine.Matched3E(e, r, r)
}

type OutSym struct {
	Symbol
}

func NewOutSym(x string) *OutSym {
	return &OutSym{Symbol{V: x}}
}

func (o *OutSym) Append(y GrammarElement) GrammarElement {
	fmt.Fprintf(os.Stdout, "%s", y)
	return o
}

func (o *OutSym) Match(engine *Engine, r GrammarElement) bool {
	fmt.Fprintf(os.Stdout, "%s", r)
	return engine.Matched3E(o, r, r)
}

type UriSym struct {
	Symbol
}

func NewUriSym(x string) *UriSym {
	return &UriSym{Symbol{V: x}}
}

func (u *UriSym) Append(y GrammarElement) GrammarElement {
	fmt.Fprintf(os.Stdout, "%s", y.ToEncode())
	return u
}

func (u *UriSym) Match(engine *Engine, r GrammarElement) bool {
	fmt.Fprintf(os.Stdout, "%s", r.ToEncode())
	return engine.Matched3E(u, r, r)
}

type UrdSym struct {
	Symbol
}

func NewUrdSym(x string) *UrdSym {
	return &UrdSym{Symbol{V: x}}
}

func (u *UrdSym) Append(y GrammarElement) GrammarElement {
	fmt.Fprintf(os.Stdout, "%s", y.ToDecode())
	return u
}

func (u *UrdSym) Match(engine *Engine, r GrammarElement) bool {
	fmt.Fprintf(os.Stdout, "%s", r.ToDecode())
	return engine.Matched3E(u, r, r)
}

type SpSym struct {
	Symbol
}

func NewSpSym(x string) *SpSym {
	return &SpSym{Symbol{V: x}}
}

func (s *SpSym) ToEncode() string {
	return " "
}

type NlSym struct {
	Symbol
}

func NewNlSym(x string) *NlSym {
	return &NlSym{Symbol{V: x}}
}

func (n *NlSym) ToEncode() string {
	return "\n"
}

type GetF struct {
	Symbol
}

func NewGetF(x string) *GetF {
	return &GetF{Symbol{V: x}}
}

func (g *GetF) Dump() {
	fmt.Print("g ")
}

func (g *GetF) Act(sr *Stream, s GenMode) GenMode {
	return sr.Getx(s)
}

type TrueSym struct {
	Symbol
}

func NewTrueSym(x string) *TrueSym {
	return &TrueSym{Symbol{V: x}}
}

func (t *TrueSym) ToVal() GrammarElement {
	return NewBoolean(true)
}

type FalseSym struct {
	Symbol
}

func NewFalseSym(x string) *FalseSym {
	return &FalseSym{Symbol{V: x}}
}

func (f *FalseSym) ToVal() GrammarElement {
	return NewBoolean(false)
}

type TrueF struct {
	Symbol
}

func NewTrueF(x string) *TrueF {
	return &TrueF{Symbol{V: x}}
}

func (t *TrueF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewBoolean(true))
	return s
}

type FalseF struct {
	Symbol
}

func NewFalseF(x string) *FalseF {
	return &FalseF{Symbol{V: x}}
}

func (f *FalseF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewBoolean(false))
	return s
}

type GetXF struct {
	Symbol
	V GrammarElement
}

func NewGetXF(x GrammarElement) *GetXF {
	return &GetXF{V: x}
}

func (g *GetXF) ToTrace() string {
	return g.V.ToTrace() + "p"
}

func (g *GetXF) Dump() {
	g.V.Dump()
}

func (g *GetXF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(g.V)
	return s
}

type GetBF struct {
	Symbol
	V GrammarElement
}

func NewGetBF(x GrammarElement) *GetBF {
	return &GetBF{V: x}
}

func (g *GetBF) ToTrace() string {
	return g.V.ToTrace()
}

func (g *GetBF) Dump() {
	g.V.Dump()
}

func (g *GetBF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(g.V)
	sr.SY = sr.Ssy().BindFn
	return s
}

type GetVF struct {
	Symbol
	V GrammarElement
}

func NewGetVF(x GrammarElement) *GetVF {
	return &GetVF{V: x}
}

func (g *GetVF) ToTrace() string {
	return g.V.ToTrace()
}

func (g *GetVF) Dump() {
	g.V.Dump()
}

func (g *GetVF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewLMRefFromElement(g.V, s))
	return s
}

type ActF struct {
	Symbol
}

func NewActF(x string) *ActF {
	return &ActF{Symbol{V: x}}
}

func (a *ActF) Dump() {
	fmt.Print("a ")
}

func (a *ActF) Trace(sr *Stream, t *Tracer) {
	t.TraceAct(sr, a)
}

func (a *ActF) Act(sr *Stream, s GenMode) GenMode {
	panic("assertion failed")
	//+ return nil
}

type Primitive struct {
	Symbol
}

func NewPrimitive() *Primitive {
	return &Primitive{}
}

func NewPrimitiveFromString(x string) *Primitive {
	return &Primitive{Symbol{V: x}}
}

func (p *Primitive) Dump() {
	fmt.Printf("f:%s ", string(p.V))
}

func (p *Primitive) Act(sr *Stream, s GenMode) GenMode {
	fmt.Printf("act: %s\n", string(p.V))
	return s
}

type ApplyF struct {
	Primitive
}

func NewApplyF(x string) *ApplyF {
	return &ApplyF{Primitive: *NewPrimitiveFromString(x)}
}

func (a *ApplyF) Trace(s *Stream, t *Tracer) {
	t.TraceApply(s, a)
}

func (a *ApplyF) Act(sr *Stream, s GenMode) GenMode {
	v := sr.Popx()
	return v.Act(sr, s)
}

type InjF struct {
	Symbol
}

func NewInjF(x string) *InjF {
	return &InjF{Symbol{V: x}}
}

func (i *InjF) Dump() {
	fmt.Printf("f:%s ", string(i.V))
}

func (i *InjF) Match(e *Engine, r GrammarElement) bool {
	e.PushRhx(e.Lhr.Popx())
	return e.Matched3E(i, nil, nil)
}

type StrF struct {
	Symbol
}

func NewStrF(x string) *StrF {
	return &StrF{Symbol{V: x}}
}

func (s *StrF) Dump() {
	fmt.Print("s ")
}

func (s *StrF) Act(sr *Stream, mode GenMode) GenMode {
	//  { return new stMode(s, (cast(str)sr.popx()).v, s); }
	return mode
}

type Anything struct {
	Symbol
}

func NewAnything(x string) *Anything {
	return &Anything{Symbol{V: x}}
}

func (a *Anything) Match(e *Engine, r GrammarElement) bool {
	e.Matched3E(a, r, r)
	return true
}

type AnySym struct {
	Anything
}

func NewAnySym(x string) *AnySym {
	return &AnySym{Anything: *NewAnything(x)}
}

func (a *AnySym) Match(e *Engine, r GrammarElement) bool {
	if _, ok := r.Token().(*Sym); ok {
		e.Matched3E(a, r, r)
		return true
	} else if _, ok := r.Token().(*Usr); ok {
		e.Matched3E(a, r, r)
		return true
	} else {
		return e.ResolveE(a, r)
	}
}

type AnyChr struct {
	Anything
}

func NewAnyChr(x string) *AnyChr {
	return &AnyChr{Anything: *NewAnything(x)}
}

func (a *AnyChr) Match(e *Engine, r GrammarElement) bool {
	if _, ok := r.Token().(*Chr); ok {
		e.Matched3E(a, r, r)
		return true
	} else {
		return e.ResolveE(a, r)
	}
}

type AnyNum struct {
	Anything
}

func NewAnyNum(x string) *AnyNum {
	return &AnyNum{Anything: *NewAnything(x)}
}

func (a *AnyNum) Match(e *Engine, r GrammarElement) bool {
	if _, ok := r.Token().(*Number); ok {
		e.Matched3E(a, r, r)
		return true
	} else {
		return e.ResolveE(a, r)
	}
}

type LnoSym struct {
	Symbol
}

func NewLnoSym(x string) *LnoSym {
	return &LnoSym{Symbol: *NewSymbol(x)}
}

func (l *LnoSym) Match(e *Engine, r GrammarElement) bool {
	return e.Matched3E(l, nil, NewNumber(LMNumber(e.Lineno())))
}

type IfnSym struct {
	Symbol
}

func NewIfnSym(x string) *IfnSym {
	return &IfnSym{Symbol: *NewSymbol(x)}
}

func (i *IfnSym) Match(e *Engine, r GrammarElement) bool {
	return e.Matched3E(i, nil, NewSym(e.Filename()))
}

type FlagSym struct {
	Symbol
}

func NewFlagSym(x string) *FlagSym {
	return &FlagSym{Symbol: *NewSymbol(x)}
}

func (f *FlagSym) Match(e *Engine, r GrammarElement) bool {
	e.FlagErrors++
	return e.Matched3E(f, nil, NewSym(e.Filename()+":"+string(e.Lineno())+": "))
}

type WarnSym struct {
	Symbol
}

func NewWarnSym(x string) *WarnSym {
	return &WarnSym{Symbol: *NewSymbol(x)}
}

func (w *WarnSym) Match(e *Engine, r GrammarElement) bool {
	e.WarnErrors++
	return e.Matched3E(w, nil, NewSym(e.Filename()+":"+string(e.Lineno())+": "))
}

type RepnSym struct {
	Symbol
}

func NewRepnSym(x string) *RepnSym {
	return &RepnSym{Symbol: *NewSymbol(x)}
}

func (r *RepnSym) Match(e *Engine, _ GrammarElement) bool {
	n := e.Lhr.Popx().ToVal().(*Number)
	return e.Repeat(uint(n.ToLong()))
}

type RepSym struct {
	Symbol
}

func NewRepSym(x string) *RepSym {
	return &RepSym{Symbol: *NewSymbol(x)}
}

func (r *RepSym) Match(e *Engine, _ GrammarElement) bool {
	return e.Repeat(0)
}

type OptSym struct {
	Symbol
}

func NewOptSym(x string) *OptSym {
	return &OptSym{Symbol: *NewSymbol(x)}
}

func (o *OptSym) Match(e *Engine, r GrammarElement) bool {
	return e.Repeat(1)
}

type OptxSym struct {
	Symbol
}

func NewOptxSym(x string) *OptxSym {
	return &OptxSym{Symbol: *NewSymbol(x)}
}

func (o *OptxSym) Match(e *Engine, r GrammarElement) bool {
	return e.Repeatx(1)
}

type RepxSym struct {
	Symbol
}

func NewRepxSym(x string) *RepxSym {
	return &RepxSym{Symbol: *NewSymbol(x)}
}

func (r *RepxSym) Match(e *Engine, _ GrammarElement) bool {
	return e.Repeatx(0)
}

type Lex struct {
	Symbol
	Table     map[GrammarElement]GrammarElement
	Inclusive bool
}

func NewLex(x string) *Lex {
	return &Lex{
		Symbol:    *NewSymbol(x),
		Table:     make(map[GrammarElement]GrammarElement),
		Inclusive: true,
	}
}

func NewLexFromEngine(s string, e *Engine) *Lex {
	l := NewLex(s)

	state := IN
	var prevc int

	//foreach(char c; s[1..(s.length - 1)]) {
	//for i, n := 1, len(s) - 1; i < n; {
	for c := range s[1:] {
		var x GrammarElement

		switch state {
		case IN:
			if c == '^' {
				l.Inclusive = false
				state = C1
			}

		case C1:
			if c == '\\' {
				state = E1
			} else {
				x = e.Tsy.UniqueR(rune(c))
				l.Table[x] = x
				prevc = c
				state = C2
			}

		case E1:
			switch c {
			case 'b':
				c = ' '
			case 'n':
				c = '\n'
			case 'r':
				c = '\r'
			case 't':
				c = '\t'
			case 'f':
				c = '\f'
			}
			x = e.Tsy.UniqueR(rune(c))
			l.Table[x] = x
			prevc = c
			state = C1

		case C2:
			if c == '\\' {
				state = E1
			} else if c == '-' {
				state = RN
			} else {
				x = e.Tsy.UniqueR(rune(c))
				l.Table[x] = x
				prevc = c
			}

		case RN:
			for prevc < c {
				x = e.Tsy.UniqueR(rune(c))
				l.Table[x] = x
				c--
			}
			state = C1
		}
	}
	return l
}

func (l *Lex) ToTrace() string {
	return "[" + l.V[1:len(l.V)-1] + "]"
}

func (l *Lex) ToString1() string {
	return "lex(" + l.V + ")"
}

func (l *Lex) Dump() {
	fmt.Printf("l:%s ", l.V)
}

func (l *Lex) AddRule(g *Grammar, x *Rule) {
	if l.Inclusive {
		for k := range l.Table {
			g.Add(x.Additional(k))
		}
	} else {
		panic("BadLexicalRule")
	}
}

func (l *Lex) Match(e *Engine, r GrammarElement) bool {
	if r.Token() == l {
		return e.Matched3E(l, r, nil)
	}
	if chr, ok := r.(*Chr); ok && (l.Inclusive != (l.Table[chr] == nil)) {
		return e.Matched3E(l, r, r)
	}
	return e.ResolveE(l, r)
}

type DropF struct {
	Primitive
}

func NewDropF(x string) *DropF {
	return &DropF{Primitive: *NewPrimitiveFromString(x)}
}

func (d *DropF) Act(sr *Stream, s GenMode) GenMode {
	sr.XS = nil
	return s
}

type Unary struct {
	Primitive
}

func NewUnary(x string) *Unary {
	return &Unary{Primitive: *NewPrimitiveFromString(x)}
}

func (u *Unary) Result(x GrammarElement) GrammarElement {
	return nil
}

func (u *Unary) Trace(s *Stream, t *Tracer) {
	t.TraceArithmetic(s, u)
}

func (u *Unary) Act(sr *Stream, s GenMode) GenMode {
	x := sr.Popx()
	sr.Pushx(u.Result(x.ToVal()))
	return s
}

type Arithmetic struct {
	Primitive
}

func NewArithmetic(x string) *Arithmetic {
	return &Arithmetic{Primitive: *NewPrimitiveFromString(x)}
}

func (a *Arithmetic) Result(x, y GrammarElement) GrammarElement {
	return nil
}

func (a *Arithmetic) Trace(s *Stream, t *Tracer) {
	t.TraceArithmetic(s, a)
}

func (a *Arithmetic) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(a.Result(x.ToVal(), y.ToVal()))
	return b
}

type Relation struct {
	Primitive
}

func NewRelation(x string) *Relation {
	return &Relation{Primitive: *NewPrimitiveFromString(x)}
}

func (r *Relation) Result(x, y GrammarElement) GrammarElement {
	return nil
}

func (r *Relation) Trace(s *Stream, t *Tracer) {
	t.TraceRelation(s, r)
}

func (r *Relation) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(r.Result(x.ToVal(), y.ToVal()))
	return b
}

type Assignment struct {
	Primitive
}

func NewAssignment(x string) *Assignment {
	return &Assignment{Primitive: *NewPrimitiveFromString(x)}
}

func (a *Assignment) Result(x, y GrammarElement) GrammarElement {
	return nil
}

func (a *Assignment) Trace(s *Stream, t *Tracer) {
	t.TraceAssignment(s, a)
}

func (a *Assignment) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(a.Result(x, y.ToVal()))
	return b
}

type IncDec struct {
	Primitive
}

func NewIncDec(x string) *IncDec {
	return &IncDec{Primitive: *NewPrimitiveFromString(x)}
}

func (i *IncDec) Result(x GrammarElement) GrammarElement {
	return nil
}

func (i *IncDec) Trace(s *Stream, t *Tracer) {
	t.TraceAssignment(s, i)
}

func (i *IncDec) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(i.Result(sr.Popx()))
	return b
}

type Iff struct {
	Primitive
}

func NewIff(x string) *Iff {
	return &Iff{Primitive: *NewPrimitiveFromString(x)}
}

func (i *Iff) Act(sr *Stream, b GenMode) GenMode {
	// sr.Dumpx()
	y := sr.Popx()
	x := sr.Popx()
	t := sr.Popx().ToVal()
	if t.ToBool() {
		s := x.ToVal().(*Str)
		return NewSTModeFromElements(b, s.V, b)
	} else {
		s := y.ToVal().(*Str)
		return NewSTModeFromElements(b, s.V, b)
	}
}

type OrOrf struct {
	Primitive
}

func NewOrOrf(x string) *OrOrf {
	return &OrOrf{Primitive: *NewPrimitiveFromString(x)}
}

func (o *OrOrf) Act(sr *Stream, b GenMode) GenMode {
	// sr.Dumpx()
	x := sr.Popx()
	t := sr.Popx().ToVal()
	if !t.ToBool() {
		s := x.ToVal().(*Str)
		return NewSTModeFromElements(b, s.V, b)
	}
	sr.Pushx(NewBoolean(true))
	return b
}

type AndAndf struct {
	Primitive
}

func NewAndAndf(x string) *AndAndf {
	return &AndAndf{Primitive: *NewPrimitiveFromString(x)}
}

func (a *AndAndf) Act(sr *Stream, b GenMode) GenMode {
	// sr.Dumpx()
	x := sr.Popx()
	t := sr.Popx().ToVal()
	if t.ToBool() {
		s := x.ToVal().(*Str)
		return NewSTModeFromElements(b, s.V, b)
	}
	sr.Pushx(NewBoolean(false))
	return b
}

type Index struct {
	Primitive
}

func NewIndex(x string) *Index {
	return &Index{Primitive: *NewPrimitiveFromString(x)}
}

func (i *Index) Result(x, y GrammarElement) GrammarElement {
	return nil
}

func (i *Index) Trace(s *Stream, t *Tracer) {
	t.TraceIndex(s, i)
}

func (i *Index) Act(sr *Stream, b GenMode) GenMode {
	// sr.Dumpx()
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(i.Result(x, y.ToVal()))
	return b
}

type Cellf struct {
	Primitive
}

func NewCellf(x string) *Cellf {
	return &Cellf{Primitive: *NewPrimitiveFromString(x)}
}

func (c *Cellf) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(NewLMCell(x, y))
	return b
}

type Arrayf struct {
	Primitive
}

func NewArrayf(x string) *Arrayf {
	return &Arrayf{Primitive: *NewPrimitiveFromString(x)}
}

func (a *Arrayf) Trace(s *Stream, t *Tracer) {
	t.TraceIndex(s, a)
}

func (a *Arrayf) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(NewLMArray(sr, b, b))
	return b
}

type Argsf struct {
	Primitive
}

func NewArgsf(x string) *Argsf {
	return &Argsf{Primitive: *NewPrimitiveFromString(x)}
}

func (a *Argsf) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(sr.Ssy().Mark)
	return b
}

type Funf struct {
	Primitive
}

func NewFunf(x string) *Funf {
	return &Funf{Primitive: *NewPrimitiveFromString(x)}
}

func (f *Funf) Act(sr *Stream, b GenMode) GenMode {
	v := sr.ToArgv(sr.Ssy().Mark)
	sr.Pushx(sr.System().Call(sr, b, v[0], v))
	return b
}

type LmnFuncf struct {
	F LMNFunc
}

func NewLmnFuncf(x LMNFunc) *LmnFuncf {
	return &LmnFuncf{F: x}
}

func (l *LmnFuncf) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(l.F(sr))
	return b
}

type Loopf struct {
	Primitive
}

func NewLoopf(x string) *Loopf {
	return &Loopf{Primitive: *NewPrimitiveFromString(x)}
}

func (l *Loopf) Trace(s *Stream, t *Tracer) {
	t.TraceLoop(s, l)
}

func (l *Loopf) Act(sr *Stream, b GenMode) GenMode {
	x := sr.Popx()
	s := x.ToVal().(*Str)
	return NewRPModeFromElement(b, s.V)
}

type Testf struct {
	Primitive
}

func NewTestf(x string) *Testf {
	return &Testf{Primitive: *NewPrimitiveFromString(x)}
}

func (t *Testf) Act(sr *Stream, b GenMode) GenMode {
	// sr.Dumpx()
	te := sr.Popx().ToVal()
	// tx("test", tElement)
	if !te.ToBool() {
		return b.Ends()
	}
	return b
}

type Self struct {
	Primitive
}

func NewSelf(x string) *Self {
	return &Self{Primitive: *NewPrimitiveFromString(x)}
}

func (s *Self) Act(sr *Stream, b GenMode) GenMode {
	// sr.Dumpx();
	sr.Popx()
	x := sr.Popx()
	t := sr.Popx().ToVal()

	if t.ToBool() {
		s := x.ToVal().(*Str)
		return NewSTModeFromElements(b, s.V, b)
	} else {
		s := x.ToVal().(*Str)
		return NewSTModeFromElements(b, s.V, b)
	}
	return b
}

type Foreachf struct {
	Primitive
}

func NewForeachf(x string) *Foreachf {
	return &Foreachf{Primitive: *NewPrimitiveFromString(x)}
}

type Retf struct {
	Primitive
}

func NewRetf(x string) *Retf {
	return &Retf{Primitive: *NewPrimitiveFromString(x)}
}

type Lamdaf struct {
	Primitive
}

func NewLamdaf(x string) *Lamdaf {
	return &Lamdaf{Primitive: *NewPrimitiveFromString(x)}
}

type Specf struct {
	Primitive
}

func NewSpecf(x string) *Specf {
	return &Specf{Primitive: *NewPrimitiveFromString(x)}
}

type Idxf struct {
	Index
}

func NewIdxf(x string) *Idxf {
	return &Idxf{Index: *NewIndex(x)}
}
func (i Idxf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Idxf(y)
}

type Idtf struct {
	Index
}

func NewIdtf(x string) *Idtf {
	return &Idtf{Index: *NewIndex(x)}
}
func (i Idtf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Idtf(y)
}

type StoValf struct {
	Assignment
}

func NewStoValf(x string) *StoValf {
	return &StoValf{Assignment: *NewAssignment(x)}
}
func (s StoValf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.StoValf(y)
}

type StoAddf struct {
	Assignment
}

func NewStoAddf(x string) *StoAddf {
	return &StoAddf{Assignment: *NewAssignment(x)}
}
func (s StoAddf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.StoAddf(y)
}

type StoSubf struct {
	Assignment
}

func NewStoSubf(x string) *StoSubf {
	return &StoSubf{Assignment: *NewAssignment(x)}
}
func (s StoSubf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.StoSubf(y)
}

type StoMulf struct {
	Assignment
}

func NewStoMulf(x string) *StoMulf {
	return &StoMulf{Assignment: *NewAssignment(x)}
}
func (s StoMulf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.StoMulf(y)
}

type StoDivf struct {
	Assignment
}

func NewStoDivf(x string) *StoDivf {
	return &StoDivf{Assignment: *NewAssignment(x)}
}
func (s StoDivf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.StoDivf(y)
}

type StoModf struct {
	Assignment
}

func NewStoModf(x string) *StoModf {
	return &StoModf{Assignment: *NewAssignment(x)}
}
func (s StoModf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.StoModf(y)
}

type Eeqf struct {
	Relation
}

func NewEeqf(x string) *Eeqf {
	return &Eeqf{Relation: *NewRelation(x)}
}
func (e Eeqf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Eeqf(y)
}

type Neef struct {
	Relation
}

func NewNeef(x string) *Neef {
	return &Neef{Relation: *NewRelation(x)}
}
func (n Neef) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Neef(y)
}

type Inf struct {
	Relation
}

func NewInf(x string) *Inf {
	return &Inf{Relation: *NewRelation(x)}
}
func (i Inf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Inf(y)
}

type Eqf struct {
	Relation
}

func NewEqf(x string) *Eqf {
	return &Eqf{Relation: *NewRelation(x)}
}
func (e Eqf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Eqf(y)
}

type Nef struct {
	Relation
}

func NewNef(x string) *Nef {
	return &Nef{Relation: *NewRelation(x)}
}
func (n Nef) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Nef(y)
}

type Ltf struct {
	Relation
}

func NewLtf(x string) *Ltf {
	return &Ltf{Relation: *NewRelation(x)}
}
func (l Ltf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Ltf(y)
}

type Gtf struct {
	Relation
}

func NewGtf(x string) *Gtf {
	return &Gtf{Relation: *NewRelation(x)}
}
func (g Gtf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Gtf(y)
}

type Lef struct {
	Relation
}

func NewLef(x string) *Lef {
	return &Lef{Relation: *NewRelation(x)}
}
func (l Lef) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Lef(y)
}

type Gef struct {
	Relation
}

func NewGef(x string) *Gef {
	return &Gef{Relation: *NewRelation(x)}
}
func (g Gef) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Gef(y)
}

type BitXorf struct {
	Arithmetic
}

func NewBitXorf(x string) *BitXorf {
	return &BitXorf{Arithmetic: *NewArithmetic(x)}
}
func (b BitXorf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.BitXorf(y)
}

type BitOrf struct {
	Arithmetic
}

func NewBitOrf(x string) *BitOrf {
	return &BitOrf{Arithmetic: *NewArithmetic(x)}
}
func (b BitOrf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.BitOrf(y)
}

type BitAndf struct {
	Arithmetic
}

func NewBitAndf(x string) *BitAndf {
	return &BitAndf{Arithmetic: *NewArithmetic(x)}
}
func (b BitAndf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.BitAndf(y)
}

type Addf struct {
	Arithmetic
}

func NewAddf(x string) *Addf {
	return &Addf{Arithmetic: *NewArithmetic(x)}
}
func (a Addf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Addf(y)
}

type Subf struct {
	Arithmetic
}

func NewSubf(x string) *Subf {
	return &Subf{Arithmetic: *NewArithmetic(x)}
}
func (s Subf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Subf(y)
}

type Mulf struct {
	Arithmetic
}

func NewMulf(x string) *Mulf {
	return &Mulf{Arithmetic: *NewArithmetic(x)}
}
func (m Mulf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Mulf(y)
}

type Divf struct {
	Arithmetic
}

func NewDivf(x string) *Divf {
	return &Divf{Arithmetic: *NewArithmetic(x)}
}
func (d Divf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Divf(y)
}

type Modf struct {
	Arithmetic
}

func NewModf(x string) *Modf {
	return &Modf{Arithmetic: *NewArithmetic(x)}
}
func (m Modf) Result(x GrammarElement, y GrammarElement) GrammarElement {
	return x.Modf(y)
}

type Preincf struct {
	IncDec
}

func NewPreincf(x string) *Preincf {
	return &Preincf{IncDec: *NewIncDec(x)}
}
func (p Preincf) Result(x GrammarElement) GrammarElement {
	return x.Preincf()
}

type Predecf struct {
	IncDec
}

func NewPredecf(x string) *Predecf {
	return &Predecf{IncDec: *NewIncDec(x)}
}
func (p Predecf) Result(x GrammarElement) GrammarElement {
	return x.Predecf()
}

type Postincf struct {
	IncDec
}

func NewPostincf(x string) *Postincf {
	return &Postincf{IncDec: *NewIncDec(x)}
}
func (p Postincf) Result(x GrammarElement) GrammarElement {
	return x.Postincf()
}

type Postdecf struct {
	IncDec
}

func NewPostdecf(x string) *Postdecf {
	return &Postdecf{IncDec: *NewIncDec(x)}
}
func (p Postdecf) Result(x GrammarElement) GrammarElement {
	return x.Postdecf()
}

type Negf struct {
	Unary
}

func NewNegf(x string) *Negf {
	return &Negf{Unary: *NewUnary(x)}
}
func (n Negf) Result(x GrammarElement) GrammarElement {
	return x.Negf()
}

type Notf struct {
	Unary
}

func NewNotf(x string) *Notf {
	return &Notf{Unary: *NewUnary(x)}
}
func (n Notf) Result(x GrammarElement) GrammarElement {
	return x.Notf()
}

type Invf struct {
	Unary
}

func NewInvf(x string) *Invf {
	return &Invf{Unary: *NewUnary(x)}
}
func (i Invf) Result(x GrammarElement) GrammarElement {
	return x.Invf()
}

type IOSymbol struct {
	Symbol
	H GrammarSystem
}

func NewIOSymbol(x string, handler GrammarSystem) *IOSymbol {
	iosymbol := &IOSymbol{Symbol: *NewSymbol(x), H: handler}
	handler.SetSymbol(iosymbol)
	return iosymbol
}

func (i *IOSymbol) SetHandler(handler GrammarStdio) {
	i.H = handler
}

func (i *IOSymbol) Match(e *Engine, r GrammarElement) bool {
	return i.H.Match(e, i, r)
}
