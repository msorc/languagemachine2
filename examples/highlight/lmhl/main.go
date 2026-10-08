// Command lmhl highlights source text with Language Machine rules
// (../README.md).
//
//	lmhl [flags] [file...]   highlight the files, or standard input
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/msorc/languagemachine2/examples/highlight"
	"github.com/msorc/languagemachine2/lm/lmn"
)

func main() {
	os.Exit(run(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	name := filepath.Base(args[0])
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	lang := fs.String("lang", "", "language (default: from the file name): "+strings.Join(names(), ", "))
	rules := fs.String("rules", "", "highlighting rules to use instead of a built-in language (.lmn or .lm)")
	format := fs.String("format", "ansi", "output format: ansi, html or json")
	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s [flags] [file...]\n", name)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *format != "ansi" && *format != "html" && *format != "json" {
		fmt.Fprintf(stderr, "%s: unknown format %q\n", name, *format)
		return 2
	}
	var fixed *highlight.Highlighter
	switch {
	case *rules != "":
		var err error
		if fixed, err = load(*rules); err != nil {
			fmt.Fprintf(stderr, "%s: %v\n", name, err)
			return 1
		}
	case *lang != "":
		var ok bool
		if fixed, ok = highlight.ByName(*lang); !ok {
			fmt.Fprintf(stderr, "%s: unknown language %q\n", name, *lang)
			return 2
		}
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}
	status := 0
	for _, file := range files {
		if err := highlightFile(file, fixed, *format, stdin, stdout); err != nil {
			fmt.Fprintf(stderr, "%s: %s: %v\n", name, file, err)
			status = 1
		}
	}
	return status
}

func names() []string {
	langs := highlight.Languages()
	names := make([]string, 0, len(langs))
	for _, l := range langs {
		names = append(names, l.Name)
	}
	return names
}

// load reads highlighting rules, compiling them if they are lmn source.
func load(file string) (*highlight.Highlighter, error) {
	name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	if filepath.Ext(file) == ".lmn" {
		rules, err := lmn.CompileFiles(file)
		if err != nil {
			return nil, err
		}
		return highlight.New(name, rules), nil
	}
	rules, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return highlight.New(name, string(rules)), nil
}

// highlightFile writes one file, or standard input for "-". A file that
// cannot be highlighted is still written in the ansi format, without
// colours, and the error is returned.
func highlightFile(file string, h *highlight.Highlighter, format string, stdin io.Reader, stdout io.Writer) error {
	var text []byte
	var err error
	if file == "-" {
		text, err = io.ReadAll(stdin)
	} else {
		text, err = os.ReadFile(file)
	}
	if err != nil {
		return err
	}
	src := string(text)
	if h == nil {
		var ok bool
		if h, ok = highlight.ByFilename(file); !ok {
			err = errors.New("no language for this file name: use -lang or -rules")
		}
	}
	var spans []highlight.Span
	if h != nil {
		spans, err = h.Spans(src)
	}
	if err != nil && format != "ansi" {
		return err
	}
	var werr error
	switch format {
	case "ansi":
		werr = highlight.WriteANSI(stdout, src, spans)
	case "html":
		werr = highlight.WriteHTML(stdout, src, spans)
	case "json":
		werr = highlight.WriteJSON(stdout, spans)
	}
	return errors.Join(err, werr)
}
