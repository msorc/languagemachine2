package machine

import ()

type VarElement interface {
	Element
	ScopeHolder
	Ru() *Rule
	Si() int
	Gr() *Grammar
	Gsy() Element
	Rsy() Element
	Lsy() Element
	Ifn() string
	Cp() int
	Ln() int
	Cn() int
	Deref(Element) VarElement
	Link() VarElement // next variable in the chain (vs in the original)
	KeyString() string
	ValueString() string
	AllVariables() VarElement
	Key() Element
	Value() Element
	Variables() VarElement
	ScopeReferenceContext() ScopeHolder
	ToDebug() string
}

// ScopeHolder
type Var struct {
	GenericElement
	scopeVariables VarElement
	allVariables   VarElement
	key            Element
	value          Element
	variables      VarElement
	scope          ScopeHolder
}

func NewVarFromParams(s VarElement, key, value Element, q ScopeHolder, a VarElement) VarElement {
	if q == nil {
		panic("vx cannot be nil")
	}
	v := MakeSelf[Var]()
	v.allVariables = a
	v.scopeVariables = s
	v.key = key
	v.value = value
	v.variables = q.ScopeVariables()
	v.scope = q

	return v
}

func (v *Var) AsVarE() VarElement {
	return v._self.(VarElement)
}

func (v *Var) Act(sr *Stream, s GenMode) GenMode {
	if v.value != nil {
		return v.value.Reference(sr, s, v.AsVarE())
	}
	return s
}

func (v *Var) ScopeContextLimitVariables() VarElement {
	if v.scope != nil {
		return v.scope.ScopeContextLimitVariables()
	}
	return nil
}

func (v *Var) ScopeContextMode() ContextHolder {
	if v.scope != nil {
		return v.scope.ScopeContextMode()
	}
	return nil
}

func (v *Var) MakeVar(k, ve Element, s ScopeHolder, a VarElement) VarElement {
	v.variables = NewVarFromParams(v.variables, k, ve, s, a)
	return v.variables
}

func (v *Var) KeyString() string {
	if v.key != nil {
		return v.key.ToString()
	}
	return "---"
}

func (v *Var) ValueString() string {
	if v.value != nil {
		return v.value.ToString()
	}
	return "---"
}

func (v *Var) ToString() string {
	return v.AsVarE().KeyString()
}

func (v *Var) ToDebug() string {
	return "var " + v.AsVarE().KeyString() + ": " + v.AsVarE().ValueString()
}

func (v *Var) Deref(k Element) VarElement {
	pp := v.variables
	for pp != nil && k != pp.Key() {
		pp = pp.Link()
	}
	return pp
}

func (v *Var) ToVal() Element {
	x := v.value
	if vs, ok := x.(*VarSym); ok {
		x = v.AsVarE().Deref(vs)
	}
	if x == nil {
		return v.AsVarE().NotFound()
	}
	return x.ToVal()
}

func (v *Var) ToVar() VarElement {
	return v.AsVarE()
}

func (v *Var) ToBool() bool {
	if v.value == nil {
		v.value = NewBoolean(false)
	}
	return v.value.ToBool()
}

func (v *Var) ToInt() int {
	if v.value == nil {
		v.value = NewNumber(0)
	}
	return v.value.ToInt()
}

func (v *Var) Append(y Element) Element {
	if v.value == nil || v.value == theNull() {
		v.value = NewLMBuffer()
	}
	return v.value.Append(y)
}

func (v *Var) Idxf(y Element) Element {
	return v.value.Idxf(y.ToVal())
}

func (v *Var) Idtf(y Element) Element {
	return v.value.Idtf(y.ToVal())
}

func (v *Var) StoValf(y Element) Element {
	v.value = y.ToVal()
	return v.value
}

func (v *Var) StoAddf(y Element) Element {
	v.value = v.value.Addf(y.ToVal())
	return v.value
}

