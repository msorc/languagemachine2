package lm_test

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/msorc/languagemachine2/lm"
)

// Rules for shout.lmn (internal/lmgo/testdata): '!' outputs shout("hey") and
// '#' outputs count(4, "abc", 2.5); everything else is copied.
const shoutRules = `m:shout L:0 n:1 ( c:! v:W G v:shout G f:args d:hey G f:fun w . ) ( m:eof v:W ) r
m:shout L:0 n:1 ( c:# v:N G v:count G f:args n:4 G d:abc G n:2.5 G f:fun w . ) ( m:eof v:N ) r
m:shout L:0 n:1 ( z m:out ) ( m:eof ) r
`

func shoutProgram() *lm.Program {
	return &lm.Program{
		Name:  "shout",
		Rules: shoutRules,
		Funcs: map[string]lm.Func{
			"shout": func(c *lm.Call) (lm.Value, error) { return lm.Sym(strings.ToUpper(c.Arg(0).String()) + "!"), nil },
			"count": func(c *lm.Call) (lm.Value, error) {
				if c.Name != "count" || !c.Arg(0).IsNumber() || c.Arg(2).Number() != 2.5 || c.Arg(3).String() != "null" {
					return lm.Sym("bad"), nil
				}
				return lm.Num(float64(len(c.Args))), nil
			},
		},
	}
}

func TestTranslate(t *testing.T) {
	t.Parallel()
	got, err := shoutProgram().Translate("a!b#c")
	if err != nil {
		t.Fatal(err)
	}
	if want := "aHEY!b3c"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// One Program serves many goroutines: every Translate runs its own engine.
func TestTranslateConcurrent(t *testing.T) {
	t.Parallel()
	p := shoutProgram()
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 20 {
				got, err := p.Translate("a!b#c")
				if err != nil {
					t.Error(err)
					return
				}
				if want := "aHEY!b3c"; got != want {
					t.Errorf("got %q, want %q", got, want)
					return
				}
			}
		})
	}
	wg.Wait()
}

// Without its functions the rules still run: a call of a missing function
// is reported and gives 0, as in the original.
func TestMissingFunc(t *testing.T) {
	t.Parallel()
	p := &lm.Program{Name: "shout", Rules: shoutRules}
	got, err := p.Translate("!")
	if err != nil {
		t.Fatal(err)
	}
	if want := "external not found: shout hey \n0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A Func replaces the builtin of the same name.
func TestFuncReplacesBuiltin(t *testing.T) {
	t.Parallel()
	p := &lm.Program{
		Name:  "t",
		Rules: strings.ReplaceAll(shoutRules, "v:shout", "v:ucase"),
		Funcs: map[string]lm.Func{"ucase": func(*lm.Call) (lm.Value, error) { return lm.Sym("mine"), nil }},
	}
	got, err := p.Translate("!")
	if err != nil {
		t.Fatal(err)
	}
	if got != "mine" {
		t.Errorf("got %q, want %q", got, "mine")
	}
}

func TestRunOptions(t *testing.T) {
	t.Parallel()
	var out, errOut strings.Builder
	if status := shoutProgram().Run([]string{"shout", "-version"}, nil, &out, &errOut); status != 0 {
		t.Fatalf("status %d: %s", status, errOut.String())
	}
	if !strings.Contains(out.String(), "language machine version") {
		t.Errorf("-version printed %q", out.String())
	}
	if status := shoutProgram().Run([]string{"shout", "-nosuchflag"}, nil, &out, &errOut); status != 1 {
		t.Errorf("bad option: status %d", status)
	}
	// stdin is the reader given
	out.Reset()
	if status := shoutProgram().Run([]string{"shout", "-stdin"}, strings.NewReader("a!"), &out, &errOut); status != 0 || out.String() != "aHEY!" {
		t.Errorf("-stdin: status %d, output %q", status, out.String())
	}
}

// An error from a Func ends the run and is reported; a panic in a Func is
// reported too, instead of ending the process.
func TestFuncFailures(t *testing.T) {
	t.Parallel()
	boom := errors.New("boom")
	p := &lm.Program{
		Name:  "shout",
		Rules: shoutRules,
		Funcs: map[string]lm.Func{
			"shout": func(*lm.Call) (lm.Value, error) { return lm.Null(), boom },
			"count": func(*lm.Call) (lm.Value, error) { panic("oops") },
		},
	}
	if _, err := p.Translate("a!"); err == nil || !strings.Contains(err.Error(), "shout: boom") {
		t.Errorf("error from a Func: %v", err)
	}
	if _, err := p.Translate("a#"); err == nil || !strings.Contains(err.Error(), "count: panic: oops") {
		t.Errorf("panic in a Func: %v", err)
	}
}

// The report of an invalid operation goes to the program's stderr, not to
// the process's.
func TestInvalidOpReported(t *testing.T) {
	t.Parallel()
	p := &lm.Program{Name: "t", Rules: "m:t L:0 n:1 ( z m:out ) ( m:eof ) r\nm:t L:0 n:1 ( c:u v:Nv V f:postinc . ) ( z c:U ) r\n"}
	var out, errOut strings.Builder
	if status := p.Run([]string{"t", "-input", "u"}, nil, &out, &errOut); status != 0 {
		t.Fatalf("status %d: %s", status, errOut.String())
	}
	if out.String() != "U" || !strings.Contains(errOut.String(), "BAD ++ Nv (undefined)") {
		t.Errorf("stdout %q, stderr %q", out.String(), errOut.String())
	}
}

// TranslateReader streams the input and the output; a failed analysis is
// an error.
func TestTranslateReader(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	if err := shoutProgram().TranslateReader(strings.NewReader("x!y"), &out); err != nil {
		t.Fatal(err)
	}
	if out.String() != "xHEY!y" {
		t.Errorf("got %q", out.String())
	}
	p := &lm.Program{Name: "t", Rules: "m:t L:0 n:1 ( c:a ) ( m:eof ) r\n"}
	if _, err := p.Translate("b"); err == nil || !strings.Contains(err.Error(), "exit status 1") {
		t.Errorf("failed analysis: %v", err)
	}
}
