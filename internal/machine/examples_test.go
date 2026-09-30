package machine

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Regression tests against the material shipped with lm-0.2.5, copied into
// examples/ (see examples/README.md).

const examplesDir = "../../examples"

func example(parts ...string) string {
	return filepath.Join(append([]string{examplesDir}, parts...)...)
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// runFiles loads rules and returns the stdout produced from the given input
// files, queued in order the way the lm command line does it.
func runFiles(t *testing.T, rules string, files ...string) string {
	t.Helper()
	return capture(t, rules, func(e *Engine) {
		for _, f := range files {
			e.AppendInput(NewGramInputFile(e, f))
		}
	})
}

var lmnSources = []string{example("lmn", "lmn2xfe.lmn"), example("lmn", "lmn2mbe.lmn")}

var (
	lmnOnce   sync.Once
	lmnStage2 string
)

// compiler returns the lmn compiler rebuilt from the current lmn2xfe/lmn2mbe
// sources. lmnbs.lm predates those sources, so its own output (stage 1) is
// only used to build stage 2.
func compiler(t *testing.T) string {
	t.Helper()
	lmnOnce.Do(func() {
		stage1 := runFiles(t, readFile(t, example("lmn", "lmnbs.lm")), lmnSources...)
		lmnStage2 = runFiles(t, stage1, lmnSources...)
	})
	if lmnStage2 == "" {
		t.Fatal("building the lmn compiler failed")
	}
	return lmnStage2
}

// The compiler rebuilt from its own sources must reproduce itself.
func TestLmnBootstrapFixpoint(t *testing.T) {
	stage2 := compiler(t)
	if stage3 := runFiles(t, stage2, lmnSources...); stage3 != stage2 {
		t.Errorf("stage 3 differs from stage 2 (%d vs %d bytes)", len(stage3), len(stage2))
	}
}

// Port of test-inc from lm-0.2.5/src/testing/Makefile: lmn2minc.lmn only
// .includes the compiler sources, so compiling it must give the compiler.
func TestLmnInclude(t *testing.T) {
	stage2 := compiler(t)
	t.Chdir(example("testing"))
	if got := runFiles(t, stage2, "lmn2minc.lmn"); got != stage2 {
		t.Errorf("lmn2minc output differs from the compiler (%d vs %d bytes)", len(got), len(stage2))
	}
}

// Every grammar in examples/ must compile without a runtime failure.
func TestExamplesCompile(t *testing.T) {
	stage2 := compiler(t)
	files, err := filepath.Glob(example("*", "*.lmn"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if filepath.Base(f) == "lmn2minc.lmn" {
			continue // covered by TestLmnInclude, needs its own cwd
		}
		t.Run(filepath.Base(filepath.Dir(f))+"/"+filepath.Base(f), func(t *testing.T) {
			if out := runFiles(t, stage2, f); out == "" {
				t.Errorf("no output compiling %s", f)
			}
		})
	}
}

// Samples with the reference output shipped in lm-0.2.5/src/samples.
func TestSamplesGolden(t *testing.T) {
	stage2 := compiler(t)
	cases := []struct{ grammar, input, want string }{
		{"flatten.lmn", "flatten.input", "flatten.flat"},
		{"reorder.lmn", "reorder.lmn", "reorder.reorder"},
	}
	for _, c := range cases {
		t.Run(c.grammar, func(t *testing.T) {
			rules := runFiles(t, stage2, example("samples", c.grammar))
			got := runFiles(t, rules, example("samples", c.input))
			if want := readFile(t, example("samples", c.want)); got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
}

const originalDocs = "../../docs/original"

// The lambda experiment: lct translates each .lam into the published
// .out.lmn, and the translation run with the lcm runtime prints the output
// published on the website (arithmeticoutput, listsoutput).
func TestLambdaGolden(t *testing.T) {
	stage2 := compiler(t)
	lct := runFiles(t, stage2, example("web", "lct.lmn"))
	cases := []struct{ name, output string }{
		{"arithmetic", "arithmeticoutput.wiki"},
		{"lists", "listsoutput.wiki"},
		{"fact2", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := runFiles(t, lct, example("lambda", c.name+".lam"))
			if want := readFile(t, example("lambda", c.name+".out.lmn")); got != want {
				t.Errorf("lct output differs from %s.out.lmn", c.name)
			}
			if c.output == "" {
				return
			}
			rules := runFiles(t, stage2, example("web", "lcm.lmn"), example("lambda", c.name+".lmn"))
			got = run(t, rules, "z")
			if want := readFile(t, filepath.Join(originalDocs, c.output)); got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
}

// lexicalbuffer against the results published in lexicalresults.wiki.
func TestLexicalBufferGolden(t *testing.T) {
	stage2 := compiler(t)
	page := readFile(t, filepath.Join(originalDocs, "lexicalresults.wiki"))
	_, results, ok := strings.Cut(page, "== the results ==\n")
	if !ok {
		t.Fatal("no results section in lexicalresults.wiki")
	}
	var want strings.Builder
	for _, line := range strings.SplitAfter(results, "\n") {
		want.WriteString(strings.TrimPrefix(line, "  "))
	}
	rules := runFiles(t, stage2, example("testing", "lexicalbuffer.lmn"))
	if got := runFiles(t, rules, example("testing", "lexicalbuffer.input")); got != want.String() {
		t.Errorf("got\n%s\nwant\n%s", got, want.String())
	}
}

// Small runs whose expected output was taken from the original lm-0.2.5.
func TestExamplesAgainstOriginal(t *testing.T) {
	stage2 := compiler(t)
	cases := []struct{ dir, grammar, input, want string }{
		// whitespace deletion inside a context whose goal is a terminal
		{"basics", "rpCalc.lmn", "0 5 N + 2 * =\n3 4 + =\n", "result: -10\nresult: 7\n"},
		// hexadecimal fractions are read with C's strtod
		{"basics", "calc.lmn", "0x10.8=\n", "the answer is 16.5\n"},
		// deferred patterns stored in an array and expanded with $(...)
		{"web", "grokcpp.lmn", "#define f(x,y) <<y x>>\nf(a,(b,c))\n", "\n<<(b,c) a>>\n"},
	}
	for _, c := range cases {
		t.Run(c.grammar, func(t *testing.T) {
			rules := runFiles(t, stage2, example(c.dir, c.grammar))
			if got := run(t, rules, c.input); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
