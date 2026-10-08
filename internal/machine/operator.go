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
	Primitive
	kind   opKind
	unary  func(x Element) Element
	binary func(x, y Element) Element
}

func newOperator(name string, kind opKind, unary func(Element) Element, binary func(Element, Element) Element) *operator {
	return ReSelf(&operator{Primitive: *NewPrimitiveFromString(name), kind: kind, unary: unary, binary: binary})
}

func (o *operator) Trace(s *Stream, t *Tracer) {
	switch o.kind {
	case opUnary, opArith:
		t.TraceArithmetic(s, o.Self())
	case opRelation:
		t.TraceRelation(s, o.Self())
	case opAssign, opIncDec:
		t.TraceAssignment(s, o.Self())
	case opIndex:
		t.TraceIndex(s, o.Self())
	}
}

func (o *operator) Act(sr *Stream, b GenMode) GenMode {
	switch o.kind {
	case opUnary:
		sr.Pushx(o.unary(sr.Popx().ToVal()))
	case opIncDec:
		sr.Pushx(o.unary(sr.Popx()))
	case opArith, opRelation:
		y := sr.Popx()
		x := sr.Popx()
		sr.Pushx(o.binary(x.ToVal(), y.ToVal()))
	case opAssign, opIndex:
		y := sr.Popx()
		x := sr.Popx()
		sr.Pushx(o.binary(x, y.ToVal()))
	}
	return b
}

// operators are the operator symbols, by name. Most have a word form and a
// punctuation form.
var operators = []struct {
	names  []string
	kind   opKind
	unary  func(Element) Element
	binary func(Element, Element) Element
}{
	{[]string{"idx"}, opIndex, nil, Element.Idxf},
	{[]string{"idt"}, opIndex, nil, Element.Idtf},
	{[]string{"stoVal", "="}, opAssign, nil, Element.StoValf},
	{[]string{"stoAdd", "+="}, opAssign, nil, Element.StoAddf},
	{[]string{"stoSub", "-="}, opAssign, nil, Element.StoSubf},
	{[]string{"stoMul", "*="}, opAssign, nil, Element.StoMulf},
	{[]string{"stoDiv", "/="}, opAssign, nil, Element.StoDivf},
	{[]string{"stoMod", "%="}, opAssign, nil, Element.StoModf},
	// eeq and nee are == and != as in the original's table, while === and
	// !== compare identity; lmn only produces === and !==
	{[]string{"eeq"}, opRelation, nil, Element.Eqf},
	{[]string{"==="}, opRelation, nil, Element.Eeqf},
	{[]string{"nee"}, opRelation, nil, Element.Nef},
	{[]string{"!=="}, opRelation, nil, Element.Neef},
	{[]string{"in"}, opRelation, nil, Element.Inf},
	{[]string{"eq", "=="}, opRelation, nil, Element.Eqf},
	{[]string{"ne", "!="}, opRelation, nil, Element.Nef},
	{[]string{"lt", "<"}, opRelation, nil, Element.Ltf},
	{[]string{"gt", ">"}, opRelation, nil, Element.Gtf},
	{[]string{"le", "<="}, opRelation, nil, Element.Lef},
	{[]string{"ge", ">="}, opRelation, nil, Element.Gef},
	{[]string{"bitOr", "|"}, opArith, nil, Element.BitOrf},
	{[]string{"bitXor", "^"}, opArith, nil, Element.BitXorf},
	{[]string{"bitAnd", "&"}, opArith, nil, Element.BitAndf},
	{[]string{"add", "+"}, opArith, nil, Element.Addf},
	{[]string{"sub", "-"}, opArith, nil, Element.Subf},
	{[]string{"mul", "*"}, opArith, nil, Element.Mulf},
	{[]string{"div", "/"}, opArith, nil, Element.Divf},
	{[]string{"mod", "%"}, opArith, nil, Element.Modf},
	{[]string{"preinc"}, opIncDec, Element.Preincf, nil},
	{[]string{"predec"}, opIncDec, Element.Predecf, nil},
	{[]string{"postinc"}, opIncDec, Element.Postincf, nil},
	{[]string{"postdec"}, opIncDec, Element.Postdecf, nil},
	{[]string{"inv"}, opUnary, Element.Invf, nil},
	{[]string{"not"}, opUnary, Element.Notf, nil},
	{[]string{"neg"}, opUnary, Element.Negf, nil},
}
