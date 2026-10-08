package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const rulesLM = `m:t L:0 n:1 ( c:! v:W G v:shout G f:args d:hey G f:fun w . ) ( m:eof v:W ) r
m:t L:0 n:1 ( z m:out ) ( m:eof ) r
`

func TestRun(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "t.lm2")
	if err := os.WriteFile(in, []byte(rulesLM), 0o644); err != nil {
		t.Fatal(err)
	}
	out, stubs := filepath.Join(dir, "t.go"), filepath.Join(dir, "funcs.go")
	var stdout, stderr strings.Builder
	if status := run([]string{"lm2n2go", "-o", out, "-stubs", stubs, in}, &stdout, &stderr); status != 0 {
		t.Fatalf("status %d: %s", status, stderr.String())
	}
	code, _ := os.ReadFile(out)
	if !strings.Contains(string(code), `Name:  "t",`) || !strings.Contains(string(code), `"shout": lmShout,`) {
		t.Errorf("generated code:\n%s", code)
	}
	if s, _ := os.ReadFile(stubs); !strings.Contains(string(s), "func lmShout(c *lm.Call) (lm.Value, error)") {
		t.Errorf("stubs:\n%s", s)
	}

	// existing stubs are kept
	if err := os.WriteFile(stubs, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	if status := run([]string{"lm2n2go", "-o", out, "-stubs", stubs, in}, &stdout, &stderr); status != 0 {
		t.Fatalf("status %d: %s", status, stderr.String())
	}
	if s, _ := os.ReadFile(stubs); string(s) != "mine" {
		t.Errorf("stubs overwritten: %q", s)
	}

	stderr.Reset()
	if status := run([]string{"lm2n2go", in, "x.lm2n"}, &stdout, &stderr); status != 1 || !strings.Contains(stderr.String(), "not both") {
		t.Errorf("mixed inputs: status %d, %q", status, stderr.String())
	}
	if status := run([]string{"lm2n2go"}, &stdout, &stderr); status != 2 {
		t.Errorf("no inputs: status %d", status)
	}
}
