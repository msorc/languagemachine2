package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/msorc/languagemachine2/internal/version"
)

// Rules for grammar t: the goal eof is out, which copies the input, and a is
// replaced by A.
const baseRules = `m:t L:0 n:1 ( z m:out ) ( m:eof ) r
m:t L:0 n:1 ( c:a ) ( z c:A ) r
`

// More rules for t: b is replaced by B.
const addRules = `m:t L:0 n:1 ( c:b ) ( z c:B ) r
`

// Rules whose goal recurses for ever.
const nestRules = `m:t L:0 n:0 ( z m:nest ) ( m:eof ) r
m:t L:0 n:0 ( z m:nest ) ( m:nest ) r
`

// Rules that send everything to err.
const errRules = `m:t L:0 n:1 ( z m:err ) ( m:eof ) r
`

type result struct {
	status         int
	stdout, stderr string
}

// run runs lm2 with args in a temporary directory that holds the files.
func run(t *testing.T, files map[string]string, args ...string) result {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	status := New(append([]string{"lm2"}, args...), nil, strings.NewReader(""), &stdout, &stderr).Start()
	return result{status, stdout.String(), stderr.String()}
}

func (r result) check(t *testing.T, status int, stdout string) {
	t.Helper()
	if r.status != status || r.stdout != stdout {
		t.Errorf("got status %d, stdout %q (stderr %q); want %d, %q", r.status, r.stdout, r.stderr, status, stdout)
	}
}

func TestInputs(t *testing.T) {
	files := map[string]string{"base.lm2": baseRules, "one": "a1", "two": "a2"}
	run(t, files, "-rules", "base.lm2", "-input", "xa").check(t, 0, "xA")
	// positional files are read in order, after -input
	run(t, files, "-rules", "base.lm2", "-input", "a0", "one", "two").check(t, 0, "A0A1A2")
}

func TestAddRules(t *testing.T) {
	files := map[string]string{"base.lm2": baseRules, "add.lm2": addRules}
	run(t, files, "-rules", "base.lm2", "-add", "add.lm2", "-input", "ab").check(t, 0, "AB")
	// options take effect in order: -rules replaces what -add loaded
	run(t, files, "-add", "add.lm2", "-rules", "base.lm2", "-input", "ab").check(t, 0, "Ab")
}

func TestOutputFiles(t *testing.T) {
	r := run(t, map[string]string{"base.lm2": baseRules}, "-rules", "base.lm2", "-output", "out.txt", "-input", "abc")
	r.check(t, 0, "")
	if b, err := os.ReadFile("out.txt"); err != nil || string(b) != "Abc" {
		t.Errorf("-output file: %q, %v", b, err)
	}

	r = run(t, map[string]string{"err.lm2": errRules}, "-rules", "err.lm2", "-errout", "err.txt", "-input", "oops")
	r.check(t, 0, "")
	if b, err := os.ReadFile("err.txt"); err != nil || string(b) != "oops" {
		t.Errorf("-errout file: %q, %v", b, err)
	}
	if r := run(t, map[string]string{"err.lm2": errRules}, "-rules", "err.lm2", "-input", "oops"); r.stderr != "oops" {
		t.Errorf("err without -errout: stderr %q", r.stderr)
	}
}

func TestErrors(t *testing.T) {
	files := map[string]string{"base.lm2": baseRules, "nest.lm2": nestRules, "bad.lm2": "m:t L:x r\n"}
	cases := []struct {
		args []string
		want string // in stderr
	}{
		{[]string{"-rules", "missing.lm2"}, "lm2: open missing.lm2"},
		{[]string{"-rules", "bad.lm2"}, "lm2: bad.lm2:1:5: bad priority"},
		{[]string{"-rules", "base.lm2", "missing.txt"}, "lm2: open missing.txt"},
		{[]string{"-rules", "base.lm2", "-trace", "Q", "-input", "a"}, "Q"},
		{[]string{"-nosuchflag"}, "flag provided but not defined"},
		{[]string{"-rules", "nest.lm2", "-max-depth", "20", "-input", "x"}, "lm2: input:1:1: maximum depth 20 exceeded"},
	}
	for _, c := range cases {
		r := run(t, files, c.args...)
		if r.status != 1 || !strings.Contains(r.stderr, c.want) {
			t.Errorf("%v: status %d, stderr %q; want 1 and %q", c.args, r.status, r.stderr, c.want)
		}
	}
	// a failed analysis is status 1 without an error message
	if r := run(t, files, "-input", "x"); r.status != 1 || r.stderr != "" {
		t.Errorf("no rules: status %d, stderr %q", r.status, r.stderr)
	}
}

