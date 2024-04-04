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

func (t *Tracer) TraceDebug(s string, l, r GrammarElement) {
	t.Trace(DEBUG, s, l, r)
}

func (t *Tracer) MatchSymbols(l, r GrammarElement) {
	t.Trace(SYMBOLS, "--", l, r)
}

func (t *Tracer) Resolve(l, r GrammarElement, p uint) {
	t.TraceFull(MISMATCH, "??", l, r, p)
}

func (t *Tracer) Back(l, r GrammarElement) {
	t.Trace(MISMATCH, "**", l, r)
}

func (t *Tracer) BindCvar(l, r GrammarElement) {
	t.Trace(CVAR, "cV", l, r)
}

func (t *Tracer) BindLvar(l, r GrammarElement) {
	t.Trace(LVAR, "lV", l, r)
}

func (t *Tracer) TraceAct(sr *Stream, x GrammarElement) {
	t.Dumpx(sr, ACT, "ACT", x)
}

func (t *Tracer) TraceApply(sr *Stream, x GrammarElement) {
	t.Dumpx(sr, APPLY, "APPLY", x)
}

func (t *Tracer) TraceArithmetic(sr *Stream, x GrammarElement) {
	t.Dumpx(sr, ARITHMETIC, "ARITHMETIC", x)
}

func (t *Tracer) TraceRelation(sr *Stream, x GrammarElement) {
	t.Dumpx(sr, RELATION, "RELATION", x)
}

func (t *Tracer) TraceAssignment(sr *Stream, x GrammarElement) {
	t.Dumpx(sr, ASSIGN, "ASSIGN", x)
}

func (t *Tracer) TraceIndex(sr *Stream, x GrammarElement) {
	t.Dumpx(sr, INDEX, "INDEX", x)
}

func (t *Tracer) TraceLoop(sr *Stream, x GrammarElement) {
	t.Dumpx(sr, LOOP, "LOOP", x)
}

func (t *Tracer) Dumpx(sr *Stream, bits uint, s string, x GrammarElement) {
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
		//+        t.E.Display.Repeat(i, t.E.Lhx.St.Si, t.E.Rhr.Sm.Cx.St.Si, t.E.Lhx.Cd, t.E.Rhr.Sm.Cx.Cd)
	}
}

func (t *Tracer) RuleScope(s string, st *State, pp, pq *Var) {
	if t.Flags&DIAGRAM != 0 {
		//+ t.E.Display.Replace(s, st.Si, t.E.Rhr.Sm.Cx.St.Si, t.E.Lhx.Cd, t.E.Rhr.Sm.Cx.Cd)
	} else {
		t.Dumpvars(CXSCOPE, "CXSCOPE", pp, pq)
	}
}

func (t *Tracer) BindRvar(l, r GrammarElement) {
	t.Trace(RVAR, "rV", l, r)
}

func (t *Tracer) BindRvarScope(lv, rv GrammarElement, c LMScope) {}

func (t *Tracer) BindRvarScopeVars(lv, rv GrammarElement, pp, pq *Var) {
	t.Trace(RVAR_VAR, "RVAR", lv, rv)
	t.Dumpvars(RVARSCOPE, "RVARSCOPE", pp, pq)
}

func (t *Tracer) TheRefScope(pk GrammarElement, c LMScope) {
	t.TheRefVars(pk, c.VvP(), c.VvQ())
}

func (t *Tracer) TheRefVars(pk GrammarElement, pp, pq *Var) {
	t.Dumpvar(REF, "REF", pp)
	t.Dumpvars(REFSCOPE, "REFSCOPE", pp, pq)
}

func (t *Tracer) ToValueScope(pk GrammarElement, c LMScope) {}

func (t *Tracer) ToValueVars(pk GrammarElement, pp, pq *Var) {
	t.Dumpvar(REF, "TOVALUE", pp)
	t.Dumpvars(REFSCOPE, "REFSCOPE", pp, pq)
}

func (t *Tracer) TheRefVar(pp *Var) {
	t.Dumpvar(REFVAR, "REFVAR", pp)
}

func (t *Tracer) EachRefScope(pk GrammarElement, c LMScope) {
	t.EachRefVars(pk, c.VvP(), c.VvQ())
}

func (t *Tracer) EachRefVars(pk GrammarElement, pp, pq *Var) {
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
	if sr.CV != nil {
		if sr.CI < uint(len(sr.CV)) {
			x := sr.CV[sr.CI]
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

func (t *Tracer) TraceFull(bits uint, s string, l, r GrammarElement, p uint) {
	if t.Flags&bits != 0 {
		//+ g := t.E.Lhx.St.Gr.Sy
		// gs := "---"
		// if g != nil {
		//     gs = g.ToTrace()
		// }
		// ls := "---"
		// if l != nil {
		//     ls = l.ToTrace()
		// }
		// rs := "---"
		// if r != nil {
		//     rs = r.ToTrace()
		// }
		// es := "---"
		// if t.E.Rsy != nil {
		//     es = t.E.Rsy.ToTrace()
		// }
		// pd := Grammar.Priassoc(p)
		// pv := Grammar.Privalue(p)
		// ld := t.E.Lhx.Cd
		// rd := t.E.Rhr.Sm.Cx.Cd
		// lk := uint(t.E.Lhr.Lk)
		// rk := uint(t.E.Rhr.Lk)
		// if t.Flags&DIAGRAM != 0 {
		//     t.E.Display.Trace(s, t.E.Lhr.Sm.Cx.St.Si, t.E.Rhr.Sm.Cx.St.Si, ld, rd, ls, rs, es)
		// } else {
		//     fmt.Printf("\t%4d %4s %4d %4d %5d%s %4d %4d %4d %6d %8s%12s%12s%12s\n",
		//         t.E.Lineno, s, ld, rd, pv, pd, t.E.Lhr.Cz, t.E.Lhr.Ci, t.E.Rhr.Cz, t.E.Rhr.Ci, gs, ls, rs, es)
		// }
	}
}

func (t *Tracer) Trace(bits uint, s string, l, r GrammarElement) {
	//+ t.TraceFull(bits, s, l, r, t.E.Lhx.Pr)
}

func (t *Tracer) Dumpit(bits uint, s string, x GrammarElement) {
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
