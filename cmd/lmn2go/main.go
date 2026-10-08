// Command lmn2go compiles Language Machine rules to a Go program or package,
// as lmn2d compiled them to D (docs/lmn2go.md).
//
//	lmn2go [flags] file.lmn...   compile lmn sources
//	lmn2go [flags] file.lm...    wrap rules that are already compiled
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/msorc/languagemachine2/internal/lmgo"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	name := filepath.Base(args[0])
	o, err := lmgo.ParseArgs(name, args[1:], stderr)
	switch {
	case errors.Is(err, flag.ErrHelp):
		return 0
	case err != nil:
		return 2
	}
	if err := o.Generate(stdout, stderr); err != nil {
		_, _ = fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}
	return 0
}
