package machine

import (
// "errors"
)

type LMScope interface {
	VvP() *Var               // variable reference LMScope
	VvQ() *Var               // limit of context
	VvC() EngineStateContext // variable context
	VvS() LMScope            // variable LMScope
	MakeVar(MachineElement, MachineElement, LMScope, *Var) *Var
	RfScope() LMScope
}

type GenMode interface {
	LMScope
	SR() *Stream
	SP() *Var
	SX() LMScope
	CX() EngineStateContext
	CI() uint
	CV() []MachineElement
	LK() any
	What() uint
	Ret() GenMode
	Restore() GenMode
	Advance(*Stream) GenMode
	Save() GenMode
	More() GenMode
	Ends() GenMode
	Cont() GenMode
	EndRep(GenMode) GenMode
	Trace(MachineElement)
	TraceRet(*Stream, *Tracer)
}

// LMScope
// element generator modes produce symbols for the engine to match
type Mode struct {
	sr *Stream          // stream registers
	sy MachineElement   // current symbol
	sv MachineElement   // current value
	cv []MachineElement // code vector
	ci uint             // code index
	lk any
	xs *Opnd              // operand stack
	sp *Var               // variables visible in this level
	sx LMScope            // reference context
	cx EngineStateContext // mode context
	ss GenMode            // mode stack link
}

func NewMode() *Mode {
	return &Mode{}
}

func NewModeFromVar(s GenMode, v *Var) *Mode {
	mode := Mode{
		ss: s,
		sr: s.SR(),
		sy: s.SR().SY,
		sv: s.SR().SV,
		cv: s.SR().CV,
		ci: s.SR().CI,
		xs: s.SR().XS,
		lk: s.SR().LK,
		cx: s.CX(),
		sx: v,
		sp: v,
	}

	mode.sr.SY = nil
	mode.sr.CV = make([]MachineElement, 0)
	mode.sr.CI = 0
	mode.sr.LK = nil

	return &mode
}

func NewModeFromElements(s GenMode, v []MachineElement, i uint, c EngineStateContext, x LMScope) *Mode {
	mode := Mode{
		ss: s,
		sr: s.SR(),
		sy: s.SR().SY,
		sv: s.SR().SV,
		cv: s.SR().CV,
		ci: s.SR().CI,
		xs: s.SR().XS,
		lk: s.SR().LK,
		cx: c,
		sx: x,
		sp: x.VvP(),
	}

	mode.sr.SY = nil
	// mode.sr.LK = nil
	mode.sr.CV = v
	mode.sr.CI = i

	return &mode
}

func NewModeFromMode(s GenMode) *Mode {
	return &Mode{
		ss: s,
		sr: s.SR(),
		sy: s.SR().SY,
		sv: s.SR().SV,
		cv: s.SR().CV,
		ci: s.SR().CI,
		xs: s.SR().XS,
		lk: s.SR().LK,
		sp: s.SP(),
		sx: s.SX(),
		cx: s.CX(),
	}
}

func (m *Mode) SR() *Stream            { return m.sr }
func (m *Mode) SP() *Var               { return m.sp }
func (m *Mode) SX() LMScope            { return m.sx }
func (m *Mode) CX() EngineStateContext { return m.cx }
func (m *Mode) CI() uint               { return m.ci }
func (m *Mode) CV() []MachineElement   { return m.cv }
func (m *Mode) LK() any                { return m.lk }

func (m *Mode) What() uint {
	return 1
}

func (m *Mode) Ret() GenMode {
	m.sr.SY = m.sy
	m.sr.SV = m.sv
	m.sr.CV = m.cv
	m.sr.CI = m.ci
	m.sr.LK = m.lk
	return m.ss
}

func (m *Mode) Restore() GenMode {
	m.sr.XS = m.xs
	m.sr.SY = m.sy
	m.sr.SV = m.sv
	m.sr.CV = m.cv
	m.sr.CI = m.ci
	m.sr.LK = m.lk
	return m.ss
}

func (m *Mode) Advance(s *Stream) GenMode {
	return s.Act(s, m)
}

func (m *Mode) VvP() *Var {
	return m.sp
}

func (m *Mode) VvQ() *Var {
	if m.sx != nil {
		return m.sx.VvQ()
	}
	return nil
}

func (m *Mode) VvS() LMScope {
	return m.sx
}

func (m *Mode) VvC() EngineStateContext {
	return m.cx
}

