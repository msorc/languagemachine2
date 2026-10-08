package machine

import (
	"math"
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
)

// LMNumber is the machine's number, a double as in the original.
type LMNumber float64

// Element is a symbol, a value or an instruction of the machine. Elements
// appear on both sides of rules and on the input; the engine matches them
// (Match), runs them in a mode (Act) and computes with them (Operand).
//
// Every element type embeds GenericElement, which gives a default for each
// method, and reaches the outermost type through Self (see self_pointer.go).
// An element type that has something to trace also implements traceable.
type Element interface {
	SelfPointer[Element]

	// The grammar engine.
	AddRule(*Grammar, *Rule)                            // file a rule whose left side starts with this element
	Match(*Engine, Element) bool                        // match this goal against an input symbol
	Compare(*Engine, Element) bool                      // equality for the default Match
	Act(*Stream, GenMode) GenMode                       // run as an instruction in a mode
	NewRHX(GenMode, ContextHolder, ScopeHolder) GenMode // a mode that produces this element's body
	Reference(*Stream, GenMode, ScopeHolder) GenMode    // act as the value of a variable
	Token() Element                                     // the symbol rules are filed under
	Weight() int                                        // length counted for rule ordering
	ToBody() []Element                                  // the elements of a list, or nil

	// Values.
	ToVal() Element
	ToVar() VarElement // the variable, or nil
	ToNumber() LMNumber
	IsNumber() bool
	ToInt() int
	ToBool() bool

	// Text.
	ToString() string
	ToTrace() string
	ToEncode() string
	ToDecode() string
	ToDump() string

	Operand
}

// Operand is the arithmetic of elements, called by the operators with the
// stream they act on. GenericElement reports each as invalid; numbers,
// symbols, variables, arrays and buffers implement what they support.
type Operand interface {
	Append(*Stream, Element) Element
	Inf(*Stream, Element) Element
	Idxf(*Stream, Element) Element
	Idtf(*Stream, Element) Element
	StoValf(*Stream, Element) Element
	StoAddf(*Stream, Element) Element
	StoSubf(*Stream, Element) Element
	StoMulf(*Stream, Element) Element
	StoDivf(*Stream, Element) Element
	StoModf(*Stream, Element) Element
	Eeqf(*Stream, Element) Element
	Neef(*Stream, Element) Element
	Eqf(*Stream, Element) Element
	Nef(*Stream, Element) Element
	Ltf(*Stream, Element) Element
	Gtf(*Stream, Element) Element
	Lef(*Stream, Element) Element
	Gef(*Stream, Element) Element
	BitXorf(*Stream, Element) Element
	BitOrf(*Stream, Element) Element
	BitAndf(*Stream, Element) Element
	Addf(*Stream, Element) Element
	Subf(*Stream, Element) Element
	Mulf(*Stream, Element) Element
	Divf(*Stream, Element) Element
	Modf(*Stream, Element) Element
	Preincf(*Stream) Element
	Predecf(*Stream) Element
	Postincf(*Stream) Element
	Postdecf(*Stream) Element
	Negf(*Stream) Element
	Notf(*Stream) Element
	Invf(*Stream) Element
}

// traceable is an element that writes a trace line before it acts
// (operators and statements, under their trace categories).
type traceable interface {
	Trace(*Stream, *Tracer)
}

// GenericElement gives every element its default behaviour.
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
	}
	return engine.ResolveE(e.Self(), r)
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

// ToVar of an element that is not a variable is nil; callers that need a
// variable check for it.
func (e *GenericElement) ToVar() VarElement {
	return nil
}

