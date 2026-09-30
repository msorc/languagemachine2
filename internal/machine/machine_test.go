package machine

import (
	"fmt"
	"strings"
	"testing"
)

// run loads rules, feeds input and returns what the grammar wrote to stdout.
func run(t *testing.T, rules, input string) string {
	t.Helper()
	return capture(t, rules, func(e *Engine) {
		e.AppendInput(NewGramInputBuffer(e, input))
	})
}

// capture loads rules, lets feed queue the inputs and returns what the grammar
// wrote to its output.
func capture(t *testing.T, rules string, feed func(e *Engine)) string {
	t.Helper()
	var b strings.Builder
	e := NewEngine()
	e.SetOutput(&b)
	if err := e.LoadFromString(rules); err != nil {
		t.Fatal(err)
	}
	feed(e)
	if _, err := e.Start(); err != nil {
		t.Error(err)
	}
	return b.String()
}

const calcRules = `
m:calc L:0 n:1 ( z m:error m:output ) ( m:eof ) r
m:calc L:0 n:1 ( z m:result m:output ) ( m:eof ) r
m:calc L:0 n:0 ( z m:x v:N p ) ( m:result c:result:%20 v:N c:%5Cn m:eom ) r
m:calc L:0 n:0 ( c:+ m:x v:A p m:x v:B p ) ( m:x z v:A V v:B V f:+ b ) r
m:calc L:0 n:0 ( c:/ m:x v:A p m:x v:B p ) ( m:x z v:A V v:B V f:/ b ) r
m:calc L:20 n:0 ( l:%5B0-9%5D t ( m:repeat l:%5B0-9%5D t ) m:toNum v:N p ) ( m:x z v:N p ) r
m:calc L:20 n:1 ( l:%5B%20%5Ct%5Cn%5D ) ( z ) r
m:calc R:30 n:1 ( z m:anything ) ( m:line ) r
m:calc R:30 n:0 ( m:eof ) ( m:line m:eof ) r
m:calc R:30 n:0 ( c:%5Cn ) ( m:line ) r
m:calc R:30 n:0 ( z m:line ) ( m:error c:---%20not%20understood%5Cn m:eom ) r
m:calc R:30 n:0 ( z m:eom ) ( m:output ) r
m:calc R:30 n:1 ( z m:out ) ( m:eom ) r
`

