package machine

import "fmt"

type doneF struct {
	symbol
}

func newDoneF(x string) *doneF {
	el := makeSelf[doneF]()
	el.v = x
	return el
}

func (df *doneF) match(e *Engine, r Element) bool {
	e.lhsStream.variables = e.lhsContext.Variables()
	return e.matchedWith(df.self(), nil, nil)
}

type takeF struct {
	symbol
}

func newTakeF(x string) *takeF {
	el := makeSelf[takeF]()
	el.v = x
	return el
}

func (tf *takeF) match(e *Engine, r Element) bool {
	if r.token() == tf.self() { // %  %
		e.takeTvar()
		return e.matchedWith(tf.self(), r, nil)
	}
	if _, ok := r.(*bindF); ok { // %  :
		e.pushX()
		return e.matchedWith(tf.self(), r, nil)
	}
	if e.rsLastMatchElement != nil { // %  [matched]
		e.pushR(e.rsLastMatchElement)
		return e.matchedWith(tf.self(), nil, nil)
	}
	return false
}

type bindF struct {
	symbol
}

func newBindF(x string) *bindF {
	el := makeSelf[bindF]()
	el.v = x
	return el
}

func (b *bindF) match(e *Engine, r Element) bool {
	if r.token() == b.self() {
		a := e.lhsStream.popX()
		bElem := e.rhsStream.popX().ToVal()
		if a, ok := a.(*varSym); ok {
			e.bindUvar(a, bElem)
			return e.matchedWith(b.self(), r, nil)
		}
		e.matchedWith(b.self(), r, nil)
		lh := []Element{a}
		e.lhsStream.mode = newSTMode(e.lhsStream.mode, lh, e.lhsStream.mode)
		rh := []Element{bElem}
		e.rhsStream.mode = newSTMode(e.rhsStream.mode, rh, e.rhsStream.mode)
		return true
	}
	if _, ok := r.(*takeF); ok {
		e.bindTvar()
		return e.matchedWith(b.self(), r, nil)
	}
	if e.rsLastMatchElement != nil {
		e.bindXvar(e.rsLastMatchElement)
		return e.matchedWith(b.self(), nil, r)
	}
	return false
}

type appendXSym struct {
	symbol
}

func newAppendXSym(x string) *appendXSym {
	el := makeSelf[appendXSym]()
	el.v = x
	return el
}

// if there is captured material, append it
// otherwise match one symbol and append that
func (a *appendXSym) match(e *Engine, r Element) bool {
	sr := e.lhsStream
	b := sr.popX()
	if !e.lhsStream.emptyX() {
		v := e.lhsStream.toRow()
		for _, x := range v {
			b.append(sr, x)
		}
		e.matchedWith(a.self(), nil, nil)
	} else {
		b.append(sr, r)
		e.matchedWith(a.self(), r, r)
	}
	return true
}

type errSym struct {
	symbol
	engine *Engine
}

func newErrSym(e *Engine, x string) *errSym {
	el := makeSelf[errSym]()
	el.v = x
	el.engine = e
	return el
}

func (e *errSym) append(sr *Stream, y Element) Element {
	e.engine.writeErr(y.ToString())
	return e.self()
}

func (e *errSym) match(engine *Engine, r Element) bool {
	e.engine.writeErr(r.ToString())
	return engine.matchedWith(e.self(), r, r)
}

type outSym struct {
	symbol
	engine *Engine
}

func newOutSym(e *Engine, x string) *outSym {
	el := makeSelf[outSym]()
	el.v = x
	el.engine = e
	return el
}

func (o *outSym) append(sr *Stream, y Element) Element {
	_, _ = o.engine.out.WriteString(y.ToString())
	return o.self()
}

func (o *outSym) match(engine *Engine, r Element) bool {
	_, _ = o.engine.out.WriteString(r.ToString())
	return engine.matchedWith(o.self(), r, r)
}

type uriSym struct {
	symbol
	engine *Engine
}

func newUriSym(e *Engine, x string) *uriSym {
	el := makeSelf[uriSym]()
	el.v = x
	el.engine = e
	return el
}

func (u *uriSym) append(sr *Stream, y Element) Element {
	_, _ = u.engine.out.WriteString(y.toEncode())
	return u.self()
}

