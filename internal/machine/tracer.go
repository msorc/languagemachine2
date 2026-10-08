package machine

import (
	"fmt"
	"io"
	"reflect"
)

// TraceFlag selects categories of trace output (-trace); the flags combine
// as a bit set.
type TraceFlag uint32

// The trace categories. The values are the original's.
const (
	MISMATCH   TraceFlag = 0x0000001
	SYMBOLS    TraceFlag = 0x0000002
	CXSCOPE    TraceFlag = 0x0000004
	CVAR       TraceFlag = 0x0000008
	LVAR       TraceFlag = 0x0000010
	RVAR       TraceFlag = 0x0000020
	RVAR_VAR   TraceFlag = 0x0000040
	RVARSCOPE  TraceFlag = 0x0000080
	REF        TraceFlag = 0x0000100
	REFSCOPE   TraceFlag = 0x0000200
	REFVAR     TraceFlag = 0x0000400
	EACH       TraceFlag = 0x0000800
	EACHSCOPE  TraceFlag = 0x0001000
	EACHREFVAR TraceFlag = 0x0002000
	DEBUG      TraceFlag = 0x0004000
	ACT        TraceFlag = 0x0008000
	APPLY      TraceFlag = 0x0010000
	ARITHMETIC TraceFlag = 0x0020000
	RELATION   TraceFlag = 0x0040000
	ASSIGN     TraceFlag = 0x0080000
	INDEX      TraceFlag = 0x0100000
	LOOP       TraceFlag = 0x0200000
	LOAD       TraceFlag = 0x0400000
	DIAGRAM    TraceFlag = 0x0800000
	DIAGRAMT   TraceFlag = 0x1000000
	GRAMMAR    TraceFlag = 0x2000000
)

// Has reports whether any of the flags in g is set in f.
func (f TraceFlag) Has(g TraceFlag) bool { return f&g != 0 }

// Tracer writes the trace categories selected by Flags to its engine's
// output.
type Tracer struct {
	E     *Engine
	Flags TraceFlag
}

func NewTracer(e *Engine) *Tracer {
	return &Tracer{E: e}
}

// Tracing reports whether any of bits is being traced.
func (t *Tracer) Tracing(bits TraceFlag) bool {
	return t.Flags.Has(bits)
}

func (t *Tracer) MatchSymbols(l, r Element) {
	t.Trace(SYMBOLS, "--", l, r)
}

func (t *Tracer) Resolve(l, r Element, p priority) {
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

func (t *Tracer) Dumpx(sr *Stream, bits TraceFlag, s string, x Element) {
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
			if x, ok := x.(traceable); ok {
				x.Trace(sr, t)
			}
		} else {
			b.TraceRet(t)
		}
	}
}

func (t *Tracer) TraceFull(bits TraceFlag, s string, l, r Element, p priority) {
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
		pd := p.assoc()
		pv := p.level()
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

func (t *Tracer) Trace(bits TraceFlag, s string, l, r Element) {
	t.TraceFull(bits, s, l, r, t.E.lhsContext.Priority())
}

// Dumpit names the operator or statement x that is about to act, as the
// original's writefln("%s", x) did through toString.
func (t *Tracer) Dumpit(bits TraceFlag, s string, x Element) {
	if t.Flags&bits != 0 {
		t.E.printf("\t%s\t%s\n", s, x.ToString())
	}
}

func (t *Tracer) Dumpvar(bits TraceFlag, s string, p VarElement) {
	if t.Flags&bits != 0 {
		TxE(t.E.out, s, p)
	}
}

func (t *Tracer) Dumpvars(bits TraceFlag, s string, p, q VarElement) {
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