func (m *Mode) MakeVar(k, v MachineElement, s LMScope, a *Var) *Var {
	return m.sx.MakeVar(k, v, s, a)
}

func (m *Mode) RfScope() LMScope {
	return m.sx
}

func (m *Mode) Save() GenMode {
	return NewModeFromMode(m)
}

func (m *Mode) More() GenMode {
	return m.ss.More()
}

func (m *Mode) Ends() GenMode {
	return m.ss.Ends()
}

func (m *Mode) Cont() GenMode {
	return m.ss.Cont()
}

func (m *Mode) EndRep(mode GenMode) GenMode {
	return m.Ret()
}

func (m *Mode) TraceRet(sr *Stream, t *Tracer) {
}

func (m *Mode) Trace(x MachineElement) {
	TxE("mm", x)
}

// LHS mode: symbols produced from LHS of rules that are being matched
type LHMode struct {
	Mode
}

func NewLHMode() *LHMode {
	return &LHMode{}
}

func newLHModeFromElement(s GenMode, v []MachineElement, i uint, c EngineStateContext) *LHMode {
	return &LHMode{
		Mode: *NewModeFromElements(s, v, i, c, c),
	}
}

func NewLHModeFromMode(s GenMode) *LHMode {
	return &LHMode{
		Mode: *NewModeFromMode(s),
	}
}

func (m *LHMode) Ret() GenMode {
	m.sr.SY = m.sy
	m.sr.CV = m.cv
	m.sr.CI = m.ci
	m.sr.LK = m.lk
	return nil
}

func (m *LHMode) Save() GenMode {
	return NewLHModeFromMode(m)
}

// variable reference lmScope
func (m *LHMode) VvP() *Var {
	return m.sx.VvP()
}

// limit of context
func (m *LHMode) vvq() *Var {
	if m.sx != nil {
		return m.sx.VvQ()
	}
	return nil
}

func (m *LHMode) VvS() LMScope {
	return m.sx
}

func (m *LHMode) VvC() EngineStateContext {
	return m.cx
}

func (m *LHMode) MakeVar(k, v MachineElement, s LMScope, a *Var) *Var {
	m.sr.AV = m.sx.MakeVar(k, v, s, a)
	return m.sr.AV
}

func (m *LHMode) RfScope() LMScope {
	return m.sx
}

// func (m *LHMode) Advance(s *Stream) GenMode {
//     return s.Act(s, m)
// }

func (m *LHMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("lh"), x)
}

func (m *LHMode) TraceRet(sr *Stream, t *Tracer) {
	if (t.Flags&DIAGRAM == DIAGRAM) && m.cx.Ru().Off >= m.cx.Ru().Rhlength() {
		sr.LM.Display.EndLevel("lx", m.cx.St().Si, sr.LM.Rhr.SM.CX().St().Si, m.cx.Cd(), sr.LM.Rhr.SM.CX().Cd())
	}
}

// RHS mode: input symbols and symbols produced by RHS of rules that have matched
type RHMode struct {
	Mode
}

func NewRHMode() *RHMode {
	return &RHMode{}
}

func NewRHModeFromParams(s GenMode, v []MachineElement, i uint, c EngineStateContext) *RHMode {
	return &RHMode{
		Mode: *NewModeFromElements(s, v, i, c, c),
	}
}

func NewRHModeFromParamsAndScope(s GenMode, v []MachineElement, i uint, c EngineStateContext, x LMScope) *RHMode {
	return &RHMode{
		Mode: *NewModeFromElements(s, v, i, c, x),
	}
}

func NewRHModeFromMode(s GenMode) *RHMode {
	return &RHMode{
		Mode: *NewModeFromMode(s),
	}
}

func (m *RHMode) Save() GenMode {
	return NewRHModeFromMode(m)
}

func (m *RHMode) VvP() *Var {
	return m.sx.VvP()
}

func (m *RHMode) VvQ() *Var {
	if m.sx != nil {
		return m.sx.VvQ()
	}
	return nil
}

func (m *RHMode) VvS() LMScope {
	return m.sx
}

func (m *RHMode) VvC() EngineStateContext {
	return m.cx
}

func (m *RHMode) MakeVar(k, v MachineElement, s LMScope, a *Var) *Var {
	return m.sx.MakeVar(k, v, s, a)
}

// func (m *RHMode) Advance(s *Stream) GenMode {
//     return s.act(m)
// }

func (m *RHMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("rh"), x)
}

