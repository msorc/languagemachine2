package machine

import (
	"fmt"
	"github.com/msorc/languagemachine2/internal/utils"
	"math"
	"os"
	"slices"
	"strconv"
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

type Element interface {
	SelfPointer[Element]
	AddRule(*Grammar, *Rule)
	Match(*Engine, Element) bool
	NewLHS(GenMode) GenMode
	NewRHX(GenMode, ContextHolder, ScopeHolder) GenMode
	Act(*Stream, GenMode) GenMode
	Compare(*Engine, Element) bool
	ToNumber() LMNumber
	IsNumber() bool
	ToBool() bool
	ToVar() VarElement
	ToInt() int
	Trace(*Stream, *Tracer)
	ToString() string
	ToTrace() string
	ToEncode() string
	ToDecode() string
	ToDump() string
	ToBody() []Element
	Weight() int
	Token() Element
	Priority(int) int
	Reference(*Stream, GenMode, ScopeHolder) GenMode
	InvalidOp(string) Element
	NotFound() Element
	ToVal() Element
	Append(Element) Element
	Inf(Element) Element
	Idxf(y Element) Element
	Idtf(y Element) Element
	StoValf(y Element) Element
	StoAddf(y Element) Element
	StoSubf(y Element) Element
	StoMulf(y Element) Element
	StoDivf(y Element) Element
	StoModf(y Element) Element
	Eeqf(y Element) Element
	Neef(y Element) Element
	Eqf(y Element) Element
	Nef(y Element) Element
	Ltf(y Element) Element
	Gtf(y Element) Element
	Lef(y Element) Element
	Gef(y Element) Element
	BitXorf(y Element) Element
	BitOrf(y Element) Element
	BitAndf(y Element) Element
	Addf(y Element) Element
	Subf(y Element) Element
	Mulf(y Element) Element
	Divf(y Element) Element
	Modf(y Element) Element
	Preincf() Element
	Predecf() Element
	Postincf() Element
	Postdecf() Element
	Negf() Element
	Notf() Element
	Invf() Element
	Result1(Element) Element
	Result2(Element, Element) Element
}

type GenericElement struct {
	SelfPointing[Element]
}

func (e *GenericElement) AddRule(g *Grammar, x *Rule) {
	g.Add(x)
}

func (e *GenericElement) Match(engine *Engine, r Element) bool {
	if e.Self().Compare(engine, r) {
		engine.Matched3E(e.Self(), r, r)
		return true
	} else {
		return engine.ResolveE(e.Self(), r)
	}
}

func (e *GenericElement) NewLHS(m GenMode) GenMode {
	panic("not implemented")
}

func (e *GenericElement) NewRHX(m GenMode, c ContextHolder, x ScopeHolder) GenMode {
	panic("not implemented")
}

func (e *GenericElement) Act(sr *Stream, s GenMode) GenMode {
	sr.currentSymbol = e.Self()
	return s
}

func (e *GenericElement) Compare(engine *Engine, r Element) bool {
	return false
}

// ToNumber of a non-number is NaN, as in the original (lmNumber.init), so
// numeric comparisons with it are false rather than fatal.
func (e *GenericElement) ToNumber() LMNumber {
	return LMNumber(math.NaN())
}

func (e *GenericElement) IsNumber() bool {
	return false
}

func (e *GenericElement) ToBool() bool {
	return true
}

func (e *GenericElement) ToVar() VarElement {
	panic("not implemented")
}

func (e *GenericElement) ToInt() int {
	panic("not implemented")
}

func (e *GenericElement) Trace(s *Stream, t *Tracer) {
}

func (e *GenericElement) ToString() string {
	return "element"
}

func (e *GenericElement) ToTrace() string {
	return e.Self().ToString()
}

func (e *GenericElement) ToEncode() string {
	return utils.Encode(e.Self().ToString())
}

func (e *GenericElement) ToDecode() string {
	return decodeURI(e.Self().ToString())
}

func decodeURI(s string) string {
	d, err := utils.Decode(s)
	if err != nil {
		fail("cannot URI-decode %q: %v", s, err)
	}
	return d
}

func (e *GenericElement) ToDump() string {
	return e.Self().ToEncode()
}

func (e *GenericElement) ToBody() []Element {
	return nil
}

func (e *GenericElement) Weight() int {
	return 0
}

func (e *GenericElement) Token() Element {
	return e.Self()
}

func (e *GenericElement) Priority(p int) int {
	return p
}

func (e *GenericElement) Reference(sr *Stream, s GenMode, x ScopeHolder) GenMode {
	return e.Self().Act(sr, s)
}

func (e *GenericElement) InvalidOp(f string) Element {
	TxE(os.Stderr, "BAD "+f, e.Self())
	return theNull()
}

func (e *GenericElement) NotFound() Element {
	return theNull()
}

func (e *GenericElement) ToVal() Element {
	return e.Self()
}

func (e *GenericElement) Append(y Element) Element {
	return e.Self().InvalidOp("~=")
}

func (e *GenericElement) Inf(y Element) Element {
	return theNull()
}

func (e *GenericElement) Idxf(y Element) Element {
	return theNull()
}

func (e *GenericElement) Idtf(y Element) Element {
	return theNull()
}

func (e *GenericElement) StoValf(y Element) Element {
	return e.Self().InvalidOp("=")
}

func (e *GenericElement) StoAddf(y Element) Element {
	return e.Self().InvalidOp("+=")
}

func (e *GenericElement) StoSubf(y Element) Element {
	return e.Self().InvalidOp("-=")
}

func (e *GenericElement) StoMulf(y Element) Element {
	return e.Self().InvalidOp("*=")
}

func (e *GenericElement) StoDivf(y Element) Element {
	return e.Self().InvalidOp("/=")
}

func (e *GenericElement) StoModf(y Element) Element {
	return e.Self().InvalidOp("%=")
}

func (e *GenericElement) Eeqf(y Element) Element {
	return NewBoolean(y.Token() == e.Self().Token())
}

func (e *GenericElement) Neef(y Element) Element {
	return NewBoolean(y.Token() != e.Self().Token())
}

func (e *GenericElement) Eqf(y Element) Element {
	return e.Self().InvalidOp("==")
}

func (e *GenericElement) Nef(y Element) Element {
	return e.Self().InvalidOp("!=")
}

func (e *GenericElement) Ltf(y Element) Element {
	return e.Self().InvalidOp("<")
}

func (e *GenericElement) Gtf(y Element) Element {
	return e.Self().InvalidOp(">")
}

func (e *GenericElement) Lef(y Element) Element {
	return e.Self().InvalidOp("<=")
}

func (e *GenericElement) Gef(y Element) Element {
	return e.Self().InvalidOp(">=")
}

func (e *GenericElement) BitXorf(y Element) Element {
	return e.Self().InvalidOp("^")
}

func (e *GenericElement) BitOrf(y Element) Element {
	return e.Self().InvalidOp("|")
}

func (e *GenericElement) BitAndf(y Element) Element {
	return e.Self().InvalidOp("&")
}

func (e *GenericElement) Addf(y Element) Element {
	return e.Self().InvalidOp("+")
}

func (e *GenericElement) Subf(y Element) Element {
	return e.Self().InvalidOp("-")
}

func (e *GenericElement) Mulf(y Element) Element {
	return e.Self().InvalidOp("*")
}

func (e *GenericElement) Divf(y Element) Element {
	return e.Self().InvalidOp("/")
}

func (e *GenericElement) Modf(y Element) Element {
	return e.Self().InvalidOp("%")
}

func (e *GenericElement) Preincf() Element {
	return e.Self().InvalidOp("++X")
}

func (e *GenericElement) Predecf() Element {
	return e.Self().InvalidOp("--X")
}

func (e *GenericElement) Postincf() Element {
	return e.Self().InvalidOp("X++")
}

func (e *GenericElement) Postdecf() Element {
	return e.Self().InvalidOp("X--")
}

func (e *GenericElement) Negf() Element {
	return e.Self().InvalidOp("u-")
}

func (e *GenericElement) Notf() Element {
	return NewBoolean(!e.Self().ToBool())
}

func (e *GenericElement) Invf() Element {
	return e.Self().InvalidOp("~")
}

func (e *GenericElement) Result1(Element) Element {
	panic("not implemented")
}

func (e *GenericElement) Result2(Element, Element) Element {
	panic("not implemented")
}

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

func (n *Number) ToBody() []Element {
	return nil
}

func (n *Number) Token() Element {
	return n.Self()
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

func (n *Number) Negf() Element {
	return NewNumber(-n.V)
}

func (n *Number) Invf() Element {
	return NewNumber(LMNumber(^n.Self().ToInt()))
}

func (n *Number) BitXorf(y Element) Element {
	return NewNumber(LMNumber(n.Self().ToInt() ^ y.Self().ToInt()))
}

func (n *Number) BitOrf(y Element) Element {
	return NewNumber(LMNumber(n.Self().ToInt() | y.Self().ToInt()))
}

func (n *Number) BitAndf(y Element) Element {
	return NewNumber(LMNumber(n.Self().ToInt() & y.Self().ToInt()))
}

func (n *Number) Addf(y Element) Element {
	return NewNumber(n.V + y.ToNumber())
}

func (n *Number) Subf(y Element) Element {
	return NewNumber(n.V - y.ToNumber())
}

func (n *Number) Mulf(y Element) Element {
	return NewNumber(n.V * y.ToNumber())
}

func (n *Number) Divf(y Element) Element {
	return NewNumber(n.V / y.ToNumber())
}

// Modf is % on doubles (fmod), as in the original: no panic on a zero divisor.
func (n *Number) Modf(y Element) Element {
	return NewNumber(LMNumber(math.Mod(float64(n.V), float64(y.ToNumber()))))
}

func (n *Number) Eqf(y Element) Element {
	return NewBoolean(n.V == y.ToNumber())
}

func (n *Number) Nef(y Element) Element {
	return NewBoolean(n.V != y.ToNumber())
}

func (n *Number) Ltf(y Element) Element {
	return NewBoolean(n.V < y.ToNumber())
}

func (n *Number) Gtf(y Element) Element {
	return NewBoolean(n.V > y.ToNumber())
}

func (n *Number) Lef(y Element) Element {
	return NewBoolean(n.V <= y.ToNumber())
}

func (n *Number) Gef(y Element) Element {
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

func (b *Boolean) ToUlong() int {
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

func (b *Boolean) Notf() Element {
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

func (s *Symbol) Token() Element {
	return s.Self()
}

func (s *Symbol) ToDump() string {
	return "m:" + utils.Encode(s.V)
}

func (s *Symbol) ToString() string {
	return s.V
}

func (s *Symbol) ToEncode() string {
	return utils.Encode(s.Self().ToString())
}

func (s *Symbol) ToBody() []Element {
	return nil
}

func (s *Symbol) Weight() int {
	return 1
}

func (s *Symbol) Act(sr *Stream, m GenMode) GenMode {
	sr.currentSymbol = s.Self()
	return m
}

func (s *Symbol) Match(e *Engine, r Element) bool {
	if r.Token() == s.Self() {
		return e.Matched3E(s.Self(), r, r)
	}
	return e.ResolveE(s.Self(), r)
}

func (s *Symbol) Append(y Element) Element {
	return s.Self().InvalidOp("~=")
}

func (s *Symbol) ToVal() Element {
	return s.Self()
}

func (s *Symbol) Eqf(y Element) Element {
	return NewBoolean(y.Token() == s.Self())
}

func (s *Symbol) Nef(y Element) Element {
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
	return "d:" + utils.Encode(q.V.ToString())
}

func (q *Quote) ToEncode() string {
	return utils.Encode(q.ToString())
}

func (q *Quote) ToString() string {
	return q.V.ToString()
}

func (q *Quote) Weight() int {
	return 1
}

func (q *Quote) Act(sr *Stream, m GenMode) GenMode {
	sr.currentSymbol = q.Self()
	return m
}

func (q *Quote) Match(e *Engine, r Element) bool {
	if r.Token() == q.V {
		return e.Matched3E(q.Self(), r, r)
	}
	return e.ResolveE(q.Self(), r)
}

func (q *Quote) ToVal() Element {
	return q.Self()
}

func (q *Quote) ToBool() bool {
	return q.Self().Token().ToBool()
}

func (q *Quote) Eqf(y Element) Element {
	return NewBoolean(y.Token() == q.Self().Token())
}

func (q *Quote) Nef(y Element) Element {
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
	//+ utf.encode ?
	return string(c.V)
}

func (c *Chr) ToTrace() string {
	return "'" + c.escaped(c.ToString()) + "'"
}

func (c *Chr) ToEncode() string {
	return utils.Encode(c.escaped(c.ToString()))
}

func (c *Chr) ToDump() string {
	return "d:" + utils.Encode(c.ToString())
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

func (z *ZLM) Append(x Element) Element {
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

type Sym = Symbol

func NewSym(x string) *Sym {
	el := MakeSelf[Sym]()
	el.V = x
	return el
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

func (s *Str) Weight() int {
	return 0
}

func (s *Str) NewLHS(m GenMode) GenMode {
	return NewLHModeFromElement(m, s.V, 0, m.ContextMode())
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

func (lb *LMBuffer) ToVal() Element {
	return lb.Self()
}

func (lb *LMBuffer) Append(x Element) Element {
	lb.V += x.ToString()
	return lb.Self()
}

func (lb *LMBuffer) ToString() string {
	return lb.V
}

// AArray is the table behind an array value. Keys keeps the keys in the
// order they were added, which is the order foreach visits them in.
type AArray struct {
	A    map[Element]Element
	Keys []Element
}

func NewAArray() *AArray {
	return &AArray{A: make(map[Element]Element)}
}

func (a *AArray) Set(k, v Element) {
	if _, ok := a.A[k]; !ok {
		a.Keys = append(a.Keys, k)
	}
	a.A[k] = v
}

type LMArray struct {
	GenericElement
	aa *AArray
	sx ScopeHolder
}

func NewLMArray(sr *Stream, s GenMode, z ScopeHolder) *LMArray {
	la := MakeSelf[LMArray]()
	la.aa = NewAArray()
	la.sx = s

	if la.sx == nil {
		panic("sx cannot be nil")
	}

	var v Element
	var i int
	var items []Element // the items, last first

	// cells [k: v] are assigned by key; the other items are counted, then
	// numbered 0..i-1 in order as they are popped (last first)
	sr.Operands().Each(func(o Element) bool {
		if o == sr.Engine.predefinedSymbols.mark {
			return false
		}
		if c, ok := o.(*LMCell); ok {
			la.Assign(sr, c)
		} else {
			i++
		}
		return true
	})
	for !sr.EmptyX() {
		v = sr.Popx()
		if v == sr.Engine.predefinedSymbols.mark {
			break
		}
		items = append(items, v)
		if _, ok := v.(*LMCell); !ok {
			i--
			la.AssignE(sr, i, v)
		}
	}

	// the keys in the order the items were written
	la.aa.Keys = la.aa.Keys[:0]
	seen := make(map[Element]bool)
	for j := len(items) - 1; j >= 0; j-- {
		var k Element
		if c, ok := items[j].(*LMCell); ok {
			k = sr.Engine.userSymbols.UniqueE(c.K)
		} else {
			k = sr.Engine.userSymbols.UniqueE(NewNumber(LMNumber(i)))
			i++
		}
		if !seen[k] {
			seen[k] = true
			la.aa.Keys = append(la.aa.Keys, k)
		}
	}

	return la
}

func (la *LMArray) ToVal() Element {
	return la.Self()
}

func (la *LMArray) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(la.Self())
	return s
}

// func (la *LMArray) Assign(e *Stream, c *LMCell) MachineElement {
func (la *LMArray) Assign(e *Stream, c Element) Element {
	if c != nil {
		lm, ok := c.(*LMCell)
		if !ok {
			panic("not lmcell")
		}
		la.aa.Set(e.Engine.userSymbols.UniqueE(lm.K), lm.V)
		return lm.V
	}
	return nil
}

func (la *LMArray) AssignE(e *Stream, i int, v Element) Element {
	la.aa.Set(e.Engine.userSymbols.UniqueE(NewNumber(LMNumber(i))), v)
	return v
}

func (la *LMArray) Idxf(y Element) Element {
	return NewARef(la.aa, la.sx.ScopeContextMode().State().engine.userSymbols.UniqueE(y.ToVal()), la.sx)
}

func (la *LMArray) Idtf(y Element) Element {
	return NewARef(la.aa, la.sx.ScopeContextMode().State().engine.userSymbols.UniqueE(y.ToVal()), la.sx)
}

type LMCell struct {
	GenericElement
	K Element
	V Element
}

func NewLMCell(y, z Element) *LMCell {
	el := MakeSelf[LMCell]()
	el.K = y
	el.V = z
	return el
}

func (lc *LMCell) ToString() string {
	return "LMCell:" + lc.K.ToString() + lc.V.ToString()
}

type NewVar struct {
	GenericElement
}

func NewNewVar() *NewVar {
	return MakeSelf[NewVar]()
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
	GenericElement
	K Element
}

func NewEachRef(x Element) *EachRef {
	el := MakeSelf[EachRef]()
	el.K = x
	return el
}

func (er *EachRef) Token() Element {
	return nil
}

func (er *EachRef) ToBody() []Element {
	return nil
}

func (er *EachRef) ToString() string {
	return "each " + er.K.ToString()
}

func (er *EachRef) Weight() int {
	return 0
}

func (er *EachRef) Act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.EachRef(s, er.K, s)
}

func (er *EachRef) Match(e *Engine, r Element) bool {
	return false
}

type AllRef struct {
	GenericElement
	K Element
}

func NewAllRef(x Element) *AllRef {
	el := MakeSelf[AllRef]()
	el.K = x
	return el
}

func (ar *AllRef) Token() Element {
	return nil
}

func (ar *AllRef) ToBody() []Element {
	return nil
}

func (ar *AllRef) ToString() string {
	return "all " + ar.K.ToString()
}

func (ar *AllRef) Weight() int {
	return 0
}

func (ar *AllRef) Act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.AllRef(s, ar.K, s)
}

func (ar *AllRef) Match(e *Engine, r Element) bool {
	return false
}

type VarSym struct {
	Symbol
}

func NewVarSym(x string) *VarSym {
	el := MakeSelf[VarSym]()
	el.V = x
	return el
}

func (vs *VarSym) Token() Element {
	return vs.Self()
}

func (vs *VarSym) ToDump() string {
	return "v:" + utils.Encode(vs.V)
}

func (vs *VarSym) Act(sr *Stream, s GenMode) GenMode {
	return vs.Self().Reference(sr, s, s)
}

func (vs *VarSym) Match(e *Engine, r Element) bool {
	return false
}

func (vs *VarSym) Reference(sr *Stream, s GenMode, x ScopeHolder) GenMode {
	return sr.Engine.TheRef(s, vs.Self(), x)
}

type DoneF struct {
	Symbol
}

func NewDoneF(x string) *DoneF {
	el := MakeSelf[DoneF]()
	el.V = x
	return el
}

func (df *DoneF) Match(e *Engine, r Element) bool {
	e.lhsStream.variables = e.lhsContext.Variables()
	return e.Matched3E(df.Self(), nil, nil)
}

type TakeF struct {
	Symbol
}

func NewTakeF(x string) *TakeF {
	el := MakeSelf[TakeF]()
	el.V = x
	return el
}

func (tf *TakeF) Match(e *Engine, r Element) bool {
	if r.Token() == tf.Self() { // %  %
		e.TakeTvar()
		return e.Matched3E(tf.Self(), r, nil)
	}
	if _, ok := r.(*BindF); ok { // %  :
		e.PushX()
		return e.Matched3E(tf.Self(), r, nil)
	}
	if e.rsLastMatchElement != nil { // %  [matched]
		e.PushR(e.rsLastMatchElement)
		return e.Matched3E(tf.Self(), nil, nil)
	}
	return false
}

type BindF struct {
	Symbol
}

func NewBindF(x string) *BindF {
	el := MakeSelf[BindF]()
	el.V = x
	return el
}

func (b *BindF) Match(e *Engine, r Element) bool {
	if r.Token() == b.Self() {
		a := e.lhsStream.Popx()
		bElem := e.rhsStream.Popx().ToVal()
		// tx("l", a); tx("r", bElem);
		if a, ok := a.(*VarSym); ok {
			e.BindUvar(a, bElem)
			return e.Matched3E(b.Self(), r, nil)
		}
		e.Matched3E(b.Self(), r, nil)
		lh := []Element{a}
		e.lhsStream.mode = NewSTModeFromElements(e.lhsStream.mode, lh, e.lhsStream.mode)
		rh := []Element{bElem}
		e.rhsStream.mode = NewSTModeFromElements(e.rhsStream.mode, rh, e.rhsStream.mode)
		return true
	}
	if _, ok := r.(*TakeF); ok {
		e.BindTvar()
		return e.Matched3E(b.Self(), r, nil)
	}
	if e.rsLastMatchElement != nil {
		e.BindXvarE(e.rsLastMatchElement)
		return e.Matched3E(b.Self(), nil, r)
	}
	return false
}

type AppendXSym struct {
	Symbol
}

func NewAppendXSym(x string) *AppendXSym {
	el := MakeSelf[AppendXSym]()
	el.V = x
	return el
}

// if there is captured material, append it
// otherwise match one symbol and append that
func (a *AppendXSym) Match(e *Engine, r Element) bool {
	b := e.lhsStream.Popx()
	if !e.lhsStream.EmptyX() {
		v := e.lhsStream.ToRow()
		for _, x := range v {
			b.Append(x)
		}
		e.Matched3E(a.Self(), nil, nil)
	} else {
		b.Append(r)
		e.Matched3E(a.Self(), r, r)
	}
	return true
}

type ErrSym struct {
	Symbol
	engine *Engine
}

func NewErrSym(e *Engine, x string) *ErrSym {
	el := MakeSelf[ErrSym]()
	el.V = x
	el.engine = e
	return el
}

func (e *ErrSym) Append(y Element) Element {
	e.engine.writeErr(y.ToString())
	return e.Self()
}

func (e *ErrSym) Match(engine *Engine, r Element) bool {
	e.engine.writeErr(r.ToString())
	return engine.Matched3E(e.Self(), r, r)
}

type OutSym struct {
	Symbol
	engine *Engine
}

func NewOutSym(e *Engine, x string) *OutSym {
	el := MakeSelf[OutSym]()
	el.V = x
	el.engine = e
	return el
}

func (o *OutSym) Append(y Element) Element {
	_, _ = o.engine.out.WriteString(y.ToString())
	return o.Self()
}

func (o *OutSym) Match(engine *Engine, r Element) bool {
	_, _ = o.engine.out.WriteString(r.ToString())
	return engine.Matched3E(o.Self(), r, r)
}

type UriSym struct {
	Symbol
	engine *Engine
}

func NewUriSym(e *Engine, x string) *UriSym {
	el := MakeSelf[UriSym]()
	el.V = x
	el.engine = e
	return el
}

func (u *UriSym) Append(y Element) Element {
	_, _ = u.engine.out.WriteString(y.ToEncode())
	return u.Self()
}

func (u *UriSym) Match(engine *Engine, r Element) bool {
	_, _ = u.engine.out.WriteString(r.ToEncode())
	return engine.Matched3E(u.Self(), r, r)
}

type UrdSym struct {
	Symbol
	engine *Engine
}

func NewUrdSym(e *Engine, x string) *UrdSym {
	el := MakeSelf[UrdSym]()
	el.V = x
	el.engine = e
	return el
}

func (u *UrdSym) Append(y Element) Element {
	_, _ = u.engine.out.WriteString(y.ToDecode())
	return u.Self()
}

func (u *UrdSym) Match(engine *Engine, r Element) bool {
	_, _ = u.engine.out.WriteString(r.ToDecode())
	return engine.Matched3E(u.Self(), r, r)
}

type SpSym struct {
	Symbol
}

func NewSpSym(x string) *SpSym {
	el := MakeSelf[SpSym]()
	el.V = x
	return el
}

func (s *SpSym) ToEncode() string {
	return " "
}

type NlSym struct {
	Symbol
}

func NewNlSym(x string) *NlSym {
	el := MakeSelf[NlSym]()
	el.V = x
	return el
}

func (n *NlSym) ToEncode() string {
	return "\n"
}

type GetF struct {
	Symbol
}

func NewGetF(x string) *GetF {
	el := MakeSelf[GetF]()
	el.V = x
	return el
}

func (g *GetF) Act(sr *Stream, s GenMode) GenMode {
	return sr.Getx(s)
}

type TrueSym struct {
	Symbol
}

func NewTrueSym(x string) *TrueSym {
	el := MakeSelf[TrueSym]()
	el.V = x
	return el
}

func (t *TrueSym) ToVal() Element {
	return NewBoolean(true)
}

type FalseSym struct {
	Symbol
}

func NewFalseSym(x string) *FalseSym {
	el := MakeSelf[FalseSym]()
	el.V = x
	return el
}

func (f *FalseSym) ToVal() Element {
	return NewBoolean(false)
}

type TrueF struct {
	Symbol
}

func NewTrueF(x string) *TrueF {
	el := MakeSelf[TrueF]()
	el.V = x
	return el
}

func (t *TrueF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewBoolean(true))
	return s
}

type FalseF struct {
	Symbol
}

func NewFalseF(x string) *FalseF {
	el := MakeSelf[FalseF]()
	el.V = x
	return el
}

func (f *FalseF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewBoolean(false))
	return s
}

type GetXF struct {
	Symbol
	V Element
}

func NewGetXF(x Element) *GetXF {
	el := MakeSelf[GetXF]()
	el.V = x
	return el
}

func (g *GetXF) ToTrace() string {
	return g.V.ToTrace() + "p"
}

func (g *GetXF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(g.V)
	return s
}

type GetVF struct {
	Symbol
	V Element
}

func NewGetVF(x Element) *GetVF {
	el := MakeSelf[GetVF]()
	el.V = x
	return el
}

func (g *GetVF) ToTrace() string {
	return g.V.ToTrace()
}

func (g *GetVF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewLMRefFromElement(g.V, s))
	return s
}

type ActF struct {
	Symbol
}

func NewActF(x string) *ActF {
	el := MakeSelf[ActF]()
	el.V = x
	return el
}

func (a *ActF) Trace(sr *Stream, t *Tracer) {
	t.TraceAct(sr, a.Self())
}

func (a *ActF) Act(sr *Stream, s GenMode) GenMode {
	fail("the act primitive (a) is not supported")
	return nil
}

type Primitive struct {
	Symbol
}

func NewPrimitive() *Primitive {
	return MakeSelf[Primitive]()
}

func NewPrimitiveFromString(x string) *Primitive {
	el := NewPrimitive()
	el.V = x
	return el
}

func (p *Primitive) Act(sr *Stream, s GenMode) GenMode {
	sr.Engine.printf("act: %s\n", string(p.V))
	return s
}

type ApplyF struct {
	Primitive
}

func NewApplyF(x string) *ApplyF {
	return ReSelf(&ApplyF{Primitive: *NewPrimitiveFromString(x)})
}

func (a *ApplyF) Trace(s *Stream, t *Tracer) {
	t.TraceApply(s, a.Self())
}

func (a *ApplyF) Act(sr *Stream, s GenMode) GenMode {
	v := sr.Popx()
	return v.Act(sr, s)
}

type InjF struct {
	Symbol
}

func NewInjF(x string) *InjF {
	el := MakeSelf[InjF]()
	el.V = x
	return el
}

func (i *InjF) Match(e *Engine, r Element) bool {
	e.PushRhx(e.lhsStream.Popx())
	return e.Matched3E(i.Self(), nil, nil)
}

type StrF struct {
	Symbol
}

func NewStrF(x string) *StrF {
	el := MakeSelf[StrF]()
	el.V = x
	return el
}

func (s *StrF) Act(sr *Stream, mode GenMode) GenMode {
	//  { return new stMode(s, (cast(str)sr.popx()).v, s); }
	return mode
}

type Anything struct {
	Symbol
}

func NewAnything(x string) *Anything {
	el := MakeSelf[Anything]()
	el.V = x
	return el
}

func (a *Anything) Match(e *Engine, r Element) bool {
	e.Matched3E(a.Self(), r, r)
	return true
}

type AnySym struct {
	Anything
}

func NewAnySym(x string) *AnySym {
	return ReSelf(&AnySym{Anything: *NewAnything(x)})
}

func (a *AnySym) Match(e *Engine, r Element) bool {
	if _, ok := r.Token().(*Sym); ok {
		e.Matched3E(a.Self(), r, r)
		return true
	} else {
		return e.ResolveE(a.Self(), r)
	}
}

type AnyChr struct {
	Anything
}

func NewAnyChr(x string) *AnyChr {
	return ReSelf(&AnyChr{Anything: *NewAnything(x)})
}

func (a *AnyChr) Match(e *Engine, r Element) bool {
	if _, ok := r.Token().(*Chr); ok {
		e.Matched3E(a.Self(), r, r)
		return true
	} else {
		return e.ResolveE(a.Self(), r)
	}
}

type LnoSym struct {
	Symbol
}

func NewLnoSym(x string) *LnoSym {
	return ReSelf(&LnoSym{Symbol: *NewSymbol(x)})
}

func (l *LnoSym) Match(e *Engine, r Element) bool {
	return e.Matched3E(l.Self(), nil, NewNumber(LMNumber(e.Lineno())))
}

type IfnSym struct {
	Symbol
}

func NewIfnSym(x string) *IfnSym {
	return ReSelf(&IfnSym{Symbol: *NewSymbol(x)})
}

func (i *IfnSym) Match(e *Engine, r Element) bool {
	return e.Matched3E(i, nil, NewSym(e.Filename()))
}

type FlagSym struct {
	Symbol
}

func NewFlagSym(x string) *FlagSym {
	return ReSelf(&FlagSym{Symbol: *NewSymbol(x)})
}

func (f *FlagSym) Match(e *Engine, r Element) bool {
	e.flagErrors++
	message := fmt.Sprintf("%s:%d: ", e.Filename(), e.Lineno())
	return e.Matched3E(f.Self(), nil, NewSym(message))
}

type WarnSym struct {
	Symbol
}

func NewWarnSym(x string) *WarnSym {
	return ReSelf(&WarnSym{Symbol: *NewSymbol(x)})
}

func (w *WarnSym) Match(e *Engine, r Element) bool {
	e.warnErrors++
	message := fmt.Sprintf("%s:%d: ", e.Filename(), e.Lineno())
	return e.Matched3E(w.Self(), nil, NewSym(message))
}

type RepnSym struct {
	Symbol
}

func NewRepnSym(x string) *RepnSym {
	return ReSelf(&RepnSym{Symbol: *NewSymbol(x)})
}

func (r *RepnSym) Match(e *Engine, _ Element) bool {
	n := e.lhsStream.Popx().ToVal().(*Number)
	return e.Repeat(n.ToInt())
}

type RepSym struct {
	Symbol
}

func NewRepSym(x string) *RepSym {
	return ReSelf(&RepSym{Symbol: *NewSymbol(x)})
}

func (r *RepSym) Match(e *Engine, _ Element) bool {
	return e.Repeat(0)
}

type OptSym struct {
	Symbol
}

func NewOptSym(x string) *OptSym {
	return ReSelf(&OptSym{Symbol: *NewSymbol(x)})
}

func (o *OptSym) Match(e *Engine, r Element) bool {
	return e.Repeat(1)
}

type Lex struct {
	Symbol
	Table     map[Element]Element
	Inclusive bool
}

func NewLex(x string) *Lex {
	lex := ReSelf(&Lex{Symbol: *NewSymbol(x)})
	lex.Table = make(map[Element]Element)
	lex.Inclusive = true

	return lex
}

func NewLexFromEngine(s string, e *Engine) *Lex {
	l := NewLex(s)

	state := IN
	var prevc int

	// s is "[...]": walk the characters between the brackets
	body := []rune(s)
	if len(body) >= 2 {
		body = body[1 : len(body)-1]
	} else {
		body = nil
	}
	for _, r := range body {
		var x Element
		c := int(r)

		if state == IN {
			state = C1
			if c == '^' {
				l.Inclusive = false
				continue
			}
		}

		switch state {
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
			state = C2

		case C2:
			switch c {
			case '\\':
				state = E1
			case '-':
				state = RN
			default:
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
	return "[" + utils.Encode(l.V[1:len(l.V)-1]) + "]"
}

func (l *Lex) ToString1() string {
	return "lex(" + l.V + ")"
}

func (l *Lex) AddRule(g *Grammar, x *Rule) {
	if l.Inclusive {
		for k := range l.Table {
			g.Add(x.Additional(k))
		}
	} else {
		fail("a rule cannot start with the negated lexical class %s", l.ToTrace())
	}
}

func (l *Lex) Match(e *Engine, r Element) bool {
	if r.Token() == l.Self() {
		return e.Matched3E(l.Self(), r, nil)
	}
	if chr, ok := r.(*Chr); ok && (l.Inclusive != (l.Table[chr] == nil)) {
		return e.Matched3E(l.Self(), r, r)
	}
	return e.ResolveE(l.Self(), r)
}

type DropF struct {
	Primitive
}

func NewDropF(x string) *DropF {
	return ReSelf(&DropF{Primitive: *NewPrimitiveFromString(x)})
}

func (d *DropF) Act(sr *Stream, s GenMode) GenMode {
	sr.ClearX()
	return s
}

type Unary struct {
	Primitive
}

func NewUnary(x string) *Unary {
	return ReSelf(&Unary{Primitive: *NewPrimitiveFromString(x)})
}

func (u *Unary) Result1(x Element) Element {
	return nil
}

func (u *Unary) Trace(s *Stream, t *Tracer) {
	t.TraceArithmetic(s, u.Self())
}

func (u *Unary) Act(sr *Stream, s GenMode) GenMode {
	x := sr.Popx()
	sr.Pushx(u.Self().Result1(x.ToVal()))
	return s
}

type Arithmetic struct {
	Primitive
}

func NewArithmetic(x string) *Arithmetic {
	return ReSelf(&Arithmetic{Primitive: *NewPrimitiveFromString(x)})
}

func (a *Arithmetic) Result2(x, y Element) Element {
	return nil
}

func (a *Arithmetic) Trace(s *Stream, t *Tracer) {
	t.TraceArithmetic(s, a.Self())
}

func (a *Arithmetic) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(a.Self().Result2(x.ToVal(), y.ToVal()))
	return b
}

type Relation struct {
	Primitive
}

func NewRelation(x string) *Relation {
	return ReSelf(&Relation{Primitive: *NewPrimitiveFromString(x)})
}

func (r *Relation) Result2(x, y Element) Element {
	return nil
}

func (r *Relation) Trace(s *Stream, t *Tracer) {
	t.TraceRelation(s, r.Self())
}

func (r *Relation) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(r.Self().Result2(x.ToVal(), y.ToVal()))
	return b
}

type Assignment struct {
	Primitive
}

func NewAssignment(x string) *Assignment {
	return ReSelf(&Assignment{Primitive: *NewPrimitiveFromString(x)})
}

func (a *Assignment) Result2(x, y Element) Element {
	return nil
}

func (a *Assignment) Trace(s *Stream, t *Tracer) {
	t.TraceAssignment(s, a.Self())
}

func (a *Assignment) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(a.Self().Result2(x, y.ToVal()))
	return b
}

type IncDec struct {
	Primitive
}

func NewIncDec(x string) *IncDec {
	return ReSelf(&IncDec{Primitive: *NewPrimitiveFromString(x)})
}

func (i *IncDec) Result1(x Element) Element {
	return nil
}

func (i *IncDec) Trace(s *Stream, t *Tracer) {
	t.TraceAssignment(s, i)
}

func (i *IncDec) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(i.Self().Result1(sr.Popx()))
	return b
}

type Iff struct {
	Primitive
}

func NewIff(x string) *Iff {
	return ReSelf(&Iff{Primitive: *NewPrimitiveFromString(x)})
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
	return ReSelf(&OrOrf{Primitive: *NewPrimitiveFromString(x)})
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
	return ReSelf(&AndAndf{Primitive: *NewPrimitiveFromString(x)})
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
	return ReSelf(&Index{Primitive: *NewPrimitiveFromString(x)})
}

func (i *Index) Result2(x, y Element) Element {
	return nil
}

func (i *Index) Trace(s *Stream, t *Tracer) {
	t.TraceIndex(s, i.Self())
}

func (i *Index) Act(sr *Stream, b GenMode) GenMode {
	// sr.Dumpx()
	y := sr.Popx()
	x := sr.Popx()
	sr.Pushx(i.Self().Result2(x, y.ToVal()))
	return b
}

type Cellf struct {
	Primitive
}

func NewCellf(x string) *Cellf {
	return ReSelf(&Cellf{Primitive: *NewPrimitiveFromString(x)})
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
	return ReSelf(&Arrayf{Primitive: *NewPrimitiveFromString(x)})
}

func (a *Arrayf) Trace(s *Stream, t *Tracer) {
	t.TraceIndex(s, a.Self())
}

func (a *Arrayf) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(NewLMArray(sr, b, b))
	return b
}

type Argsf struct {
	Primitive
}

func NewArgsf(x string) *Argsf {
	return ReSelf(&Argsf{Primitive: *NewPrimitiveFromString(x)})
}

func (a *Argsf) Act(sr *Stream, b GenMode) GenMode {
	sr.Pushx(sr.Engine.predefinedSymbols.mark)
	return b
}

type Funf struct {
	Primitive
}

func NewFunf(x string) *Funf {
	return ReSelf(&Funf{Primitive: *NewPrimitiveFromString(x)})
}

func (f *Funf) Act(sr *Stream, b GenMode) GenMode {
	v := sr.ToArgv(sr.Engine.predefinedSymbols.mark)
	sr.Pushx(sr.Engine.externalSystem.Call(sr, b, v[0], v))
	return b
}

type Loopf struct {
	Primitive
}

func NewLoopf(x string) *Loopf {
	return ReSelf(&Loopf{Primitive: *NewPrimitiveFromString(x)})
}

func (l *Loopf) Trace(s *Stream, t *Tracer) {
	t.TraceLoop(s, l.Self())
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
	return ReSelf(&Testf{Primitive: *NewPrimitiveFromString(x)})
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

// loopMode finds the loop that a break or continue in m belongs to. It does
// not look past the start of a rule side, so a loop in another rule is never
// found.
func loopMode(m GenMode, what string) *RPMode {
	for x := m; x != nil; x = x.StackMode() {
		switch y := x.(type) {
		case *RPMode:
			return y
		case *LHMode, *RHMode, *LZMode, *RZMode:
			fail("%s outside a loop", what)
		}
	}
	fail("%s outside a loop", what)
	return nil
}

// Forf runs a for loop: I ( <E> f:test B ) G ( N ) G f:for. The step N
// follows the body, so that continue can resume at it.
type Forf struct {
	Primitive
}

func NewForf(x string) *Forf {
	return ReSelf(&Forf{Primitive: *NewPrimitiveFromString(x)})
}

func (f *Forf) Trace(s *Stream, t *Tracer) {
	t.TraceLoop(s, f.Self())
}

func (f *Forf) Act(sr *Stream, b GenMode) GenMode {
	next := sr.Popx().ToVal().(*Str).V
	body := sr.Popx().ToVal().(*Str).V
	v := make([]Element, 0, len(body)+len(next))
	v = append(append(v, body...), next...)
	m := NewRPModeFromElement(b, v)
	m.next = len(body)
	return m
}

// Breakf ends the innermost loop.
type Breakf struct {
	Primitive
}

func NewBreakf(x string) *Breakf {
	return ReSelf(&Breakf{Primitive: *NewPrimitiveFromString(x)})
}

func (f *Breakf) Act(sr *Stream, b GenMode) GenMode {
	return loopMode(b, "break").Return()
}

// Continuef starts the next iteration of the innermost loop, at the step of a
// for loop or at the test of a while loop.
type Continuef struct {
	Primitive
}

func NewContinuef(x string) *Continuef {
	return ReSelf(&Continuef{Primitive: *NewPrimitiveFromString(x)})
}

func (f *Continuef) Act(sr *Stream, b GenMode) GenMode {
	m := loopMode(b, "continue")
	// return from the modes inside the loop (if blocks), back to its body
	for x := b; x != GenMode(m); {
		y := x.StackMode()
		if y == GenMode(m) {
			x.Return()
		}
		x = y
	}
	sr.codeIndex = m.next
	return m
}

// Rulef defines a rule while the rules run: rule(G, P) { lhs <- rhs }
// compiles to G P n:N ( lhs ) G ( rhs ) G f:rule, the operands of the r
// opcode. P is encoded as it is there. The value is the grammar symbol.
type Rulef struct {
	Primitive
}

func NewRulef(x string) *Rulef {
	return ReSelf(&Rulef{Primitive: *NewPrimitiveFromString(x)})
}

func (f *Rulef) Act(sr *Stream, b GenMode) GenMode {
	v := make([]Element, 5)
	for i := len(v) - 1; i >= 0; i-- {
		v[i] = sr.Popx().ToVal()
	}
	if len(v[3].ToBody()) == 0 {
		fail("rule(%s, %s) has an empty left side", v[0].ToString(), v[1].ToString())
	}
	sr.Engine.AddRule(v, "rule", sr.Engine.ruleNumbers)
	sr.Pushx(v[0])
	return b
}

// EachX is each (expr), the E opcode: the value of expr names the variable.
type EachX struct {
	Primitive
}

func NewEachX(x string) *EachX {
	return ReSelf(&EachX{Primitive: *NewPrimitiveFromString(x)})
}

func (x *EachX) Act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.EachRef(s, sr.Engine.varKey(sr.Popx()), s)
}

// AllX is all (expr), the bare B opcode.
type AllX struct {
	Primitive
}

func NewAllX(x string) *AllX {
	return ReSelf(&AllX{Primitive: *NewPrimitiveFromString(x)})
}

func (x *AllX) Act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.AllRef(s, sr.Engine.varKey(sr.Popx()), s)
}

