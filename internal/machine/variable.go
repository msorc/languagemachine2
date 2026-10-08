package machine

// scopeHolder is a scope in which variables are found and made.
type scopeHolder interface {
	ScopeVariables() varElement             // variable reference scopeHolder
	scopeContextLimitVariables() varElement // limit of context
	scopeContextMode() contextHolder        // variable context
	scopeReferenceContext() scopeHolder     // variable scopeHolder
	makeVar(Element, Element, scopeHolder, varElement) varElement
}

type varElement interface {
	Element
	scopeHolder
	bindingRule() *rule
	stateIndex() int
	bindingGrammar() *grammar
	grammarSymbol() Element
	rhsSymbol() Element
	lhsSymbol() Element
	fileName() string
	charPos() int
	lineNo() int
	charNo() int
	deref(Element) varElement
	link() varElement // next variable in the chain (vs in the original)
	keyString() string
	valueString() string
	AllVariables() varElement
	Key() Element
	Value() Element
	Variables() varElement
	scopeReferenceContext() scopeHolder
	toDebug() string
}

// Var is a variable binding. Bindings are linked into chains that record
// where each was made (state, grammar, input position).
type binding struct {
	genericElement
	scopeVariables varElement
	allVariables   varElement
	key            Element
	value          Element
	variables      varElement
	scope          scopeHolder
}

func newBinding(s varElement, key, value Element, q scopeHolder, a varElement) varElement {
	if q == nil {
		panic("vx cannot be nil")
	}
	v := makeSelf[binding]()
	v.allVariables = a
	v.scopeVariables = s
	v.key = key
	v.value = value
	v.variables = q.ScopeVariables()
	v.scope = q

	return v
}

func (v *binding) asVarE() varElement {
	return v._self.(varElement)
}

func (v *binding) act(sr *Stream, s GenMode) GenMode {
	if v.value != nil {
		return v.value.reference(sr, s, v.asVarE())
	}
	return s
}

func (v *binding) scopeContextLimitVariables() varElement {
	if v.scope != nil {
		return v.scope.scopeContextLimitVariables()
	}
	return nil
}

func (v *binding) scopeContextMode() contextHolder {
	if v.scope != nil {
		return v.scope.scopeContextMode()
	}
	return nil
}

func (v *binding) makeVar(k, ve Element, s scopeHolder, a varElement) varElement {
	v.variables = newBinding(v.variables, k, ve, s, a)
	return v.variables
}

func (v *binding) keyString() string {
	if v.key != nil {
		return v.key.ToString()
	}
	return "---"
}

func (v *binding) valueString() string {
	if v.value != nil {
		return v.value.ToString()
	}
	return "---"
}

func (v *binding) ToString() string {
	return v.asVarE().keyString()
}

func (v *binding) toDebug() string {
	return "var " + v.asVarE().keyString() + ": " + v.asVarE().valueString()
}

func (v *binding) deref(k Element) varElement {
	pp := v.variables
	for pp != nil && k != pp.Key() {
		pp = pp.link()
	}
	return pp
}

func (v *binding) ToVal() Element {
	x := v.value
	if vs, ok := x.(*varSym); ok {
		x = v.asVarE().deref(vs)
	}
	if x == nil {
		return Null()
	}
	return x.ToVal()
}

func (v *binding) toVar() varElement {
	return v.asVarE()
}

func (v *binding) ToBool() bool {
	if v.value == nil {
		v.value = newBoolean(false)
	}
	return v.value.ToBool()
}

func (v *binding) toInt() int {
	if v.value == nil {
		v.value = newNumber(0)
	}
	return v.value.toInt()
}

func (v *binding) append(sr *Stream, y Element) Element {
	if v.value == nil || v.value == Null() {
		v.value = newBufferValue()
	}
	return v.value.append(sr, y)
}

func (v *binding) idxf(sr *Stream, y Element) Element {
	return v.value.idxf(sr, y.ToVal())
}

func (v *binding) idtf(sr *Stream, y Element) Element {
	return v.value.idtf(sr, y.ToVal())
}

func (v *binding) stoValf(sr *Stream, y Element) Element {
	v.value = y.ToVal()
	return v.value
}

func (v *binding) stoAddf(sr *Stream, y Element) Element {
	v.value = v.value.addf(sr, y.ToVal())
	return v.value
}

