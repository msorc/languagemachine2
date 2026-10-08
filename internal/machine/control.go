package machine

import "slices"

// block is the code of a { ... } operand; anything else is a fault in the
// rules.
func block(x Element) []Element {
	s, ok := x.ToVal().(*Str)
	if !ok {
		fail("expected a block, found %s", x.ToVal().ToString())
	}
	return s.V
}

type Iff struct {
	Primitive
}

func NewIff(x string) *Iff {
	return ReSelf(&Iff{Primitive: *NewPrimitiveFromString(x)})
}

func (i *Iff) Act(sr *Stream, b GenMode) GenMode {
	y := sr.Popx()
	x := sr.Popx()
	t := sr.Popx().ToVal()
	if t.ToBool() {
		return NewSTModeFromElements(b, block(x), b)
	}
	return NewSTModeFromElements(b, block(y), b)
}

type OrOrf struct {
	Primitive
}

func NewOrOrf(x string) *OrOrf {
	return ReSelf(&OrOrf{Primitive: *NewPrimitiveFromString(x)})
}

func (o *OrOrf) Act(sr *Stream, b GenMode) GenMode {
	x := sr.Popx()
	t := sr.Popx().ToVal()
	if !t.ToBool() {
		return NewSTModeFromElements(b, block(x), b)
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
	x := sr.Popx()
	t := sr.Popx().ToVal()
	if t.ToBool() {
		return NewSTModeFromElements(b, block(x), b)
	}
	sr.Pushx(NewBoolean(false))
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
	return NewRPModeFromElement(b, block(sr.Popx()))
}

type Testf struct {
	Primitive
}

func NewTestf(x string) *Testf {
	return ReSelf(&Testf{Primitive: *NewPrimitiveFromString(x)})
}

func (t *Testf) Act(sr *Stream, b GenMode) GenMode {
	te := sr.Popx().ToVal()
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
	next := block(sr.Popx())
	body := block(sr.Popx())
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
	for i, side := range []string{"left", "right"} {
		if len(v[3+i].ToBody()) == 0 {
			fail("rule(%s, %s) has an empty %s side", v[0].ToString(), v[1].ToString(), side)
		}
	}
	sr.Engine.AddRule(v, sr.Engine.ruleNumbers)
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
		return NewSTModeFromElements(b, block(x), b)
	}
	return NewSTModeFromElements(b, block(y), b)
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
	body := block(sr.Popx())
	e := sr.Popx().ToVal()
	value := sr.Popx()
	key := sr.Popx()
	step := NewForeachStep(value, key)
	switch a := e.(type) {
	case *LMArray:
		step.a = a.aa
		step.keys = slices.Clone(a.aa.Keys)
	case *ZLM:
	default:
		invalidOp(sr, "foreach over "+e.ToString(), f.Self())
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
		r.StoValf(sr, key)
	}
	if r, ok := fs.v.(*LMRef); ok {
		r.StoValf(sr, fs.a.A[key])
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
