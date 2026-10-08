package machine

import "github.com/msorc/languagemachine2/internal/conv"

type NewVar struct {
	GenericElement
}

func NewNewVar() *NewVar {
	return MakeSelf[NewVar]()
}

func (nv *NewVar) ToString() string {
	return "newvar"
}

func (nv *NewVar) Act(sr *Stream, s GenMode) GenMode {
	v := sr.Popx().ToVal()
	k := sr.Popx()
	s.MakeVar(k, v, s, sr.variables)
	return s
}

type EachRef struct {
	GenericElement
	K Element
}

func NewEachRef(x Element) *EachRef {
	el := MakeSelf[EachRef]()
	el.K = x
	return el
}

func (er *EachRef) ToString() string {
	return "each " + er.K.ToString()
}

func (er *EachRef) Act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.EachRef(s, er.K, s)
}

func (er *EachRef) Match(e *Engine, r Element) bool {
	return false
}

type AllRef struct {
	GenericElement
	K Element
}

func NewAllRef(x Element) *AllRef {
	el := MakeSelf[AllRef]()
	el.K = x
	return el
}

func (ar *AllRef) ToString() string {
	return "all " + ar.K.ToString()
}

func (ar *AllRef) Act(sr *Stream, s GenMode) GenMode {
	return sr.Engine.AllRef(s, ar.K, s)
}

func (ar *AllRef) Match(e *Engine, r Element) bool {
	return false
}

type VarSym struct {
	Symbol
}

func NewVarSym(x string) *VarSym {
	el := MakeSelf[VarSym]()
	el.V = x
	return el
}

func (vs *VarSym) ToDump() string {
	return "v:" + conv.Encode(vs.V)
}

func (vs *VarSym) Act(sr *Stream, s GenMode) GenMode {
	return vs.Self().Reference(sr, s, s)
}

func (vs *VarSym) Match(e *Engine, r Element) bool {
	return false
}

func (vs *VarSym) Reference(sr *Stream, s GenMode, x ScopeHolder) GenMode {
	return sr.Engine.TheRef(s, vs.Self(), x)
}
