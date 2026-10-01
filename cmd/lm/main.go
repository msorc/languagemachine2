package main

import (
	"github.com/msorc/languagemachine2/internal/application"
	"os"
)

func main() {
	a := application.NewApplication(os.Args)
	// like the original: 1 when the analysis fails or flagError was used
	os.Exit(a.Start())
}
