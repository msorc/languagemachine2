package machine

import "slices"

// block is the code of a { ... } operand; anything else is a fault in the
// rules.
func block(x Element) []Element {
	s, ok := x.ToVal().(*str)
	if !ok {
		fail("expected a block, found %s", x.ToVal().ToString())
	}
	return s.v
}

type iff struct {
	primitive
}

func newIff(x string) *iff {
	return reSelf(&iff{primitive: *newPrimitive(x)})
}

func (i *iff) act(sr *Stream, b GenMode) GenMode {
	y := sr.popX()
	x := sr.popX()
	t := sr.popX().ToVal()
	if t.ToBool() {
		return newSTMode(b, block(x), b)
	}
	return newSTMode(b, block(y), b)
}

type orOrf struct {
	primitive
}

func newOrOrf(x string) *orOrf {
	return reSelf(&orOrf{primitive: *newPrimitive(x)})
}

func (o *orOrf) act(sr *Stream, b GenMode) GenMode {
	x := sr.popX()
	t := sr.popX().ToVal()
	if !t.ToBool() {
		return newSTMode(b, block(x), b)
	}
	sr.pushX(newBoolean(true))
	return b
}

type andAndf struct {
	primitive
}

func newAndAndf(x string) *andAndf {
	return reSelf(&andAndf{primitive: *newPrimitive(x)})
}

func (a *andAndf) act(sr *Stream, b GenMode) GenMode {
	x := sr.popX()
	t := sr.popX().ToVal()
	if t.ToBool() {
		return newSTMode(b, block(x), b)
	}
	sr.pushX(newBoolean(false))
	return b
}

type cellf struct {
	primitive
}

func newCellf(x string) *cellf {
	return reSelf(&cellf{primitive: *newPrimitive(x)})
}

func (c *cellf) act(sr *Stream, b GenMode) GenMode {
	y := sr.popX()
	x := sr.popX()
	sr.pushX(newCell(x, y))
	return b
}

type arrayf struct {
	primitive
}

func newArrayf(x string) *arrayf {
	return reSelf(&arrayf{primitive: *newPrimitive(x)})
}

func (a *arrayf) trace(s *Stream, t *tracer) {
	t.traceIndex(s, a.self())
}

func (a *arrayf) act(sr *Stream, b GenMode) GenMode {
	sr.pushX(newArrayValue(sr, b))
	return b
}

type argsf struct {
	primitive
}

func newArgsf(x string) *argsf {
	return reSelf(&argsf{primitive: *newPrimitive(x)})
}

func (a *argsf) act(sr *Stream, b GenMode) GenMode {
	sr.pushX(sr.Engine.predefinedSymbols.mark)
	return b
}

type funf struct {
	primitive
}

func newFunf(x string) *funf {
	return reSelf(&funf{primitive: *newPrimitive(x)})
}

func (f *funf) act(sr *Stream, b GenMode) GenMode {
	v := sr.toArgv(sr.Engine.predefinedSymbols.mark)
	sr.pushX(sr.Engine.externalSystem.call(sr, b, v[0], v))
	return b
}

type loopf struct {
	primitive
}

func newLoopf(x string) *loopf {
	return reSelf(&loopf{primitive: *newPrimitive(x)})
}

func (l *loopf) trace(s *Stream, t *tracer) {
	t.traceLoop(s, l.self())
}

func (l *loopf) act(sr *Stream, b GenMode) GenMode {
	return newRPMode(b, block(sr.popX()))
}

type testf struct {
	primitive
}

func newTestf(x string) *testf {
	return reSelf(&testf{primitive: *newPrimitive(x)})
}

func (t *testf) act(sr *Stream, b GenMode) GenMode {
	te := sr.popX().ToVal()
	if !te.ToBool() {
		return b.ends()
	}
	return b
}

// loopMode finds the loop that a break or continue in m belongs to. It does
// not look past the start of a rule side, so a loop in another rule is never
// found.
func loopMode(m GenMode, what string) *rpMode {
	for x := m; x != nil; x = x.StackMode() {
		switch y := x.(type) {
		case *rpMode:
			return y
		case *lhMode, *rhMode, *lzMode, *rzMode:
			fail("%s outside a loop", what)
		}
	}
	fail("%s outside a loop", what)
	return nil
}

// forf runs a for loop: I ( <E> f:test B ) G ( N ) G f:for. The step N
// follows the body, so that continue can resume at it.
type forf struct {
	primitive
}

func newForf(x string) *forf {
	return reSelf(&forf{primitive: *newPrimitive(x)})
}

func (f *forf) trace(s *Stream, t *tracer) {
	t.traceLoop(s, f.self())
}

func (f *forf) act(sr *Stream, b GenMode) GenMode {
	next := block(sr.popX())
	body := block(sr.popX())
	v := make([]Element, 0, len(body)+len(next))
	v = append(append(v, body...), next...)
	m := newRPMode(b, v)
	m.next = len(body)
	return m
}

// breakf ends the innermost loop.
type breakf struct {
	primitive
}

func newBreakf(x string) *breakf {
	return reSelf(&breakf{primitive: *newPrimitive(x)})
}

func (f *breakf) act(sr *Stream, b GenMode) GenMode {
	return loopMode(b, "break").exit()
}

