package lm_test

import (
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
			"shout": func(c *lm.Call) lm.Value { return lm.Sym(strings.ToUpper(c.Arg(0).String()) + "!") },
			"count": func(c *lm.Call) lm.Value {
				if c.Name != "count" || !c.Arg(0).IsNumber() || c.Arg(2).Number() != 2.5 || c.Arg(3).String() != "null" {
					return lm.Sym("bad")
				}
				return lm.Num(float64(len(c.Args)))
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
		Funcs: map[string]lm.Func{"ucase": func(c *lm.Call) lm.Value { return lm.Sym("mine") }},
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
	if status := shoutProgram().Run([]string{"shout", "-version"}, &out, &errOut); status != 0 {
		t.Fatalf("status %d: %s", status, errOut.String())
	}
	if !strings.Contains(out.String(), "language machine version") {
		t.Errorf("-version printed %q", out.String())
	}
	if _, err := shoutProgram().Translate("x", "-nosuchflag"); err == nil {
		t.Error("bad option: no error")
	}
}
