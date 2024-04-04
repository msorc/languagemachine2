package machine

import (
	"fmt"
	"strings"
)

func lmFormat(sr *Stream, m GenMode, args []GrammarElement) GrammarElement {
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

func doFormat(putc func(rune), args []GrammarElement) {
	for _, arg := range args {
		str := fmt.Sprintf("%s", arg.ToString())
		for _, r := range str {
			putc(r)
		}
	}
}

type extFn func(*Stream, GenMode, []GrammarElement) GrammarElement

type LMExternal struct {
	Table map[string]extFn
}

func NewLMExternal() *LMExternal {
	lm := &LMExternal{Table: make(map[string]extFn)}

	lm.Set("octal", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Octal(sr, args[1]) })
	lm.Set("binary", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Binary(sr, args[1]) })
	lm.Set("hex", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Hex(sr, args[1]) })
	lm.Set("num", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Num(sr, args[1]) })
	lm.Set("usym", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Usym(sr, args[1]) })
	lm.Set("ulsym", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Ulsym(sr, args[1]) })
	lm.Set("uusym", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Uusym(sr, args[1]) })
	lm.Set("ssym", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Ssym(sr, args[1]) })
	lm.Set("slsym", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Slsym(sr, args[1]) })
	lm.Set("susym", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Susym(sr, args[1]) })
	lm.Set("variable", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Variable(sr, args[1]) })
	lm.Set("urn", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Urn(sr, args[1]) })
	lm.Set("urd", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Urd(sr, args[1]) })
	lm.Set("lcase", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Lcase(sr, args[1]) })
	lm.Set("ucase", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Ucase(sr, args[1]) })
	lm.Set("stripl", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Stripl(sr, args[1]) })
	lm.Set("stripr", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Stripr(sr, args[1]) })
	lm.Set("strip", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return Strip(sr, args[1]) })
	lm.Set("format", lmFormat)
	lm.Set("trOn", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return sr.LM.SetTrace(args) })
	lm.Set("trOff", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return sr.LM.UnsetTrace(args) })
	lm.Set("include", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return sr.LM.Include(args) })
	lm.Set("use", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement {
		return sr.LM.SetGrammarElements(args)
	})
	lm.Set("toChars", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return ToChars(sr, args[1].ToVal()) })
	lm.Set("varSi", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarSi(sr, args[1].ToVar()) })
	lm.Set("varGsy", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarGsy(sr, args[1].ToVar()) })
	lm.Set("varLsy", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarLsy(sr, args[1].ToVar()) })
	lm.Set("varRsy", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarRsy(sr, args[1].ToVar()) })
	lm.Set("varIfn", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarIfn(sr, args[1].ToVar()) })
	lm.Set("varCp", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarCp(sr, args[1].ToVar()) })
	lm.Set("varLn", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarLn(sr, args[1].ToVar()) })
	lm.Set("varCn", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return VarCn(sr, args[1].ToVar()) })
	lm.Set("lmVersion", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return LmVersion(sr) })
	lm.Set("lmDate", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return LmDate(sr) })
	lm.Set("buffer", func(sr *Stream, m GenMode, args []GrammarElement) GrammarElement { return NewLMBuffer() })

	return lm
}

func (lm *LMExternal) Set(k string, f extFn) {
	lm.Table[k] = f
}

func (lm *LMExternal) Call(sr *Stream, m GenMode, f GrammarElement, args []GrammarElement) GrammarElement {
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
