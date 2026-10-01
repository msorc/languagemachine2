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
	"strings"

	"github.com/msorc/languagemachine2/internal/lmgo"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	name := filepath.Base(args[0])
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	var cfg lmgo.Config
	out := fs.String("o", "", "write the Go file here (default standard output)")
	stubs := fs.String("stubs", "", "write stubs for the functions the rules call to this file, unless it exists")
	compiler := fs.String("compiler", "", "lmn compiler in .lm bytecode (default: built from the embedded sources)")
	emitLM := fs.String("emit-lm", "", "also write the compiled rules (.lm) to this file")
	fs.StringVar(&cfg.Package, "pkg", "main", "package name; main also writes func main")
	fs.StringVar(&cfg.Var, "var", "Program", "name of the Program variable")
	fs.StringVar(&cfg.Name, "name", "", "program name (default: the first file's base name)")
	fs.StringVar(&cfg.Prefix, "prefix", "lm", "prefix of the Go names of the functions the rules call")
	fs.StringVar(&cfg.Import, "import", lmgo.DefaultImport, "import path of the runtime package lm")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s [flags] file.lmn... | file.lm...\n", name)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return 2
	}
	if err := generate(&cfg, fs.Args(), *out, *stubs, *compiler, *emitLM, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}
	return 0
}

func generate(cfg *lmgo.Config, files []string, out, stubs, compiler, emitLM string, stdout, stderr io.Writer) error {
	rules, err := rules(files, compiler)
	if err != nil {
		return err
	}
	if emitLM != "" {
		if err := os.WriteFile(emitLM, []byte(rules), 0o644); err != nil {
			return err
		}
	}
	for _, f := range files {
		cfg.Sources = append(cfg.Sources, filepath.Base(f))
	}
	if cfg.Name == "" {
		cfg.Name = strings.TrimSuffix(cfg.Sources[0], filepath.Ext(cfg.Sources[0]))
	}
	res, err := lmgo.Generate(rules, *cfg)
	if err != nil {
		return err
	}
	if out == "" {
		_, err = stdout.Write(res.Code)
	} else {
		err = os.WriteFile(out, res.Code, 0o644)
	}
	if err != nil {
		return err
	}
	if res.Dynamic > 0 {
		fmt.Fprintf(stderr, "note: %d call(s) name their function through an expression; they are bound at run time\n", res.Dynamic)
	}
	if len(res.Funcs) == 0 {
		return nil
	}
	if stubs == "" {
		var ids []string
		for _, f := range res.Funcs {
			ids = append(ids, fmt.Sprintf("%s (%s)", f.Ident, f.Name))
		}
		fmt.Fprintf(stderr, "note: package %s must define func(*lm.Call) lm.Value: %s\n", cfg.Package, strings.Join(ids, ", "))
		return nil
	}
	if _, err := os.Stat(stubs); err == nil {
		return nil // never overwrite the user's implementations
	}
	code, err := lmgo.Stubs(res.Funcs, *cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(stubs, code, 0o644)
}

// rules returns the bytecode for files: .lm files are read and joined; any
// other files are lmn sources and are compiled together.
func rules(files []string, compiler string) (string, error) {
	nlm := 0
	for _, f := range files {
		if filepath.Ext(f) == ".lm" {
			nlm++
		}
	}
	switch nlm {
	case 0:
		if compiler != "" {
			b, err := os.ReadFile(compiler)
			if err != nil {
				return "", err
			}
			compiler = string(b)
		}
		return lmgo.Compile(compiler, files...)
	case len(files):
		var b strings.Builder
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return "", err
			}
			b.Write(data)
			if len(data) > 0 && data[len(data)-1] != '\n' {
				b.WriteByte('\n')
			}
		}
		return b.String(), nil
	default:
		return "", errors.New("give either .lm files or lmn sources, not both")
	}
}