func (m *RHMode) TraceRet(sr *Stream, t *Tracer) {
	if (t.Flags & DIAGRAM) == DIAGRAM {
		sr.LM.Display.EndLevel("rx", sr.LM.Lhx.St().Si, m.cx.St().Si, sr.LM.Lhx.Cd(), m.cx.Cd())
	}
}

type LZMode struct {
	Mode
}

func NewLZModeFromContext(z EngineStateContext, s *Stream) *LZMode {
	return &LZMode{Mode{cx: z, sr: s}}
}

func NewLZModeFromMode(s GenMode) *LZMode {
	return &LZMode{Mode: *NewModeFromMode(s)}
}

func (m *LZMode) Save() GenMode {
	return NewLZModeFromMode(m)
}

func (m *LZMode) VvP() *Var {
	return m.sp
}

func (m *LZMode) VvQ() *Var {
	if m.sx != nil {
		return m.sx.VvQ()
	}
	return nil
}

func (m *LZMode) VvS() LMScope {
	return m.sx
}

func (m *LZMode) VvC() EngineStateContext {
	return m.cx
}

func (m *LZMode) Advance(s *Stream) GenMode {
	s.SY = s.Ssy().EOF
	s.CI++
	if s.CI > 0 {
		return nil
	}
	return m
}

func (m *LZMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("lz"), x)
}

type RZMode struct {
	Mode
}

func NewRZModeFromContext(z EngineStateContext, s *Stream) *RZMode {
	return &RZMode{Mode: Mode{cx: z, sr: s}}
}

func NewRZModeFromMode(s GenMode) *RZMode {
	return &RZMode{Mode: *NewModeFromMode(s)}
}

func (m *RZMode) Save() GenMode {
	return NewRZModeFromMode(m)
}

func (m *RZMode) VvP() *Var {
	return m.sp
}

func (m *RZMode) VvQ() *Var {
	if m.sx != nil {
		return m.sx.VvQ()
	}
	return nil
}

func (m *RZMode) Vvs() LMScope {
	return m.sx
}

func (m *RZMode) Vvc() EngineStateContext {
	return m.cx
}

func (m *RZMode) Advance(s *Stream) GenMode {
	s.SY = m.cx.St().GetChr(s.CI)
	s.CI++
	return m
}

func (m *RZMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("rz"), x)
}

type STMode struct {
	Mode
}

func NewSTMode() *STMode {
	return &STMode{}
}

func NewSTModeFromElements(s GenMode, v []MachineElement, x LMScope) *STMode {
	return &STMode{Mode: *NewModeFromElements(s, v, 0, s.CX(), x)}
}

func NewSTModeFromMode(s GenMode) *STMode {
	return &STMode{Mode: *NewModeFromMode(s)}
}

func (m *STMode) Save() GenMode {
	return NewSTModeFromMode(m)
}

func (m *STMode) VvP() *Var {
	return m.sp
}

func (m *STMode) VvQ() *Var {
	if m.sx != nil {
		return m.sx.VvQ()
	}
	return nil
}

func (m *STMode) VvS() LMScope {
	return m.sx
}

func (m *STMode) VvC() EngineStateContext {
	return m.cx
}

func (m *STMode) MakeVar(k, v MachineElement, s LMScope, a *Var) *Var {
	return m.sx.MakeVar(k, v, s, a)
}

func (m *STMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("st"), x)
}

type RPMode struct {
	Mode
}

func NewRPMode() *RPMode {
	return &RPMode{}
}

func NewRPModeFromElement(s GenMode, v []MachineElement) *RPMode {
	return &RPMode{Mode: *NewModeFromElements(s, v, 0, s.CX(), s.CX())}
}

func NewRPModeFromMode(s GenMode) *RPMode {
	return &RPMode{Mode: *NewModeFromMode(s)}
}

func (m *RPMode) Save() GenMode {
	return NewRPModeFromMode(m)
}

func (m *RPMode) More() GenMode {
	return m
}

func (m *RPMode) Ends() GenMode {
	return m.Ret()
}

func (m *RPMode) Cont() GenMode {
	m.ci = 0
	return m
}

func (m *RPMode) Advance(s *Stream) GenMode {
	return s.Rep(s, m)
}

func (m *RPMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("rp"), x)
}

type RFMode struct {
	Mode
}

func NewRFMode() *RFMode {
	return &RFMode{}
}

