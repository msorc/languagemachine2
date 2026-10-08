package machine

import (
	"fmt"
	"io"
	"reflect"
)

const (
	MISMATCH   int = 0x0000001
	SYMBOLS    int = 0x0000002
	CXSCOPE    int = 0x0000004
	CVAR       int = 0x0000008
	LVAR       int = 0x0000010
	RVAR       int = 0x0000020
	RVAR_VAR   int = 0x0000040
	RVARSCOPE  int = 0x0000080
	REF        int = 0x0000100
	REFSCOPE   int = 0x0000200
	REFVAR     int = 0x0000400
	EACH       int = 0x0000800
	EACHSCOPE  int = 0x0001000
	EACHREFVAR int = 0x0002000
	DEBUG      int = 0x0004000
	ACT        int = 0x0008000
	APPLY      int = 0x0010000
	ARITHMETIC int = 0x0020000
	RELATION   int = 0x0040000
	ASSIGN     int = 0x0080000
	INDEX      int = 0x0100000
	LOOP       int = 0x0200000
	LOAD       int = 0x0400000
	DIAGRAM    int = 0x0800000
	DIAGRAMT   int = 0x1000000
	GRAMMAR    int = 0x2000000
)

func priValue(pri int) int { return (pri & PRIMASK) / 2 }

func priAssoc(pri int) string {
	if pri == 0 {
		return "L"
	}
	if pri&BRACKET != 0 {
		return "B"
	}
	if pri&1 != 0 {
		return "R"
	}
	return "L"
}

type Tracer struct {
	E     *Engine
	Flags int
}

func NewTracer(e *Engine) *Tracer {
	return &Tracer{E: e}
}

// Tracing reports whether any of bits is being traced.
func (t *Tracer) Tracing(bits int) bool {
	return t.Flags&bits != 0
}

func (t *Tracer) MatchSymbols(l, r Element) {
	t.Trace(SYMBOLS, "--", l, r)
}

func (t *Tracer) Resolve(l, r Element, p int) {
	t.TraceFull(MISMATCH, "??", l, r, p)
}

func (t *Tracer) Back(l, r Element) {
	t.Trace(MISMATCH, "**", l, r)
}

func (t *Tracer) BindCvar(l, r Element) {
	t.Trace(CVAR, "cV", l, r)
}

func (t *Tracer) BindLvar(l, r Element) {
	t.Trace(LVAR, "lV", l, r)
}

func (t *Tracer) TraceAct(sr *Stream, x Element) {
	t.Dumpx(sr, ACT, "ACT", x)
}

func (t *Tracer) TraceApply(sr *Stream, x Element) {
	t.Dumpx(sr, APPLY, "APPLY", x)
}

func (t *Tracer) TraceArithmetic(sr *Stream, x Element) {
	t.Dumpx(sr, ARITHMETIC, "ARITHMETIC", x)
}

func (t *Tracer) TraceRelation(sr *Stream, x Element) {
	t.Dumpx(sr, RELATION, "RELATION", x)
}

func (t *Tracer) TraceAssignment(sr *Stream, x Element) {
	t.Dumpx(sr, ASSIGN, "ASSIGN", x)
}

func (t *Tracer) TraceIndex(sr *Stream, x Element) {
	t.Dumpx(sr, INDEX, "INDEX", x)
}

func (t *Tracer) TraceLoop(sr *Stream, x Element) {
	t.Dumpx(sr, LOOP, "LOOP", x)
}

func (t *Tracer) Dumpx(sr *Stream, bits int, s string, x Element) {
	if t.Flags&bits != 0 {
		t.Dumpit(bits, s, x)
		sr.Dumpx()
	}
}

func (t *Tracer) Dumpg(gr *Grammar) {
	if t.Flags&GRAMMAR != 0 {
		gr.Dump(t.E.out)
	}
}

func (t *Tracer) Repeat(i int) {
	if t.Flags&DIAGRAM != 0 {
		t.E.display.Repeat(i, t.E.lhsContext.State().stateIndex, t.E.rhsStream.mode.ContextMode().State().stateIndex, t.E.lhsContext.NestingDepth(), t.E.rhsStream.mode.ContextMode().NestingDepth())
	}
}

func (t *Tracer) RuleScope(s string, st *State, pp, pq VarElement) {
	if t.Flags&DIAGRAM != 0 {
		t.E.display.Replace(s, st.stateIndex, t.E.rhsStream.mode.ContextMode().State().stateIndex, t.E.lhsContext.NestingDepth(), t.E.rhsStream.mode.ContextMode().NestingDepth())
	} else {
		t.Dumpvars(CXSCOPE, "CXSCOPE", pp, pq)
	}
}

func (t *Tracer) BindRvar(l, r Element) {
	t.Trace(RVAR, "rV", l, r)
}