func (u *uriSym) match(engine *Engine, r Element) bool {
	_, _ = u.engine.out.WriteString(r.toEncode())
	return engine.matchedWith(u.self(), r, r)
}

type urdSym struct {
	symbol
	engine *Engine
}

func newUrdSym(e *Engine, x string) *urdSym {
	el := makeSelf[urdSym]()
	el.v = x
	el.engine = e
	return el
}

func (u *urdSym) append(sr *Stream, y Element) Element {
	_, _ = u.engine.out.WriteString(y.toDecode())
	return u.self()
}

func (u *urdSym) match(engine *Engine, r Element) bool {
	_, _ = u.engine.out.WriteString(r.toDecode())
	return engine.matchedWith(u.self(), r, r)
}

type spSym struct {
	symbol
}

func newSpSym(x string) *spSym {
	el := makeSelf[spSym]()
	el.v = x
	return el
}

func (s *spSym) toEncode() string {
	return " "
}

type nlSym struct {
	symbol
}

func newNlSym(x string) *nlSym {
	el := makeSelf[nlSym]()
	el.v = x
	return el
}

func (n *nlSym) toEncode() string {
	return "\n"
}

type getF struct {
	symbol
}

func newGetF(x string) *getF {
	el := makeSelf[getF]()
	el.v = x
	return el
}

func (g *getF) act(sr *Stream, s GenMode) GenMode {
	return sr.getX(s)
}

type trueSym struct {
	symbol
}

func newTrueSym(x string) *trueSym {
	el := makeSelf[trueSym]()
	el.v = x
	return el
}

func (t *trueSym) ToVal() Element {
	return newBoolean(true)
}

type falseSym struct {
	symbol
}

func newFalseSym(x string) *falseSym {
	el := makeSelf[falseSym]()
	el.v = x
	return el
}

func (f *falseSym) ToVal() Element {
	return newBoolean(false)
}

type trueF struct {
	symbol
}

func newTrueF(x string) *trueF {
	el := makeSelf[trueF]()
	el.v = x
	return el
}

func (t *trueF) act(sr *Stream, s GenMode) GenMode {
	sr.pushX(newBoolean(true))
	return s
}

type falseF struct {
	symbol
}

func newFalseF(x string) *falseF {
	el := makeSelf[falseF]()
	el.v = x
	return el
}

func (f *falseF) act(sr *Stream, s GenMode) GenMode {
	sr.pushX(newBoolean(false))
	return s
}

type getXF struct {
	symbol
	v Element
}

func newGetXF(x Element) *getXF {
	el := makeSelf[getXF]()
	el.v = x
	return el
}

func (g *getXF) toTrace() string {
	return g.v.toTrace() + "p"
}

func (g *getXF) act(sr *Stream, s GenMode) GenMode {
	sr.pushX(g.v)
	return s
}

type getVF struct {
	symbol
	v Element
}

func newGetVF(x Element) *getVF {
	el := makeSelf[getVF]()
	el.v = x
	return el
}

func (g *getVF) toTrace() string {
	return g.v.toTrace()
}

func (g *getVF) act(sr *Stream, s GenMode) GenMode {
	sr.pushX(newVarRefOf(g.v, s))
	return s
}

type actF struct {
	symbol
}

func newActF(x string) *actF {
	el := makeSelf[actF]()
	el.v = x
	return el
}

func (a *actF) trace(sr *Stream, t *tracer) {
	t.traceAct(sr, a.self())
}

func (a *actF) act(sr *Stream, s GenMode) GenMode {
	fail("the act primitive (a) is not supported")
	return nil
}

type primitive struct {
	symbol
}

func allocPrimitive() *primitive {
	return makeSelf[primitive]()
}

func newPrimitive(x string) *primitive {
	el := allocPrimitive()
	el.v = x
	return el
}

func (p *primitive) act(sr *Stream, s GenMode) GenMode {
	sr.Engine.printf("act: %s\n", p.v)
	return s
}

type applyF struct {
	primitive
}

func newApplyF(x string) *applyF {
	return reSelf(&applyF{primitive: *newPrimitive(x)})
}

func (a *applyF) trace(s *Stream, t *tracer) {
	t.traceApply(s, a.self())
}

func (a *applyF) act(sr *Stream, s GenMode) GenMode {
	v := sr.popX()
	return v.act(sr, s)
}

