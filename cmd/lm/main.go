package main

import (
	"languagemachine2/internal/application"
	"os"
)

func main() {
	a := application.NewApplication(os.Args)
	a.Start()
}
