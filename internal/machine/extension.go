package machine

import (
	"fmt"
	"maps"
	"strings"
)

func lmFormat(sr *Stream, m GenMode, args []Element) Element {
	if len(args) < 2 {
		fail("format needs a format string")
	}
	return NewSym(doFormat(args[1:]))
}

// doFormat follows the original format builtin: the first argument is a printf-style format
// string; arguments left over after it are appended using their default
// formatting.
func doFormat(args []Element) string {
	var rs strings.Builder
	if len(args) == 0 {
		return ""
	}
	f := args[0].ToVal().ToString()
	av := args[1:]
	for i := 0; i < len(f); i++ {
		if f[i] != '%' {
			rs.WriteByte(f[i])
			continue
		}
		j := i + 1
		for j < len(f) && strings.IndexByte("+- #0123456789.", f[j]) >= 0 {
			j++
		}
		if j >= len(f) {
			rs.WriteString(f[i:])
			break
		}
		verb := f[j]
		spec := f[i:j]
		i = j
		if verb == '%' {
			rs.WriteByte('%')
			continue
		}
		if len(av) == 0 {
			rs.WriteString(spec + string(verb))
			continue
		}
		a := av[0].ToVal()
		av = av[1:]
		switch verb {
		case 'd', 'i', 'u':
			fmt.Fprintf(&rs, spec+"d", int64(a.ToNumber()))
		case 'x', 'X', 'o', 'b':
			fmt.Fprintf(&rs, spec+string(verb), int64(a.ToNumber()))
		case 'c':
			fmt.Fprintf(&rs, spec+"c", rune(a.ToNumber()))
		case 'e', 'E', 'f', 'F', 'g', 'G':
			fmt.Fprintf(&rs, spec+string(verb), float64(a.ToNumber()))
		default: // 's' and anything else
			fmt.Fprintf(&rs, spec+"s", a.ToString())
		}
	}
	for _, a := range av {
		rs.WriteString(a.ToVal().ToString())
	}
	return rs.String()
}

// ExtFn is a function that rules call by name (f:fun). args[0] is the function
// symbol and the arguments follow it.
type ExtFn func(*Stream, GenMode, []Element) Element

// arg1 adapts a builtin of one argument; Call supplies null when it is
// missing.
func arg1(f func(*Stream, Element) Element) ExtFn {
	return func(sr *Stream, _ GenMode, args []Element) Element { return f(sr, args[1]) }
}

// varArg adapts a builtin that takes a variable: called with anything else
// it gives null instead of failing.
func varArg(f func(*Stream, VarElement) Element) ExtFn {
	return func(sr *Stream, _ GenMode, args []Element) Element {
		v := args[1].ToVar()
		if v == nil {
			return Null()
		}
		return f(sr, v)
	}
}

// engineFn adapts a builtin that the engine implements on the whole argument
// vector.
func engineFn(f func(*Engine, []Element) Element) ExtFn {
	return func(sr *Stream, _ GenMode, args []Element) Element { return f(sr.Engine, args) }
}

// builtins are the functions that rules can call without registering them.
// A table is a copy of it, so an engine can add to its own.
var builtins = map[string]ExtFn{
	"octal":     arg1(Octal),
	"binary":    arg1(Binary),
	"hex":       arg1(Hex),
	"num":       arg1(Num),
	"usym":      arg1(Usym),
	"ulsym":     arg1(Ulsym),
	"uusym":     arg1(Uusym),
	"ssym":      arg1(Ssym),
	"slsym":     arg1(Slsym),
	"susym":     arg1(Susym),
	"variable":  arg1(Variable),
	"urn":       arg1(Urn),
	"urd":       arg1(Urd),
	"lcase":     arg1(Lcase),
	"ucase":     arg1(Ucase),
	"stripl":    arg1(Stripl),
	"stripr":    arg1(Stripr),
	"strip":     arg1(Strip),
	"toChars":   arg1(ToChars),
	"format":    lmFormat,
	"trOn":      engineFn((*Engine).SetTrace),
	"trOff":     engineFn((*Engine).UnsetTrace),
	"include":   engineFn((*Engine).Include),
	"use":       engineFn((*Engine).SetMachineElements),
	"varSi":     varArg(VarSi),
	"varGsy":    varArg(VarGsy),
	"varLsy":    varArg(VarLsy),
	"varRsy":    varArg(VarRsy),
	"varIfn":    varArg(VarIfn),
	"varCp":     varArg(VarCp),
	"varLn":     varArg(VarLn),
	"varCn":     varArg(VarCn),
	"lmVersion": func(sr *Stream, _ GenMode, _ []Element) Element { return LmVersion(sr) },
	"lmDate":    func(sr *Stream, _ GenMode, _ []Element) Element { return LmDate(sr) },
	"buffer":    func(*Stream, GenMode, []Element) Element { return NewLMBuffer() },
}

// LMExternal is an engine's table of the functions that rules call by name:
// the builtins, and the functions registered with Set.
type LMExternal struct {
	Table map[string]ExtFn
}

func NewLMExternal() *LMExternal {
	return &LMExternal{Table: maps.Clone(builtins)}
}

func (lm *LMExternal) Set(k string, f ExtFn) {
	lm.Table[k] = f
}

func (lm *LMExternal) Call(sr *Stream, m GenMode, f Element, args []Element) Element {
	k := f.ToString()
	if fn, exists := lm.Table[k]; exists {
		// builtins read args[1] unchecked: supply a null for a missing argument
		for len(args) < 2 {
			args = append(args, Null())
		}
		return fn(sr, m, args)
	}
	sr.Engine.printf("external not found: ")
	for _, x := range args {
		sr.Engine.printf("%s ", x.ToString())
	}
	sr.Engine.newline()
	return NewNumber(0)
}