// continuef starts the next iteration of the innermost loop, at the step of a
// for loop or at the test of a while loop.
type continuef struct {
	primitive
}

func newContinuef(x string) *continuef {
	return reSelf(&continuef{primitive: *newPrimitive(x)})
}

func (f *continuef) act(sr *Stream, b GenMode) GenMode {
	m := loopMode(b, "continue")
	// return from the modes inside the loop (if blocks), back to its body
	for x := b; x != GenMode(m); {
		y := x.StackMode()
		if y == GenMode(m) {
			x.exit()
		}
		x = y
	}
	sr.codeIndex = m.next
	return m
}

// rulef defines a rule while the rules run: rule(G, P) { lhs <- rhs }
// compiles to G P n:N ( lhs ) G ( rhs ) G f:rule, the operands of the r
// opcode. P is encoded as it is there. The value is the grammar symbol.
type rulef struct {
	primitive
}

func newRulef(x string) *rulef {
	return reSelf(&rulef{primitive: *newPrimitive(x)})
}

func (f *rulef) act(sr *Stream, b GenMode) GenMode {
	v := make([]Element, 5)
	for i := len(v) - 1; i >= 0; i-- {
		v[i] = sr.popX().ToVal()
	}
	for i, side := range []string{"left", "right"} {
		if len(v[3+i].toBody()) == 0 {
			fail("rule(%s, %s) has an empty %s side", v[0].ToString(), v[1].ToString(), side)
		}
	}
	sr.Engine.addRule(v, sr.Engine.ruleNumbers)
	sr.pushX(v[0])
	return b
}

// eachX is each (expr), the E opcode: the value of expr names the variable.
type eachX struct {
	primitive
}

func newEachX(x string) *eachX {
	return reSelf(&eachX{primitive: *newPrimitive(x)})
}

func (x *eachX) act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.eachRef(s, sr.Engine.varKey(sr.popX()), s)
}

// allX is all (expr), the bare B opcode.
type allX struct {
	primitive
}

func newAllX(x string) *allX {
	return reSelf(&allX{primitive: *newPrimitive(x)})
}

func (x *allX) act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.allRef(s, sr.Engine.varKey(sr.popX()), s)
}

type selF struct {
	primitive
}

func newSelF(x string) *selF {
	return reSelf(&selF{primitive: *newPrimitive(x)})
}

// c ? a : b compiles to <c> ( <a> ) G ( <b> ) G f:sel
func (s *selF) act(sr *Stream, b GenMode) GenMode {
	y := sr.popX()
	x := sr.popX()
	t := sr.popX().ToVal()

	if t.ToBool() {
		return newSTMode(b, block(x), b)
	}
	return newSTMode(b, block(y), b)
}

// foreachf runs foreach (K, V; E) B, compiled to <K> <V> <E> ( B ) G
// f:foreach, where <K> and <V> are variable references (<K> is null in
// foreach (V; E)). For each key of the array E, in the order the keys were
// added, it assigns the key to K and the value to V and runs B. Keys added by
// B are not visited.
type foreachf struct {
	primitive
}

func newForeachf(x string) *foreachf {
	return reSelf(&foreachf{primitive: *newPrimitive(x)})
}

func (f *foreachf) trace(s *Stream, t *tracer) {
	t.traceLoop(s, f.self())
}

func (f *foreachf) act(sr *Stream, b GenMode) GenMode {
	body := block(sr.popX())
	e := sr.popX().ToVal()
	value := sr.popX()
	key := sr.popX()
	step := newForeachStep(value, key)
	switch a := e.(type) {
	case *arrayValue:
		step.a = a.aa
		step.keys = slices.Clone(a.aa.keys)
	case *nullSym:
	default:
		invalidOp(sr, "foreach over "+e.ToString(), f.self())
	}
	v := make([]Element, 0, len(body)+1)
	v = append(append(v, step), body...)
	return newRPMode(b, v)
}

// foreachStep starts each pass of a foreach loop, so continue goes to the
// next key; after the last key it ends the loop.
type foreachStep struct {
	genericElement
	k, v Element // the loop variables (references); k may be null
	a    *assocArray
	keys []Element
	i    int
}

func newForeachStep(v, k Element) *foreachStep {
	el := makeSelf[foreachStep]()
	el.k = k
	el.v = v
	return el
}

func (fs *foreachStep) ToString() string {
	return "foreach step"
}

func (fs *foreachStep) act(sr *Stream, m GenMode) GenMode {
	if fs.i >= len(fs.keys) {
		return m.ends()
	}
	key := fs.keys[fs.i]
	fs.i++
	if r, ok := fs.k.(*varRef); ok {
		r.stoValf(sr, key)
	}
	if r, ok := fs.v.(*varRef); ok {
		r.stoValf(sr, fs.a.a[key])
	}
	return m
}

type retf struct {
	primitive
}

func newRetf(x string) *retf {
	return reSelf(&retf{primitive: *newPrimitive(x)})
}

type lamdaf struct {
	primitive
}

func newLamdaf(x string) *lamdaf {
	return reSelf(&lamdaf{primitive: *newPrimitive(x)})
}

type specf struct {
	primitive
}

func newSpecf(x string) *specf {
	return reSelf(&specf{primitive: *newPrimitive(x)})
}