func (v *Var) StoSubf(y Element) Element {
	v.value = v.value.Subf(y.ToVal())
	return v.value
}

func (v *Var) StoMulf(y Element) Element {
	v.value = v.value.Mulf(y.ToVal())
	return v.value
}

func (v *Var) StoDivf(y Element) Element {
	v.value = v.value.Divf(y.ToVal())
	return v.value
}

func (v *Var) StoModf(y Element) Element {
	v.value = v.value.Modf(y.ToVal())
	return v.value
}

func (v *Var) Preincf() Element {
	return v.Self().StoAddf(NewNumber(1))
}

func (v *Var) Predecf() Element {
	return v.Self().StoSubf(NewNumber(1))
}

func (v *Var) Postincf() Element {
	r := v.value.ToVal()
	v.Self().Preincf()
	return r
}

func (v *Var) Postdecf() Element {
	r := v.value.ToVal()
	v.Self().Predecf()
	return r
}

func (v *Var) Ru() *Rule {
	return v.ScopeContextMode().Rule()
}

func (v *Var) Si() int {
	return v.ScopeContextMode().State().stateIndex
}

func (v *Var) Gr() *Grammar {
	return v.ScopeContextMode().State().grammar
}

func (v *Var) Gsy() Element {
	return v.ScopeContextMode().State().grammar.symbol
}

func (v *Var) Rsy() Element {
	return v.ScopeContextMode().State().rsy
}

func (v *Var) Lsy() Element {
	return v.ScopeContextMode().State().lsy
}

func (v *Var) Ifn() string {
	return v.ScopeContextMode().State().input.Filename()
}

func (v *Var) Cp() int {
	return v.ScopeContextMode().State().charPosition
}

func (v *Var) Ln() int {
	return v.ScopeContextMode().State().lineNumber
}

func (v *Var) Cn() int {
	return v.ScopeContextMode().State().charNumber
}

// ScopeVariables is the variable chain seen from this variable as a scope
// (vvp in the original), not the next link in its own chain.
func (v *Var) ScopeVariables() VarElement {
	return v.variables
}

func (v *Var) Link() VarElement {
	return v.scopeVariables
}

func (v *Var) AllVariables() VarElement {
	return v.allVariables
}

func (v *Var) Key() Element {
	return v.key
}

func (v *Var) Value() Element {
	return v.value
}

func (v *Var) Variables() VarElement {
	return v.variables
}

func (v *Var) ScopeReferenceContext() ScopeHolder {
	return v.scope
}

type LMRef struct {
	Var
}

func NewLMRef() *LMRef {
	return MakeSelf[LMRef]()
}

func NewLMRefFromElement(k Element, q ScopeHolder) *LMRef {
	lm := NewLMRef()
	lm.key = k
	lm.variables = q.ScopeVariables() // Assuming Vvp() returns Var
	lm.scope = q
	if lm.scope == nil {
		panic("vx is null")
	}
	lm.value = lm.Deref(lm.key)
	return lm
}

func (lm *LMRef) ToString() string {
	return lm.AsVarE().KeyString()
}

func (lm *LMRef) ToDebug() string {
	return "LMRef " + lm.AsVarE().KeyString() + ": " + lm.AsVarE().ValueString()
}

func (lm *LMRef) ToVal() Element {
	return lm.Var.ToVal()
}

func (lm *LMRef) ToRef() VarElement {
	return lm.AsVarE()
}

// target is the referenced variable; an undefined name has no target.
func (lm *LMRef) target(op string) (Element, bool) {
	if lm.value == nil {
		lm.InvalidOp(op + " " + lm.AsVarE().KeyString() + " (undefined)")
		return theNull(), false
	}
	return lm.value, true
}

func (lm *LMRef) Append(y Element) Element {
	if t, ok := lm.target("~="); ok {
		return t.Append(y)
	}
	return theNull()
}

func (lm *LMRef) Funf(y Element) Element {
	return nil
}

