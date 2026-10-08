// Command lm2 runs Language Machine rules (.lm2 bytecode) on its input.
package main

import (
	"os"

	"github.com/msorc/languagemachine2/internal/application"
)

func main() {
	a := application.New(os.Args, nil, os.Stdin, os.Stdout, os.Stderr)
	// like the original: 1 when the analysis fails or flagError was used
	os.Exit(a.Start())
}