func (t *Tracer) BindRvarScopeVars(lv, rv Element, pp, pq VarElement) {
	t.Trace(RVAR_VAR, "RVAR", lv, rv)
	t.Dumpvars(RVARSCOPE, "RVARSCOPE", pp, pq)
}

func (t *Tracer) TheRefVars(pk Element, pp, pq VarElement) {
	t.Dumpvar(REF, "REF", pp)
	t.Dumpvars(REFSCOPE, "REFSCOPE", pp, pq)
}

func (t *Tracer) TheRefVar(pp VarElement) {
	t.Dumpvar(REFVAR, "REFVAR", pp)
}

func (t *Tracer) EachRefVars(pk Element, pp, pq VarElement) {
	t.Dumpvar(EACH, "EACH", pp)
	t.Dumpvars(EACHSCOPE, "EACHSCOPE", pp, pq)
}

func (t *Tracer) EachRefVar(pp VarElement) {
	t.Dumpvar(EACHREFVAR, "EACHREF", pp)
}

func (t *Tracer) TraceShort(b GenMode) {
	sr := b.Stream()
	if sr.codeVector != nil {
		if sr.codeIndex < len(sr.codeVector) {
			x := sr.codeVector[sr.codeIndex]
			if t.Flags&DEBUG != 0 {
				b.Trace(x)
			}
			if x != nil {
				x.Trace(sr, t)
			}
		} else {
			b.TraceRet(t)
		}
	}
}

func (t *Tracer) TraceFull(bits int, s string, l, r Element, p int) {
	if t.Flags&bits != 0 {
		g := t.E.lhsContext.State().grammar.symbol
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
		ld := t.E.lhsContext.NestingDepth()
		rd := t.E.rhsStream.mode.ContextMode().NestingDepth()
		if t.Flags&DIAGRAM != 0 {
			t.E.display.Trace(s, t.E.lhsStream.mode.ContextMode().State().stateIndex, t.E.rhsStream.mode.ContextMode().State().stateIndex, ld, rd, ls, rs, es)
		} else {
			// the zero columns are the code indexes of compiled (C/D) rules
			// in the original, which interpreted rules leave at 0
			t.E.printf("\t%4d %4s %4d %4d %5d%s %4d %4d %4d %6d %8s%12s%12s%12s\n",
				t.E.Lineno(), s, ld, rd, pv, pd, 0, t.E.lhsStream.codeIndex, 0, t.E.rhsStream.codeIndex, gs, ls, rs, es)
		}
	}
}

func (t *Tracer) Trace(bits int, s string, l, r Element) {
	t.TraceFull(bits, s, l, r, t.E.lhsContext.Priority())
}

// Dumpit names the operator or statement x that is about to act, as the
// original's writefln("%s", x) did through toString.
func (t *Tracer) Dumpit(bits int, s string, x Element) {
	if t.Flags&bits != 0 {
		t.E.printf("\t%s\t%s\n", s, x.ToString())
	}
}

func (t *Tracer) Dumpvar(bits int, s string, p VarElement) {
	if t.Flags&bits != 0 {
		TxE(t.E.out, s, p)
	}
}

func (t *Tracer) Dumpvars(bits int, s string, p, q VarElement) {
	if t.Flags&bits != 0 {
		t.E.printf("VARIABLES: %s\n", s)
		for p != nil {
			marker := "-"
			if p == q {
				marker = "*"
			}
			TxV(t.E.out, "VV", marker, p)
			p = p.Link()
		}
		t.E.printf("---------\n")
	}
}

// addr is the address traced for x, 0 for nil, as in the original.
func addr(x any) uintptr {
	if x == nil {
		return 0
	}
	if v := reflect.ValueOf(x); v.Kind() == reflect.Pointer {
		return v.Pointer()
	}
	return 0
}

// TxE writes a trace line for x (its address and trace form) and returns x.
func TxE(w io.Writer, s string, x Element) Element {
	var xtrace string
	if x != nil {
		xtrace = x.ToTrace()
	} else {
		xtrace = "---"
	}
	_, _ = fmt.Fprintf(w, "\t%6s:     %8X %24s\n", s, addr(x), xtrace)
	return x
}

// TxV writes a trace line for the variable w and its links, and returns w.
func TxV(out io.Writer, r, s string, w VarElement) Element {
	var n string

	if w != nil {
		n = w.ToDump()
	} else {
		n = "---"
	}

	_, _ = fmt.Fprintf(out, "\t%6s:%4s %8X %24s %8X %8X %8X %8X\n", r, s, addr(w), n, addr(w.Value()), addr(w.Variables()), addr(w.ScopeReferenceContext()), addr(w.Link()))

	return w
}
