package highlight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/msorc/languagemachine2/internal/lmgo"
)

// The generated files must be what the //go:generate lines write now.
func TestUpToDate(t *testing.T) {
	t.Parallel()
	for _, out := range []string{"go_lm.go", "lmn_lm.go"} {
		if err := lmgo.UpToDate("highlight.go", out); err != nil {
			t.Error(err)
		}
	}
}

// classes returns the spans as "class:text" strings.
func classes(t *testing.T, lang, src string) []string {
	t.Helper()
	h, ok := ByName(lang)
	if !ok {
		t.Fatalf("no language %q", lang)
	}
	spans, err := h.Spans(src)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, s := range spans {
		got = append(got, s.Class+":"+src[s.Start:s.End])
	}
	return got
}

func TestSpans(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, lang, src string
		want            []string
	}{
		{"words", "go", "func add(a int) bool { return len(s) > 0x1F || nil }",
			[]string{"keyword:func", "function:add", "type:int", "type:bool", "keyword:return", "builtin:len", "number:0x1F", "constant:nil"}},
		{"word boundaries", "go", "function iffy int8x café", nil},
		{"conversion and call", "go", "string(x) + pkg.Do(y) + f (z)",
			[]string{"type:string", "function:Do"}},
		{"comments", "go", "x // one\n/* two\nlines */ y /**/",
			[]string{"comment:// one", "comment:/* two\nlines */", "comment:/**/"}},
		{"strings", "go", `"a\"b" + '\'' + ` + "`raw\n\\`" + ` + "ü" // "c"`,
			[]string{`string:"a\"b"`, `string:'\''`, "string:`raw\n\\`", `string:"ü"`, `comment:// "c"`}},
		{"not closed", "go", "a = \"abc\nb = 'x\nc = `raw\n/* end",
			[]string{`string:"abc`, "string:'x", "string:`raw\n/* end"}},
		{"comment not closed", "go", "x /* end", []string{"comment:/* end"}},
		{"lmn", "lmn", "wiki 'text'\n .[0-9] % toNum :N <- x :N; // done\nmore\n",
			[]string{"comment:wiki 'text'", "regexp:.[0-9]", "builtin:toNum", "variable:N", "operator:<-", "variable:N", "comment:// done", "comment:more"}},
		{"lmn code first", "lmn", " 'a\\'' \"b\" repeat /* x /* y */ z */ eof",
			[]string{`string:'a\''`, `string:"b"`, "keyword:repeat", "comment:/* x /* y */ z */", "builtin:eof"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classes(t, tt.lang, tt.src)
			if strings.Join(got, "|") != strings.Join(tt.want, "|") {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

// Whatever the input is, the rules copy all of it: no source may make them
// fail. The sources of this repository and some awkward inputs are tried.
func TestCopiesEverything(t *testing.T) {
	t.Parallel()
	inputs := map[string][]string{
		"go":  {"", "\n", "-", "\r\n\"a\r\n", "\\", "'", "\"\\", "/", "/*", "*/", "é(", "\x00\x01 �", "0", "a.b.c(", "((("},
		"lmn": {"", "\n", " ", "x", " x", " .[", " '\\", " /* /* */", " <", "\n\n x\ny"},
	}
	for lang, glob := range map[string]string{"go": "../../internal/machine/*.go", "lmn": "../*/*.lmn"} {
		files, err := filepath.Glob(glob)
		if err != nil || len(files) == 0 {
			t.Fatalf("%s: no files (%v)", glob, err)
		}
		for _, f := range files {
			text, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			inputs[lang] = append(inputs[lang], string(text))
		}
	}
	for lang, srcs := range inputs {
		h, _ := ByName(lang)
		for _, src := range srcs {
			t.Run(lang, func(t *testing.T) {
				t.Parallel()
				spans, err := h.Spans(src)
				if err != nil {
					t.Fatalf("%.40q: %v", src, err)
				}
				end := 0
				for _, s := range spans {
					if s.Start < end || s.End <= s.Start || s.End > len(src) {
						t.Fatalf("%.40q: bad span %+v after offset %d", src, s, end)
					}
					end = s.End
				}
			})
		}
	}
}

func TestInvalidInput(t *testing.T) {
	t.Parallel()
	h, _ := ByName("go")
	if _, err := h.Spans("x \xff y"); err == nil {
		t.Error("no error for input that is not UTF-8")
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	mark := func(class, text string) string { return markOpen + class + markText + text + markClose }
	spans, err := parse("a "+mark("k", "if")+" b"+mark("", "c"), "a if bc")
	if err != nil || len(spans) != 1 || spans[0] != (Span{2, 4, "k"}) {
		t.Errorf("spans %+v, error %v", spans, err)
	}
	for name, out := range map[string]string{
		"changed":   "a IF b",
		"dropped":   "a if",
		"added":     "a if b c",
		"unclosed":  "a " + markOpen + "k" + markText + "if b",
		"nested":    "a " + mark("k", mark("k", "if")) + " b",
		"no marker": "a " + markOpen + "if b",
	} {
		if _, err := parse(out, "a if b"); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestRender(t *testing.T) {
	t.Parallel()
	src := "if x < \"a\" /* b\nc */"
	spans := []Span{{0, 2, "keyword"}, {7, 10, "string"}, {11, 20, "comment"}}
	var b strings.Builder
	if err := WriteANSI(&b, src, spans); err != nil {
		t.Fatal(err)
	}
	if want := "\x1b[35mif\x1b[0m x < \x1b[32m\"a\"\x1b[0m \x1b[90m/* b\x1b[0m\n\x1b[90mc */\x1b[0m"; b.String() != want {
		t.Errorf("ansi: got %q\nwant %q", b.String(), want)
	}
	b.Reset()
	if err := WriteHTML(&b, src, spans); err != nil {
		t.Fatal(err)
	}
	if want := `<pre class="hl"><span class="hl-keyword">if</span> x &lt; <span class="hl-string">&#34;a&#34;</span> <span class="hl-comment">/* b` + "\n" + `c */</span></pre>` + "\n"; b.String() != want {
		t.Errorf("html: got %q\nwant %q", b.String(), want)
	}
	b.Reset()
	if err := WriteJSON(&b, nil); err != nil || b.String() != "[]\n" {
		t.Errorf("json: got %q, error %v", b.String(), err)
	}
}