// ToInt of a non-number is 0, as C's conversions of unparsable text are.
func (e *GenericElement) ToInt() int {
	return 0
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

func (e *GenericElement) Reference(sr *Stream, s GenMode, x ScopeHolder) GenMode {
	return e.Self().Act(sr, s)
}

// invalidOp reports an operation that x does not support on the engine's
// error output, after the output written so far, and gives null.
func invalidOp(sr *Stream, op string, x Element) Element {
	var b strings.Builder
	TxE(&b, "BAD "+op, x)
	sr.Engine.writeErr(b.String())
	return Null()
}

func (e *GenericElement) ToVal() Element {
	return e.Self()
}

func (e *GenericElement) Append(sr *Stream, y Element) Element {
	return invalidOp(sr, "~=", e.Self())
}

func (e *GenericElement) Inf(sr *Stream, y Element) Element {
	return Null()
}

func (e *GenericElement) Idxf(sr *Stream, y Element) Element {
	return Null()
}

func (e *GenericElement) Idtf(sr *Stream, y Element) Element {
	return Null()
}

func (e *GenericElement) StoValf(sr *Stream, y Element) Element {
	return invalidOp(sr, "=", e.Self())
}

func (e *GenericElement) StoAddf(sr *Stream, y Element) Element {
	return invalidOp(sr, "+=", e.Self())
}

func (e *GenericElement) StoSubf(sr *Stream, y Element) Element {
	return invalidOp(sr, "-=", e.Self())
}

func (e *GenericElement) StoMulf(sr *Stream, y Element) Element {
	return invalidOp(sr, "*=", e.Self())
}

func (e *GenericElement) StoDivf(sr *Stream, y Element) Element {
	return invalidOp(sr, "/=", e.Self())
}

func (e *GenericElement) StoModf(sr *Stream, y Element) Element {
	return invalidOp(sr, "%=", e.Self())
}

func (e *GenericElement) Eeqf(sr *Stream, y Element) Element {
	return NewBoolean(y.Token() == e.Self().Token())
}

func (e *GenericElement) Neef(sr *Stream, y Element) Element {
	return NewBoolean(y.Token() != e.Self().Token())
}

func (e *GenericElement) Eqf(sr *Stream, y Element) Element {
	return invalidOp(sr, "==", e.Self())
}

func (e *GenericElement) Nef(sr *Stream, y Element) Element {
	return invalidOp(sr, "!=", e.Self())
}

func (e *GenericElement) Ltf(sr *Stream, y Element) Element {
	return invalidOp(sr, "<", e.Self())
}

func (e *GenericElement) Gtf(sr *Stream, y Element) Element {
	return invalidOp(sr, ">", e.Self())
}

func (e *GenericElement) Lef(sr *Stream, y Element) Element {
	return invalidOp(sr, "<=", e.Self())
}

func (e *GenericElement) Gef(sr *Stream, y Element) Element {
	return invalidOp(sr, ">=", e.Self())
}

func (e *GenericElement) BitXorf(sr *Stream, y Element) Element {
	return invalidOp(sr, "^", e.Self())
}

func (e *GenericElement) BitOrf(sr *Stream, y Element) Element {
	return invalidOp(sr, "|", e.Self())
}

func (e *GenericElement) BitAndf(sr *Stream, y Element) Element {
	return invalidOp(sr, "&", e.Self())
}

func (e *GenericElement) Addf(sr *Stream, y Element) Element {
	return invalidOp(sr, "+", e.Self())
}

func (e *GenericElement) Subf(sr *Stream, y Element) Element {
	return invalidOp(sr, "-", e.Self())
}

func (e *GenericElement) Mulf(sr *Stream, y Element) Element {
	return invalidOp(sr, "*", e.Self())
}

func (e *GenericElement) Divf(sr *Stream, y Element) Element {
	return invalidOp(sr, "/", e.Self())
}

func (e *GenericElement) Modf(sr *Stream, y Element) Element {
	return invalidOp(sr, "%", e.Self())
}

func (e *GenericElement) Preincf(sr *Stream) Element {
	return invalidOp(sr, "++X", e.Self())
}

func (e *GenericElement) Predecf(sr *Stream) Element {
	return invalidOp(sr, "--X", e.Self())
}

func (e *GenericElement) Postincf(sr *Stream) Element {
	return invalidOp(sr, "X++", e.Self())
}

func (e *GenericElement) Postdecf(sr *Stream) Element {
	return invalidOp(sr, "X--", e.Self())
}

func (e *GenericElement) Negf(sr *Stream) Element {
	return invalidOp(sr, "u-", e.Self())
}

func (e *GenericElement) Notf(sr *Stream) Element {
	return NewBoolean(!e.Self().ToBool())
}

func (e *GenericElement) Invf(sr *Stream) Element {
	return invalidOp(sr, "~", e.Self())
}