type SelF struct {
	Primitive
}

func NewSelF(x string) *SelF {
	return ReSelf(&SelF{Primitive: *NewPrimitiveFromString(x)})
}

// c ? a : b compiles to <c> ( <a> ) G ( <b> ) G f:sel
func (s *SelF) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	t := sr.Popx().ToVal()

	if t.ToBool() {
		return NewSTModeFromElements(b, x.ToVal().(*Str).V, b)
	}
	return NewSTModeFromElements(b, y.ToVal().(*Str).V, b)
}

// Foreachf runs foreach (K, V; E) B, compiled to <K> <V> <E> ( B ) G
// f:foreach, where <K> and <V> are variable references (<K> is null in
// foreach (V; E)). For each key of the array E, in the order the keys were
// added, it assigns the key to K and the value to V and runs B. Keys added by
// B are not visited.
type Foreachf struct {
	Primitive
}

func NewForeachf(x string) *Foreachf {
	return ReSelf(&Foreachf{Primitive: *NewPrimitiveFromString(x)})
}

func (f *Foreachf) Trace(s *Stream, t *Tracer) {
	t.TraceLoop(s, f.Self())
}

func (f *Foreachf) Act(sr *Stream, b GenMode) GenMode {
	body := sr.Popx().ToVal().(*Str).V
	e := sr.Popx().ToVal()
	step := NewForeachStep(sr.Popx(), sr.Popx())
	switch a := e.(type) {
	case *LMArray:
		step.a = a.aa
		step.keys = slices.Clone(a.aa.Keys)
	case *ZLM:
	default:
		f.InvalidOp("foreach over " + e.ToString())
	}
	v := make([]Element, 0, len(body)+1)
	v = append(append(v, step), body...)
	return NewRPModeFromElement(b, v)
}

