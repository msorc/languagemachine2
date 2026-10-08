package machine

import (
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
	"github.com/msorc/languagemachine2/internal/version"
)

// The numeric conversions give 0 for text they cannot parse (like C's
// strtol/strtod), never a nil element.

func octal(s *Stream, x Element) Element {
	return newNumber(LMNumber(conv.ScanOctal(x.ToVal().ToString())))
}

func binary(s *Stream, x Element) Element {
	return newNumber(LMNumber(conv.ScanBinary(x.ToVal().ToString())))
}

func hex(s *Stream, x Element) Element {
	return newNumber(LMNumber(conv.Strtod(x.ToVal().ToString())))
}

func num(s *Stream, x Element) Element {
	return newNumber(LMNumber(conv.Strtod(x.ToVal().ToString())))
}

func usym(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	return s.Engine.userSymbols.uniqueE(newSym(t))
}

func ulsym(s *Stream, x Element) Element {
	t := strings.ToLower(x.ToVal().ToString())
	return s.Engine.userSymbols.uniqueE(newSym(t))
}

func uusym(s *Stream, x Element) Element {
	t := strings.ToUpper(x.ToVal().ToString())
	return s.Engine.userSymbols.uniqueE(newSym(t))
}

func ssym(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	return s.Engine.nonTerminalSymbols.uniqueE(newSym(t))
}

func slsym(s *Stream, x Element) Element {
	t := strings.ToLower(x.ToVal().ToString())
	return s.Engine.nonTerminalSymbols.uniqueE(newSym(t))
}

func susym(s *Stream, x Element) Element {
	t := strings.ToUpper(x.ToVal().ToString())
	return s.Engine.nonTerminalSymbols.uniqueE(newSym(t))
}

func variableSym(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	return s.Engine.varSymbols.uniqueE(newSym(t))
}

func urn(s *Stream, x Element) Element {
	t := conv.EncodeComponent(x.ToVal().ToString())
	return newSym(t)
}

func urd(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	su := decodeURI(t)
	return newSym(su)
}

func lcase(s *Stream, x Element) Element {
	t := strings.ToLower(x.ToVal().ToString())
	return newSym(t)
}

func ucase(s *Stream, x Element) Element {
	t := strings.ToUpper(x.ToVal().ToString())
	return newSym(t)
}

func stripl(s *Stream, x Element) Element {
	t := strings.TrimLeft(x.ToVal().ToString(), " ")
	return newSym(t)
}

func stripr(s *Stream, x Element) Element {
	t := strings.TrimRight(x.ToVal().ToString(), " ")
	return newSym(t)
}

func strip(s *Stream, x Element) Element {
	t := strings.Trim(x.ToVal().ToString(), " ")
	return newSym(t)
}

func toChars(s *Stream, x Element) Element {
	t := x.ToVal().ToString()
	v := make([]Element, 0, len(t))
	for _, r := range t {
		v = append(v, s.Engine.terminalSymbols.uniqueR(r))
	}
	return newChrStr(v)
}

func varSi(s *Stream, v varElement) Element {
	return newNumber(LMNumber(v.stateIndex()))
}

func varGsy(s *Stream, v varElement) Element {
	return v.grammarSymbol()
}

func varLsy(s *Stream, v varElement) Element {
	return v.lhsSymbol()
}

func varRsy(s *Stream, v varElement) Element {
	return v.rhsSymbol()
}

func varIfn(s *Stream, v varElement) Element {
	return newQuote(s.Engine.nonTerminalSymbols.uniqueE(newSym(v.fileName())))
}

func varCp(s *Stream, v varElement) Element {
	return newNumber(LMNumber(v.charPos()))
}

func varLn(s *Stream, v varElement) Element {
	return newNumber(LMNumber(v.lineNo()))
}

func varCn(s *Stream, v varElement) Element {
	return newNumber(LMNumber(v.charNo()))
}

func lmVersion(s *Stream) Element {
	return newQuote(s.Engine.nonTerminalSymbols.uniqueE(newSym(version.Version())))
}

func lmDate(s *Stream) Element {
	return newQuote(s.Engine.nonTerminalSymbols.uniqueE(newSym(version.Date())))
}
