// Package lm runs Language Machine rulesets built into a Go program. It is
// the runtime for the code that cmd/lmn2go generates: a generated file
// declares a Program, holding the compiled rules and the Go functions the
// rules call, and either runs it as a command (Main) or leaves it to the
// caller (Run, Translate).
package lm

import (
	"bytes"
	"fmt"
	"io"
	"os"

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
func Sym(s string) Value { return Value{machine.NewSym(s)} }

// Num returns the number x.
func Num(x float64) Value { return Value{machine.NewNumber(machine.LMNumber(x))} }

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
// is the command name and the rest are lm options and input files. It
// returns the exit status.
func (p *Program) Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{p.Name}
	}
	return application.NewProgramApplication(args, p.app(), stdout, stderr).Start()
}

// Main runs the program on the process's command line and exits.
func Main(p *Program) {
	os.Exit(p.Run(os.Args, os.Stdout, os.Stderr))
}

// Translate runs the program on input and returns what it writes to its
// standard output. Extra lm options, such as "-trace", "m", may precede the
// input.
func (p *Program) Translate(input string, options ...string) (string, error) {
	var out, errOut bytes.Buffer
	args := append([]string{p.Name}, options...)
	args = append(args, "-input", input)
	if status := p.Run(args, &out, &errOut); status != 0 {
		return out.String(), fmt.Errorf("%s: exit status %d: %s", p.Name, status, bytes.TrimSpace(errOut.Bytes()))
	}
	return out.String(), nil
}
