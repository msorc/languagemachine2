package machine

import (
	"slices"
	"testing"
)

func TestCalls(t *testing.T) {
	t.Parallel()
	// A(1), g(h(x)) and f() (URI-encoded) call by name; var B = [a] is an
	// array literal, not a call; T[0]() calls a table entry and V() a
	// variable's value
	rules := `#!/usr/bin/lm2 -r
m:t L:0 n:1 ( z v:A G f:args n:1 G f:fun . v:g G f:args v:h G f:args v:x G f:fun f:fun . ) ( m:eof ) r
m:t L:0 n:1 ( z v:B G f:args v:a G f:array w . v:T V n:0 G f:idx f:args f:fun . v:V V f:args f:fun . ) ( m:eof ) r
m:t L:0 n:1 ( z v:%66 G f:args f:fun . ) ( m:eof ) r
`
	names, dynamic, err := Calls(rules)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"A", "f", "g", "h"}; !slices.Equal(names, want) {
		t.Errorf("names = %q, want %q", names, want)
	}
	if dynamic != 2 {
		t.Errorf("dynamic = %d, want 2", dynamic)
	}
	if _, _, err := Calls("v:f G f:fun"); err == nil {
		t.Error("f:fun without f:args: no error")
	}
}
