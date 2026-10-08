package lmgo

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Options are the lmn2go command line: the generator's Config, the input
// files and where the results go.
type Options struct {
	Config
	Files    []string // .lmn sources, or .lm bytecode
	Out      string   // the Go file, or "" for standard output
	Stubs    string   // the file of stubs to write, unless it exists
	Compiler string   // a .lm file of the lmn compiler, or "" for the built-in one
	EmitLM   string   // where to write the compiled rules too, or ""
}

// ErrUsage is the error of ParseArgs when no input files are given; the
// usage has been printed.
var ErrUsage = errors.New("usage")

// ParseArgs parses the lmn2go arguments args (without the command name),
// writing usage and flag errors to stderr. It returns flag.ErrHelp for -h.
func ParseArgs(name string, args []string, stderr io.Writer) (*Options, error) {
	o := &Options{}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&o.Out, "o", "", "write the Go file here (default standard output)")
	fs.StringVar(&o.Stubs, "stubs", "", "write stubs for the functions the rules call to this file, unless it exists")
	fs.StringVar(&o.Compiler, "compiler", "", "lmn compiler in .lm bytecode (default: built from the embedded sources)")
	fs.StringVar(&o.EmitLM, "emit-lm", "", "also write the compiled rules (.lm) to this file")
	fs.StringVar(&o.Package, "pkg", "main", "package name; main also writes func main")
	fs.StringVar(&o.Var, "var", "Program", "name of the Program variable")
	fs.StringVar(&o.Name, "name", "", "program name (default: the first file's base name)")
	fs.StringVar(&o.Prefix, "prefix", "lm", "prefix of the Go names of the functions the rules call")
	fs.StringVar(&o.Import, "import", DefaultImport, "import path of the runtime package lm")
	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "usage: %s [flags] file.lmn... | file.lm...\n", name)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return nil, ErrUsage
	}
	o.Files = fs.Args()
	return o, nil
}

// DirectiveArgs finds the //go:generate line in src that runs lmn2go with
// -o out, and returns its lmn2go arguments, so that a test can check that
// the generated file is up to date with exactly the arguments go generate
// uses. The arguments must not be quoted.
func DirectiveArgs(src, out string) ([]string, error) {
	for line := range strings.Lines(src) {
		f := strings.Fields(line)
		if len(f) == 0 || f[0] != "//go:generate" {
			continue
		}
		for i, w := range f {
			if filepath.Base(w) == "lmn2go" || w == "lmn2go" {
				args := f[i+1:]
				for j := 0; j+1 < len(args); j++ {
					if args[j] == "-o" && args[j+1] == out {
						return args, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("no //go:generate lmn2go line writes %s", out)
}

// Generate compiles the files (or reads them, if they are .lm), writes the
// Go file and the stubs, and reports notes on stderr.
func (o *Options) Generate(stdout, stderr io.Writer) error {
	rules, err := o.rules()
	if err != nil {
		return err
	}
	if o.EmitLM != "" {
		if err := os.WriteFile(o.EmitLM, []byte(rules), 0o644); err != nil {
			return err
		}
	}
	cfg := o.Config
	for _, f := range o.Files {
		cfg.Sources = append(cfg.Sources, filepath.Base(f))
	}
	if cfg.Name == "" {
		cfg.Name = strings.TrimSuffix(cfg.Sources[0], filepath.Ext(cfg.Sources[0]))
	}
	res, err := Generate(rules, cfg)
	if err != nil {
		return err
	}
	if o.Out == "" {
		_, err = stdout.Write(res.Code)
	} else {
		err = os.WriteFile(o.Out, res.Code, 0o644)
	}
	if err != nil {
		return err
	}
	if res.Dynamic > 0 {
		_, _ = fmt.Fprintf(stderr, "note: %d call(s) name their function through an expression; they are bound at run time\n", res.Dynamic)
	}
	if len(res.Funcs) == 0 {
		return nil
	}
	if o.Stubs == "" {
		var ids []string
		for _, f := range res.Funcs {
			ids = append(ids, fmt.Sprintf("%s (%s)", f.Ident, f.Name))
		}
		_, _ = fmt.Fprintf(stderr, "note: package %s must define func(*lm.Call) (lm.Value, error): %s\n", cfg.Package, strings.Join(ids, ", "))
		return nil
	}
	if _, err := os.Stat(o.Stubs); err == nil {
		return nil // never overwrite the user's implementations
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	code, err := Stubs(res.Funcs, cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(o.Stubs, code, 0o644)
}

// rules returns the bytecode for the files: .lm files are read and joined;
// any other files are lmn sources and are compiled together.
func (o *Options) rules() (string, error) {
	nlm := 0
	for _, f := range o.Files {
		if filepath.Ext(f) == ".lm" {
			nlm++
		}
	}
	switch nlm {
	case 0:
		compiler := ""
		if o.Compiler != "" {
			b, err := os.ReadFile(o.Compiler)
			if err != nil {
				return "", err
			}
			compiler = string(b)
		}
		return Compile(compiler, o.Files...)
	case len(o.Files):
		var b strings.Builder
		for _, f := range o.Files {
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

// UpToDate reports an error unless the Go file out is what the
// //go:generate lmn2go line in the Go file src writes now. Paths are
// relative to the current directory, as they are for go generate.
func UpToDate(src, out string) error {
	text, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	args, err := DirectiveArgs(string(text), out)
	if err != nil {
		return err
	}
	o, err := ParseArgs("lmn2go", args, io.Discard)
	if err != nil {
		return fmt.Errorf("%s: the lmn2go arguments: %w", src, err)
	}
	o.Out, o.Stubs, o.EmitLM = "", "", ""
	var code strings.Builder
	if err := o.Generate(&code, io.Discard); err != nil {
		return err
	}
	have, err := os.ReadFile(out)
	if err != nil {
		return err
	}
	if string(have) != code.String() {
		return fmt.Errorf("%s is out of date: run go generate", out)
	}
	return nil
}
