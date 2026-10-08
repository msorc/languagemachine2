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
// Every element type embeds genericElement, which gives a default for each
// method, and reaches the outermost type through Self (see self_pointer.go).
// An element type that has something to trace also implements traceable.
type Element interface {
	selfPointer[Element]

	// The grammar engine.
	addRule(*grammar, *rule)                            // file a rule whose left side starts with this element
	match(*Engine, Element) bool                        // match this goal against an input symbol
	compare(*Engine, Element) bool                      // equality for the default Match
	act(*Stream, GenMode) GenMode                       // run as an instruction in a mode
	newRHX(GenMode, contextHolder, scopeHolder) GenMode // a mode that produces this element's body
	reference(*Stream, GenMode, scopeHolder) GenMode    // act as the value of a variable
	token() Element                                     // the symbol rules are filed under
	weight() int                                        // length counted for rule ordering
	toBody() []Element                                  // the elements of a list, or nil

	// Values.
	ToVal() Element
	toVar() varElement // the variable, or nil
	ToNumber() LMNumber
	IsNumber() bool
	toInt() int
	ToBool() bool

	// Text.
	ToString() string
	toTrace() string
	toEncode() string
	toDecode() string
	toDump() string

	operand
}

// Operand is the arithmetic of elements, called by the operators with the
// stream they act on. genericElement reports each as invalid; numbers,
// symbols, variables, arrays and buffers implement what they support.
type operand interface {
	append(*Stream, Element) Element
	inf(*Stream, Element) Element
	idxf(*Stream, Element) Element
	idtf(*Stream, Element) Element
	stoValf(*Stream, Element) Element
	stoAddf(*Stream, Element) Element
	stoSubf(*Stream, Element) Element
	stoMulf(*Stream, Element) Element
	stoDivf(*Stream, Element) Element
	stoModf(*Stream, Element) Element
	eeqf(*Stream, Element) Element
	neef(*Stream, Element) Element
	eqf(*Stream, Element) Element
	nef(*Stream, Element) Element
	ltf(*Stream, Element) Element
	gtf(*Stream, Element) Element
	lef(*Stream, Element) Element
	gef(*Stream, Element) Element
	bitXorf(*Stream, Element) Element
	bitOrf(*Stream, Element) Element
	bitAndf(*Stream, Element) Element
	addf(*Stream, Element) Element
	subf(*Stream, Element) Element
	mulf(*Stream, Element) Element
	divf(*Stream, Element) Element
	modf(*Stream, Element) Element
	preincf(*Stream) Element
	predecf(*Stream) Element
	postincf(*Stream) Element
	postdecf(*Stream) Element
	negf(*Stream) Element
	notf(*Stream) Element
	invf(*Stream) Element
}

// traceable is an element that writes a trace line before it acts
// (operators and statements, under their trace categories).
type traceable interface {
	trace(*Stream, *tracer)
}

// genericElement gives every element its default behaviour.
type genericElement struct {
	selfPointing[Element]
}

func (e *genericElement) addRule(g *grammar, x *rule) {
	g.add(x)
}

func (e *genericElement) match(engine *Engine, r Element) bool {
	if e.self().compare(engine, r) {
		engine.matchedWith(e.self(), r, r)
		return true
	}
	return engine.resolve(e.self(), r)
}

func (e *genericElement) newRHX(m GenMode, c contextHolder, x scopeHolder) GenMode {
	panic("not implemented")
}

func (e *genericElement) act(sr *Stream, s GenMode) GenMode {
	sr.currentSymbol = e.self()
	return s
}

func (e *genericElement) compare(engine *Engine, r Element) bool {
	return false
}

// ToNumber of a non-number is NaN, as in the original (lmNumber.init), so
// numeric comparisons with it are false rather than fatal.
func (e *genericElement) ToNumber() LMNumber {
	return LMNumber(math.NaN())
}

func (e *genericElement) IsNumber() bool {
	return false
}

func (e *genericElement) ToBool() bool {
	return true
}

// toVar of an element that is not a variable is nil; callers that need a
// variable check for it.
func (e *genericElement) toVar() varElement {
	return nil
}

// toInt of a non-number is 0, as C's conversions of unparsable text are.
func (e *genericElement) toInt() int {
	return 0
}

func (e *genericElement) ToString() string {
	return "element"
}

func (e *genericElement) toTrace() string {
	return e.self().ToString()
}

func (e *genericElement) toEncode() string {
	return conv.Encode(e.self().ToString())
}

func (e *genericElement) toDecode() string {
	return decodeURI(e.self().ToString())
}

func decodeURI(s string) string {
	d, err := conv.Decode(s)
	if err != nil {
		fail("cannot URI-decode %q: %v", s, err)
	}
	return d
}

