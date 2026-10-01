// Command lm runs Language Machine rules (.lm bytecode) on its input.
package main

import (
	"os"

	"github.com/msorc/languagemachine2/internal/application"
)

func main() {
	a := application.NewApplication(os.Args)
	// like the original: 1 when the analysis fails or flagError was used
	os.Exit(a.Start())
}