func NewRFModeFromVar(s GenMode, v *Var) *RFMode {
	return &RFMode{Mode: *NewModeFromVar(s, v)}
}

func NewRFModeFromMode(s GenMode) *RFMode {
	return &RFMode{Mode: *NewModeFromMode(s)}
}

func (m *RFMode) Save() GenMode {
	return NewRFModeFromMode(m)
}

func (m *RFMode) Advance(s *Stream) GenMode {
	return m.sp.Vv.Reference(s, m.Ret(), m.sp.VvS())
}

func (m *RFMode) VvP() *Var {
	return m.sp
}

func (m *RFMode) VvQ() *Var {
	if m.sx != nil {
		return m.sx.VvQ()
	}
	return nil
}

func (m *RFMode) Vvs() LMScope {
	return m.sx
}

func (m *RFMode) Vvc() EngineStateContext {
	return m.cx
}

func (m *RFMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("rf"), x)
}

type APMode struct {
	Mode
	v MachineElement
}

func NewAPMode() *APMode {
	return &APMode{}
}

func NewAPModeFromElement(s GenMode, x MachineElement) *APMode {
	mode := &APMode{Mode: *NewModeFromMode(s)}
	mode.v = x
	return mode
}

func NewAPModeFromMode(s GenMode) *APMode {
	return &APMode{Mode: *NewModeFromMode(s)}
}

func (m *APMode) Save() GenMode {
	return NewAPModeFromMode(m)
}

func (m *APMode) Advance(s *Stream) GenMode {
	if m.v != nil {
		return m.v.Act(s, m.Ret())
	}
	return m.Ret()
}

func (m *APMode) Trace(x MachineElement) {
	TxE(m.cx.Trace("ap"), x)
}

// Operand stack - pushdown list of operand elements
type Opnd struct {
	S *Opnd          // operand stack link
	V MachineElement // operand element
}

func NewOpnd(p *Opnd, x MachineElement) *Opnd {
	return &Opnd{
		S: p,
		V: x,
	}
}

func (o *Opnd) ToString() string {
	if o.V != nil {
		return o.V.ToString()
	}
	return "---"
}

// state information that can be fixed at the start of a context, ie when a mismatch occurs
type State struct {
	lm    *Engine        // the engine - for access to global properties
	Gr    *Grammar       // current grammar
	lsy   MachineElement // lh symbol at mismatch
	rsy   MachineElement // rh symbol at mismatch
	input GrammarStdio   // input source object
	cp    uint           // absolute char position in file
	ln    uint           // line number
	cn    uint           // char number in line
	Si    uint           // state index or identity
}

func NewState(e *Engine, g *Grammar, l, r MachineElement, i GrammarStdio, p uint, n uint, c uint, x uint) *State {
	return &State{
		lm:    e,
		Gr:    g,
		lsy:   l,
		rsy:   r,
		input: i,
		cp:    p,
		ln:    n,
		cn:    c,
		Si:    x,
	}
}

// Method to get character element
func (s *State) GetChr(ci uint) MachineElement {
	return s.lm.Rhz.GetChr(s.lm, ci)
	return nil
}

type EngineStateContext interface {
	LMScope
	Ru() *Rule
	St() *State
	Pr() uint
	Xs() Opnd
	Cp() *Var
	Cq() *Var
	Cd() uint
	Cs() EngineStateContext
	CheckDepth(uint) error
	Trace(string) string
}

// LMScope
// contexts: the state of the engine as rules are applied
type Context struct {
	st *State             // state at start of context
	ru *Rule              // rule
	pr uint               // context priority
	xs Opnd               // operand stack
	cp *Var               // variables
	cq *Var               // limit of context
	cd uint               // context nesting depth
	cs EngineStateContext // context stack
}

func NewContextFromState(s *State) *Context {
	return &Context{
		st: s,
	}
}

func NewContextFromParams(s *State, c EngineStateContext, x *Rule, n uint, p, q *Var) *Context {
	return &Context{
		st: s,
		cs: c,
		pr: n,
		ru: x,
		cp: p,
		cq: q,
		cd: c.Cd() + 1,
	}
}

func NewContextFromContext(x EngineStateContext) *Context {
	return &Context{
		st: x.St(),
		ru: x.Ru(),
		pr: x.Pr(),
		xs: x.Xs(),
		cp: x.Cp(),
		cq: x.Cq(),
		cs: x.Cs(),
	}
}