// ForeachStep starts each pass of a foreach loop, so continue goes to the
// next key; after the last key it ends the loop.
type ForeachStep struct {
	GenericElement
	k, v Element // the loop variables (references); k may be null
	a    *AArray
	keys []Element
	i    int
}

func NewForeachStep(v, k Element) *ForeachStep {
	el := MakeSelf[ForeachStep]()
	el.k = k
	el.v = v
	return el
}

func (fs *ForeachStep) ToString() string {
	return "foreach step"
}

func (fs *ForeachStep) Act(sr *Stream, m GenMode) GenMode {
	if fs.i >= len(fs.keys) {
		return m.Ends()
	}
	key := fs.keys[fs.i]
	fs.i++
	if r, ok := fs.k.(*LMRef); ok {
		r.StoValf(key)
	}
	if r, ok := fs.v.(*LMRef); ok {
		r.StoValf(fs.a.A[key])
	}
	return m
}

type Retf struct {
	Primitive
}

func NewRetf(x string) *Retf {
	return ReSelf(&Retf{Primitive: *NewPrimitiveFromString(x)})
}

type Lamdaf struct {
	Primitive
}

func NewLamdaf(x string) *Lamdaf {
	return ReSelf(&Lamdaf{Primitive: *NewPrimitiveFromString(x)})
}