func TestCalc(t *testing.T) {
	t.Parallel()
	got := run(t, calcRules, "+ 2 + 30 4\nz\n/ 100 3\n")
	want := "result: 36\n--- not understood\nresult: 33.3333\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Shorter rules added after a longer one in the same group must not be lost.
func TestRuleOrderWithinGroup(t *testing.T) {
	t.Parallel()
	rules := `
m:t L:0 n:1 ( z m:w m:out ) ( m:eof ) r
m:t L:0 n:0 ( c:a c:b c:c ) ( m:w c:X ) r
m:t L:0 n:0 ( c:a c:b ) ( m:w c:Y ) r
m:t L:0 n:0 ( c:a ) ( m:w c:Z ) r
`
	for in, want := range map[string]string{"abc": "X", "ab": "Y", "a": "Z"} {
		if got := run(t, rules, in); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

// Material grabbed by a failed repeat iteration is given back.
func TestRepeatReleasesFailedIteration(t *testing.T) {
	t.Parallel()
	rules := `
m:t L:0 n:1 ( z m:out ) ( m:eof ) r
m:t L:0 n:1 ( c:a t ( m:repeat c:b t c:c t ) m:toUstr v:S p ) ( z c:%5B v:S c:%5D ) r
`
	if got, want := run(t, rules, "abcbd"), "[ABC]bd"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpressions(t *testing.T) {
	t.Parallel()
	rules := `
m:t L:0 n:1 ( z m:out ) ( m:eof ) r
m:t L:0 n:1 ( c:a ) ( z v:format G f:args d:x=%25d%20y=%25s%20z=%25g G n:3 G d:q G n:2.5 G f:fun f:apply ) r
m:t L:0 n:1 ( c:b ) ( z n:0 G ( c:Y ) G ( c:N ) G f:sel ) r
m:t L:0 n:1 ( c:c ) ( z n:1 G ( c:Y ) G ( c:N ) G f:sel ) r
m:t L:0 n:1 ( c:d ) ( z v:hex G f:args d:0xff G f:fun f:apply c:%20 v:num G f:args d:3.5 G f:fun f:apply ) r
m:t L:0 n:1 ( c:e ) ( z v:hex G f:args d:ff G f:fun f:apply c:%20 v:hex G f:args d:0x10.8 G f:fun f:apply ) r
`
	// hex is C's strtod, as in the original: "ff" without 0x is 0
	for in, want := range map[string]string{"a": "x=3 y=q z=2.5", "b": "N", "c": "Y", "d": "255 3.5", "e": "0 16.5"} {
		if got := run(t, rules, in); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

// A maximal-priority rule starts, but nothing can nest inside it.
func TestMaximalPriority(t *testing.T) {
	t.Parallel()
	rules := `
m:t L:0 n:1 ( z m:out ) ( m:eof ) r
m:t %s n:1 ( c:a m:b ) ( z c:X ) r
m:t L:0 n:0 ( c:c ) ( m:b ) r
`
	for pri, want := range map[string]string{"M:0": "ac", "L:0": "X"} {
		if got := run(t, fmt.Sprintf(rules, pri), "ac"); got != want {
			t.Errorf("%s: got %q, want %q", pri, got, want)
		}
	}
}

// As in the original, -lexpri is not applied when the goal is a terminal: a
// prioritised rule (here whitespace deletion) can start inside a token.
func TestLexicalPriority(t *testing.T) {
	t.Parallel()
	rules := `
m:t L:0 n:1 ( z m:out ) ( m:eof ) r
m:t L:0 n:1 ( c:abc ) ( z c:X ) r
m:t L:5 n:1 ( c:%20 ) ( z ) r
`
	for in, want := range map[string]string{"abc abc": "XX", "a bc": "X"} {
		if got := run(t, rules, in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

func TestLexClass(t *testing.T) {
	t.Parallel()
	e := NewEngine()
	e.Load()
	l := NewLexFromEngine("[a-c\\n]", e)
	for _, c := range "abc\n" {
		if l.Table[e.terminalSymbols.UniqueR(c)] == nil {
			t.Errorf("%q missing from class", c)
		}
	}
	if len(l.Table) != 4 {
		t.Errorf("class has %d members, want 4", len(l.Table))
	}
	if x := NewLexFromEngine("[^0-9]", e); x.Inclusive || len(x.Table) != 10 {
		t.Errorf("negated class: inclusive=%v size=%d", x.Inclusive, len(x.Table))
	}
}

func TestLoaderEach(t *testing.T) {
	t.Parallel()
	e := NewEngine()
	if err := e.LoadFromString("m:t L:0 n:0 ( z m:y ) ( m:x ( v:X e ) p ) r\n"); err != nil {
		t.Fatal(err)
	}
	r := e.initGrammar.Get(e.predefinedSymbols.nil, e.nonTerminalSymbols.GetByString("x"))
	if r == nil {
		t.Fatal("rule not defined")
	}
	x := r.rhs[1].(*GetXF).V.(*Str).V[0] // ( v:X e ) p
	if _, ok := x.(*EachRef); !ok {
		t.Errorf("e did not build an EachRef: %T", x)
	}
}

// Loading with reset replaces earlier rules, including their initial grammar.
func TestLoadResetReplacesInitialGrammar(t *testing.T) {
	t.Parallel()
	base := "m:t L:0 n:1 ( z m:out ) ( m:eof ) r\nm:t L:0 n:1 ( c:a ) ( z c:A ) r\n"
	add := "m:u L:0 n:1 ( c:b ) ( z c:B ) r\n"
	got := capture(t, add, func(e *Engine) {
		if err := e.LoadFromString(base); err != nil {
			t.Fatal(err)
		}
		e.AppendInput(NewGramInputBuffer(e, "a"))
	})
	if got != "A" {
		t.Errorf("got %q, want %q", got, "A")
	}
}

// Failures caused by the rules or the input are returned as errors.
func TestErrors(t *testing.T) {
	t.Parallel()
	e := NewEngine()
	if err := e.LoadFromString("m:t L:0 n:1 ( c:a ) ( z c:A ) r\nm:t L:x n:0 ( c:b ) ( z ) r\n"); err == nil {
		t.Error("bad priority: no load error")
	}
	if err := e.LoadFromString("m:t L:0 n:1 ( c:%zz ) ( z ) r\n"); err == nil {
		t.Error("bad escape: no load error")
	}

	nest := "m:t L:0 n:0 ( z m:nest ) ( m:eof ) r\nm:t L:0 n:0 ( z m:nest ) ( m:nest ) r\n"
	e = NewEngine()
	if err := e.LoadFromString(nest); err != nil {
		t.Fatal(err)
	}
	e.SetMaxDepth(50)
	e.AppendInput(NewGramInputBuffer(e, "x"))
	if status, err := e.Start(); status != 1 || err == nil || !strings.Contains(err.Error(), "maximum depth") {
		t.Errorf("max depth: status %d, err %v", status, err)
	}

	if _, err := NewGramInputFile(e, "does-not-exist"); err == nil {
		t.Error("missing input file: no error")
	}
	include := "m:t L:0 n:1 ( z m:out ) ( m:eof ) r\nm:t L:0 n:1 ( c:i ) ( z v:include G f:args d:does-not-exist G f:fun ) r\n"
	e = NewEngine()
	if err := e.LoadFromString(include); err != nil {
		t.Fatal(err)
	}
	e.AppendInput(NewGramInputBuffer(e, "i"))
	if _, err := e.Start(); err == nil || !strings.Contains(err.Error(), "include") {
		t.Errorf("include of a missing file: err %v", err)
	}
}

// -trace G lists the rules in definition order, the same on every run.
func TestGrammarDumpOrder(t *testing.T) {
	t.Parallel()
	var first string
	for i := 0; i < 5; i++ {
		e := NewEngine()
		if err := e.LoadFromString(calcRules); err != nil {
			t.Fatal(err)
		}
		var b strings.Builder
		e.initGrammar.Dump(&b)
		if i == 0 {
			first = b.String()
			if !strings.HasPrefix(first, "line    0:") {
				t.Errorf("dump does not start with rule 0:\n%s", first)
			}
		} else if b.String() != first {
			t.Fatalf("dump differs between runs:\n%s\n---\n%s", first, b.String())
		}
	}
}
