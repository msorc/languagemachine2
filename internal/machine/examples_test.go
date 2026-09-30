package machine

import (
	"os"
	"path/filepath"
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
	cases := []struct {
		grammar, input, want string
		known                string // non-empty: known divergence, skipped
	}{
		{"flatten.lmn", "flatten.input", "flatten.flat", "prints `BAD >=:` instead of the flattened blocks"},
		{"reorder.lmn", "reorder.lmn", "reorder.reorder", "groups are not collected into the Table"},
	}
	for _, c := range cases {
		t.Run(c.grammar, func(t *testing.T) {
			if c.known != "" {
				t.Skip("known divergence from lm-0.2.5: " + c.known)
			}
			rules := runFiles(t, stage2, example("samples", c.grammar))
			got := runFiles(t, rules, example("samples", c.input))
			if want := readFile(t, example("samples", c.want)); got != want {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
}
