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

type MachineElement interface {
	AddRule(*Grammar, *Rule)
	Match(*Engine, MachineElement) bool
	NewLHS(GenMode) GenMode
	NewRHX(GenMode, EngineStateContext, LMScope) GenMode
	Act(*Stream, GenMode) GenMode
	Compare(*Engine, MachineElement) bool
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
	ToBody() []MachineElement
	Weight() uint
	Token() MachineElement
	Priority(uint) uint
	Reference(*Stream, GenMode, LMScope) GenMode
	ToExplore() MachineElement
	InvalidOp(string) MachineElement
	NotFound() MachineElement
	ToVal() MachineElement
	ToDeref(*Var) *Var
	Append(MachineElement) MachineElement
	Inf(MachineElement) MachineElement
	Idxf(y MachineElement) MachineElement
	Idtf(y MachineElement) MachineElement
	StoValf(y MachineElement) MachineElement
	StoAddf(y MachineElement) MachineElement
	StoSubf(y MachineElement) MachineElement
	StoMulf(y MachineElement) MachineElement
	StoDivf(y MachineElement) MachineElement
	StoModf(y MachineElement) MachineElement
	StoAndf(y MachineElement) MachineElement
	StoOrf(y MachineElement) MachineElement
	StoXorf(y MachineElement) MachineElement
	StoShlf(y MachineElement) MachineElement
	StoShrf(y MachineElement) MachineElement
	Eeqf(y MachineElement) MachineElement
	Neef(y MachineElement) MachineElement
	Eqf(y MachineElement) MachineElement
	Nef(y MachineElement) MachineElement
	Ltf(y MachineElement) MachineElement
	Gtf(y MachineElement) MachineElement
	Lef(y MachineElement) MachineElement
	Gef(y MachineElement) MachineElement
	BitXorf(y MachineElement) MachineElement
	BitOrf(y MachineElement) MachineElement
	BitAndf(y MachineElement) MachineElement
	OrOrf(y MachineElement) MachineElement
	AndAndf(y MachineElement) MachineElement
	Addf(y MachineElement) MachineElement
	Subf(y MachineElement) MachineElement
	Mulf(y MachineElement) MachineElement
	Divf(y MachineElement) MachineElement
	Modf(y MachineElement) MachineElement
	Preincf() MachineElement
	Predecf() MachineElement
	Postincf() MachineElement
	Postdecf() MachineElement
	Posf() MachineElement
	Negf() MachineElement
	Notf() MachineElement
	Invf() MachineElement
	OpAdd(MachineElement) MachineElement
	OpShl(MachineElement) MachineElement
	OpShr(MachineElement) MachineElement
	OpUShr(MachineElement) MachineElement
	OpCat(MachineElement) MachineElement
	OpEquals(MachineElement) bool
	OpCmp(MachineElement) int
	OpAndAssign(MachineElement) MachineElement
	OpOrAssign(MachineElement) MachineElement
	OpXorAssign(MachineElement) MachineElement
	OpShlAssign(MachineElement) MachineElement
	OpShrAssign(MachineElement) MachineElement
	OpUShrAssign(MachineElement) MachineElement
	OpCatAssign(MachineElement) MachineElement
	OpCall() MachineElement
	OpIndex() MachineElement
	OpIndexElement(MachineElement) MachineElement
	OpIndexAssign() MachineElement
	OpIndexAssignElement(MachineElement, MachineElement) MachineElement
	OpSlice() MachineElement
}

type Element struct {
}

func (e *Element) AddRule(g *Grammar, x *Rule) {
	g.Add(x)
}

func (e *Element) Match(engine *Engine, r MachineElement) bool {
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
	sr.currentSymbol = e
	return s
}

