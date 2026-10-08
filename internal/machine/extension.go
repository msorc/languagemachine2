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
	return newSym(doFormat(args[1:]))
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
func varArg(f func(*Stream, varElement) Element) ExtFn {
	return func(sr *Stream, _ GenMode, args []Element) Element {
		v := args[1].toVar()
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
	"octal":     arg1(octal),
	"binary":    arg1(binary),
	"hex":       arg1(hex),
	"num":       arg1(num),
	"usym":      arg1(usym),
	"ulsym":     arg1(ulsym),
	"uusym":     arg1(uusym),
	"ssym":      arg1(ssym),
	"slsym":     arg1(slsym),
	"susym":     arg1(susym),
	"variable":  arg1(variableSym),
	"urn":       arg1(urn),
	"urd":       arg1(urd),
	"lcase":     arg1(lcase),
	"ucase":     arg1(ucase),
	"stripl":    arg1(stripl),
	"stripr":    arg1(stripr),
	"strip":     arg1(strip),
	"toChars":   arg1(toChars),
	"format":    lmFormat,
	"trOn":      engineFn((*Engine).setTrace),
	"trOff":     engineFn((*Engine).unsetTrace),
	"include":   engineFn((*Engine).include),
	"use":       engineFn((*Engine).setMachineElements),
	"varSi":     varArg(varSi),
	"varGsy":    varArg(varGsy),
	"varLsy":    varArg(varLsy),
	"varRsy":    varArg(varRsy),
	"varIfn":    varArg(varIfn),
	"varCp":     varArg(varCp),
	"varLn":     varArg(varLn),
	"varCn":     varArg(varCn),
	"lmVersion": func(sr *Stream, _ GenMode, _ []Element) Element { return lmVersion(sr) },
	"lmDate":    func(sr *Stream, _ GenMode, _ []Element) Element { return lmDate(sr) },
	"buffer":    func(*Stream, GenMode, []Element) Element { return newBufferValue() },
}

// External is an engine's table of the functions that rules call by name:
// the builtins, and the functions registered with Set.
type External struct {
	table map[string]ExtFn
}

func newExternal() *External {
	return &External{table: maps.Clone(builtins)}
}

func (lm *External) Set(k string, f ExtFn) {
	lm.table[k] = f
}

func (lm *External) call(sr *Stream, m GenMode, f Element, args []Element) Element {
	k := f.ToString()
	if fn, exists := lm.table[k]; exists {
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
	return newNumber(0)
}