func (c *Context) Copy(x EngineStateContext) *Context {
	c.st = x.St()
	c.ru = x.Ru()
	c.pr = x.Pr()
	c.xs = x.Xs()
	c.cp = x.Cp()
	c.cq = x.Cq()
	c.cs = x.Cs()
	return c
}

func (c *Context) Dup() *Context {
	return NewContextFromContext(c)
}

func (c *Context) CheckDepth(max uint) error {
	if max == 0 || c.cd < max {
		return nil
	}
	panic("maxDepthError")
	//return errors.New("maxDepthError")
}

func (c *Context) VvP() *Var {
	return c.cp
}

func (c *Context) VvQ() *Var {
	return c.cq
}

func (c *Context) VvS() LMScope {
	return c
}

func (c *Context) VvC() EngineStateContext {
	return c
}

func (c *Context) Ru() *Rule              { return c.ru }
func (c *Context) St() *State             { return c.st }
func (c *Context) Pr() uint               { return c.pr }
func (c *Context) Xs() Opnd               { return c.xs }
func (c *Context) Cp() *Var               { return c.cp }
func (c *Context) Cq() *Var               { return c.cq }
func (c *Context) Cd() uint               { return c.cd }
func (c *Context) Cs() EngineStateContext { return c.cs }

func (c *Context) MakeVar(k, v MachineElement, s LMScope, a *Var) *Var {
	c.cp = NewVarFromParams(c.cp, k, v, s, a)
	return c.cp
}

func (c *Context) RfScope() LMScope {
	return c
}

func (c *Context) Trace(s string) string {
	return "C-" + s
}

// LHS context: the context in which a rule is being tried
type LHContext struct {
	Context
}

func NewLHContext() *LHContext {
	return &LHContext{}
}

func NewLHContextFromState(s *State) *LHContext {
	return &LHContext{
		Context: *NewContextFromState(s),
	}
}

func NewLHContextFromContext(c EngineStateContext) *LHContext {
	return &LHContext{
		Context: *NewContextFromContext(c),
	}
}

func NewLHContextFromRule(s *State, c EngineStateContext, x *Rule) *LHContext {
	return &LHContext{
		Context: *NewContextFromParams(s, c, x, x.Cxtpri(c.Pr()), c.Cp(), c.Cp()),
	}
}

func (lh *LHContext) Dup() *LHContext {
	return NewLHContextFromContext(lh)
}

func (lh *LHContext) VvP() *Var {
	return lh.cp
}

func (lh *LHContext) VvQ() *Var {
	return lh.cq
}

func (lh *LHContext) VvS() LMScope {
	return lh
}

func (lh *LHContext) VvC() EngineStateContext {
	return lh
}

func (lh *LHContext) MakeVar(k, v MachineElement, s LMScope, a *Var) *Var {
	lh.cp = NewVarFromParams(lh.cp, k, v, s, a)
	return lh.cp
}

func (lh *LHContext) RfScope() LMScope {
	return lh
}

func (lh *LHContext) Trace(s string) string {
	return "L-" + s
}

// RHS context: used to provide information to RHS modes and to variables
type RHContext struct {
	Context
}

func NewRHContext() *RHContext {
	return &RHContext{}
}

func NewRHContextFromState(s *State) *RHContext {
	return &RHContext{
		Context: *NewContextFromState(s),
	}
}

func NewRHContextFromContext(c EngineStateContext) *RHContext {
	return &RHContext{
		Context: *NewContextFromContext(c),
	}
}

func NewRHContextFromStateContexts(s *State, c, l EngineStateContext) *RHContext {
	return &RHContext{
		Context: *NewContextFromParams(s, c, l.Ru(), l.Pr(), l.Cp(), l.Cq()),
	}
}

func (rh *RHContext) Dup() *RHContext {
	return NewRHContextFromContext(rh)
}

func (rh *RHContext) VvP() *Var {
	return rh.cp
}

func (rh *RHContext) VvQ() *Var {
	return rh.cq
}

func (rh *RHContext) VvS() LMScope {
	return rh
}

func (rh *RHContext) VvC() EngineStateContext {
	return rh
}

func (rh *RHContext) MakeVar(k, v MachineElement, s LMScope, a *Var) *Var {
	rh.cp = NewVarFromParams(rh.cp, k, v, s, a)
	return rh.cp
}

func (rh *RHContext) RfScope() LMScope {
	return rh
}

func (rh *RHContext) Trace(s string) string {
	return "R-" + s
}