func (v *binding) stoSubf(sr *Stream, y Element) Element {
	v.value = v.value.subf(sr, y.ToVal())
	return v.value
}

func (v *binding) stoMulf(sr *Stream, y Element) Element {
	v.value = v.value.mulf(sr, y.ToVal())
	return v.value
}

func (v *binding) stoDivf(sr *Stream, y Element) Element {
	v.value = v.value.divf(sr, y.ToVal())
	return v.value
}

func (v *binding) stoModf(sr *Stream, y Element) Element {
	v.value = v.value.modf(sr, y.ToVal())
	return v.value
}

func (v *binding) preincf(sr *Stream) Element {
	return v.self().stoAddf(sr, newNumber(1))
}

func (v *binding) predecf(sr *Stream) Element {
	return v.self().stoSubf(sr, newNumber(1))
}

func (v *binding) postincf(sr *Stream) Element {
	r := v.value.ToVal()
	v.self().preincf(sr)
	return r
}

func (v *binding) postdecf(sr *Stream) Element {
	r := v.value.ToVal()
	v.self().predecf(sr)
	return r
}

func (v *binding) bindingRule() *rule {
	return v.scopeContextMode().Rule()
}

func (v *binding) stateIndex() int {
	return v.scopeContextMode().State().stateIndex
}

func (v *binding) bindingGrammar() *grammar {
	return v.scopeContextMode().State().grammar
}

func (v *binding) grammarSymbol() Element {
	return v.scopeContextMode().State().grammar.symbol
}

func (v *binding) rhsSymbol() Element {
	return v.scopeContextMode().State().rsy
}

func (v *binding) lhsSymbol() Element {
	return v.scopeContextMode().State().lsy
}

func (v *binding) fileName() string {
	return v.scopeContextMode().State().input.Filename()
}

func (v *binding) charPos() int {
	return v.scopeContextMode().State().charPosition
}

func (v *binding) lineNo() int {
	return v.scopeContextMode().State().lineNumber
}

func (v *binding) charNo() int {
	return v.scopeContextMode().State().charNumber
}

// ScopeVariables is the variable chain seen from this variable as a scope
// (vvp in the original), not the next link in its own chain.
func (v *binding) ScopeVariables() varElement {
	return v.variables
}

func (v *binding) link() varElement {
	return v.scopeVariables
}

func (v *binding) AllVariables() varElement {
	return v.allVariables
}

func (v *binding) Key() Element {
	return v.key
}

func (v *binding) Value() Element {
	return v.value
}

func (v *binding) Variables() varElement {
	return v.variables
}

func (v *binding) scopeReferenceContext() scopeHolder {
	return v.scope
}

type varRef struct {
	binding
}

func newVarRef() *varRef {
	return makeSelf[varRef]()
}

func newVarRefOf(k Element, q scopeHolder) *varRef {
	lm := newVarRef()
	lm.key = k
	lm.variables = q.ScopeVariables()
	lm.scope = q
	if lm.scope == nil {
		panic("vx is null")
	}
	lm.value = lm.deref(lm.key)
	return lm
}

func (lm *varRef) ToString() string {
	return lm.asVarE().keyString()
}

func (lm *varRef) toDebug() string {
	return "LMRef " + lm.asVarE().keyString() + ": " + lm.asVarE().valueString()
}

func (lm *varRef) ToVal() Element {
	return lm.binding.ToVal()
}

func (lm *varRef) toRef() varElement {
	return lm.asVarE()
}

// target is the referenced variable; an undefined name has no target.
func (lm *varRef) target(sr *Stream, op string) (Element, bool) {
	if lm.value == nil {
		invalidOp(sr, op+" "+lm.asVarE().keyString()+" (undefined)", lm.self())
		return Null(), false
	}
	return lm.value, true
}

func (lm *varRef) append(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "~="); ok {
		return t.append(sr, y)
	}
	return Null()
}

func (lm *varRef) inf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "in"); ok {
		return t.inf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) idxf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "[]"); ok {
		return t.idxf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) idtf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "."); ok {
		return t.idtf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) stoValf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "="); ok {
		return t.stoValf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) stoAddf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "+="); ok {
		return t.stoAddf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) stoSubf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "-="); ok {
		return t.stoSubf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) stoMulf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "*="); ok {
		return t.stoMulf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) stoDivf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "/="); ok {
		return t.stoDivf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) stoModf(sr *Stream, y Element) Element {
	if t, ok := lm.target(sr, "%="); ok {
		return t.stoModf(sr, y.ToVal())
	}
	return Null()
}

