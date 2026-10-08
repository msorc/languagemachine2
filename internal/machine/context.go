package machine

// state information that can be fixed at the start of a context, ie when a mismatch occurs
type State struct {
	engine       *Engine   // the engine - for access to global properties
	grammar      *Grammar  // current grammar
	lsy          Element   // lh symbol at mismatch
	rsy          Element   // rh symbol at mismatch
	input        GrammarIO // input source object
	charPosition int       // absolute char position in file
	lineNumber   int       // line number
	charNumber   int       // char number in line
	stateIndex   int       // state index or identity
}

func NewState(e *Engine, g *Grammar, l, r Element, i GrammarIO, p int, n int, c int, x int) *State {
	return &State{
		engine:       e,
		grammar:      g,
		lsy:          l,
		rsy:          r,
		input:        i,
		charPosition: p,
		lineNumber:   n,
		charNumber:   c,
		stateIndex:   x,
	}
}

// Method to get character element
func (s *State) GetChr(ci int) Element {
	return s.engine.rhsBuffer.GetChr(s.engine, ci)
}

type ContextHolder interface {
	ScopeHolder
	SelfPointer[ContextHolder]
	Rule() *Rule
	State() *State
	Priority() int
	Variables() VarElement
	SetVariables(VarElement)
	ContextLimitVariable() VarElement
	NestingDepth() int
	Trace(string) string
}

type ContextType int

const (
	LHContext ContextType = iota
	RHContext
)

// Context is the state of the engine as a rule is applied: the state at the
// mismatch it resolves, the rule, its priority and the variables it binds.
type Context struct {
	SelfPointing[ContextHolder]
	state                *State      // state at start of context
	rule                 *Rule       // rule
	priority             int         // context priority
	variables            VarElement  // variables
	contextLimitVariable VarElement  // limit of context
	nestingDepth         int         // context nesting depth
	contextType          ContextType // LH or RH
}

func NewContextFromState(ct ContextType, s *State) *Context {
	return ReSelf(&Context{
		contextType: ct,
		state:       s,
	})
}

func NewContextFromParams(ct ContextType, s *State, c ContextHolder, x *Rule, n int, p, q VarElement) *Context {
	return ReSelf(&Context{
		contextType:          ct,
		state:                s,
		priority:             n,
		rule:                 x,
		variables:            p,
		contextLimitVariable: q,
		nestingDepth:         c.NestingDepth() + 1,
	})
}

func NewLHContextFromRule(s *State, c ContextHolder, x *Rule) *Context {
	return NewContextFromParams(LHContext, s, c, x, x.Cxtpri(c.Priority()), c.Variables(), c.Variables())
}
func NewRHContextFromStateContext(s *State, c, l ContextHolder) *Context {
	return NewContextFromParams(RHContext, s, c, l.Rule(), l.Priority(), l.Variables(), l.ContextLimitVariable())
}

func (c *Context) ScopeVariables() VarElement {
	return c.variables
}

func (c *Context) ScopeContextLimitVariables() VarElement {
	return c.contextLimitVariable
}

func (c *Context) ScopeReferenceContext() ScopeHolder {
	return c.Self()
}

func (c *Context) ScopeContextMode() ContextHolder {
	return c.Self()
}

func (c *Context) Rule() *Rule                      { return c.rule }
func (c *Context) State() *State                    { return c.state }
func (c *Context) Priority() int                    { return c.priority }
func (c *Context) Variables() VarElement            { return c.variables }
func (c *Context) SetVariables(v VarElement)        { c.variables = v }
func (c *Context) ContextLimitVariable() VarElement { return c.contextLimitVariable }
func (c *Context) NestingDepth() int                { return c.nestingDepth }

func (c *Context) MakeVar(k, v Element, s ScopeHolder, a VarElement) VarElement {
	c.variables = NewVarFromParams(c.variables, k, v, s, a)
	return c.variables
}

func (c *Context) typeName() string {
	switch c.contextType {
	case LHContext:
		return "L"
	case RHContext:
		return "R"
	default:
		return "C"
	}
}

func (c *Context) Trace(s string) string {
	return c.typeName() + "-" + s
}