type Specf struct {
	Primitive
}

func NewSpecf(x string) *Specf {
	return ReSelf(&Specf{Primitive: *NewPrimitiveFromString(x)})
}

type Idxf struct {
	Index
}

func NewIdxf(x string) *Idxf {
	return ReSelf(&Idxf{Index: *NewIndex(x)})
}
func (i *Idxf) Result2(x Element, y Element) Element {
	return x.Idxf(y)
}

type Idtf struct {
	Index
}

func NewIdtf(x string) *Idtf {
	return ReSelf(&Idtf{Index: *NewIndex(x)})
}
func (i *Idtf) Result2(x Element, y Element) Element {
	return x.Idtf(y)
}

type StoValf struct {
	Assignment
}

func NewStoValf(x string) *StoValf {
	return ReSelf(&StoValf{Assignment: *NewAssignment(x)})
}
func (s *StoValf) Result2(x Element, y Element) Element {
	return x.StoValf(y)
}

type StoAddf struct {
	Assignment
}

func NewStoAddf(x string) *StoAddf {
	return ReSelf(&StoAddf{Assignment: *NewAssignment(x)})
}
func (s *StoAddf) Result2(x Element, y Element) Element {
	return x.StoAddf(y)
}

type StoSubf struct {
	Assignment
}

