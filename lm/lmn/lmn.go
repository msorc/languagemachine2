// Package lmn compiles lmn, the notation of Language Machine rules, to the
// .lm bytecode that lm.Program runs. It uses the lmn compiler built into the
// module, which is itself a ruleset (see docs/lm/07-compilation-and-bytecode.md).
package lmn

import (
	"fmt"

	"github.com/msorc/languagemachine2/internal/lmgo"
)

// Compile compiles the lmn source src. name names the source in errors.
func Compile(name, src string) (string, error) {
	lm, err := lmgo.CompileInputs("", lmgo.Input{Text: src})
	if err != nil {
		return "", fmt.Errorf("compiling %s: %w", name, err)
	}
	return lm, nil
}

// CompileFiles compiles lmn source files, read in order as one input, as
// the lmn command does.
func CompileFiles(files ...string) (string, error) {
	return lmgo.Compile("", files...)
}
