package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/msorc/languagemachine2/internal/lm2go"
)

var sources = []string{
	filepath.Join("..", "..", "internal", "lm2nsrc", "lm2n2xfe.lm2n"),
	filepath.Join("..", "..", "internal", "lm2nsrc", "lm2n2mbe.lm2n"),
}

// lm2n.go must be what go generate writes now: the compiler built from the
// current sources, with the arguments of the //go:generate line.
func TestUpToDate(t *testing.T) {
	t.Parallel()
	if err := lm2go.UpToDate("generate.go", "lm2n.go"); err != nil {
		t.Error(err)
	}
}

// The built-in compiler compiles its own sources to itself.
func TestFixpoint(t *testing.T) {
	t.Parallel()
	var out, errOut strings.Builder
	if status := Program.Run(append([]string{"lmn"}, sources...), nil, &out, &errOut); status != 0 {
		t.Fatalf("status %d: %s", status, errOut.String())
	}
	want, err := lm2go.Compiler()
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != want {
		t.Errorf("output differs from the compiler (%d vs %d bytes)", out.Len(), len(want))
	}
}
