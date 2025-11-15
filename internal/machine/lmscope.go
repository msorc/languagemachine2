package machine

type LMScope interface {
	ScopeVariables() VarElement             // variable reference LMScope
	ScopeContextLimitVariables() VarElement // limit of context
	ScopeContextMode() ContextHolder   // variable context
	ScopeReferenceContext() LMScope         // variable LMScope
	MakeVar(Element, Element, LMScope, VarElement) VarElement
	RfScope() LMScope
}
