package machine

type ScopeHolder interface {
	ScopeVariables() VarElement             // variable reference ScopeHolder
	ScopeContextLimitVariables() VarElement // limit of context
	ScopeContextMode() ContextHolder   // variable context
	ScopeReferenceContext() ScopeHolder         // variable ScopeHolder
	MakeVar(Element, Element, ScopeHolder, VarElement) VarElement
	RfScope() ScopeHolder
}
