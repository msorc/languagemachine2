// Package lm runs Language Machine rulesets built into a Go program. It is
// the runtime for the code that cmd/lmn2go generates: a generated file
// declares a Program, holding the compiled rules and the Go functions the
// rules call, and either runs it as a command (Main) or leaves it to the
// caller (Run, Translate, TranslateReader).
package lm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/msorc/languagemachine2/internal/application"
	"github.com/msorc/languagemachine2/internal/machine"
)

// Program is a compiled ruleset. Run and Translate may be called from several
// goroutines at once: each call loads the rules into a machine of its own. The
// Funcs are then called concurrently too, so they must be safe for that.
type Program struct {
	// Name is the command name used in messages when args are not given.
	Name string
	// Rules is the ruleset in .lm bytecode (docs/bytecode.md).
	Rules string
	// Funcs are the functions that the rules call by name, in addition to
	// the builtins (docs/lm/04-special-symbols-and-builtins.md). A Func
	// replaces a builtin of the same name.
	Funcs map[string]Func
}

// Func is a function the rules call by name, as f(a, b) in lmn. An error
// stops the run, which reports it the way it reports a fault in the rules; a
// panic in a Func is reported the same way instead of ending the process.
type Func func(c *Call) (Value, error)

// Call is one call of a Func from the rules.
type Call struct {
	// Name is the name the rules called the function by.
	Name string
	// Args are the arguments, without the function name.
	Args []Value
}

// Arg returns argument i, or the null value when there are fewer arguments.
func (c *Call) Arg(i int) Value {
	if i < 0 || i >= len(c.Args) {
		return Null()
	}
	return c.Args[i]
}

// Value is a value passed between the rules and Go: a symbol, a number, or
// another element of the machine such as an array or a buffer.
type Value struct {
	e machine.Element
}

// Sym returns the symbol s.
func Sym(s string) Value { return Value{machine.Symbol(s)} }

// Num returns the number x.
func Num(x float64) Value { return Value{machine.Number(x)} }

// Null returns the null value, which is what an unset variable holds.
func Null() Value { return Value{machine.Null()} }

func (v Value) val() machine.Element {
	if v.e == nil {
		return machine.Null()
	}
	return v.e.ToVal()
}

// String returns the value as text.
func (v Value) String() string { return v.val().ToString() }

// Number returns the value as a number.
func (v Value) Number() float64 { return float64(v.val().ToNumber()) }

// IsNumber reports whether the value is a number.
func (v Value) IsNumber() bool { return v.val().IsNumber() }

// Bool returns the truth of the value, as an if in the rules sees it.
func (v Value) Bool() bool { return v.val().ToBool() }

func (f Func) ext(name string) machine.ExtFn {
	return func(_ *machine.Stream, _ machine.GenMode, args []machine.Element) machine.Element {
		c := &Call{Name: name}
		if len(args) > 1 {
			c.Args = make([]Value, len(args)-1)
			for i, a := range args[1:] {
				c.Args[i] = Value{a}
			}
		}
		r, err := f.call(c)
		if err != nil {
			// the engine's own way to abort a run from inside matching
			panic(&machine.Error{Msg: fmt.Sprintf("%s: %v", name, err)})
		}
		if r.e == nil {
			return machine.Null()
		}
		return r.e
	}
}

// call runs f and turns a panic into an error.
func (f Func) call(c *Call) (v Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return f(c)
}

func (p *Program) app() *application.Program {
	ap := &application.Program{Rules: p.Rules, Funcs: make(map[string]machine.ExtFn, len(p.Funcs))}
	for name, f := range p.Funcs {
		ap.Funcs[name] = f.ext(name)
	}
	return ap
}

// Run runs the program as the lm command would run it with -rules: args[0]
// is the command name and the rest are lm options and input files. It reads
// stdin when the options say so (-stdin) or there are no inputs, and returns
// the exit status. A nil stdin is an empty input.
func (p *Program) Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{p.Name}
	}
	return application.New(args, p.app(), stdin, stdout, stderr).Start()
}

// Main runs the program on the process's command line and exits.
func Main(p *Program) {
	os.Exit(p.Run(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

// TranslateReader runs the program on the input r and writes what the rules
// write to standard output to w. It returns an error when the rules fail, as
// the lm command exits with status 1, with what the rules wrote to err.
func (p *Program) TranslateReader(r io.Reader, w io.Writer) error {
	var errOut bytes.Buffer
	e, err := p.app().Engine(w, &errOut)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Name, err)
	}
	e.AppendInput(machine.NewReaderInput(e, "input", r))
	if err := e.Run(); err != nil {
		if msg := bytes.TrimSpace(errOut.Bytes()); len(msg) > 0 {
			return fmt.Errorf("%s: %w: %s", p.Name, err, msg)
		}
		return fmt.Errorf("%s: %w", p.Name, err)
	}
	return nil
}

// Translate runs the program on input and returns what it writes to its
// standard output; see TranslateReader.
func (p *Program) Translate(input string) (string, error) {
	var out strings.Builder
	err := p.TranslateReader(strings.NewReader(input), &out)
	return out.String(), err
}
