package machine

import (
	"math"
	"os"

	"github.com/msorc/languagemachine2/internal/conv"
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
	return conv.Encode(e.Self().ToString())
}

func (e *GenericElement) ToDecode() string {
	return decodeURI(e.Self().ToString())
}

func decodeURI(s string) string {
	d, err := conv.Decode(s)
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
