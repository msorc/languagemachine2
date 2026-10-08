package machine

// state information that can be fixed at the start of a context, ie when a mismatch occurs
type state struct {
	engine       *Engine  // the engine - for access to global properties
	grammar      *grammar // current grammar
	lsy          Element  // lh symbol at mismatch
	rsy          Element  // rh symbol at mismatch
	input        Input    // input source object
	charPosition int      // absolute char position in file
	lineNumber   int      // line number
	charNumber   int      // char number in line
	stateIndex   int      // state index or identity
}

func newState(e *Engine, g *grammar, l, r Element, i Input, p int, n int, c int, x int) *state {
	return &state{
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
func (s *state) getChr(ci int) Element {
	return s.engine.rhsBuffer.getChr(s.engine, ci)
}

type contextHolder interface {
	scopeHolder
	Rule() *rule
	State() *state
	Priority() priority
	Variables() varElement
	setVariables(varElement)
	ContextLimitVariable() varElement
	NestingDepth() int
	trace(string) string
}

type contextType int

const (
	lhContext contextType = iota
	rhContext
)

// Context is the state of the engine as a rule is applied: the state at the
// mismatch it resolves, the rule, its priority and the variables it binds.
type context struct {
	state                *state      // state at start of context
	rule                 *rule       // rule
	priority             priority    // context priority
	variables            varElement  // variables
	contextLimitVariable varElement  // limit of context
	nestingDepth         int         // context nesting depth
	contextType          contextType // LH or RH
}

func newRootContext(ct contextType, s *state) *context {
	return &context{
		contextType: ct,
		state:       s,
	}
}

func newContext(ct contextType, s *state, c contextHolder, x *rule, n priority, p, q varElement) *context {
	return &context{
		contextType:          ct,
		state:                s,
		priority:             n,
		rule:                 x,
		variables:            p,
		contextLimitVariable: q,
		nestingDepth:         c.NestingDepth() + 1,
	}
}

func newLHContext(s *state, c contextHolder, x *rule) *context {
	return newContext(lhContext, s, c, x, x.priority.context(c.Priority()), c.Variables(), c.Variables())
}
func newRHContext(s *state, c, l contextHolder) *context {
	return newContext(rhContext, s, c, l.Rule(), l.Priority(), l.Variables(), l.ContextLimitVariable())
}

func (c *context) ScopeVariables() varElement {
	return c.variables
}

func (c *context) scopeContextLimitVariables() varElement {
	return c.contextLimitVariable
}

func (c *context) scopeReferenceContext() scopeHolder {
	return c
}

func (c *context) scopeContextMode() contextHolder {
	return c
}

func (c *context) Rule() *rule                      { return c.rule }
func (c *context) State() *state                    { return c.state }
func (c *context) Priority() priority               { return c.priority }
func (c *context) Variables() varElement            { return c.variables }
func (c *context) setVariables(v varElement)        { c.variables = v }
func (c *context) ContextLimitVariable() varElement { return c.contextLimitVariable }
func (c *context) NestingDepth() int                { return c.nestingDepth }

func (c *context) makeVar(k, v Element, s scopeHolder, a varElement) varElement {
	c.variables = newBinding(c.variables, k, v, s, a)
	return c.variables
}

func (c *context) typeName() string {
	switch c.contextType {
	case lhContext:
		return "L"
	case rhContext:
		return "R"
	default:
		return "C"
	}
}

func (c *context) trace(s string) string {
	return c.typeName() + "-" + s
}