func (e *Element) Compare(engine *Engine, r MachineElement) bool {
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

func (e *Element) ToBody() []MachineElement {
	return nil
}

func (e *Element) Weight() uint {
	return 0
}

func (e *Element) Token() MachineElement {
	return e
}

func (e *Element) Priority(p uint) uint {
	return p
}

func (e *Element) Reference(sr *Stream, s GenMode, x LMScope) GenMode {
	return e.Act(sr, s)
}

func (e *Element) ToExplore() MachineElement {
	return TxE("E", e)
}

func (e *Element) InvalidOp(f string) MachineElement {
	TxE("BAD "+f, e)
	return theNull()
}

func (e *Element) NotFound() MachineElement {
	return theNull()
}

func (e *Element) ToVal() MachineElement {
	return e
}

func (e *Element) ToDeref(x *Var) *Var {
	return nil
}

func (e *Element) Append(y MachineElement) MachineElement {
	return e.InvalidOp("~=")
}

func (e *Element) Inf(y MachineElement) MachineElement {
	return theNull()
}

func (e *Element) Idxf(y MachineElement) MachineElement {
	return theNull()
}

func (e *Element) Idtf(y MachineElement) MachineElement {
	return theNull()
}

func (e *Element) StoValf(y MachineElement) MachineElement {
	return e.InvalidOp("=")
}

func (e *Element) StoAddf(y MachineElement) MachineElement {
	return e.InvalidOp("+=")
}

func (e *Element) StoSubf(y MachineElement) MachineElement {
	return e.InvalidOp("-=")
}

func (e *Element) StoMulf(y MachineElement) MachineElement {
	return e.InvalidOp("*=")
}

func (e *Element) StoDivf(y MachineElement) MachineElement {
	return e.InvalidOp("/=")
}

func (e *Element) StoModf(y MachineElement) MachineElement {
	return e.InvalidOp("%=")
}

func (e *Element) StoAndf(y MachineElement) MachineElement {
	return e.InvalidOp("&=")
}

func (e *Element) StoOrf(y MachineElement) MachineElement {
	return e.InvalidOp("|=")
}

func (e *Element) StoXorf(y MachineElement) MachineElement {
	return e.InvalidOp("^=")
}

func (e *Element) StoShlf(y MachineElement) MachineElement {
	return e.InvalidOp("<<=")
}

func (e *Element) StoShrf(y MachineElement) MachineElement {
	return e.InvalidOp(">>=")
}

func (e *Element) Eeqf(y MachineElement) MachineElement {
	return NewBoolean(y.Token() == e.Token())
}

func (e *Element) Neef(y MachineElement) MachineElement {
	return NewBoolean(y.Token() != e.Token())
}

func (e *Element) Eqf(y MachineElement) MachineElement {
	return e.InvalidOp("==")
}

func (e *Element) Nef(y MachineElement) MachineElement {
	return e.InvalidOp("!=")
}

func (e *Element) Ltf(y MachineElement) MachineElement {
	return e.InvalidOp("<")
}

func (e *Element) Gtf(y MachineElement) MachineElement {
	return e.InvalidOp(">")
}

func (e *Element) Lef(y MachineElement) MachineElement {
	return e.InvalidOp("<=")
}

func (e *Element) Gef(y MachineElement) MachineElement {
	return e.InvalidOp(">=")
}

func (e *Element) BitXorf(y MachineElement) MachineElement {
	return e.InvalidOp("^")
}

func (e *Element) BitOrf(y MachineElement) MachineElement {
	return e.InvalidOp("|")
}

func (e *Element) BitAndf(y MachineElement) MachineElement {
	return e.InvalidOp("&")
}

func (e *Element) OrOrf(y MachineElement) MachineElement {
	return e.InvalidOp("||")
}

func (e *Element) AndAndf(y MachineElement) MachineElement {
	return e.InvalidOp("&&")
}

func (e *Element) Addf(y MachineElement) MachineElement {
	return e.InvalidOp("+")
}

func (e *Element) Subf(y MachineElement) MachineElement {
	return e.InvalidOp("-")
}

func (e *Element) Mulf(y MachineElement) MachineElement {
	return e.InvalidOp("*")
}

func (e *Element) Divf(y MachineElement) MachineElement {
	return e.InvalidOp("/")
}

func (e *Element) Modf(y MachineElement) MachineElement {
	return e.InvalidOp("%")
}

func (e *Element) Preincf() MachineElement {
	return e.InvalidOp("++X")
}

func (e *Element) Predecf() MachineElement {
	return e.InvalidOp("--X")
}

func (e *Element) Postincf() MachineElement {
	return e.InvalidOp("X++")
}

func (e *Element) Postdecf() MachineElement {
	return e.InvalidOp("X--")
}

func (e *Element) Posf() MachineElement {
	return e.InvalidOp("u+")
}

func (e *Element) Negf() MachineElement {
	return e.InvalidOp("u-")
}

func (e *Element) Notf() MachineElement {
	return NewBoolean(!e.ToBool())
}

func (e *Element) Invf() MachineElement {
	return e.InvalidOp("~")
}

func (e *Element) OpAdd(y MachineElement) MachineElement {
	return e.Addf(y.ToVal())
}

func (e *Element) OpShl(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpShr(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpUShr(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpCat(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpEquals(y MachineElement) bool {
	panic("badType")
}

func (e *Element) OpCmp(y MachineElement) int {
	panic("badType")
}

func (e *Element) OpAndAssign(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpOrAssign(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpXorAssign(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpShlAssign(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpShrAssign(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpUShrAssign(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpCatAssign(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpCall() MachineElement {
	panic("badType")
}

func (e *Element) OpIndex() MachineElement {
	panic("badType")
}

func (e *Element) OpIndexElement(y MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpIndexAssign() MachineElement {
	panic("badType")
}

func (e *Element) OpIndexAssignElement(y, z MachineElement) MachineElement {
	panic("badType")
}

func (e *Element) OpSlice() MachineElement {
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

func (n *Number) ToBody() []MachineElement {
	return nil
}

func (n *Number) Token() MachineElement {
	return n
}

func (n *Number) Compare(e *Engine, r MachineElement) bool {
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

func (n *Number) Posf() MachineElement {
	return n
}

func (n *Number) Negf() MachineElement {
	return NewNumber(-n.V)
}

func (n *Number) Invf() MachineElement {
	return NewNumber(LMNumber(^n.ToUlong()))
}

func (n *Number) BitXorf(y MachineElement) MachineElement {
	return NewNumber(LMNumber(n.ToUlong() ^ y.ToUlong()))
}

func (n *Number) BitOrf(y MachineElement) MachineElement {
	return NewNumber(LMNumber(n.ToUlong() | y.ToUlong()))
}

func (n *Number) BitAndf(y MachineElement) MachineElement {
	return NewNumber(LMNumber(n.ToUlong() & y.ToUlong()))
}

func (n *Number) Addf(y MachineElement) MachineElement {
	return NewNumber(n.V + y.ToNumber())
}

func (n *Number) Subf(y MachineElement) MachineElement {
	return NewNumber(n.V - y.ToNumber())
}

func (n *Number) Mulf(y MachineElement) MachineElement {
	return NewNumber(n.V * y.ToNumber())
}

func (n *Number) Divf(y MachineElement) MachineElement {
	return NewNumber(n.V / y.ToNumber())
}

func (n *Number) Modf(y MachineElement) MachineElement {
	return NewNumber(LMNumber(int64(n.V) % y.ToLong()))
}

func (n *Number) Eqf(y MachineElement) MachineElement {
	return NewBoolean(n.V == y.ToNumber())
}

func (n *Number) Nef(y MachineElement) MachineElement {
	return NewBoolean(n.V != y.ToNumber())
}

func (n *Number) Ltf(y MachineElement) MachineElement {
	return NewBoolean(n.V < y.ToNumber())
}

func (n *Number) Gtf(y MachineElement) MachineElement {
	return NewBoolean(n.V > y.ToNumber())
}

func (n *Number) Lef(y MachineElement) MachineElement {
	return NewBoolean(n.V <= y.ToNumber())
}

func (n *Number) Gef(y MachineElement) MachineElement {
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

func (b *Boolean) Compare(e *Engine, r MachineElement) bool {
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

func (b *Boolean) Notf() MachineElement {
	return NewBoolean(!b.V)
}

type Symbol struct {
	Element
	V string
}

func NewSymbol(x string) *Symbol {
	return &Symbol{V: x}
}

func (s *Symbol) Token() MachineElement {
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

func (s *Symbol) ToBody() []MachineElement {
	return nil
}

func (s *Symbol) Weight() uint {
	return 1
}

func (s *Symbol) Act(sr *Stream, m GenMode) GenMode {
	sr.currentSymbol = s
	return m
}

func (s *Symbol) Match(e *Engine, r MachineElement) bool {
	if r.Token() == s {
		return e.Matched3E(s, r, r)
	}
	return e.ResolveE(s, r)
}

func (s *Symbol) Append(y MachineElement) MachineElement {
	return s.InvalidOp("~=")
}

func (s *Symbol) ToVal() MachineElement {
	return s
}

func (s *Symbol) ToDeref(x *Var) *Var {
	return x.Deref(s)
}

func (s *Symbol) Eqf(y MachineElement) MachineElement {
	return NewBoolean(y.Token() == s)
}

func (s *Symbol) Nef(y MachineElement) MachineElement {
	return NewBoolean(y.Token() != s)
}

type Quote struct {
	Element
	V MachineElement
}

func NewQuote(x MachineElement) *Quote {
	return &Quote{V: x}
}

func (q *Quote) Token() MachineElement {
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
	sr.currentSymbol = q
	return m
}

func (q *Quote) Match(e *Engine, r MachineElement) bool {
	if r.Token() == q.V {
		return e.Matched3E(q, r, r)
	}
	return e.ResolveE(q, r)
}

func (q *Quote) ToVal() MachineElement {
	return q
}

func (q *Quote) ToDeref(x *Var) *Var {
	return nil
}

func (q *Quote) ToBool() bool {
	return q.Token().ToBool()
}

func (q *Quote) Eqf(y MachineElement) MachineElement {
	return NewBoolean(y.Token() == q.Token())
}

func (q *Quote) Nef(y MachineElement) MachineElement {
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

func (z *ZLM) Append(x MachineElement) MachineElement {
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
	V []MachineElement
}

func NewStr(x []MachineElement) *Str {
	return &Str{V: x}
}

func (s *Str) Token() MachineElement {
	return nil
}

func (s *Str) ToBody() []MachineElement {
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
	isChr := func(ge MachineElement) bool { _, ok := ge.(*Chr); return ok }
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
	return newLHModeFromElement(m, s.V, 0, m.ContextMode())
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

func NewChrStr(x []MachineElement) *ChrStr {
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

func NewLMBufferFromElement(x MachineElement) *LMBuffer {
	return &LMBuffer{V: x.ToString()}
}

func (lb *LMBuffer) ToVal() MachineElement {
	return lb
}

func (lb *LMBuffer) Append(x MachineElement) MachineElement {
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
	Vk MachineElement
	Vv MachineElement
	Vi uint
	Vp *Var
	Vx LMScope
}

func NewVarDefault() *Var {
	return &Var{}
}

func NewVarFromParams(s *Var, k, v MachineElement, q LMScope, a *Var) *Var {
	if q == nil {
		panic("vx cannot be nil")
	}
	return &Var{
		Va: a,
		Vs: s,
		Vk: k,
		Vv: v,
		Vp: q.ScopeVariables(),
		Vx: q,
	}
}

func (v *Var) Act(sr *Stream, s GenMode) GenMode {
	if v.Vv != nil {
		return v.Vv.Reference(sr, s, v)
	}
	return s
}

func (v *Var) ScopeVariables() *Var {
	return v.Vp
}

func (v *Var) ScopeContextLimitVariables() *Var {
	if v.Vx != nil {
		return v.Vx.ScopeContextLimitVariables()
	}
	return nil
}

func (v *Var) ScopeReferenceContext() LMScope {
	return v.Vx
}

func (v *Var) ScopeContextMode() EngineStateContext {
	if v.Vx != nil {
		return v.Vx.ScopeContextMode()
	}
	return nil
}

func (v *Var) MakeVar(k, ve MachineElement, s LMScope, a *Var) *Var {
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

func (v *Var) Deref(k MachineElement) *Var {
	pp := v.Vp
	for (pp != nil) && !(k == pp.Vk) {
		pp = pp.Vs
	}
	return pp
}

func (v *Var) ToExplore() MachineElement {
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

func (v *Var) ToVal() MachineElement {
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

func (v *Var) Append(y MachineElement) MachineElement {
	if v.Vv == nil || v.Vv == theNull() {
		v.Vv = NewLMBuffer()
	}
	return v.Vv.Append(y)
}

func (v *Var) Idxf(y MachineElement) MachineElement {
	return v.Vv.Idxf(y.ToVal())
}

func (v *Var) Idtf(y MachineElement) MachineElement {
	return v.Vv.Idtf(y.ToVal())
}

func (v *Var) StoValf(y MachineElement) MachineElement {
	v.Vv = y.ToVal()
	return v.Vv
}

func (v *Var) StoAddf(y MachineElement) MachineElement {
	v.Vv = v.Vv.Addf(y.ToVal())
	return v.Vv
}

func (v *Var) StoSubf(y MachineElement) MachineElement {
	v.Vv = v.Vv.Subf(y.ToVal())
	return v.Vv
}

func (v *Var) StoMulf(y MachineElement) MachineElement {
	v.Vv = v.Vv.Mulf(y.ToVal())
	return v.Vv
}

func (v *Var) StoDivf(y MachineElement) MachineElement {
	v.Vv = v.Vv.Divf(y.ToVal())
	return v.Vv
}

func (v *Var) StoModf(y MachineElement) MachineElement {
	v.Vv = v.Vv.Modf(y.ToVal())
	return v.Vv
}

func (v *Var) Preincf() MachineElement {
	return v.StoAddf(NewNumber(1))
}

func (v *Var) Predecf() MachineElement {
	return v.StoSubf(NewNumber(1))
}

func (v *Var) Postincf() MachineElement {
	r := v.Vv.ToVal()
	v.Preincf()
	return r
}

func (v *Var) Postdecf() MachineElement {
	r := v.Vv.ToVal()
	v.Predecf()
	return r
}

func (v *Var) Ru() *Rule {
	return v.ScopeContextMode().Rule()
}

func (v *Var) Si() uint {
	return v.ScopeContextMode().State().stateIndex
}

func (v *Var) Gr() *Grammar {
	return v.ScopeContextMode().State().grammar
}

func (v *Var) Gsy() MachineElement {
	return v.ScopeContextMode().State().grammar.symbol
}

func (v *Var) Rsy() MachineElement {
	return v.ScopeContextMode().State().rsy
}

func (v *Var) Lsy() MachineElement {
	return v.ScopeContextMode().State().lsy
}

func (v *Var) Ifn() string {
	return v.ScopeContextMode().State().input.Filename()
}

func (v *Var) Cp() uint {
	return v.ScopeContextMode().State().charPosition
}

func (v *Var) Ln() uint {
	return v.ScopeContextMode().State().lineNumber
}

func (v *Var) Cn() uint {
	return v.ScopeContextMode().State().charNumber
}

type LMRef struct {
	Var
}

func NewLMRef() *LMRef {
	return &LMRef{}
}

func NewLMRefFromElement(k MachineElement, q LMScope) *LMRef {
	lm := &LMRef{}
	lm.Vk = k
	lm.Vp = q.ScopeVariables() // Assuming Vvp() returns Var
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

func (lm *LMRef) ToExplore() MachineElement {
	TxV("R", " ", &lm.Var)
	if lm.Vv != nil {
		lm.Vv.ToExplore()
	}
	return lm
}

func (lm *LMRef) ToVal() MachineElement {
	return lm.Var.ToVal()
}

func (lm *LMRef) ToRef() *Var {
	return &lm.Var
}

func (lm *LMRef) ToDeref(v *Var) *Var {
	return nil
}

func (lm *LMRef) Append(y MachineElement) MachineElement {
	return lm.Vv.Append(y)
}

func (lm *LMRef) Funf(y MachineElement) MachineElement {
	return nil
}

func (lm *LMRef) Inf(y MachineElement) MachineElement {
	return lm.Vv.Inf(y.ToVal())
}

func (lm *LMRef) Idxf(y MachineElement) MachineElement {
	return lm.Vv.Idxf(y.ToVal())
}

func (lm *LMRef) Idtf(y MachineElement) MachineElement {
	return lm.Vv.Idtf(y.ToVal())
}

func (lm *LMRef) StoValf(y MachineElement) MachineElement {
	return lm.Vv.StoValf(y.ToVal())
}

func (lm *LMRef) StoAddf(y MachineElement) MachineElement {
	return lm.Vv.StoAddf(y.ToVal())
}

func (lm *LMRef) StoSubf(y MachineElement) MachineElement {
	return lm.Vv.StoSubf(y.ToVal())
}

func (lm *LMRef) StoMulf(y MachineElement) MachineElement {
	return lm.Vv.StoMulf(y.ToVal())
}

func (lm *LMRef) StoDivf(y MachineElement) MachineElement {
	return lm.Vv.StoDivf(y.ToVal())
}

func (lm *LMRef) StoModf(y MachineElement) MachineElement {
	return lm.Vv.StoModf(y.ToVal())
}

type ARef struct {
	Var
	A *AArray
	K MachineElement
}

func NewARef(x *AArray, y MachineElement, z LMScope) *ARef {
	ar := &ARef{}
	ar.A = x
	ar.K = y
	ar.Vv = z.ScopeVariables()
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

func (ar *ARef) ToVal() MachineElement {
	if val, ok := ar.A.A[ar.K]; ok {
		return val
	}
	//+ return NotFound()
	return nil
}

func (ar *ARef) ToDeref(v *Var) *Var {
	return nil
}

func (ar *ARef) Append(y MachineElement) MachineElement {
	return ar.StoValf(ar.ToVal().Append(y))
}

func (ar *ARef) Inf(y MachineElement) MachineElement {
	return ar.ToVal().Inf(y)
}

func (ar *ARef) Idxf(y MachineElement) MachineElement {
	return ar.ToVal().Idxf(y)
}

func (ar *ARef) Idtf(y MachineElement) MachineElement {
	return ar.ToVal().Idtf(y)
}

func (ar *ARef) StoValf(y MachineElement) MachineElement {
	ar.A.A[ar.K] = y
	return y
}

func (ar *ARef) StoAddf(y MachineElement) MachineElement {
	r := ar.ToVal().Addf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoSubf(y MachineElement) MachineElement {
	r := ar.ToVal().Subf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoMulf(y MachineElement) MachineElement {
	r := ar.ToVal().Mulf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoDivf(y MachineElement) MachineElement {
	r := ar.ToVal().Divf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoModf(y MachineElement) MachineElement {
	r := ar.ToVal().Modf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) Preincf() MachineElement {
	return ar.StoAddf(NewNumber(1))
}

func (ar *ARef) Predecf() MachineElement {
	return ar.StoSubf(NewNumber(1))
}

func (ar *ARef) Postincf() MachineElement {
	r := ar.ToVal()
	ar.Preincf()
	return r
}

func (ar *ARef) Postdecf() MachineElement {
	r := ar.ToVal()
	ar.Predecf()
	return r
}

type AArray struct {
	A map[MachineElement]MachineElement
}

func NewAArray() *AArray {
	return &AArray{A: make(map[MachineElement]MachineElement)}
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
	var v MachineElement
	var i uint

	for x, i = sr.operandsStack, 0; x != nil && x.V != sr.PredefinedSymbols().mark; x = x.S {
		if la.Assign(sr, x.V) == nil {
			i++
		}
	}
	for v = sr.Popx(); sr.operandsStack != nil && v != sr.PredefinedSymbols().mark; v = sr.Popx() {
		if sr.operandsStack.V == nil {
			i -= 1
			la.AssignE(sr, i, v)
		}
	}

	return &la
}

func (la *LMArray) ToVal() MachineElement {
	return la
}

func (la *LMArray) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(la)
	return s
}

// func (la *LMArray) Assign(e *Stream, c *LMCell) MachineElement {
func (la *LMArray) Assign(e *Stream, c MachineElement) MachineElement {
	if c != nil {
		lm, ok := c.(*LMCell)
		if !ok {
			panic("not lmcell")
			return nil
		}
		la.aa.A[e.UserSymbols().UniqueE(lm.K)] = lm.V
		return lm.V
	}
	return nil
}

func (la *LMArray) AssignE(e *Stream, i uint, v MachineElement) MachineElement {
	la.aa.A[e.UserSymbols().UniqueE(NewNumber(LMNumber(i)))] = v
	return v
}

func (la *LMArray) Idxf(y MachineElement) MachineElement {
	return NewARef(la.aa, la.sx.ScopeContextMode().State().engine.userSymbols.UniqueE(y.ToVal()), la.sx)
}

func (la *LMArray) Idtf(y MachineElement) MachineElement {
	return NewARef(la.aa, la.sx.ScopeContextMode().State().engine.userSymbols.UniqueE(y.ToVal()), la.sx)
}

type LMCell struct {
	Element
	K MachineElement
	V MachineElement
}

func NewLMCell(y, z MachineElement) *LMCell {
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
	s.MakeVar(k, v, s, sr.variables)
	return s
}

type EachRef struct {
	Element
	K MachineElement
}

func NewEachRef(x MachineElement) *EachRef {
	return &EachRef{K: x}
}

func (er *EachRef) Token() MachineElement {
	return nil
}

func (er *EachRef) ToBody() []MachineElement {
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

func (er *EachRef) Match(e *Engine, r MachineElement) bool {
	return false
}

type AllRef struct {
	Element
	K MachineElement
}

func NewAllRef(x MachineElement) *AllRef {
	return &AllRef{K: x}
}

func (ar *AllRef) Token() MachineElement {
	return nil
}

func (ar *AllRef) ToBody() []MachineElement {
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

func (ar *AllRef) Match(e *Engine, r MachineElement) bool {
	return false
}

type VarSym struct {
	Symbol
}

func NewVarSym(x string) *VarSym {
	return &VarSym{Symbol{V: x}}
}

func (vs *VarSym) Token() MachineElement {
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

func (vs *VarSym) Match(e *Engine, r MachineElement) bool {
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

func (df *DoneF) Match(e *Engine, r MachineElement) bool {
	e.lhsStream.variables = e.lhsContext.Variables()
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

func (tf *TakeF) Match(e *Engine, r MachineElement) bool {
	if r.Token() == tf { // %  %
		e.TakeTvar()
		return e.Matched3E(tf, r, nil)
	}
	if _, ok := r.(*BindF); ok { // %  :
		e.PushX()
		return e.Matched3E(tf, r, nil)
	}
	if e.rsLastMatchElement != nil { // %  [matched]
		e.PushR(e.rsLastMatchElement)
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

func (b *BindF) Match(e *Engine, r MachineElement) bool {
	if r.Token() == b {
		a := e.lhsStream.Popx()
		bElem := e.rhsStream.Popx().ToVal()
		// tx("l", a); tx("r", bElem);
		if a, ok := a.(*VarSym); ok {
			e.BindUvar(a, bElem)
			return e.Matched3E(b, r, nil)
		}
		e.Matched3E(b, r, nil)
		lh := []MachineElement{a}
		e.lhsStream.mode = NewSTModeFromElements(e.lhsStream.mode, lh, e.lhsStream.mode)
		rh := []MachineElement{bElem}
		e.rhsStream.mode = NewSTModeFromElements(e.rhsStream.mode, rh, e.rhsStream.mode)
		return true
	}
	if _, ok := r.(*TakeF); ok {
		e.BindTvar()
		return e.Matched3E(b, r, nil)
	}
	if e.rsLastMatchElement != nil {
		e.BindXvarE(e.rsLastMatchElement)
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

func (a *AppendSym) Match(e *Engine, r MachineElement) bool {
	e.lhsStream.Popx().Append(r)
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
func (a *AppendXSym) Match(e *Engine, r MachineElement) bool {
	b := e.lhsStream.Popx()
	if e.lhsStream.operandsStack != nil {
		v := e.lhsStream.ToRow()
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

func (e *ErrSym) Append(y MachineElement) MachineElement {
	fmt.Fprintf(os.Stderr, "%s", y)
	return e
}

func (e *ErrSym) Match(engine *Engine, r MachineElement) bool {
	fmt.Fprintf(os.Stderr, "%s", r)
	return engine.Matched3E(e, r, r)
}

type OutSym struct {
	Symbol
}

func NewOutSym(x string) *OutSym {
	return &OutSym{Symbol{V: x}}
}

func (o *OutSym) Append(y MachineElement) MachineElement {
	fmt.Fprintf(os.Stdout, "%s", y)
	return o
}

func (o *OutSym) Match(engine *Engine, r MachineElement) bool {
	fmt.Fprintf(os.Stdout, "%s", r)
	return engine.Matched3E(o, r, r)
}

type UriSym struct {
	Symbol
}

func NewUriSym(x string) *UriSym {
	return &UriSym{Symbol{V: x}}
}

func (u *UriSym) Append(y MachineElement) MachineElement {
	fmt.Fprintf(os.Stdout, "%s", y.ToEncode())
	return u
}

func (u *UriSym) Match(engine *Engine, r MachineElement) bool {
	fmt.Fprintf(os.Stdout, "%s", r.ToEncode())
	return engine.Matched3E(u, r, r)
}

type UrdSym struct {
	Symbol
}

func NewUrdSym(x string) *UrdSym {
	return &UrdSym{Symbol{V: x}}
}

func (u *UrdSym) Append(y MachineElement) MachineElement {
	fmt.Fprintf(os.Stdout, "%s", y.ToDecode())
	return u
}

func (u *UrdSym) Match(engine *Engine, r MachineElement) bool {
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

func (t *TrueSym) ToVal() MachineElement {
	return NewBoolean(true)
}

type FalseSym struct {
	Symbol
}

func NewFalseSym(x string) *FalseSym {
	return &FalseSym{Symbol{V: x}}
}

func (f *FalseSym) ToVal() MachineElement {
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
	V MachineElement
}

func NewGetXF(x MachineElement) *GetXF {
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
	V MachineElement
}

func NewGetBF(x MachineElement) *GetBF {
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
	sr.currentSymbol = sr.PredefinedSymbols().bindFn
	return s
}

type GetVF struct {
	Symbol
	V MachineElement
}

func NewGetVF(x MachineElement) *GetVF {
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

func (i *InjF) Match(e *Engine, r MachineElement) bool {
	e.PushRhx(e.lhsStream.Popx())
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

func (a *Anything) Match(e *Engine, r MachineElement) bool {
	e.Matched3E(a, r, r)
	return true
}

type AnySym struct {
	Anything
}

func NewAnySym(x string) *AnySym {
	return &AnySym{Anything: *NewAnything(x)}
}

func (a *AnySym) Match(e *Engine, r MachineElement) bool {
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

func (a *AnyChr) Match(e *Engine, r MachineElement) bool {
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

func (a *AnyNum) Match(e *Engine, r MachineElement) bool {
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

func (l *LnoSym) Match(e *Engine, r MachineElement) bool {
	return e.Matched3E(l, nil, NewNumber(LMNumber(e.Lineno())))
}

type IfnSym struct {
	Symbol
}

func NewIfnSym(x string) *IfnSym {
	return &IfnSym{Symbol: *NewSymbol(x)}
}

func (i *IfnSym) Match(e *Engine, r MachineElement) bool {
	return e.Matched3E(i, nil, NewSym(e.Filename()))
}

type FlagSym struct {
	Symbol
}

func NewFlagSym(x string) *FlagSym {
	return &FlagSym{Symbol: *NewSymbol(x)}
}

func (f *FlagSym) Match(e *Engine, r MachineElement) bool {
	e.flagErrors++
	return e.Matched3E(f, nil, NewSym(e.Filename()+":"+string(e.Lineno())+": "))
}

type WarnSym struct {
	Symbol
}

func NewWarnSym(x string) *WarnSym {
	return &WarnSym{Symbol: *NewSymbol(x)}
}

func (w *WarnSym) Match(e *Engine, r MachineElement) bool {
	e.warnErrors++
	return e.Matched3E(w, nil, NewSym(e.Filename()+":"+string(e.Lineno())+": "))
}

type RepnSym struct {
	Symbol
}

func NewRepnSym(x string) *RepnSym {
	return &RepnSym{Symbol: *NewSymbol(x)}
}

func (r *RepnSym) Match(e *Engine, _ MachineElement) bool {
	n := e.lhsStream.Popx().ToVal().(*Number)
	return e.Repeat(uint(n.ToLong()))
}

type RepSym struct {
	Symbol
}

func NewRepSym(x string) *RepSym {
	return &RepSym{Symbol: *NewSymbol(x)}
}

func (r *RepSym) Match(e *Engine, _ MachineElement) bool {
	return e.Repeat(0)
}

type OptSym struct {
	Symbol
}

func NewOptSym(x string) *OptSym {
	return &OptSym{Symbol: *NewSymbol(x)}
}

func (o *OptSym) Match(e *Engine, r MachineElement) bool {
	return e.Repeat(1)
}

type OptxSym struct {
	Symbol
}

func NewOptxSym(x string) *OptxSym {
	return &OptxSym{Symbol: *NewSymbol(x)}
}

func (o *OptxSym) Match(e *Engine, r MachineElement) bool {
	return e.Repeatx(1)
}

type RepxSym struct {
	Symbol
}

func NewRepxSym(x string) *RepxSym {
	return &RepxSym{Symbol: *NewSymbol(x)}
}

func (r *RepxSym) Match(e *Engine, _ MachineElement) bool {
	return e.Repeatx(0)
}

type Lex struct {
	Symbol
	Table     map[MachineElement]MachineElement
	Inclusive bool
}

func NewLex(x string) *Lex {
	return &Lex{
		Symbol:    *NewSymbol(x),
		Table:     make(map[MachineElement]MachineElement),
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
		var x MachineElement

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
				x = e.terminalSymbols.UniqueR(rune(c))
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
			x = e.terminalSymbols.UniqueR(rune(c))
			l.Table[x] = x
			prevc = c
			state = C1

		case C2:
			if c == '\\' {
				state = E1
			} else if c == '-' {
				state = RN
			} else {
				x = e.terminalSymbols.UniqueR(rune(c))
				l.Table[x] = x
				prevc = c
			}

		case RN:
			for prevc < c {
				x = e.terminalSymbols.UniqueR(rune(c))
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

func (l *Lex) Match(e *Engine, r MachineElement) bool {
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
	sr.operandsStack = nil
	return s
}

type Unary struct {
	Primitive
}

func NewUnary(x string) *Unary {
	return &Unary{Primitive: *NewPrimitiveFromString(x)}
}

func (u *Unary) Result(x MachineElement) MachineElement {
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

func (a *Arithmetic) Result(x, y MachineElement) MachineElement {
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

func (r *Relation) Result(x, y MachineElement) MachineElement {
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

func (a *Assignment) Result(x, y MachineElement) MachineElement {
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

func (i *IncDec) Result(x MachineElement) MachineElement {
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

func (i *Index) Result(x, y MachineElement) MachineElement {
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
	sr.Pushx(sr.PredefinedSymbols().mark)
	return b
}

type Funf struct {
	Primitive
}

func NewFunf(x string) *Funf {
	return &Funf{Primitive: *NewPrimitiveFromString(x)}
}

func (f *Funf) Act(sr *Stream, b GenMode) GenMode {
	v := sr.ToArgv(sr.PredefinedSymbols().mark)
	sr.Pushx(sr.ExternalSystem().Call(sr, b, v[0], v))
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
func (i Idxf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Idxf(y)
}

type Idtf struct {
	Index
}

func NewIdtf(x string) *Idtf {
	return &Idtf{Index: *NewIndex(x)}
}
func (i Idtf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Idtf(y)
}

type StoValf struct {
	Assignment
}

func NewStoValf(x string) *StoValf {
	return &StoValf{Assignment: *NewAssignment(x)}
}
func (s StoValf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.StoValf(y)
}

type StoAddf struct {
	Assignment
}

func NewStoAddf(x string) *StoAddf {
	return &StoAddf{Assignment: *NewAssignment(x)}
}
func (s StoAddf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.StoAddf(y)
}

type StoSubf struct {
	Assignment
}

func NewStoSubf(x string) *StoSubf {
	return &StoSubf{Assignment: *NewAssignment(x)}
}
func (s StoSubf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.StoSubf(y)
}

type StoMulf struct {
	Assignment
}

func NewStoMulf(x string) *StoMulf {
	return &StoMulf{Assignment: *NewAssignment(x)}
}
func (s StoMulf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.StoMulf(y)
}

type StoDivf struct {
	Assignment
}

func NewStoDivf(x string) *StoDivf {
	return &StoDivf{Assignment: *NewAssignment(x)}
}
func (s StoDivf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.StoDivf(y)
}

type StoModf struct {
	Assignment
}

func NewStoModf(x string) *StoModf {
	return &StoModf{Assignment: *NewAssignment(x)}
}
func (s StoModf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.StoModf(y)
}

type Eeqf struct {
	Relation
}

func NewEeqf(x string) *Eeqf {
	return &Eeqf{Relation: *NewRelation(x)}
}
func (e Eeqf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Eeqf(y)
}

type Neef struct {
	Relation
}

func NewNeef(x string) *Neef {
	return &Neef{Relation: *NewRelation(x)}
}
func (n Neef) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Neef(y)
}

type Inf struct {
	Relation
}

func NewInf(x string) *Inf {
	return &Inf{Relation: *NewRelation(x)}
}
func (i Inf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Inf(y)
}

type Eqf struct {
	Relation
}

func NewEqf(x string) *Eqf {
	return &Eqf{Relation: *NewRelation(x)}
}
func (e Eqf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Eqf(y)
}

type Nef struct {
	Relation
}

func NewNef(x string) *Nef {
	return &Nef{Relation: *NewRelation(x)}
}
func (n Nef) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Nef(y)
}

type Ltf struct {
	Relation
}

func NewLtf(x string) *Ltf {
	return &Ltf{Relation: *NewRelation(x)}
}
func (l Ltf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Ltf(y)
}

type Gtf struct {
	Relation
}

func NewGtf(x string) *Gtf {
	return &Gtf{Relation: *NewRelation(x)}
}
func (g Gtf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Gtf(y)
}

type Lef struct {
	Relation
}

func NewLef(x string) *Lef {
	return &Lef{Relation: *NewRelation(x)}
}
func (l Lef) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Lef(y)
}

type Gef struct {
	Relation
}

func NewGef(x string) *Gef {
	return &Gef{Relation: *NewRelation(x)}
}
func (g Gef) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Gef(y)
}

type BitXorf struct {
	Arithmetic
}

func NewBitXorf(x string) *BitXorf {
	return &BitXorf{Arithmetic: *NewArithmetic(x)}
}
func (b BitXorf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.BitXorf(y)
}

type BitOrf struct {
	Arithmetic
}

func NewBitOrf(x string) *BitOrf {
	return &BitOrf{Arithmetic: *NewArithmetic(x)}
}
func (b BitOrf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.BitOrf(y)
}

type BitAndf struct {
	Arithmetic
}

func NewBitAndf(x string) *BitAndf {
	return &BitAndf{Arithmetic: *NewArithmetic(x)}
}
func (b BitAndf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.BitAndf(y)
}

type Addf struct {
	Arithmetic
}

func NewAddf(x string) *Addf {
	return &Addf{Arithmetic: *NewArithmetic(x)}
}
func (a Addf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Addf(y)
}

type Subf struct {
	Arithmetic
}

func NewSubf(x string) *Subf {
	return &Subf{Arithmetic: *NewArithmetic(x)}
}
func (s Subf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Subf(y)
}

type Mulf struct {
	Arithmetic
}

func NewMulf(x string) *Mulf {
	return &Mulf{Arithmetic: *NewArithmetic(x)}
}
func (m Mulf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Mulf(y)
}

type Divf struct {
	Arithmetic
}

func NewDivf(x string) *Divf {
	return &Divf{Arithmetic: *NewArithmetic(x)}
}
func (d Divf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Divf(y)
}

type Modf struct {
	Arithmetic
}

func NewModf(x string) *Modf {
	return &Modf{Arithmetic: *NewArithmetic(x)}
}
func (m Modf) Result(x MachineElement, y MachineElement) MachineElement {
	return x.Modf(y)
}

type Preincf struct {
	IncDec
}

func NewPreincf(x string) *Preincf {
	return &Preincf{IncDec: *NewIncDec(x)}
}
func (p Preincf) Result(x MachineElement) MachineElement {
	return x.Preincf()
}

type Predecf struct {
	IncDec
}

func NewPredecf(x string) *Predecf {
	return &Predecf{IncDec: *NewIncDec(x)}
}
func (p Predecf) Result(x MachineElement) MachineElement {
	return x.Predecf()
}

type Postincf struct {
	IncDec
}

func NewPostincf(x string) *Postincf {
	return &Postincf{IncDec: *NewIncDec(x)}
}
func (p Postincf) Result(x MachineElement) MachineElement {
	return x.Postincf()
}

type Postdecf struct {
	IncDec
}

func NewPostdecf(x string) *Postdecf {
	return &Postdecf{IncDec: *NewIncDec(x)}
}
func (p Postdecf) Result(x MachineElement) MachineElement {
	return x.Postdecf()
}

type Negf struct {
	Unary
}

func NewNegf(x string) *Negf {
	return &Negf{Unary: *NewUnary(x)}
}
func (n Negf) Result(x MachineElement) MachineElement {
	return x.Negf()
}

type Notf struct {
	Unary
}

func NewNotf(x string) *Notf {
	return &Notf{Unary: *NewUnary(x)}
}
func (n Notf) Result(x MachineElement) MachineElement {
	return x.Notf()
}

type Invf struct {
	Unary
}

func NewInvf(x string) *Invf {
	return &Invf{Unary: *NewUnary(x)}
}
func (i Invf) Result(x MachineElement) MachineElement {
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

func (i *IOSymbol) Match(e *Engine, r MachineElement) bool {
	return i.H.Match(e, i, r)
}
