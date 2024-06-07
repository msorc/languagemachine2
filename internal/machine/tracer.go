package machine

import (
	"fmt"
)

const (
	MISMATCH   uint = 0x0000001
	SYMBOLS    uint = 0x0000002
	CXSCOPE    uint = 0x0000004
	CVAR       uint = 0x0000008
	LVAR       uint = 0x0000010
	RVAR       uint = 0x0000020
	RVAR_VAR   uint = 0x0000040
	RVARSCOPE  uint = 0x0000080
	REF        uint = 0x0000100
	REFSCOPE   uint = 0x0000200
	REFVAR     uint = 0x0000400
	EACH       uint = 0x0000800
	EACHSCOPE  uint = 0x0001000
	EACHREFVAR uint = 0x0002000
	DEBUG      uint = 0x0004000
	ACT        uint = 0x0008000
	APPLY      uint = 0x0010000
	ARITHMETIC uint = 0x0020000
	RELATION   uint = 0x0040000
	ASSIGN     uint = 0x0080000
	INDEX      uint = 0x0100000
	LOOP       uint = 0x0200000
	LOAD       uint = 0x0400000
	DIAGRAM    uint = 0x0800000
	DIAGRAMT   uint = 0x1000000
	GRAMMAR    uint = 0x2000000
)

type Tracer struct {
	E     *Engine
	Flags uint
}

func NewTracer(e *Engine) *Tracer {
	return &Tracer{E: e}
}

func (t *Tracer) Tracing(bits uint) uint {
	return t.Flags & bits
}

func (t *Tracer) TraceDebug(s string, l, r MachineElement) {
	t.Trace(DEBUG, s, l, r)
}

func (t *Tracer) MatchSymbols(l, r MachineElement) {
	t.Trace(SYMBOLS, "--", l, r)
}

func (t *Tracer) Resolve(l, r MachineElement, p uint) {
	t.TraceFull(MISMATCH, "??", l, r, p)
}

func (t *Tracer) Back(l, r MachineElement) {
	t.Trace(MISMATCH, "**", l, r)
}

func (t *Tracer) BindCvar(l, r MachineElement) {
	t.Trace(CVAR, "cV", l, r)
}

func (t *Tracer) BindLvar(l, r MachineElement) {
	t.Trace(LVAR, "lV", l, r)
}

func (t *Tracer) TraceAct(sr *Stream, x MachineElement) {
	t.Dumpx(sr, ACT, "ACT", x)
}

func (t *Tracer) TraceApply(sr *Stream, x MachineElement) {
	t.Dumpx(sr, APPLY, "APPLY", x)
}

func (t *Tracer) TraceArithmetic(sr *Stream, x MachineElement) {
	t.Dumpx(sr, ARITHMETIC, "ARITHMETIC", x)
}

func (t *Tracer) TraceRelation(sr *Stream, x MachineElement) {
	t.Dumpx(sr, RELATION, "RELATION", x)
}

func (t *Tracer) TraceAssignment(sr *Stream, x MachineElement) {
	t.Dumpx(sr, ASSIGN, "ASSIGN", x)
}

func (t *Tracer) TraceIndex(sr *Stream, x MachineElement) {
	t.Dumpx(sr, INDEX, "INDEX", x)
}

func (t *Tracer) TraceLoop(sr *Stream, x MachineElement) {
	t.Dumpx(sr, LOOP, "LOOP", x)
}

func (t *Tracer) Dumpx(sr *Stream, bits uint, s string, x MachineElement) {
	if t.Flags&bits != 0 {
		t.Dumpit(bits, s, x)
		sr.Dumpx()
	}
}

func (t *Tracer) Dumpg(gr *Grammar) {
	if t.Flags&GRAMMAR != 0 {
		gr.Dump()
	}
}

func (t *Tracer) Repeat(i uint) {
	if t.Flags&DIAGRAM != 0 {
		t.E.display.Repeat(i, t.E.lhsContext.St().Si, t.E.rhsStream.mode.CX().St().Si, t.E.lhsContext.Cd(), t.E.rhsStream.mode.CX().Cd())
	}
}

func (t *Tracer) RuleScope(s string, st *State, pp, pq *Var) {
	if t.Flags&DIAGRAM != 0 {
		t.E.display.Replace(s, st.Si, t.E.rhsStream.mode.CX().St().Si, t.E.lhsContext.Cd(), t.E.rhsStream.mode.CX().Cd())
	} else {
		t.Dumpvars(CXSCOPE, "CXSCOPE", pp, pq)
	}
}

func (t *Tracer) BindRvar(l, r MachineElement) {
	t.Trace(RVAR, "rV", l, r)
}

func (t *Tracer) BindRvarScope(lv, rv MachineElement, c LMScope) {}

