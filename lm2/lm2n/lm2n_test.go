package lm2n_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/msorc/languagemachine2/lm2"
	"github.com/msorc/languagemachine2/lm2/lm2n"
)

const shout = `  - out <- eof - ;
  'a' <- eof - "A" ;
`

func TestCompile(t *testing.T) {
	t.Parallel()
	rules, err := lm2n.Compile("shout", shout)
	if err != nil {
		t.Fatal(err)
	}
	got, err := (&lm2.Program{Name: "shout", Rules: rules}).Translate("bab")
	if err != nil {
		t.Fatal(err)
	}
	if got != "bAb" {
		t.Errorf("got %q, want %q", got, "bAb")
	}
}

func TestCompileFiles(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "shout.lm2n")
	if err := os.WriteFile(file, []byte(shout), 0o644); err != nil {
		t.Fatal(err)
	}
	fromFile, err := lm2n.CompileFiles(file)
	if err != nil {
		t.Fatal(err)
	}
	fromText, err := lm2n.Compile("shout", shout)
	if err != nil {
		t.Fatal(err)
	}
	if fromFile != fromText {
		t.Error("a file and its text compile differently")
	}
	if _, err := lm2n.CompileFiles(filepath.Join(t.TempDir(), "missing.lm2n")); err == nil || !strings.Contains(err.Error(), "missing.lm2n") {
		t.Errorf("missing file: %v", err)
	}
}
