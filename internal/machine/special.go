package machine

import "fmt"

type DoneF struct {
	Symbol
}

func NewDoneF(x string) *DoneF {
	el := MakeSelf[DoneF]()
	el.V = x
	return el
}

func (df *DoneF) Match(e *Engine, r Element) bool {
	e.lhsStream.variables = e.lhsContext.Variables()
	return e.Matched3E(df.Self(), nil, nil)
}

type TakeF struct {
	Symbol
}

func NewTakeF(x string) *TakeF {
	el := MakeSelf[TakeF]()
	el.V = x
	return el
}

func (tf *TakeF) Match(e *Engine, r Element) bool {
	if r.Token() == tf.Self() { // %  %
		e.TakeTvar()
		return e.Matched3E(tf.Self(), r, nil)
	}
	if _, ok := r.(*BindF); ok { // %  :
		e.PushX()
		return e.Matched3E(tf.Self(), r, nil)
	}
	if e.rsLastMatchElement != nil { // %  [matched]
		e.PushR(e.rsLastMatchElement)
		return e.Matched3E(tf.Self(), nil, nil)
	}
	return false
}

type BindF struct {
	Symbol
}

func NewBindF(x string) *BindF {
	el := MakeSelf[BindF]()
	el.V = x
	return el
}

func (b *BindF) Match(e *Engine, r Element) bool {
	if r.Token() == b.Self() {
		a := e.lhsStream.Popx()
		bElem := e.rhsStream.Popx().ToVal()
		if a, ok := a.(*VarSym); ok {
			e.BindUvar(a, bElem)
			return e.Matched3E(b.Self(), r, nil)
		}
		e.Matched3E(b.Self(), r, nil)
		lh := []Element{a}
		e.lhsStream.mode = NewSTModeFromElements(e.lhsStream.mode, lh, e.lhsStream.mode)
		rh := []Element{bElem}
		e.rhsStream.mode = NewSTModeFromElements(e.rhsStream.mode, rh, e.rhsStream.mode)
		return true
	}
	if _, ok := r.(*TakeF); ok {
		e.BindTvar()
		return e.Matched3E(b.Self(), r, nil)
	}
	if e.rsLastMatchElement != nil {
		e.BindXvarE(e.rsLastMatchElement)
		return e.Matched3E(b.Self(), nil, r)
	}
	return false
}

type AppendXSym struct {
	Symbol
}

func NewAppendXSym(x string) *AppendXSym {
	el := MakeSelf[AppendXSym]()
	el.V = x
	return el
}

// if there is captured material, append it
// otherwise match one symbol and append that
func (a *AppendXSym) Match(e *Engine, r Element) bool {
	b := e.lhsStream.Popx()
	if !e.lhsStream.EmptyX() {
		v := e.lhsStream.ToRow()
		for _, x := range v {
			b.Append(x)
		}
		e.Matched3E(a.Self(), nil, nil)
	} else {
		b.Append(r)
		e.Matched3E(a.Self(), r, r)
	}
	return true
}

type ErrSym struct {
	Symbol
	engine *Engine
}

func NewErrSym(e *Engine, x string) *ErrSym {
	el := MakeSelf[ErrSym]()
	el.V = x
	el.engine = e
	return el
}

func (e *ErrSym) Append(y Element) Element {
	e.engine.writeErr(y.ToString())
	return e.Self()
}

func (e *ErrSym) Match(engine *Engine, r Element) bool {
	e.engine.writeErr(r.ToString())
	return engine.Matched3E(e.Self(), r, r)
}

type OutSym struct {
	Symbol
	engine *Engine
}

func NewOutSym(e *Engine, x string) *OutSym {
	el := MakeSelf[OutSym]()
	el.V = x
	el.engine = e
	return el
}

func (o *OutSym) Append(y Element) Element {
	_, _ = o.engine.out.WriteString(y.ToString())
	return o.Self()
}

func (o *OutSym) Match(engine *Engine, r Element) bool {
	_, _ = o.engine.out.WriteString(r.ToString())
	return engine.Matched3E(o.Self(), r, r)
}

type UriSym struct {
	Symbol
	engine *Engine
}

func NewUriSym(e *Engine, x string) *UriSym {
	el := MakeSelf[UriSym]()
	el.V = x
	el.engine = e
	return el
}

func (u *UriSym) Append(y Element) Element {
	_, _ = u.engine.out.WriteString(y.ToEncode())
	return u.Self()
}