func (lm *LMRef) Inf(y Element) Element {
	if t, ok := lm.target("in"); ok {
		return t.Inf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) Idxf(y Element) Element {
	if t, ok := lm.target("[]"); ok {
		return t.Idxf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) Idtf(y Element) Element {
	if t, ok := lm.target("."); ok {
		return t.Idtf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) StoValf(y Element) Element {
	if t, ok := lm.target("="); ok {
		return t.StoValf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) StoAddf(y Element) Element {
	if t, ok := lm.target("+="); ok {
		return t.StoAddf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) StoSubf(y Element) Element {
	if t, ok := lm.target("-="); ok {
		return t.StoSubf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) StoMulf(y Element) Element {
	if t, ok := lm.target("*="); ok {
		return t.StoMulf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) StoDivf(y Element) Element {
	if t, ok := lm.target("/="); ok {
		return t.StoDivf(y.ToVal())
	}
	return theNull()
}

func (lm *LMRef) StoModf(y Element) Element {
	if t, ok := lm.target("%="); ok {
		return t.StoModf(y.ToVal())
	}
	return theNull()
}

type ARef struct {
	Var
	A *AArray
	K Element
}

func NewARef(x *AArray, y Element, z ScopeHolder) *ARef {
	ar := MakeSelf[ARef]()
	ar.A = x
	ar.K = y
	ar.value = z.ScopeVariables()
	ar.scope = z
	if ar.scope == nil {
		panic("vx is null")
	}
	return ar
}

func (ar *ARef) Act(sr *Stream, s GenMode) GenMode {
	if _, ok := ar.A.A[ar.K]; ok {
		ar.value = ar.A.A[ar.K]
		return ar.value.Reference(sr, s, ar)
	}
	return s
}

func (ar *ARef) KeyString() string {
	if ar.K != nil {
		return ar.K.ToString()
	}
	return "---"
}

func (ar *ARef) ValueString() string {
	if _, ok := ar.A.A[ar.K]; ok {
		return ar.A.A[ar.K].ToString()
	}
	return "---"
}

func (ar *ARef) ToString() string {
	return "aref " + ar.AsVarE().KeyString() + ": " + ar.AsVarE().ValueString()
}

func (ar *ARef) ToRef() *ARef {
	return ar
}

func (ar *ARef) ToVal() Element {
	if val, ok := ar.A.A[ar.K]; ok {
		return val
	}
	return ar.Self().NotFound()
}

func (ar *ARef) Append(y Element) Element {
	return ar.Self().StoValf(ar.Self().ToVal().Append(y))
}

func (ar *ARef) Inf(y Element) Element {
	return ar.Self().ToVal().Inf(y)
}

func (ar *ARef) Idxf(y Element) Element {
	return ar.Self().ToVal().Idxf(y)
}

func (ar *ARef) Idtf(y Element) Element {
	return ar.Self().ToVal().Idtf(y)
}

func (ar *ARef) StoValf(y Element) Element {
	ar.A.A[ar.K] = y
	return y
}

func (ar *ARef) StoAddf(y Element) Element {
	r := ar.Self().ToVal().Addf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoSubf(y Element) Element {
	r := ar.Self().ToVal().Subf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoMulf(y Element) Element {
	r := ar.Self().ToVal().Mulf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoDivf(y Element) Element {
	r := ar.Self().ToVal().Divf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) StoModf(y Element) Element {
	r := ar.Self().ToVal().Modf(y)
	ar.A.A[ar.K] = r
	return r
}

func (ar *ARef) Preincf() Element {
	return ar.Self().StoAddf(NewNumber(1))
}

func (ar *ARef) Predecf() Element {
	return ar.Self().StoSubf(NewNumber(1))
}

func (ar *ARef) Postincf() Element {
	r := ar.Self().ToVal()
	ar.Self().Preincf()
	return r
}

func (ar *ARef) Postdecf() Element {
	r := ar.Self().ToVal()
	ar.Self().Predecf()
	return r
}
