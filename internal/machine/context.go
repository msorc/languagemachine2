package machine

import "github.com/liyue201/gostl/ds/list/bidlist"

// state information that can be fixed at the start of a context, ie when a mismatch occurs
type State struct {
	engine       *Engine   // the engine - for access to global properties
	grammar      *Grammar  // current grammar
	lsy          Element   // lh symbol at mismatch
	rsy          Element   // rh symbol at mismatch
	input        GrammarIO // input source object
	charPosition uint      // absolute char position in file
	lineNumber   uint      // line number
	charNumber   uint      // char number in line
	stateIndex   uint      // state index or identity
}

func NewState(e *Engine, g *Grammar, l, r Element, i GrammarIO, p uint, n uint, c uint, x uint) *State {
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
func (s *State) GetChr(ci uint) Element {
	return s.engine.rhsBuffer.GetChr(s.engine, ci)
}

type EngineStateContext interface {
	LMScope
	SelfPointer[EngineStateContext]
	Rule() *Rule
	State() *State
	Priority() uint
	Operands() bidlist.List[Element]
	Variables() VarElement
	ContextLimitVariable() VarElement
	NestingDepth() uint
	ContextStack() EngineStateContext
	CheckDepth(uint) error
	Trace(string) string
}

// LMScope
// contexts: the state of the engine as rules are applied
type Context struct {
	SelfPointing[EngineStateContext]
	state                *State // state at start of context
	rule                 *Rule  // rule
	priority             uint   // context priority
	operands             bidlist.List[Element]
	variables            VarElement         // variables
	contextLimitVariable VarElement         // limit of context
	nestingDepth         uint               // context nesting depth
	contextStack         EngineStateContext // context stack
}

func NewContextFromState(s *State) *Context {
	return ReSelf(&Context{
		state: s,
	})
}

func NewContextFromParams(s *State, c EngineStateContext, x *Rule, n uint, p, q VarElement) *Context {
	return ReSelf(&Context{
		state:                s,
		contextStack:         c,
		priority:             n,
		rule:                 x,
		variables:            p,
		contextLimitVariable: q,
		nestingDepth:         c.NestingDepth() + 1,
	})
}

func NewContextFromContext(x EngineStateContext) *Context {
	return ReSelf(&Context{
		state:                x.State(),
		rule:                 x.Rule(),
		priority:             x.Priority(),
		operands:             x.Operands(),
		variables:            x.Variables(),
		contextLimitVariable: x.ContextLimitVariable(),
		contextStack:         x.ContextStack(),
	})
}

func (c *Context) Copy(x EngineStateContext) EngineStateContext {
	c.state = x.State()
	c.rule = x.Rule()
	c.priority = x.Priority()
	c.operands = x.Operands()
	c.variables = x.Variables()
	c.contextLimitVariable = x.ContextLimitVariable()
	c.contextStack = x.ContextStack()
	return c.Self()
}

func (c *Context) Dup() EngineStateContext {
	return NewContextFromContext(c.Self())
}

func (c *Context) CheckDepth(max uint) error {
	if max == 0 || c.nestingDepth < max {
		return nil
	}
	panic("maxDepthError")
	//return errors.New("maxDepthError")
}

func (c *Context) ScopeVariables() VarElement {
	return c.variables
}

func (c *Context) ScopeContextLimitVariables() VarElement {
	return c.contextLimitVariable
}

func (c *Context) ScopeReferenceContext() LMScope {
	return c.Self()
}

func (c *Context) ScopeContextMode() EngineStateContext {
	return c.Self()
}

func (c *Context) Rule() *Rule                      { return c.rule }
func (c *Context) State() *State                    { return c.state }
func (c *Context) Priority() uint                   { return c.priority }
func (c *Context) Operands() bidlist.List[Element]  { return c.operands }
func (c *Context) Variables() VarElement            { return c.variables }
func (c *Context) ContextLimitVariable() VarElement { return c.contextLimitVariable }
func (c *Context) NestingDepth() uint               { return c.nestingDepth }
func (c *Context) ContextStack() EngineStateContext { return c.contextStack }

func (c *Context) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	c.variables = NewVarFromParams(c.variables, k, v, s, a)
	return c.variables
}

func (c *Context) RfScope() LMScope {
	return c.Self()
}

func (c *Context) Trace(s string) string {
	return "C-" + s
}

// LHS context: the context in which a rule is being tried
type LHContext struct {
	Context
}

func NewLHContext() *LHContext {
	return ReSelf(&LHContext{})
}

func NewLHContextFromState(s *State) *LHContext {
	return ReSelf(&LHContext{
		Context: *NewContextFromState(s),
	})
}

func NewLHContextFromContext(c EngineStateContext) *LHContext {
	return ReSelf(&LHContext{
		Context: *NewContextFromContext(c),
	})
}

func NewLHContextFromRule(s *State, c EngineStateContext, x *Rule) *LHContext {
	return ReSelf(&LHContext{
		Context: *NewContextFromParams(s, c, x, x.Cxtpri(c.Priority()), c.Variables(), c.Variables()),
	})
}

func (lh *LHContext) Dup() EngineStateContext {
	return NewLHContextFromContext(lh.Self())
}

func (lh *LHContext) ScopeVariables() VarElement {
	return lh.variables
}

func (lh *LHContext) ScopeContextLimitVariables() VarElement {
	return lh.contextLimitVariable
}

func (lh *LHContext) ScopeReferenceContext() LMScope {
	return lh.Self()
}

func (lh *LHContext) ScopeContextMode() EngineStateContext {
	return lh.Self()
}

func (lh *LHContext) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	lh.variables = NewVarFromParams(lh.variables, k, v, s, a)
	return lh.variables
}

func (lh *LHContext) RfScope() LMScope {
	return lh.Self()
}

func (lh *LHContext) Trace(s string) string {
	return "L-" + s
}

// RHS context: used to provide information to RHS modes and to variables
type RHContext struct {
	Context
}

func NewRHContext() *RHContext {
	return ReSelf(&RHContext{})
}

func NewRHContextFromState(s *State) *RHContext {
	return ReSelf(&RHContext{
		Context: *NewContextFromState(s),
	})
}

func NewRHContextFromContext(c EngineStateContext) *RHContext {
	return ReSelf(&RHContext{
		Context: *NewContextFromContext(c),
	})
}

func NewRHContextFromStateContexts(s *State, c, l EngineStateContext) *RHContext {
	return ReSelf(&RHContext{
		Context: *NewContextFromParams(s, c, l.Rule(), l.Priority(), l.Variables(), l.ContextLimitVariable()),
	})
}

func (rh *RHContext) Dup() EngineStateContext {
	return NewRHContextFromContext(rh.Self())
}

func (rh *RHContext) ScopeVariables() VarElement {
	return rh.variables
}

func (rh *RHContext) ScopeContextLimitVariables() VarElement {
	return rh.contextLimitVariable
}

func (rh *RHContext) ScopeReferenceContext() LMScope {
	return rh.Self()
}

func (rh *RHContext) ScopeContextMode() EngineStateContext {
	return rh.Self()
}

func (rh *RHContext) MakeVar(k, v Element, s LMScope, a VarElement) VarElement {
	rh.variables = NewVarFromParams(rh.variables, k, v, s, a)
	return rh.variables
}

func (rh *RHContext) RfScope() LMScope {
	return rh.Self()
}

func (rh *RHContext) Trace(s string) string {
	return "R-" + s
}