func TestInformation(t *testing.T) {
	r := run(t, nil, "-h")
	if r.status != 0 || !strings.Contains(r.stderr, "Usage of lm2:") || !strings.Contains(r.stderr, "-trace") {
		t.Errorf("-h: status %d, stderr %q", r.status, r.stderr)
	}
	r = run(t, map[string]string{"base.lm2": baseRules}, "-version", "-rules", "base.lm2", "-input", "a")
	if r.status != 0 || !strings.HasPrefix(r.stdout, "lm2: language machine version "+version.Version()) || !strings.HasSuffix(r.stdout, "\nA") {
		t.Errorf("-version: status %d, stdout %q", r.status, r.stdout)
	}
	r = run(t, map[string]string{"base.lm2": baseRules}, "-shebang", "/usr/bin/lm2", "-rules", "base.lm2", "-input", "")
	if !strings.HasPrefix(r.stdout, "#! /usr/bin/lm2 -rules\n") || !strings.HasSuffix(r.stdout, "\n") {
		t.Errorf("-shebang: stdout %q", r.stdout)
	}
	// the header ends its line, so the rules that follow it all load
	run(t, map[string]string{"script.lm2": r.stdout + baseRules}, "-rules", "script.lm2", "-input", "xa").check(t, 0, "xA")
}

func TestTraceOptions(t *testing.T) {
	files := map[string]string{"base.lm2": baseRules}
	// codes combine as CSV or by repeating the flag
	csv := run(t, files, "-rules", "base.lm2", "-trace", "m,s", "-input", "a")
	rep := run(t, files, "-rules", "base.lm2", "-trace", "m", "-trace", "s", "-input", "a")
	if csv.status != 0 || csv.stdout != rep.stdout || !strings.Contains(csv.stdout, "??") || !strings.Contains(csv.stdout, "--") {
		t.Errorf("-trace m,s and -trace m -trace s differ or lack events:\n%s\n---\n%s", csv.stdout, rep.stdout)
	}

	// -dwidth applies only to a diagram started after it
	width := func(r result) int {
		n := 0
		for l := range strings.SplitSeq(r.stdout, "\n") {
			n = max(n, len([]rune(l)))
		}
		return n
	}
	narrow := run(t, files, "-rules", "base.lm2", "-dwidth", "40", "-trace", "D", "-input", "a")
	late := run(t, files, "-rules", "base.lm2", "-trace", "D", "-dwidth", "40", "-input", "a")
	if width(narrow) >= width(late) {
		t.Errorf("-dwidth before -trace D: width %d, after: %d", width(narrow), width(late))
	}
}

// More rules for t: c is replaced by C.
const add2Rules = `m:t L:0 n:1 ( c:c ) ( z c:C ) r
`

// Every occurrence of a flag takes effect, in command-line order.
func TestRepeatedFlags(t *testing.T) {
	files := map[string]string{"base.lm2": baseRules, "add.lm2": addRules, "add2.lm2": add2Rules}
	run(t, files, "-rules", "base.lm2", "-input", "a", "-input", "b").check(t, 0, "Ab")
	run(t, files, "-rules", "base.lm2", "-add", "add.lm2", "-add", "add2.lm2", "-input", "abc").check(t, 0, "ABC")
	// each -rules replaces what was loaded before it
	run(t, files, "-rules", "add.lm2", "-rules", "base.lm2", "-input", "ab").check(t, 0, "Ab")
	run(t, files, "-rules", "base.lm2", "-add", "add.lm2", "-rules", "base.lm2", "-input", "ab").check(t, 0, "Ab")
}

// Trace codes take effect in order, and z turns the earlier ones off.
func TestTraceOrder(t *testing.T) {
	files := map[string]string{"base.lm2": baseRules}
	off := run(t, files, "-rules", "base.lm2", "-trace", "m", "-trace", "z", "-input", "a")
	off.check(t, 0, "A")
	only := run(t, files, "-rules", "base.lm2", "-trace", "m,z,s", "-input", "a")
	if only.status != 0 || strings.Contains(only.stdout, "??") || !strings.Contains(only.stdout, "--") {
		t.Errorf("-trace m,z,s: status %d, stdout %q", only.status, only.stdout)
	}
	// every code in the table is accepted and described in the usage
	help := run(t, nil, "-h").stderr
	for _, c := range traceCodes {
		if r := run(t, files, "-rules", "base.lm2", "-trace", c.code, "-input", "a"); r.status != 0 {
			t.Errorf("-trace %s: status %d, stderr %q", c.code, r.status, r.stderr)
		}
		if !strings.Contains(help, "  "+c.code+"  "+c.name) {
			t.Errorf("-h does not describe trace code %s (%s)", c.code, c.name)
		}
	}
}

// A diagram too narrow to draw is rejected instead of crashing.
func TestDiagramWidth(t *testing.T) {
	files := map[string]string{"base.lm2": baseRules}
	r := run(t, files, "-rules", "base.lm2", "-dwidth", "10", "-trace", "D", "-input", "a")
	if r.status != 1 || !strings.Contains(r.stderr, "at least 20") {
		t.Errorf("-dwidth 10: status %d, stderr %q", r.status, r.stderr)
	}
	if r := run(t, files, "-rules", "base.lm2", "-dwidth", "20", "-trace", "D", "-input", "a"); r.status != 0 {
		t.Errorf("-dwidth 20: status %d, stderr %q", r.status, r.stderr)
	}
	if r := run(t, files, "-rules", "base.lm2", "-dwidth", "x"); r.status != 1 || !strings.Contains(r.stderr, "invalid value") {
		t.Errorf("-dwidth x: status %d, stderr %q", r.status, r.stderr)
	}
}
