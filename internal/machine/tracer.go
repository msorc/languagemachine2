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
	TraceMismatch     TraceFlag = 0x0000001
	TraceSymbols      TraceFlag = 0x0000002
	TraceContextScope TraceFlag = 0x0000004
	TraceCVar         TraceFlag = 0x0000008
	TraceLVar         TraceFlag = 0x0000010
	TraceRVar         TraceFlag = 0x0000020
	TraceRVarVar      TraceFlag = 0x0000040
	TraceRVarScope    TraceFlag = 0x0000080
	TraceRef          TraceFlag = 0x0000100
	TraceRefScope     TraceFlag = 0x0000200
	TraceRefVar       TraceFlag = 0x0000400
	TraceEach         TraceFlag = 0x0000800
	TraceEachScope    TraceFlag = 0x0001000
	TraceEachRefVar   TraceFlag = 0x0002000
	TraceDebug        TraceFlag = 0x0004000
	TraceAct          TraceFlag = 0x0008000
	TraceApply        TraceFlag = 0x0010000
	TraceArithmetic   TraceFlag = 0x0020000
	TraceRelation     TraceFlag = 0x0040000
	TraceAssign       TraceFlag = 0x0080000
	TraceIndex        TraceFlag = 0x0100000
	TraceLoop         TraceFlag = 0x0200000
	TraceLoad         TraceFlag = 0x0400000
	TraceDiagram      TraceFlag = 0x0800000
	TraceDiagramText  TraceFlag = 0x1000000
	TraceGrammar      TraceFlag = 0x2000000
)

// Has reports whether any of the flags in g is set in f.
func (f TraceFlag) Has(g TraceFlag) bool { return f&g != 0 }

// Tracer writes the trace categories selected by Flags to its engine's
// output.
type tracer struct {
	e     *Engine
	flags TraceFlag
}

func newTracer(e *Engine) *tracer {
	return &tracer{e: e}
}

// Tracing reports whether any of bits is being traced; a nil Tracer traces
// nothing.
func (t *tracer) tracing(bits TraceFlag) bool {
	return t.on(bits)
}

// on is Tracing for the methods below. Every method of Tracer may be called
// on a nil Tracer, which is the engine's tracer when nothing is traced.
func (t *tracer) on(bits TraceFlag) bool {
	return t != nil && t.flags&bits != 0
}

func (t *tracer) matchSymbols(l, r Element) {
	t.trace(TraceSymbols, "--", l, r)
}

func (t *tracer) resolve(l, r Element, p priority) {
	t.traceFull(TraceMismatch, "??", l, r, p)
}

func (t *tracer) back(l, r Element) {
	t.trace(TraceMismatch, "**", l, r)
}

func (t *tracer) bindCvar(l, r Element) {
	t.trace(TraceCVar, "cV", l, r)
}

func (t *tracer) bindLvar(l, r Element) {
	t.trace(TraceLVar, "lV", l, r)
}

func (t *tracer) traceAct(sr *Stream, x Element) {
	t.dumpX(sr, TraceAct, "ACT", x)
}

func (t *tracer) traceApply(sr *Stream, x Element) {
	t.dumpX(sr, TraceApply, "APPLY", x)
}

func (t *tracer) traceArithmetic(sr *Stream, x Element) {
	t.dumpX(sr, TraceArithmetic, "ARITHMETIC", x)
}

func (t *tracer) traceRelation(sr *Stream, x Element) {
	t.dumpX(sr, TraceRelation, "RELATION", x)
}

func (t *tracer) traceAssignment(sr *Stream, x Element) {
	t.dumpX(sr, TraceAssign, "ASSIGN", x)
}

func (t *tracer) traceIndex(sr *Stream, x Element) {
	t.dumpX(sr, TraceIndex, "INDEX", x)
}

func (t *tracer) traceLoop(sr *Stream, x Element) {
	t.dumpX(sr, TraceLoop, "LOOP", x)
}

func (t *tracer) dumpX(sr *Stream, bits TraceFlag, s string, x Element) {
	if t.on(bits) {
		t.dumpOp(bits, s, x)
		sr.dumpX()
	}
}

func (t *tracer) dumpGrammar(gr *grammar) {
	if t.on(TraceGrammar) {
		gr.dump(t.e.out)
	}
}

func (t *tracer) repeat(i int) {
	if t.on(TraceDiagram) {
		t.e.display.repeat(i, t.e.lhsContext.State().stateIndex, t.e.rhsStream.mode.ContextMode().State().stateIndex, t.e.lhsContext.NestingDepth(), t.e.rhsStream.mode.ContextMode().NestingDepth())
	}
}

func (t *tracer) ruleScope(s string, st *state, pp, pq varElement) {
	if t.on(TraceDiagram) {
		t.e.display.replace(s, st.stateIndex, t.e.rhsStream.mode.ContextMode().State().stateIndex, t.e.lhsContext.NestingDepth(), t.e.rhsStream.mode.ContextMode().NestingDepth())
	} else {
		t.dumpVars(TraceContextScope, "CXSCOPE", pp, pq)
	}
}

