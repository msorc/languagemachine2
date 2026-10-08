// Package lm2n compiles lm2n, the notation of Language Machine rules, to the
// .lm2 bytecode that lm2.Program runs. It uses the lm2n compiler built into the
// module, which is itself a ruleset (see docs/lm2/07-compilation-and-bytecode.md).
package lm2n

import (
	"fmt"

	"github.com/msorc/languagemachine2/internal/lm2go"
)

// Compile compiles the lm2n source src. name names the source in errors.
func Compile(name, src string) (string, error) {
	lm, err := lm2go.CompileInputs("", lm2go.Input{Text: src})
	if err != nil {
		return "", fmt.Errorf("compiling %s: %w", name, err)
	}
	return lm, nil
}

// CompileFiles compiles lmn source files, read in order as one input, as
// the lmn command does.
func CompileFiles(files ...string) (string, error) {
	return lm2go.Compile("", files...)
}