func (e *genericElement) toDump() string {
	return e.self().toEncode()
}

func (e *genericElement) toBody() []Element {
	return nil
}

func (e *genericElement) weight() int {
	return 0
}

func (e *genericElement) token() Element {
	return e.self()
}

func (e *genericElement) reference(sr *Stream, s GenMode, x scopeHolder) GenMode {
	return e.self().act(sr, s)
}

// invalidOp reports an operation that x does not support on the engine's
// error output, after the output written so far, and gives null.
func invalidOp(sr *Stream, op string, x Element) Element {
	var b strings.Builder
	traceElement(&b, "BAD "+op, x)
	sr.Engine.writeErr(b.String())
	return Null()
}

func (e *genericElement) ToVal() Element {
	return e.self()
}

func (e *genericElement) append(sr *Stream, y Element) Element {
	return invalidOp(sr, "~=", e.self())
}

func (e *genericElement) inf(sr *Stream, y Element) Element {
	return Null()
}

func (e *genericElement) idxf(sr *Stream, y Element) Element {
	return Null()
}

func (e *genericElement) idtf(sr *Stream, y Element) Element {
	return Null()
}

func (e *genericElement) stoValf(sr *Stream, y Element) Element {
	return invalidOp(sr, "=", e.self())
}

func (e *genericElement) stoAddf(sr *Stream, y Element) Element {
	return invalidOp(sr, "+=", e.self())
}

func (e *genericElement) stoSubf(sr *Stream, y Element) Element {
	return invalidOp(sr, "-=", e.self())
}

func (e *genericElement) stoMulf(sr *Stream, y Element) Element {
	return invalidOp(sr, "*=", e.self())
}

func (e *genericElement) stoDivf(sr *Stream, y Element) Element {
	return invalidOp(sr, "/=", e.self())
}

func (e *genericElement) stoModf(sr *Stream, y Element) Element {
	return invalidOp(sr, "%=", e.self())
}

func (e *genericElement) eeqf(sr *Stream, y Element) Element {
	return newBoolean(y.token() == e.self().token())
}

func (e *genericElement) neef(sr *Stream, y Element) Element {
	return newBoolean(y.token() != e.self().token())
}

func (e *genericElement) eqf(sr *Stream, y Element) Element {
	return invalidOp(sr, "==", e.self())
}

func (e *genericElement) nef(sr *Stream, y Element) Element {
	return invalidOp(sr, "!=", e.self())
}

func (e *genericElement) ltf(sr *Stream, y Element) Element {
	return invalidOp(sr, "<", e.self())
}

func (e *genericElement) gtf(sr *Stream, y Element) Element {
	return invalidOp(sr, ">", e.self())
}

func (e *genericElement) lef(sr *Stream, y Element) Element {
	return invalidOp(sr, "<=", e.self())
}

func (e *genericElement) gef(sr *Stream, y Element) Element {
	return invalidOp(sr, ">=", e.self())
}

func (e *genericElement) bitXorf(sr *Stream, y Element) Element {
	return invalidOp(sr, "^", e.self())
}

func (e *genericElement) bitOrf(sr *Stream, y Element) Element {
	return invalidOp(sr, "|", e.self())
}

func (e *genericElement) bitAndf(sr *Stream, y Element) Element {
	return invalidOp(sr, "&", e.self())
}

func (e *genericElement) addf(sr *Stream, y Element) Element {
	return invalidOp(sr, "+", e.self())
}

func (e *genericElement) subf(sr *Stream, y Element) Element {
	return invalidOp(sr, "-", e.self())
}

func (e *genericElement) mulf(sr *Stream, y Element) Element {
	return invalidOp(sr, "*", e.self())
}

func (e *genericElement) divf(sr *Stream, y Element) Element {
	return invalidOp(sr, "/", e.self())
}

func (e *genericElement) modf(sr *Stream, y Element) Element {
	return invalidOp(sr, "%", e.self())
}

func (e *genericElement) preincf(sr *Stream) Element {
	return invalidOp(sr, "++X", e.self())
}

func (e *genericElement) predecf(sr *Stream) Element {
	return invalidOp(sr, "--X", e.self())
}

func (e *genericElement) postincf(sr *Stream) Element {
	return invalidOp(sr, "X++", e.self())
}

func (e *genericElement) postdecf(sr *Stream) Element {
	return invalidOp(sr, "X--", e.self())
}

func (e *genericElement) negf(sr *Stream) Element {
	return invalidOp(sr, "u-", e.self())
}

func (e *genericElement) notf(sr *Stream) Element {
	return newBoolean(!e.self().ToBool())
}

func (e *genericElement) invf(sr *Stream) Element {
	return invalidOp(sr, "~", e.self())
}
