package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lmhl(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()
	var out, errOut strings.Builder
	status := run(append([]string{"lmhl"}, args...), strings.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), status
}

func TestStdin(t *testing.T) {
	t.Parallel()
	out, errOut, status := lmhl(t, "if x", "-lang", "go", "-format", "json")
	if want := `[{"start":0,"end":2,"class":"keyword"}]` + "\n"; out != want || status != 0 {
		t.Errorf("got %q, status %d, stderr %q", out, status, errOut)
	}
}

// A file is highlighted by its extension, and with rules given as lmn source.
func TestFiles(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "x.go")
	if err := os.WriteFile(file, []byte("return 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := "\x1b[35mreturn\x1b[0m \x1b[31m1\x1b[0m\n"
	for _, args := range [][]string{{file}, {"-rules", filepath.Join("..", "go.lmn"), file}} {
		if out, errOut, status := lmhl(t, "", args...); out != want || status != 0 {
			t.Errorf("%v: got %q, status %d, stderr %q", args, out, status, errOut)
		}
	}
}

// Text that cannot be highlighted is written as it is, with an error.
func TestNoLanguage(t *testing.T) {
	t.Parallel()
	out, errOut, status := lmhl(t, "if x")
	if out != "if x" || status != 1 || !strings.Contains(errOut, "no language") {
		t.Errorf("got %q, status %d, stderr %q", out, status, errOut)
	}
	if out, _, status := lmhl(t, "if x", "-format", "json"); out != "" || status != 1 {
		t.Errorf("json: got %q, status %d", out, status)
	}
	if _, _, status := lmhl(t, "", "-lang", "nope"); status != 2 {
		t.Errorf("unknown language: status %d", status)
	}
}