func NewStoSubf(x string) *StoSubf {
	return ReSelf(&StoSubf{Assignment: *NewAssignment(x)})
}
func (s *StoSubf) Result2(x Element, y Element) Element {
	return x.StoSubf(y)
}

type StoMulf struct {
	Assignment
}

func NewStoMulf(x string) *StoMulf {
	return ReSelf(&StoMulf{Assignment: *NewAssignment(x)})
}
func (s *StoMulf) Result2(x Element, y Element) Element {
	return x.StoMulf(y)
}

type StoDivf struct {
	Assignment
}

func NewStoDivf(x string) *StoDivf {
	return ReSelf(&StoDivf{Assignment: *NewAssignment(x)})
}
func (s *StoDivf) Result2(x Element, y Element) Element {
	return x.StoDivf(y)
}

type StoModf struct {
	Assignment
}

func NewStoModf(x string) *StoModf {
	return ReSelf(&StoModf{Assignment: *NewAssignment(x)})
}
func (s *StoModf) Result2(x Element, y Element) Element {
	return x.StoModf(y)
}

type Eeqf struct {
	Relation
}

func NewEeqf(x string) *Eeqf {
	return ReSelf(&Eeqf{Relation: *NewRelation(x)})
}
func (e *Eeqf) Result2(x Element, y Element) Element {
	return x.Eeqf(y)
}