func (t *Tracer) BindRvarScopeVars(lv, rv MachineElement, pp, pq *Var) {
	t.Trace(RVAR_VAR, "RVAR", lv, rv)
	t.Dumpvars(RVARSCOPE, "RVARSCOPE", pp, pq)
}

func (t *Tracer) TheRefScope(pk MachineElement, c LMScope) {
	t.TheRefVars(pk, c.VvP(), c.VvQ())
}

func (t *Tracer) TheRefVars(pk MachineElement, pp, pq *Var) {
	t.Dumpvar(REF, "REF", pp)
	t.Dumpvars(REFSCOPE, "REFSCOPE", pp, pq)
}

func (t *Tracer) ToValueScope(pk MachineElement, c LMScope) {}

func (t *Tracer) ToValueVars(pk MachineElement, pp, pq *Var) {
	t.Dumpvar(REF, "TOVALUE", pp)
	t.Dumpvars(REFSCOPE, "REFSCOPE", pp, pq)
}

func (t *Tracer) TheRefVar(pp *Var) {
	t.Dumpvar(REFVAR, "REFVAR", pp)
}

func (t *Tracer) EachRefScope(pk MachineElement, c LMScope) {
	t.EachRefVars(pk, c.VvP(), c.VvQ())
}

func (t *Tracer) EachRefVars(pk MachineElement, pp, pq *Var) {
	t.Dumpvar(EACH, "EACH", pp)
	t.Dumpvars(EACHSCOPE, "EACHSCOPE", pp, pq)
}

func (t *Tracer) EachRefVar(pp *Var) {
	t.Dumpvar(EACHREFVAR, "EACHREF", pp)
}

func (t *Tracer) T0(bits uint, s string) {
	if t.Flags&bits != 0 {
		fmt.Printf("\tt0: %s\n", s)
	}
}

func (t *Tracer) Depth(c EngineStateContext) uint {
	var n uint
	for n = 0; c != nil; n++ {
		c = c.Cs()
	}
	return n
}

func (t *Tracer) TraceShort(sr *Stream, b GenMode) {
	if sr.codeVector != nil {
		if sr.codeIndex < uint(len(sr.codeVector)) {
			x := sr.codeVector[sr.codeIndex]
			if t.Flags&DEBUG != 0 {
				b.Trace(x)
			}
			if x != nil {
				x.Trace(sr, t)
			}
		} else {
			b.TraceRet(sr, t)
		}
	}
}

func (t *Tracer) TraceFull(bits uint, s string, l, r MachineElement, p uint) {
	if t.Flags&bits != 0 {
		g := t.E.lhsContext.St().Gr.Sy
		gs := "---"
		if g != nil {
			gs = g.ToTrace()
		}
		ls := "---"
		if l != nil {
			ls = l.ToTrace()
		}
		rs := "---"
		if r != nil {
			rs = r.ToTrace()
		}
		es := "---"
		if t.E.rsLastMatchElement != nil {
			es = t.E.rsLastMatchElement.ToTrace()
		}
		pd := priAssoc(p)
		pv := priValue(p)
		ld := t.E.lhsContext.Cd()
		rd := t.E.rhsStream.mode.CX().Cd()
		// lk := uint(t.E.lhsStream.LK)
		// rk := uint(t.E.rhsStream.LK)
		if t.Flags&DIAGRAM != 0 {
			t.E.display.Trace(s, t.E.lhsStream.mode.CX().St().Si, t.E.rhsStream.mode.CX().St().Si, ld, rd, ls, rs, es)
		} else {
			fmt.Printf("\t%4d %4s %4d %4d %5d%s %4d %4d %4d %6d %8s%12s%12s%12s\n",
				t.E.Lineno, s, ld, rd, pv, pd, t.E.lhsStream.compiledRulesCodeIndex, t.E.lhsStream.codeIndex, t.E.rhsStream.compiledRulesCodeIndex, t.E.rhsStream.codeIndex, gs, ls, rs, es)
		}
	}
}

func (t *Tracer) Trace(bits uint, s string, l, r MachineElement) {
	t.TraceFull(bits, s, l, r, t.E.lhsContext.Pr())
}

func (t *Tracer) Dumpit(bits uint, s string, x MachineElement) {
	if t.Flags&bits != 0 {
		fmt.Printf("\t%s\t%s\n", s, x)
	}
}

func (t *Tracer) Dumpvar(bits uint, s string, p *Var) {
	if t.Flags&bits != 0 {
		TxE(s, p)
	}
}

func (t *Tracer) Dumpvars(bits uint, s string, p, q *Var) {
	if t.Flags&bits != 0 {
		fmt.Printf("VARIABLES: %s\n", s)
		for p != nil {
			marker := "-"
			if p == q {
				marker = "*"
			}
			TxV("VV", marker, p)
			p = p.Vs
		}
		fmt.Println("---------")
	}
}
