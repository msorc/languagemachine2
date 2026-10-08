package machine

import (
	"fmt"
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

type LMExternal struct {
	Table map[string]ExtFn
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

func NewLMExternal() *LMExternal {
	lm := &LMExternal{Table: make(map[string]ExtFn)}

	lm.Set("octal", func(sr *Stream, m GenMode, args []Element) Element { return Octal(sr, args[1]) })
	lm.Set("binary", func(sr *Stream, m GenMode, args []Element) Element { return Binary(sr, args[1]) })
	lm.Set("hex", func(sr *Stream, m GenMode, args []Element) Element { return Hex(sr, args[1]) })
	lm.Set("num", func(sr *Stream, m GenMode, args []Element) Element { return Num(sr, args[1]) })
	lm.Set("usym", func(sr *Stream, m GenMode, args []Element) Element { return Usym(sr, args[1]) })
	lm.Set("ulsym", func(sr *Stream, m GenMode, args []Element) Element { return Ulsym(sr, args[1]) })
	lm.Set("uusym", func(sr *Stream, m GenMode, args []Element) Element { return Uusym(sr, args[1]) })
	lm.Set("ssym", func(sr *Stream, m GenMode, args []Element) Element { return Ssym(sr, args[1]) })
	lm.Set("slsym", func(sr *Stream, m GenMode, args []Element) Element { return Slsym(sr, args[1]) })
	lm.Set("susym", func(sr *Stream, m GenMode, args []Element) Element { return Susym(sr, args[1]) })
	lm.Set("variable", func(sr *Stream, m GenMode, args []Element) Element { return Variable(sr, args[1]) })
	lm.Set("urn", func(sr *Stream, m GenMode, args []Element) Element { return Urn(sr, args[1]) })
	lm.Set("urd", func(sr *Stream, m GenMode, args []Element) Element { return Urd(sr, args[1]) })
	lm.Set("lcase", func(sr *Stream, m GenMode, args []Element) Element { return Lcase(sr, args[1]) })
	lm.Set("ucase", func(sr *Stream, m GenMode, args []Element) Element { return Ucase(sr, args[1]) })
	lm.Set("stripl", func(sr *Stream, m GenMode, args []Element) Element { return Stripl(sr, args[1]) })
	lm.Set("stripr", func(sr *Stream, m GenMode, args []Element) Element { return Stripr(sr, args[1]) })
	lm.Set("strip", func(sr *Stream, m GenMode, args []Element) Element { return Strip(sr, args[1]) })
	lm.Set("format", lmFormat)
	lm.Set("trOn", func(sr *Stream, m GenMode, args []Element) Element { return sr.Engine.SetTrace(args) })
	lm.Set("trOff", func(sr *Stream, m GenMode, args []Element) Element { return sr.Engine.UnsetTrace(args) })
	lm.Set("include", func(sr *Stream, m GenMode, args []Element) Element { return sr.Engine.Include(args) })
	lm.Set("use", func(sr *Stream, m GenMode, args []Element) Element {
		return sr.Engine.SetMachineElements(args)
	})
	lm.Set("toChars", func(sr *Stream, m GenMode, args []Element) Element { return ToChars(sr, args[1].ToVal()) })
	lm.Set("varSi", varArg(VarSi))
	lm.Set("varGsy", varArg(VarGsy))
	lm.Set("varLsy", varArg(VarLsy))
	lm.Set("varRsy", varArg(VarRsy))
	lm.Set("varIfn", varArg(VarIfn))
	lm.Set("varCp", varArg(VarCp))
	lm.Set("varLn", varArg(VarLn))
	lm.Set("varCn", varArg(VarCn))
	lm.Set("lmVersion", func(sr *Stream, m GenMode, args []Element) Element { return LmVersion(sr) })
	lm.Set("lmDate", func(sr *Stream, m GenMode, args []Element) Element { return LmDate(sr) })
	lm.Set("buffer", func(sr *Stream, m GenMode, args []Element) Element { return NewLMBuffer() })

	return lm
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