type Neef struct {
	Relation
}

func NewNeef(x string) *Neef {
	return ReSelf(&Neef{Relation: *NewRelation(x)})
}
func (n *Neef) Result2(x Element, y Element) Element {
	return x.Neef(y)
}

type Inf struct {
	Relation
}

func NewInf(x string) *Inf {
	return ReSelf(&Inf{Relation: *NewRelation(x)})
}
func (i *Inf) Result2(x Element, y Element) Element {
	return x.Inf(y)
}

type Eqf struct {
	Relation
}

func NewEqf(x string) *Eqf {
	return ReSelf(&Eqf{Relation: *NewRelation(x)})
}
func (e Eqf) Result2(x Element, y Element) Element {
	return x.Eqf(y)
}

type Nef struct {
	Relation
}

func NewNef(x string) *Nef {
	return ReSelf(&Nef{Relation: *NewRelation(x)})
}
func (n *Nef) Result2(x Element, y Element) Element {
	return x.Nef(y)
}

type Ltf struct {
	Relation
}

func NewLtf(x string) *Ltf {
	return ReSelf(&Ltf{Relation: *NewRelation(x)})
}
func (l *Ltf) Result2(x Element, y Element) Element {
	return x.Ltf(y)
}

type Gtf struct {
	Relation
}