func (u *UriSym) Match(engine *Engine, r Element) bool {
	_, _ = u.engine.out.WriteString(r.ToEncode())
	return engine.Matched3E(u.Self(), r, r)
}

type UrdSym struct {
	Symbol
	engine *Engine
}

func NewUrdSym(e *Engine, x string) *UrdSym {
	el := MakeSelf[UrdSym]()
	el.V = x
	el.engine = e
	return el
}

func (u *UrdSym) Append(y Element) Element {
	_, _ = u.engine.out.WriteString(y.ToDecode())
	return u.Self()
}

func (u *UrdSym) Match(engine *Engine, r Element) bool {
	_, _ = u.engine.out.WriteString(r.ToDecode())
	return engine.Matched3E(u.Self(), r, r)
}

type SpSym struct {
	Symbol
}

func NewSpSym(x string) *SpSym {
	el := MakeSelf[SpSym]()
	el.V = x
	return el
}

func (s *SpSym) ToEncode() string {
	return " "
}

type NlSym struct {
	Symbol
}

func NewNlSym(x string) *NlSym {
	el := MakeSelf[NlSym]()
	el.V = x
	return el
}

func (n *NlSym) ToEncode() string {
	return "\n"
}

type GetF struct {
	Symbol
}

func NewGetF(x string) *GetF {
	el := MakeSelf[GetF]()
	el.V = x
	return el
}

func (g *GetF) Act(sr *Stream, s GenMode) GenMode {
	return sr.Getx(s)
}

type TrueSym struct {
	Symbol
}

func NewTrueSym(x string) *TrueSym {
	el := MakeSelf[TrueSym]()
	el.V = x
	return el
}

func (t *TrueSym) ToVal() Element {
	return NewBoolean(true)
}

type FalseSym struct {
	Symbol
}

func NewFalseSym(x string) *FalseSym {
	el := MakeSelf[FalseSym]()
	el.V = x
	return el
}

func (f *FalseSym) ToVal() Element {
	return NewBoolean(false)
}

type TrueF struct {
	Symbol
}

func NewTrueF(x string) *TrueF {
	el := MakeSelf[TrueF]()
	el.V = x
	return el
}

func (t *TrueF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewBoolean(true))
	return s
}

type FalseF struct {
	Symbol
}

func NewFalseF(x string) *FalseF {
	el := MakeSelf[FalseF]()
	el.V = x
	return el
}

func (f *FalseF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewBoolean(false))
	return s
}

type GetXF struct {
	Symbol
	V Element
}

func NewGetXF(x Element) *GetXF {
	el := MakeSelf[GetXF]()
	el.V = x
	return el
}

func (g *GetXF) ToTrace() string {
	return g.V.ToTrace() + "p"
}

func (g *GetXF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(g.V)
	return s
}

type GetVF struct {
	Symbol
	V Element
}

func NewGetVF(x Element) *GetVF {
	el := MakeSelf[GetVF]()
	el.V = x
	return el
}

func (g *GetVF) ToTrace() string {
	return g.V.ToTrace()
}

func (g *GetVF) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(NewLMRefFromElement(g.V, s))
	return s
}

type ActF struct {
	Symbol
}

func NewActF(x string) *ActF {
	el := MakeSelf[ActF]()
	el.V = x
	return el
}

func (a *ActF) Trace(sr *Stream, t *Tracer) {
	t.TraceAct(sr, a.Self())
}

func (a *ActF) Act(sr *Stream, s GenMode) GenMode {
	fail("the act primitive (a) is not supported")
	return nil
}

type Primitive struct {
	Symbol
}

func NewPrimitive() *Primitive {
	return MakeSelf[Primitive]()
}

func NewPrimitiveFromString(x string) *Primitive {
	el := NewPrimitive()
	el.V = x
	return el
}

func (p *Primitive) Act(sr *Stream, s GenMode) GenMode {
	sr.Engine.printf("act: %s\n", string(p.V))
	return s
}

type ApplyF struct {
	Primitive
}

func NewApplyF(x string) *ApplyF {
	return ReSelf(&ApplyF{Primitive: *NewPrimitiveFromString(x)})
}

func (a *ApplyF) Trace(s *Stream, t *Tracer) {
	t.TraceApply(s, a.Self())
}

func (a *ApplyF) Act(sr *Stream, s GenMode) GenMode {
	v := sr.Popx()
	return v.Act(sr, s)
}

type InjF struct {
	Symbol
}

func NewInjF(x string) *InjF {
	el := MakeSelf[InjF]()
	el.V = x
	return el
}