func (lm *varRef) preincf(sr *Stream) Element {
	if t, ok := lm.target(sr, "++"); ok {
		return t.preincf(sr)
	}
	return Null()
}

func (lm *varRef) predecf(sr *Stream) Element {
	if t, ok := lm.target(sr, "--"); ok {
		return t.predecf(sr)
	}
	return Null()
}

func (lm *varRef) postincf(sr *Stream) Element {
	if t, ok := lm.target(sr, "++"); ok {
		return t.postincf(sr)
	}
	return Null()
}

func (lm *varRef) postdecf(sr *Stream) Element {
	if t, ok := lm.target(sr, "--"); ok {
		return t.postdecf(sr)
	}
	return Null()
}

type arrayRef struct {
	binding
	a *assocArray
	k Element
}

func newArrayRef(x *assocArray, y Element, z scopeHolder) *arrayRef {
	ar := makeSelf[arrayRef]()
	ar.a = x
	ar.k = y
	ar.value = z.ScopeVariables()
	ar.scope = z
	if ar.scope == nil {
		panic("vx is null")
	}
	return ar
}

func (ar *arrayRef) act(sr *Stream, s GenMode) GenMode {
	if _, ok := ar.a.a[ar.k]; ok {
		ar.value = ar.a.a[ar.k]
		return ar.value.reference(sr, s, ar)
	}
	return s
}

func (ar *arrayRef) keyString() string {
	if ar.k != nil {
		return ar.k.ToString()
	}
	return "---"
}

func (ar *arrayRef) valueString() string {
	if _, ok := ar.a.a[ar.k]; ok {
		return ar.a.a[ar.k].ToString()
	}
	return "---"
}

func (ar *arrayRef) ToString() string {
	return "aref " + ar.asVarE().keyString() + ": " + ar.asVarE().valueString()
}

func (ar *arrayRef) toRef() *arrayRef {
	return ar
}

func (ar *arrayRef) ToVal() Element {
	if val, ok := ar.a.a[ar.k]; ok {
		return val
	}
	return Null()
}

func (ar *arrayRef) append(sr *Stream, y Element) Element {
	return ar.self().stoValf(sr, ar.self().ToVal().append(sr, y))
}

func (ar *arrayRef) inf(sr *Stream, y Element) Element {
	return ar.self().ToVal().inf(sr, y)
}

func (ar *arrayRef) idxf(sr *Stream, y Element) Element {
	return ar.self().ToVal().idxf(sr, y)
}

func (ar *arrayRef) idtf(sr *Stream, y Element) Element {
	return ar.self().ToVal().idtf(sr, y)
}

func (ar *arrayRef) stoValf(sr *Stream, y Element) Element {
	ar.a.Set(ar.k, y)
	return y
}

func (ar *arrayRef) stoAddf(sr *Stream, y Element) Element {
	r := ar.self().ToVal().addf(sr, y)
	ar.a.Set(ar.k, r)
	return r
}

func (ar *arrayRef) stoSubf(sr *Stream, y Element) Element {
	r := ar.self().ToVal().subf(sr, y)
	ar.a.Set(ar.k, r)
	return r
}

func (ar *arrayRef) stoMulf(sr *Stream, y Element) Element {
	r := ar.self().ToVal().mulf(sr, y)
	ar.a.Set(ar.k, r)
	return r
}

func (ar *arrayRef) stoDivf(sr *Stream, y Element) Element {
	r := ar.self().ToVal().divf(sr, y)
	ar.a.Set(ar.k, r)
	return r
}

func (ar *arrayRef) stoModf(sr *Stream, y Element) Element {
	r := ar.self().ToVal().modf(sr, y)
	ar.a.Set(ar.k, r)
	return r
}

func (ar *arrayRef) preincf(sr *Stream) Element {
	return ar.self().stoAddf(sr, newNumber(1))
}

func (ar *arrayRef) predecf(sr *Stream) Element {
	return ar.self().stoSubf(sr, newNumber(1))
}

func (ar *arrayRef) postincf(sr *Stream) Element {
	r := ar.self().ToVal()
	ar.self().preincf(sr)
	return r
}

func (ar *arrayRef) postdecf(sr *Stream) Element {
	r := ar.self().ToVal()
	ar.self().predecf(sr)
	return r
}
