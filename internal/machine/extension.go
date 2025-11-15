package machine

import (
	"fmt"
	"strings"
)

func lmFormat(sr *Stream, m GenMode, args []Element) Element {
	if len(args) < 2 {
		panic("bad format args")
	}

	av := args[1:]
	var rs strings.Builder

	putc := func(c rune) {
		rs.WriteRune(c)
	}

	doFormat(putc, av)

	return NewSym(rs.String())
}

func doFormat(putc func(rune), args []Element) {
	for _, arg := range args {
		str := arg.ToString()
		for _, r := range str {
			putc(r)
		}
	}
}

type extFn func(*Stream, GenMode, []Element) Element

type LMExternal struct {
	Table map[string]extFn
}

func NewLMExternal() *LMExternal {
	lm := &LMExternal{Table: make(map[string]extFn)}

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
	lm.Set("varSi", func(sr *Stream, m GenMode, args []Element) Element { return VarSi(sr, args[1].ToVar()) })
	lm.Set("varGsy", func(sr *Stream, m GenMode, args []Element) Element { return VarGsy(sr, args[1].ToVar()) })
	lm.Set("varLsy", func(sr *Stream, m GenMode, args []Element) Element { return VarLsy(sr, args[1].ToVar()) })
	lm.Set("varRsy", func(sr *Stream, m GenMode, args []Element) Element { return VarRsy(sr, args[1].ToVar()) })
	lm.Set("varIfn", func(sr *Stream, m GenMode, args []Element) Element { return VarIfn(sr, args[1].ToVar()) })
	lm.Set("varCp", func(sr *Stream, m GenMode, args []Element) Element { return VarCp(sr, args[1].ToVar()) })
	lm.Set("varLn", func(sr *Stream, m GenMode, args []Element) Element { return VarLn(sr, args[1].ToVar()) })
	lm.Set("varCn", func(sr *Stream, m GenMode, args []Element) Element { return VarCn(sr, args[1].ToVar()) })
	lm.Set("lmVersion", func(sr *Stream, m GenMode, args []Element) Element { return LmVersion(sr) })
	lm.Set("lmDate", func(sr *Stream, m GenMode, args []Element) Element { return LmDate(sr) })
	lm.Set("buffer", func(sr *Stream, m GenMode, args []Element) Element { return NewLMBuffer() })

	return lm
}

func (lm *LMExternal) Set(k string, f extFn) {
	lm.Table[k] = f
}

func (lm *LMExternal) Call(sr *Stream, m GenMode, f Element, args []Element) Element {
	k := f.ToString()
	if fn, exists := lm.Table[k]; exists {
		return fn(sr, m, args)
	}
	fmt.Print("external not found: ")
	for _, x := range args {
		fmt.Printf("%s ", x)
	}
	fmt.Println()
	return NewNumber(0)
}
