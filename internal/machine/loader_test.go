package machine

import (
	"errors"
	"strings"
	"testing"
)

// Malformed bytecode is reported as an *Error with its position, never as a
// runtime panic.
func TestLoaderMalformed(t *testing.T) {
	t.Parallel()
	cases := []struct{ rules, want string }{
		{"m", "1:1: opcode m needs a value"},
		{"m:t L:0 n:1 ( ) ( m:eof ) r", "1:27: the left side of a rule is empty"},
		{"m:t L:0 n:1 ( c:a ) ( ) r", "1:25: the right side of a rule is empty"},
		{"m:t d:x n:1 ( c:a ) ( z ) r", "the priority of a rule is not a number"},
		{"m:t L:0 d:x ( c:a ) ( z ) r", "the offset of a rule is not a number"},
		{"m:t L:0 n:1 c:a ( z ) r", "the left side of a rule is not a list"},
		{"m:t L:x n:1 ( c:a ) ( z ) r", "1:5: bad priority `L:x`"},
		{"m:t L:0 n:x ( c:a ) ( z ) r", "bad number `n:x`"},
		{")", "1:1: operand stack underflow"},
		{"(", "1 operands left over"},
		{"m:t L:0 n:1 ( c:a ( z ) r", "the left side of a rule is not a list"},
		{"r", "a rule needs 5 operands, found 0"},
		{"m:t L:0 n:1 ( c:a ) ( z ) r r", "1:29: a rule needs 5 operands, found 0"},
		{"m:t L:0 n:1 ( c:%zz ) ( z ) r", "bad rule text"},
		{"m:t L:0 n:1 ( c:a ) ( z ) T", "unsupported opcode `T`"},
		{"m:t L:0 n:1 ( c:a ) ( z )\nQ", "2:1: bad load format `Q`"},
		{"m:t L:0 n:1 ( c:a ) ( z )", "5 operands left over"},
		{"# comment\n  m:t\n  L:0 n:1 ( c:a ) ( z ) r\n  m:t L:0 n:1 ( c:b", "unclosed"},
	}
	for _, c := range cases {
		t.Run(c.rules, func(t *testing.T) {
			t.Parallel()
			err := NewEngine().LoadFromString(c.rules)
			var lmErr *Error
			if !errors.As(err, &lmErr) || !strings.Contains(err.Error(), c.want) {
				t.Errorf("got %v, want an *Error containing %q", err, c.want)
			}
		})
	}
}

func FuzzLoad(f *testing.F) {
	f.Add(calcRules)
	f.Add(exprRules)
	for _, s := range []string{"m", "(", ")", "r", "m:t L:0 n:1 ( ) ( m:eof ) r", "c:%zz", "T", "n:1e400"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, rules string) {
		// any outcome but a panic is fine
		_ = NewEngine().LoadFromString(rules)
	})
}

// A load that fails leaves the engine as it was, and a later load works.
func TestLoadFailureKeepsEngineUsable(t *testing.T) {
	t.Parallel()
	good := "m:t L:0 n:1 ( z m:out ) ( m:eof ) r\nm:t L:0 n:1 ( c:a ) ( z c:A ) r\n"
	bad := "m:u L:0 n:1 ( z m:out ) ( m:eof ) r\nm:u L:0 n:1 ( c:a ( z c:B ) r\n"
	got := capture(t, good, func(e *Engine) {
		if err := e.LoadFromString(bad); err == nil {
			t.Fatal("bad rules loaded")
		}
		e.AppendInput(NewGramInputBuffer(e, "a"))
	})
	if got != "A" {
		t.Errorf("after a failed load: got %q, want %q", got, "A")
	}

	e := NewEngine()
	if err := e.LoadFromString(bad); err == nil {
		t.Fatal("bad rules loaded")
	}
	if err := e.LoadFromString(good); err != nil {
		t.Fatalf("load after a failed load: %v", err)
	}
	if n := len(e.loader.operands); n != 0 {
		t.Errorf("%d operands left on the loader after a failed load", n)
	}
}

// Rules added from a file that fails to load are taken out again.
func TestAddFailureKeepsOldRules(t *testing.T) {
	t.Parallel()
	good := "m:t L:0 n:1 ( z m:out ) ( m:eof ) r\nm:t L:0 n:1 ( c:a ) ( z c:A ) r\n"
	bad := "m:t L:0 n:1 ( c:a c:a ) ( z c:B ) r\nm:t L:0 n:1 ( c:b ) ( z c:C ) r\nm:u L:0 n:1 ( c:x ) ( z ) r\nm:t L:0 n:1 ( c:c\n"
	got := capture(t, good, func(e *Engine) {
		if err := e.LoadFromStringReset(bad, false); err == nil {
			t.Fatal("bad rules loaded")
		}
		if e.grammars.Get("u") != nil {
			t.Error("the grammar u of the failed load is still defined")
		}
		e.AppendInput(NewGramInputBuffer(e, "aab"))
	})
	if got != "AAb" {
		t.Errorf("after a failed -add: got %q, want %q", got, "AAb")
	}
}
