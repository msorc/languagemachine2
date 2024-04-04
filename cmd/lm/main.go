package main

import (
	"os"
	"languagemachine2/internal/application"
)

func main() {
	a := application.NewApplication(os.Args, "")
	a.Start()
}
