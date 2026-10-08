package machine

import "github.com/msorc/languagemachine2/internal/conv"

type declareVar struct {
	genericElement
}

func newDeclareVar() *declareVar {
	return makeSelf[declareVar]()
}

func (nv *declareVar) ToString() string {
	return "newvar"
}

func (nv *declareVar) act(sr *Stream, s GenMode) GenMode {
	v := sr.popX().ToVal()
	k := sr.popX()
	s.makeVar(k, v, s, sr.variables)
	return s
}

type eachRef struct {
	genericElement
	k Element
}

func newEachRef(x Element) *eachRef {
	el := makeSelf[eachRef]()
	el.k = x
	return el
}

func (er *eachRef) ToString() string {
	return "each " + er.k.ToString()
}

func (er *eachRef) act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.eachRef(s, er.k, s)
}

func (er *eachRef) match(e *Engine, r Element) bool {
	return false
}

type allRef struct {
	genericElement
	k Element
}

func newAllRef(x Element) *allRef {
	el := makeSelf[allRef]()
	el.k = x
	return el
}

func (ar *allRef) ToString() string {
	return "all " + ar.k.ToString()
}

func (ar *allRef) act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.allRef(s, ar.k, s)
}

func (ar *allRef) match(e *Engine, r Element) bool {
	return false
}

type varSym struct {
	symbol
}

func newVarSym(x string) *varSym {
	el := makeSelf[varSym]()
	el.v = x
	return el
}

func (vs *varSym) toDump() string {
	return "v:" + conv.Encode(vs.v)
}

func (vs *varSym) act(sr *Stream, s GenMode) GenMode {
	return vs.self().reference(sr, s, s)
}

func (vs *varSym) match(e *Engine, r Element) bool {
	return false
}

func (vs *varSym) reference(sr *Stream, s GenMode, x scopeHolder) GenMode {
	return sr.Engine.theRef(s, vs.self(), x)
}