type injF struct {
	symbol
}

func newInjF(x string) *injF {
	el := makeSelf[injF]()
	el.v = x
	return el
}

func (i *injF) match(e *Engine, r Element) bool {
	e.pushRHX(e.lhsStream.popX())
	return e.matchedWith(i.self(), nil, nil)
}

type strF struct {
	symbol
}

func newStrF(x string) *strF {
	el := makeSelf[strF]()
	el.v = x
	return el
}

func (s *strF) act(sr *Stream, mode GenMode) GenMode {
	//  { return new stMode(s, (cast(str)sr.popx()).v, s); }
	return mode
}

type anything struct {
	symbol
}

func newAnything(x string) *anything {
	el := makeSelf[anything]()
	el.v = x
	return el
}

func (a *anything) match(e *Engine, r Element) bool {
	e.matchedWith(a.self(), r, r)
	return true
}

type anySym struct {
	anything
}

func newAnySym(x string) *anySym {
	return reSelf(&anySym{anything: *newAnything(x)})
}

func (a *anySym) match(e *Engine, r Element) bool {
	if _, ok := r.token().(*symbol); ok {
		e.matchedWith(a.self(), r, r)
		return true
	} else {
		return e.resolve(a.self(), r)
	}
}

type anyChr struct {
	anything
}

func newAnyChr(x string) *anyChr {
	return reSelf(&anyChr{anything: *newAnything(x)})
}

func (a *anyChr) match(e *Engine, r Element) bool {
	if _, ok := r.token().(*chr); ok {
		e.matchedWith(a.self(), r, r)
		return true
	} else {
		return e.resolve(a.self(), r)
	}
}

type lnoSym struct {
	symbol
}

func newLnoSym(x string) *lnoSym {
	return reSelf(&lnoSym{symbol: *newSymbol(x)})
}

func (l *lnoSym) match(e *Engine, r Element) bool {
	return e.matchedWith(l.self(), nil, newNumber(LMNumber(e.lineNo())))
}

type ifnSym struct {
	symbol
}

func newIfnSym(x string) *ifnSym {
	return reSelf(&ifnSym{symbol: *newSymbol(x)})
}

func (i *ifnSym) match(e *Engine, r Element) bool {
	return e.matchedWith(i, nil, newSym(e.Filename()))
}

type flagSym struct {
	symbol
}

func newFlagSym(x string) *flagSym {
	return reSelf(&flagSym{symbol: *newSymbol(x)})
}

func (f *flagSym) match(e *Engine, r Element) bool {
	e.flagErrors++
	message := fmt.Sprintf("%s:%d: ", e.Filename(), e.lineNo())
	return e.matchedWith(f.self(), nil, newSym(message))
}

type warnSym struct {
	symbol
}

func newWarnSym(x string) *warnSym {
	return reSelf(&warnSym{symbol: *newSymbol(x)})
}

func (w *warnSym) match(e *Engine, r Element) bool {
	e.warnErrors++
	message := fmt.Sprintf("%s:%d: ", e.Filename(), e.lineNo())
	return e.matchedWith(w.self(), nil, newSym(message))
}

type repnSym struct {
	symbol
}

func newRepnSym(x string) *repnSym {
	return reSelf(&repnSym{symbol: *newSymbol(x)})
}

func (r *repnSym) match(e *Engine, _ Element) bool {
	n := e.lhsStream.popX().ToVal()
	if !n.IsNumber() {
		fail("repeat count is not a number: %s", n.ToString())
	}
	return e.repeat(n.toInt())
}

type repSym struct {
	symbol
}

func newRepSym(x string) *repSym {
	return reSelf(&repSym{symbol: *newSymbol(x)})
}

func (r *repSym) match(e *Engine, _ Element) bool {
	return e.repeat(0)
}

type optSym struct {
	symbol
}

func newOptSym(x string) *optSym {
	return reSelf(&optSym{symbol: *newSymbol(x)})
}

func (o *optSym) match(e *Engine, r Element) bool {
	return e.repeat(1)
}

type dropF struct {
	primitive
}

func newDropF(x string) *dropF {
	return reSelf(&dropF{primitive: *newPrimitive(x)})
}

func (d *dropF) act(sr *Stream, s GenMode) GenMode {
	sr.clearX()
	return s
}
