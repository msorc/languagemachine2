package machine

import (
	"github.com/msorc/languagemachine2/internal/version"
	"github.com/msorc/languagemachine2/internal/conv"
	"strings"
)

// The numeric conversions give 0 for text they cannot parse (like C's
// strtol/strtod), never a nil element.

func Octal(s *Stream, x Element) Element {
	return NewNumber(LMNumber(conv.ScanOctal(x.ToVal().ToString())))
}

func Binary(s *Stream, x Element) Element {
	return NewNumber(LMNumber(conv.ScanBinary(x.ToVal().ToString())))
}

func Hex(s *Stream, x Element) Element {
	return NewNumber(LMNumber(conv.Strtod(x.ToVal().ToString())))
}

func Num(s *Stream, x Element) Element {
	return NewNumber(LMNumber(conv.Strtod(x.ToVal().ToString())))
}

func Usym(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	return s.Engine.userSymbols.UniqueE(NewSym(t))
}

func Ulsym(s *Stream, x Element) Element {
	t := strings.ToLower(x.ToVal().ToString())
	return s.Engine.userSymbols.UniqueE(NewSym(t))
}

func Uusym(s *Stream, x Element) Element {
	t := strings.ToUpper(x.ToVal().ToString())
	return s.Engine.userSymbols.UniqueE(NewSym(t))
}

func Ssym(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	return s.Engine.nonTerminalSymbols.UniqueE(NewSym(t))
}

func Slsym(s *Stream, x Element) Element {
	t := strings.ToLower(x.ToVal().ToString())
	return s.Engine.nonTerminalSymbols.UniqueE(NewSym(t))
}

func Susym(s *Stream, x Element) Element {
	t := strings.ToUpper(x.ToVal().ToString())
	return s.Engine.nonTerminalSymbols.UniqueE(NewSym(t))
}

func Variable(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	return s.Engine.varSymbols.UniqueE(NewSym(t))
}

func Urn(s *Stream, x Element) Element {
	t := conv.EncodeComponent(x.ToVal().ToString())
	return NewSym(t)
}

func Urd(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	su := decodeURI(t)
	return NewSym(su)
}

func Lcase(s *Stream, x Element) Element {
	t := strings.ToLower(x.ToVal().ToString())
	return NewSym(t)
}

func Ucase(s *Stream, x Element) Element {
	t := strings.ToUpper(x.ToVal().ToString())
	return NewSym(t)
}

func Stripl(s *Stream, x Element) Element {
	t := strings.TrimLeft(x.ToVal().ToString(), " ")
	return NewSym(t)
}

func Stripr(s *Stream, x Element) Element {
	t := strings.TrimRight(x.ToVal().ToString(), " ")
	return NewSym(t)
}

func Strip(s *Stream, x Element) Element {
	t := strings.Trim(x.ToVal().ToString(), " ")
	return NewSym(t)
}

func ToChars(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	v := make([]Element, len(t))
	n := 0
	for _, te := range t {
		v[n] = s.Engine.terminalSymbols.UniqueR(te)
		n++
	}
	v = v[:n]
	return NewChrStr(v)
}

func VarSi(s *Stream, v VarElement) Element {
	return NewNumber(LMNumber(v.Si()))
}

func VarGsy(s *Stream, v VarElement) Element {
	return v.Gsy()
}

func VarLsy(s *Stream, v VarElement) Element {
	return v.Lsy()
}

func VarRsy(s *Stream, v VarElement) Element {
	return v.Rsy()
}

func VarIfn(s *Stream, v VarElement) Element {
	return NewQuote(s.Engine.nonTerminalSymbols.UniqueE(NewSym(v.Ifn())))
}

func VarCp(s *Stream, v VarElement) Element {
	return NewNumber(LMNumber(v.Cp()))
}

func VarLn(s *Stream, v VarElement) Element {
	return NewNumber(LMNumber(v.Ln()))
}

func VarCn(s *Stream, v VarElement) Element {
	return NewNumber(LMNumber(v.Cn()))
}

func LmVersion(s *Stream) Element {
	return NewQuote(s.Engine.nonTerminalSymbols.UniqueE(NewSym(version.Version)))
}

func LmDate(s *Stream) Element {
	return NewQuote(s.Engine.nonTerminalSymbols.UniqueE(NewSym(version.Date)))
}