func NewGtf(x string) *Gtf {
	return ReSelf(&Gtf{Relation: *NewRelation(x)})
}
func (g *Gtf) Result2(x Element, y Element) Element {
	return x.Gtf(y)
}

type Lef struct {
	Relation
}

func NewLef(x string) *Lef {
	return ReSelf(&Lef{Relation: *NewRelation(x)})
}
func (l *Lef) Result2(x Element, y Element) Element {
	return x.Lef(y)
}

type Gef struct {
	Relation
}

func NewGef(x string) *Gef {
	return ReSelf(&Gef{Relation: *NewRelation(x)})
}
func (g *Gef) Result2(x Element, y Element) Element {
	return x.Gef(y)
}

type BitXorf struct {
	Arithmetic
}

func NewBitXorf(x string) *BitXorf {
	return ReSelf(&BitXorf{Arithmetic: *NewArithmetic(x)})
}
func (b *BitXorf) Result2(x Element, y Element) Element {
	return x.BitXorf(y)
}

type BitOrf struct {
	Arithmetic
}

func NewBitOrf(x string) *BitOrf {
	return ReSelf(&BitOrf{Arithmetic: *NewArithmetic(x)})
}
func (b *BitOrf) Result2(x Element, y Element) Element {
	return x.BitOrf(y)
}