func (t *tracer) bindRvar(l, r Element) {
	t.trace(TraceRVar, "rV", l, r)
}

func (t *tracer) bindRvarScopeVars(lv, rv Element, pp, pq varElement) {
	t.trace(TraceRVarVar, "RVAR", lv, rv)
	t.dumpVars(TraceRVarScope, "RVARSCOPE", pp, pq)
}

func (t *tracer) theRefVars(pk Element, pp, pq varElement) {
	t.dumpVar(TraceRef, "REF", pp)
	t.dumpVars(TraceRefScope, "REFSCOPE", pp, pq)
}

func (t *tracer) theRefVar(pp varElement) {
	t.dumpVar(TraceRefVar, "REFVAR", pp)
}

func (t *tracer) eachRefVars(pk Element, pp, pq varElement) {
	t.dumpVar(TraceEach, "EACH", pp)
	t.dumpVars(TraceEachScope, "EACHSCOPE", pp, pq)
}

func (t *tracer) eachRefVar(pp varElement) {
	t.dumpVar(TraceEachRefVar, "EACHREF", pp)
}

func (t *tracer) traceShort(b GenMode) {
	if t == nil {
		return
	}
	sr := b.Stream()
	if sr.codeVector != nil {
		if sr.codeIndex < len(sr.codeVector) {
			x := sr.codeVector[sr.codeIndex]
			if t.on(TraceDebug) {
				b.trace(x)
			}
			if x, ok := x.(traceable); ok {
				x.trace(sr, t)
			}
		} else {
			b.traceRet(t)
		}
	}
}

func (t *tracer) traceFull(bits TraceFlag, s string, l, r Element, p priority) {
	if t.on(bits) {
		g := t.e.lhsContext.State().grammar.symbol
		gs := "---"
		if g != nil {
			gs = g.toTrace()
		}
		ls := "---"
		if l != nil {
			ls = l.toTrace()
		}
		rs := "---"
		if r != nil {
			rs = r.toTrace()
		}
		es := "---"
		if t.e.rsLastMatchElement != nil {
			es = t.e.rsLastMatchElement.toTrace()
		}
		pd := p.assoc()
		pv := p.level()
		ld := t.e.lhsContext.NestingDepth()
		rd := t.e.rhsStream.mode.ContextMode().NestingDepth()
		if t.on(TraceDiagram) {
			t.e.display.trace(s, t.e.lhsStream.mode.ContextMode().State().stateIndex, t.e.rhsStream.mode.ContextMode().State().stateIndex, ld, rd, ls, rs, es)
		} else {
			// the zero columns are the code indexes of compiled (C/D) rules
			// in the original, which interpreted rules leave at 0
			t.e.printf("\t%4d %4s %4d %4d %5d%s %4d %4d %4d %6d %8s%12s%12s%12s\n",
				t.e.lineNo(), s, ld, rd, pv, pd, 0, t.e.lhsStream.codeIndex, 0, t.e.rhsStream.codeIndex, gs, ls, rs, es)
		}
	}
}

func (t *tracer) trace(bits TraceFlag, s string, l, r Element) {
	if !t.on(bits) {
		return
	}
	t.traceFull(bits, s, l, r, t.e.lhsContext.Priority())
}

// dumpOp names the operator or statement x that is about to act, as the
// original's writefln("%s", x) did through toString.
func (t *tracer) dumpOp(bits TraceFlag, s string, x Element) {
	if t.on(bits) {
		t.e.printf("\t%s\t%s\n", s, x.ToString())
	}
}

func (t *tracer) dumpVar(bits TraceFlag, s string, p varElement) {
	if t.on(bits) {
		traceElement(t.e.out, s, p)
	}
}

func (t *tracer) dumpVars(bits TraceFlag, s string, p, q varElement) {
	if t.on(bits) {
		t.e.printf("VARIABLES: %s\n", s)
		for p != nil {
			marker := "-"
			if p == q {
				marker = "*"
			}
			traceVar(t.e.out, "VV", marker, p)
			p = p.link()
		}
		t.e.printf("---------\n")
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

// traceElement writes a trace line for x (its address and trace form) and returns x.
func traceElement(w io.Writer, s string, x Element) Element {
	var xtrace string
	if x != nil {
		xtrace = x.toTrace()
	} else {
		xtrace = "---"
	}
	_, _ = fmt.Fprintf(w, "\t%6s:     %8X %24s\n", s, addr(x), xtrace)
	return x
}

// traceVar writes a trace line for the variable w and its links, and returns w.
func traceVar(out io.Writer, r, s string, w varElement) Element {
	var n string

	if w != nil {
		n = w.toDump()
	} else {
		n = "---"
	}

	_, _ = fmt.Fprintf(out, "\t%6s:%4s %8X %24s %8X %8X %8X %8X\n", r, s, addr(w), n, addr(w.Value()), addr(w.Variables()), addr(w.scopeReferenceContext()), addr(w.link()))

	return w
}
