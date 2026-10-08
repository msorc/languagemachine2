package machine

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// Regression tests against the material from the original Language Machine
// release, copied into examples/ (see examples/README.md).

const examplesDir = "../../examples"

func example(parts ...string) string {
	return filepath.Join(append([]string{examplesDir}, parts...)...)
}

func readFile(t testing.TB, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// runFiles loads rules and returns the stdout produced from the given input
// files, queued in order the way the lm command line does it.
func runFiles(t testing.TB, rules string, files ...string) string {
	t.Helper()
	out, err := runFilesErr(rules, files...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func runFilesErr(rules string, files ...string) (string, error) {
	return captureErr(rules, func(e *Engine) error {
		for _, f := range files {
			g, err := NewFileInput(e, f)
			if err != nil {
				return err
			}
			e.AppendInput(g)
		}
		return nil
	})
}

// lmnSrc names a file of the lmn compiler sources.
func lmnSrc(name string) string {
	return filepath.Join("..", "lmnsrc", name)
}

var lmnSources = []string{lmnSrc("lmn2xfe.lmn"), lmnSrc("lmn2mbe.lmn")}

var (
	lmnOnce   sync.Once
	lmnStage2 string
)

// compiler returns the lmn compiler rebuilt from the current lmn2xfe/lmn2mbe
// sources. lmnbs.lm predates those sources, so its own output (stage 1) is
// only used to build stage 2.
func compiler(t testing.TB) string {
	t.Helper()
	lmnOnce.Do(func() {
		stage1 := runFiles(t, readFile(t, lmnSrc("lmnbs.lm")), lmnSources...)
		lmnStage2 = runFiles(t, stage1, lmnSources...)
	})
	if lmnStage2 == "" {
		t.Fatal("building the lmn compiler failed")
	}
	return lmnStage2
}

// The compiler rebuilt from its own sources must reproduce itself.
func TestLmnBootstrapFixpoint(t *testing.T) {
	t.Parallel()
	stage2 := compiler(t)
	if stage3 := runFiles(t, stage2, lmnSources...); stage3 != stage2 {
		t.Errorf("stage 3 differs from stage 2 (%d vs %d bytes)", len(stage3), len(stage2))
	}
}

// Port of the original test-inc make target: lmn2minc.lmn only
// .includes the compiler sources, so compiling it must give the compiler.
// Not parallel: the include paths are relative to the working directory.
func TestLmnInclude(t *testing.T) {
	stage2 := compiler(t)
	t.Chdir(example("testing"))
	if got := runFiles(t, stage2, "lmn2minc.lmn"); got != stage2 {
		t.Errorf("lmn2minc output differs from the compiler (%d vs %d bytes)", len(got), len(stage2))
	}
}

// Every grammar in examples/, and every lmn compiler source, must compile,
// and the result must load.
func TestExamplesCompile(t *testing.T) {
	t.Parallel()
	stage2 := compiler(t)
	files, err := filepath.Glob(example("*", "*.lmn"))
	if err != nil {
		t.Fatal(err)
	}
	compilers, err := filepath.Glob(lmnSrc("*.lmn"))
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, compilers...)
	for _, f := range files {
		if filepath.Base(f) == "lmn2minc.lmn" {
			continue // covered by TestLmnInclude, needs its own cwd
		}
		t.Run(filepath.Base(filepath.Dir(f))+"/"+filepath.Base(f), func(t *testing.T) {
			t.Parallel()
			out := runFiles(t, stage2, f)
			if out == "" {
				t.Fatalf("no output compiling %s", f)
			}
			if err := NewEngine().LoadFromString(out); err != nil {
				t.Errorf("compiled %s does not load: %v", f, err)
			}
		})
	}
}

// Samples with the reference output shipped with the original release.
func TestSamplesGolden(t *testing.T) {
	t.Parallel()
	stage2 := compiler(t)
	cases := []struct{ grammar, input, want string }{
		{"flatten.lmn", "flatten.input", "flatten.flat"},
		{"reorder.lmn", "reorder.lmn", "reorder.reorder"},
	}
	for _, c := range cases {
		t.Run(c.grammar, func(t *testing.T) {
			t.Parallel()
			rules := runFiles(t, stage2, example("samples", c.grammar))
			got := runFiles(t, rules, example("samples", c.input))
			if want := readFile(t, example("samples", c.want)); got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
}

// golog.lmn adds a _log twin to every function of a Go source file. The
// sample covers receivers, variadic and unnamed parameters, a multi-line
// parameter list, braces in strings and comments, and the functions that
// are left alone (generic, no body, main).
func TestGologGolden(t *testing.T) {
	t.Parallel()
	rules := runFiles(t, compiler(t), example("golang", "golog.lmn"))
	got := runFiles(t, rules, example("golang", "sample.go.txt"))
	if want := readFile(t, example("golang", "sample.golog.txt")); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

const originalDocs = "../../docs/original"

// The lambda experiment: lct translates each .lam into the published
// .out.lmn, and the translation run with the lcm runtime prints the output
// published on the website (arithmeticoutput, listsoutput).
func TestLambdaGolden(t *testing.T) {
	t.Parallel()
	stage2 := compiler(t)
	lct := runFiles(t, stage2, example("web", "lct.lmn"))
	cases := []struct{ name, output string }{
		{"arithmetic", "arithmeticoutput.wiki"},
		{"lists", "listsoutput.wiki"},
		{"fact2", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
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

// The control statements and rule values that 0.2.5 compiled but could not
// run (testdata/control.lmn).
func TestControlStatements(t *testing.T) {
	t.Parallel()
	rules := runFiles(t, compiler(t), filepath.Join("testdata", "control.lmn"))
	for _, c := range []struct{ input, want string }{
		{"w1:", "while:  Av = 0 Bv = 1 Cv = 5 Sv = null\n"},
		{"w2:", "while:  Av = 0 Bv = 20 Cv = 4 Sv = null\n"},
		{"f1:", "break:  Av = 0 Bv = 1 Cv = 3 Sv = null\n"},
		{"f2:", "continue:  Av = 0 Bv = 3 Cv = 4 Sv = null\n"},
		{"f3:", "nested:  Av = 3 Bv = 6 Cv = 2 Sv = null\n"},
		{"h1:", "foreach:  Av = 0 Bv = 3142 Cv = 2 Sv = 0\n"},
		{"h2:", "foreach:  Av = 7 Bv = 13 Cv = 4 Sv = 0\n"},
		{"h3:", "foreach:  Av = 0 Bv = 0 Cv = null Sv = null\n"},
		{"e1:abc;", "each: abc|abc|c|c|| Av = 0 Bv = 1 Cv = null Sv = X\n"},
		{"r1:zz", "defined:  Av = 0 Bv = 1 Cv = null Sv = tt\nrule:  Av = 0 Bv = 1 Cv = null Sv = tt\n"},
	} {
		if got := run(t, rules, c.input); got != c.want {
			t.Errorf("%s: got %q, want %q", c.input, got, c.want)
		}
	}

	e := NewEngine()
	if err := e.LoadFromString(rules); err != nil {
		t.Fatal(err)
	}
	e.AppendInput(NewStringInput(e, "b1:"))
	if _, err := e.Start(); err == nil || !strings.Contains(err.Error(), "break outside a loop") {
		t.Errorf("break outside a loop: err %v", err)
	}
}

// The Java front end parses hello.java and dumps its intermediate
// representation (hello.java.dump). It has no back end, so its output is
// the dump of the tree.
func TestJavaFrontEnd(t *testing.T) {
	t.Parallel()
	rules := runFiles(t, compiler(t), example("translators", "j2xfe.lmn"))
	got := runFiles(t, rules, example("translators", "hello.java"))
	if want := readFile(t, example("translators", "hello.java.dump")); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

// Small runs whose expected output was taken from the original engine.
func TestExamplesAgainstOriginal(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			rules := runFiles(t, stage2, example(c.dir, c.grammar))
			if got := run(t, rules, c.input); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// asciiDiagram maps the port's box drawing to the original's ASCII.
var asciiDiagram = strings.NewReplacer("┌", ".", "┐", ".", "└", "'", "┘", "'", "│", "|", "─", "-")

// maskAddr hides the addresses that variable traces print, which differ
// between runs.
var maskAddr = regexp.MustCompile(`\b[0-9A-F]{6,}\b`)

// golden compares got with testdata/<name>, or rewrites the file when the
// tests run with LM_UPDATE_GOLDEN=1.
func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if os.Getenv("LM_UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want := readFile(t, path)
	if got == want {
		return
	}
	gl, wl := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(gl) && i < len(wl); i++ {
		if gl[i] != wl[i] {
			t.Fatalf("line %d differs:\ngot  %q\nwant %q", i+1, gl[i], wl[i])
		}
	}
	t.Fatalf("got %d lines, want %d", len(gl), len(wl))
}

// Traces and lm-diagrams against the output of the original engine, kept in
// testdata/trace (see its README). The cases marked port are the Go port's own
// output, pinned so that refactoring cannot change it unnoticed.
func TestTraceGolden(t *testing.T) {
	t.Parallel()
	stage2 := compiler(t)
	cats := "the cat likes the dog .\n"
	fp := "/ 307 241\n"
	allVars := TraceContextScope | TraceCVar | TraceLVar | TraceRVar | TraceRVarVar | TraceRVarScope | TraceRef | TraceRefScope | TraceRefVar | TraceEach | TraceEachScope | TraceEachRefVar | TraceDebug
	cases := []struct {
		golden, grammar string
		width           int
		flags           TraceFlag
		input           string
		port            bool
		rules           string // bytecode to run instead of compiling grammar
	}{
		{"cats.diagram", "web/cats.lmn", 40, TraceDiagram, cats, false, ""},
		{"cats.diagram-text", "web/cats.lmn", 40, TraceDiagramText, cats, false, ""},
		{"cats.mismatch-symbols", "web/cats.lmn", 0, TraceMismatch | TraceSymbols, cats, false, ""},
		{"fpCalc.diagram", "basics/fpCalc.lmn", 50, TraceDiagram, fp, false, ""},
		{"fpCalc.mismatch-symbols", "basics/fpCalc.lmn", 0, TraceMismatch | TraceSymbols, fp, false, ""},
		{"rpCalc.diagram", "basics/rpCalc.lmn", 40, TraceDiagram, "0 5 N + 2 * =\n", false, ""},
		{"cats.grammar", "web/cats.lmn", 0, TraceGrammar, cats, true, ""},
		{"cats.vars", "web/cats.lmn", 0, allVars, cats, true, ""},
		{"fpCalc.arith-assign", "basics/fpCalc.lmn", 0, TraceArithmetic | TraceAssign, fp, true, ""},
		{"expr.apply", "", 0, TraceApply, "a", true, exprRules},
		{"control.arith-relation", "testdata/control.lmn", 0, TraceArithmetic | TraceRelation, "w2:", true, ""},
		{"control.assign-loop", "testdata/control.lmn", 0, TraceAssign | TraceLoop, "f3:", true, ""},
		{"control.index", "testdata/control.lmn", 0, TraceIndex, "h2:", true, ""},
		{"control.vars-each", "testdata/control.lmn", 0, allVars, "e1:abc;", true, ""},
		{"control.vars-foreach", "testdata/control.lmn", 0, allVars, "h2:", true, ""},
	}
	for _, c := range cases {
		t.Run(c.golden, func(t *testing.T) {
			t.Parallel()
			rules := c.rules
			if rules == "" {
				src := example(strings.Split(c.grammar, "/")...)
				if dir, name, ok := strings.Cut(c.grammar, "/"); ok && dir == "testdata" {
					src = filepath.Join(dir, name)
				}
				rules = runFiles(t, stage2, src)
			}
			got := capture(t, rules, func(e *Engine) {
				if c.width > 0 {
					e.SetDiagramWidth(c.width)
				}
				e.SetTraceFlag(c.flags)
				e.AppendInput(NewStringInput(e, c.input))
			})
			got = asciiDiagram.Replace(got)
			if c.port {
				got = maskAddr.ReplaceAllString(got, "ADDR")
			}
			golden(t, filepath.Join("trace", c.golden+".txt"), got)
		})
	}
}