func (i *InjF) Match(e *Engine, r Element) bool {
	e.PushRhx(e.lhsStream.Popx())
	return e.Matched3E(i.Self(), nil, nil)
}

type StrF struct {
	Symbol
}

func NewStrF(x string) *StrF {
	el := MakeSelf[StrF]()
	el.V = x
	return el
}

func (s *StrF) Act(sr *Stream, mode GenMode) GenMode {
	//  { return new stMode(s, (cast(str)sr.popx()).v, s); }
	return mode
}

type Anything struct {
	Symbol
}

func NewAnything(x string) *Anything {
	el := MakeSelf[Anything]()
	el.V = x
	return el
}

func (a *Anything) Match(e *Engine, r Element) bool {
	e.Matched3E(a.Self(), r, r)
	return true
}

type AnySym struct {
	Anything
}

func NewAnySym(x string) *AnySym {
	return ReSelf(&AnySym{Anything: *NewAnything(x)})
}

func (a *AnySym) Match(e *Engine, r Element) bool {
	if _, ok := r.Token().(*Symbol); ok {
		e.Matched3E(a.Self(), r, r)
		return true
	} else {
		return e.ResolveE(a.Self(), r)
	}
}

type AnyChr struct {
	Anything
}

func NewAnyChr(x string) *AnyChr {
	return ReSelf(&AnyChr{Anything: *NewAnything(x)})
}

func (a *AnyChr) Match(e *Engine, r Element) bool {
	if _, ok := r.Token().(*Chr); ok {
		e.Matched3E(a.Self(), r, r)
		return true
	} else {
		return e.ResolveE(a.Self(), r)
	}
}

type LnoSym struct {
	Symbol
}

func NewLnoSym(x string) *LnoSym {
	return ReSelf(&LnoSym{Symbol: *NewSymbol(x)})
}

func (l *LnoSym) Match(e *Engine, r Element) bool {
	return e.Matched3E(l.Self(), nil, NewNumber(LMNumber(e.Lineno())))
}

type IfnSym struct {
	Symbol
}

func NewIfnSym(x string) *IfnSym {
	return ReSelf(&IfnSym{Symbol: *NewSymbol(x)})
}

func (i *IfnSym) Match(e *Engine, r Element) bool {
	return e.Matched3E(i, nil, NewSym(e.Filename()))
}

type FlagSym struct {
	Symbol
}

func NewFlagSym(x string) *FlagSym {
	return ReSelf(&FlagSym{Symbol: *NewSymbol(x)})
}

func (f *FlagSym) Match(e *Engine, r Element) bool {
	e.flagErrors++
	message := fmt.Sprintf("%s:%d: ", e.Filename(), e.Lineno())
	return e.Matched3E(f.Self(), nil, NewSym(message))
}

type WarnSym struct {
	Symbol
}

func NewWarnSym(x string) *WarnSym {
	return ReSelf(&WarnSym{Symbol: *NewSymbol(x)})
}

func (w *WarnSym) Match(e *Engine, r Element) bool {
	e.warnErrors++
	message := fmt.Sprintf("%s:%d: ", e.Filename(), e.Lineno())
	return e.Matched3E(w.Self(), nil, NewSym(message))
}

type RepnSym struct {
	Symbol
}

func NewRepnSym(x string) *RepnSym {
	return ReSelf(&RepnSym{Symbol: *NewSymbol(x)})
}

func (r *RepnSym) Match(e *Engine, _ Element) bool {
	n := e.lhsStream.Popx().ToVal()
	if !n.IsNumber() {
		fail("repeat count is not a number: %s", n.ToString())
	}
	return e.Repeat(n.ToInt())
}

type RepSym struct {
	Symbol
}

func NewRepSym(x string) *RepSym {
	return ReSelf(&RepSym{Symbol: *NewSymbol(x)})
}

func (r *RepSym) Match(e *Engine, _ Element) bool {
	return e.Repeat(0)
}

type OptSym struct {
	Symbol
}

func NewOptSym(x string) *OptSym {
	return ReSelf(&OptSym{Symbol: *NewSymbol(x)})
}

func (o *OptSym) Match(e *Engine, r Element) bool {
	return e.Repeat(1)
}

type DropF struct {
	Primitive
}

func NewDropF(x string) *DropF {
	return ReSelf(&DropF{Primitive: *NewPrimitiveFromString(x)})
}

func (d *DropF) Act(sr *Stream, s GenMode) GenMode {
	sr.ClearX()
	return s
}
