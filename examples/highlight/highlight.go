// Package highlight does syntax highlighting with Language Machine rules. A
// highlighting ruleset copies its input to its output and calls
// hl(class, text) for each piece that has a class; the package turns that
// output into spans and renders them (README.md). The rulesets for
// the built-in languages are go.lmn and lmn.lmn.
package highlight

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/msorc/languagemachine2/lm"
)

//go:generate go run ../../cmd/lmn2go -pkg highlight -var goProgram -name go -o go_lm.go go.lmn
//go:generate go run ../../cmd/lmn2go -pkg highlight -var lmnProgram -name lmn -o lmn_lm.go lmn.lmn

// Span is a piece of the source that has a class. Start and End are byte
// offsets, and the spans of a source are in order and do not overlap.
type Span struct {
	Start int    `json:"start"`
	End   int    `json:"end"`
	Class string `json:"class"`
}

// Highlighter is a highlighting ruleset. It may be used from several
// goroutines at once.
type Highlighter struct {
	prog *lm.Program
}

// Language is a built-in highlighter.
type Language struct {
	Name string
	Exts []string // file name extensions, with the dot
	*Highlighter
}

var languages = []Language{
	{"go", []string{".go"}, &Highlighter{goProgram}},
	{"lmn", []string{".lmn"}, &Highlighter{lmnProgram}},
}

// Languages returns the built-in highlighters.
func Languages() []Language { return languages }

// ByName returns the built-in highlighter for a language.
func ByName(name string) (*Highlighter, bool) {
	for _, l := range languages {
		if l.Name == name {
			return l.Highlighter, true
		}
	}
	return nil, false
}

// ByFilename returns the built-in highlighter for a file, by its extension.
func ByFilename(file string) (*Highlighter, bool) {
	for _, l := range languages {
		for _, ext := range l.Exts {
			if strings.HasSuffix(file, ext) {
				return l.Highlighter, true
			}
		}
	}
	return nil, false
}

// New returns a highlighter for rules in .lm bytecode.
func New(name, rules string) *Highlighter {
	return &Highlighter{&lm.Program{Name: name, Rules: rules, Funcs: map[string]lm.Func{"hl": lmHl}}}
}

// The markers that hl puts around a piece. They are bytes that valid UTF-8
// never holds, so they cannot be confused with the source.
const (
	markOpen  = "\xff" // before the class
	markText  = "\xfe" // between the class and the text
	markClose = "\xfd" // after the text
)

// lmHl implements hl(class, text) for the rules.
func lmHl(c *lm.Call) (lm.Value, error) {
	return lm.Sym(markOpen + c.Arg(0).String() + markText + c.Arg(1).String() + markClose), nil
}

// Spans highlights src. It is an error if the rules fail, or if their output
// is not src: the rules must copy everything, in order.
func (h *Highlighter) Spans(src string) ([]Span, error) {
	if src == "" {
		return nil, nil
	}
	if !utf8.ValidString(src) {
		return nil, errors.New("highlight: the input is not valid UTF-8")
	}
	out, err := h.prog.Translate(src)
	if err != nil {
		return nil, fmt.Errorf("highlight: %w", err)
	}
	return parse(out, src)
}

// parse reads the spans from the marked output of the rules.
func parse(out, src string) ([]Span, error) {
	var spans []Span
	pos := 0 // offset in src of what out has copied so far
	copied := func(text string) error {
		if !strings.HasPrefix(src[pos:], text) {
			return fmt.Errorf("highlight: the rules changed the text at offset %d", pos+diff(src[pos:], text))
		}
		pos += len(text)
		return nil
	}
	for out != "" {
		plain, rest, marked := strings.Cut(out, markOpen)
		if err := copied(plain); err != nil {
			return nil, err
		}
		if !marked {
			break
		}
		class, rest, ok := strings.Cut(rest, markText)
		text, rest, ok2 := strings.Cut(rest, markClose)
		// the markers are invalid UTF-8 on purpose, so look for their bytes
		if !ok || !ok2 || strings.Contains(text, markOpen) || strings.Contains(text, markText) {
			return nil, fmt.Errorf("highlight: bad marker at offset %d", pos)
		}
		start := pos
		if err := copied(text); err != nil {
			return nil, err
		}
		if class != "" && pos > start {
			spans = append(spans, Span{Start: start, End: pos, Class: class})
		}
		out = rest
	}
	if pos != len(src) {
		return nil, fmt.Errorf("highlight: the rules stopped at offset %d of %d", pos, len(src))
	}
	return spans, nil
}

// diff returns the length of the common prefix of a and b.
func diff(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}
