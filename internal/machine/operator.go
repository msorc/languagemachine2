package machine

// opKind says how an operator takes its operands and which trace category it
// belongs to.
type opKind int

const (
	opUnary    opKind = iota // f(value of x)
	opArith                  // f(value of x, value of y)
	opRelation               // f(value of x, value of y)
	opAssign                 // f(x, value of y): x is the variable assigned to
	opIncDec                 // f(x): x is the variable changed
	opIndex                  // f(x, value of y): x is the array or table indexed
)

// operator is an arithmetic, relational, assignment or index operator. It
// pops its operands, applies its function and pushes the result.
type operator struct {
	primitive
	kind   opKind
	unary  func(x Element, sr *Stream) Element
	binary func(x Element, sr *Stream, y Element) Element
}

func newOperator(name string, kind opKind, unary func(Element, *Stream) Element, binary func(Element, *Stream, Element) Element) *operator {
	return reSelf(&operator{primitive: *newPrimitive(name), kind: kind, unary: unary, binary: binary})
}

func (o *operator) trace(s *Stream, t *tracer) {
	switch o.kind {
	case opUnary, opArith:
		t.traceArithmetic(s, o.self())
	case opRelation:
		t.traceRelation(s, o.self())
	case opAssign, opIncDec:
		t.traceAssignment(s, o.self())
	case opIndex:
		t.traceIndex(s, o.self())
	}
}

func (o *operator) act(sr *Stream, b GenMode) GenMode {
	switch o.kind {
	case opUnary:
		sr.pushX(o.unary(sr.popX().ToVal(), sr))
	case opIncDec:
		sr.pushX(o.unary(sr.popX(), sr))
	case opArith, opRelation:
		y := sr.popX()
		x := sr.popX()
		sr.pushX(o.binary(x.ToVal(), sr, y.ToVal()))
	case opAssign, opIndex:
		y := sr.popX()
		x := sr.popX()
		sr.pushX(o.binary(x, sr, y.ToVal()))
	}
	return b
}

// operators are the operator symbols, by name. Most have a word form and a
// punctuation form.
var operators = []struct {
	names  []string
	kind   opKind
	unary  func(Element, *Stream) Element
	binary func(Element, *Stream, Element) Element
}{
	{[]string{"idx"}, opIndex, nil, Element.idxf},
	{[]string{"idt"}, opIndex, nil, Element.idtf},
	{[]string{"stoVal", "="}, opAssign, nil, Element.stoValf},
	{[]string{"stoAdd", "+="}, opAssign, nil, Element.stoAddf},
	{[]string{"stoSub", "-="}, opAssign, nil, Element.stoSubf},
	{[]string{"stoMul", "*="}, opAssign, nil, Element.stoMulf},
	{[]string{"stoDiv", "/="}, opAssign, nil, Element.stoDivf},
	{[]string{"stoMod", "%="}, opAssign, nil, Element.stoModf},
	// eeq and nee are == and != as in the original's table, while === and
	// !== compare identity; lmn only produces === and !==
	{[]string{"eeq"}, opRelation, nil, Element.eqf},
	{[]string{"==="}, opRelation, nil, Element.eeqf},
	{[]string{"nee"}, opRelation, nil, Element.nef},
	{[]string{"!=="}, opRelation, nil, Element.neef},
	{[]string{"in"}, opRelation, nil, Element.inf},
	{[]string{"eq", "=="}, opRelation, nil, Element.eqf},
	{[]string{"ne", "!="}, opRelation, nil, Element.nef},
	{[]string{"lt", "<"}, opRelation, nil, Element.ltf},
	{[]string{"gt", ">"}, opRelation, nil, Element.gtf},
	{[]string{"le", "<="}, opRelation, nil, Element.lef},
	{[]string{"ge", ">="}, opRelation, nil, Element.gef},
	{[]string{"bitOr", "|"}, opArith, nil, Element.bitOrf},
	{[]string{"bitXor", "^"}, opArith, nil, Element.bitXorf},
	{[]string{"bitAnd", "&"}, opArith, nil, Element.bitAndf},
	{[]string{"add", "+"}, opArith, nil, Element.addf},
	{[]string{"sub", "-"}, opArith, nil, Element.subf},
	{[]string{"mul", "*"}, opArith, nil, Element.mulf},
	{[]string{"div", "/"}, opArith, nil, Element.divf},
	{[]string{"mod", "%"}, opArith, nil, Element.modf},
	{[]string{"preinc"}, opIncDec, Element.preincf, nil},
	{[]string{"predec"}, opIncDec, Element.predecf, nil},
	{[]string{"postinc"}, opIncDec, Element.postincf, nil},
	{[]string{"postdec"}, opIncDec, Element.postdecf, nil},
	{[]string{"inv"}, opUnary, Element.invf, nil},
	{[]string{"not"}, opUnary, Element.notf, nil},
	{[]string{"neg"}, opUnary, Element.negf, nil},
}
