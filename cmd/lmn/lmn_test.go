package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/msorc/languagemachine2/internal/lmgo"
)

var sources = []string{
	filepath.Join("..", "..", "internal", "lmnsrc", "lmn2xfe.lmn"),
	filepath.Join("..", "..", "internal", "lmnsrc", "lmn2mbe.lmn"),
}

// lmn.go must be what go generate writes now: the compiler built from the
// current sources.
func TestUpToDate(t *testing.T) {
	t.Parallel()
	rules, err := lmgo.Compile("", sources...)
	if err != nil {
		t.Fatal(err)
	}
	res, err := lmgo.Generate(rules, lmgo.Config{Name: "lmn", Sources: []string{"lmn2xfe.lmn", "lmn2mbe.lmn"}})
	if err != nil {
		t.Fatal(err)
	}
	have, err := os.ReadFile("lmn.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(have) != string(res.Code) {
		t.Error("lmn.go is out of date with the compiler sources: run go generate ./cmd/lmn")
	}
}

// The built-in compiler compiles its own sources to itself.
func TestFixpoint(t *testing.T) {
	t.Parallel()
	var out, errOut strings.Builder
	if status := Program.Run(append([]string{"lmn"}, sources...), nil, &out, &errOut); status != 0 {
		t.Fatalf("status %d: %s", status, errOut.String())
	}
	want, err := lmgo.Compiler()
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != want {
		t.Errorf("output differs from the compiler (%d vs %d bytes)", out.Len(), len(want))
	}
}