type BitAndf struct {
	Arithmetic
}

func NewBitAndf(x string) *BitAndf {
	return ReSelf(&BitAndf{Arithmetic: *NewArithmetic(x)})
}
func (b *BitAndf) Result2(x Element, y Element) Element {
	return x.BitAndf(y)
}

type Addf struct {
	Arithmetic
}

func NewAddf(x string) *Addf {
	return ReSelf(&Addf{Arithmetic: *NewArithmetic(x)})
}
func (a *Addf) Result2(x Element, y Element) Element {
	return x.Addf(y)
}

type Subf struct {
	Arithmetic
}

func NewSubf(x string) *Subf {
	return ReSelf(&Subf{Arithmetic: *NewArithmetic(x)})
}
func (s *Subf) Result2(x Element, y Element) Element {
	return x.Subf(y)
}

type Mulf struct {
	Arithmetic
}

func NewMulf(x string) *Mulf {
	return ReSelf(&Mulf{Arithmetic: *NewArithmetic(x)})
}
func (m *Mulf) Result2(x Element, y Element) Element {
	return x.Mulf(y)
}

type Divf struct {
	Arithmetic
}

func NewDivf(x string) *Divf {
	return ReSelf(&Divf{Arithmetic: *NewArithmetic(x)})
}
func (d *Divf) Result2(x Element, y Element) Element {
	return x.Divf(y)
}

type Modf struct {
	Arithmetic
}

func NewModf(x string) *Modf {
	return ReSelf(&Modf{Arithmetic: *NewArithmetic(x)})
}
func (m *Modf) Result2(x Element, y Element) Element {
	return x.Modf(y)
}

type Preincf struct {
	IncDec
}

func NewPreincf(x string) *Preincf {
	return ReSelf(&Preincf{IncDec: *NewIncDec(x)})
}
func (p *Preincf) Result1(x Element) Element {
	return x.Preincf()
}

type Predecf struct {
	IncDec
}

func NewPredecf(x string) *Predecf {
	return ReSelf(&Predecf{IncDec: *NewIncDec(x)})
}
func (p *Predecf) Result1(x Element) Element {
	return x.Predecf()
}

type Postincf struct {
	IncDec
}

func NewPostincf(x string) *Postincf {
	return ReSelf(&Postincf{IncDec: *NewIncDec(x)})
}
func (p *Postincf) Result1(x Element) Element {
	return x.Postincf()
}

type Postdecf struct {
	IncDec
}

func NewPostdecf(x string) *Postdecf {
	return ReSelf(&Postdecf{IncDec: *NewIncDec(x)})
}
func (p *Postdecf) Result1(x Element) Element {
	return x.Postdecf()
}

type Negf struct {
	Unary
}

func NewNegf(x string) *Negf {
	return ReSelf(&Negf{Unary: *NewUnary(x)})
}
func (n *Negf) Result1(x Element) Element {
	return x.Negf()
}

type Notf struct {
	Unary
}

func NewNotf(x string) *Notf {
	return ReSelf(&Notf{Unary: *NewUnary(x)})
}
func (n *Notf) Result1(x Element) Element {
	return x.Notf()
}

type Invf struct {
	Unary
}

func NewInvf(x string) *Invf {
	return ReSelf(&Invf{Unary: *NewUnary(x)})
}
func (i *Invf) Result1(x Element) Element {
	return x.Invf()
}

type IOSymbol struct {
	Symbol
	H GrammarSystem
}

func NewIOSymbol(x string, handler GrammarSystem) *IOSymbol {
	iosymbol := ReSelf(&IOSymbol{Symbol: *NewSymbol(x)})
	iosymbol.H = handler
	handler.SetSymbol(iosymbol)
	return iosymbol
}

func (i *IOSymbol) Match(e *Engine, r Element) bool {
	return i.H.Match(e, i, r)
}
