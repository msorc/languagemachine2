package machine

import (
	"languagemachine2/internal/summary"
	"strconv"
	"strings"
)

func Octal(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	n, err := strconv.ParseInt(t, 8, 64)
	if err != nil {
		return nil
	}
	return NewNumber(LMNumber(int(n)))
}

func Binary(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	var n int64
	b := int64(1)
	for i := len(t); i > 0; {
		if t[i-1] == '1' {
			n += b
		}
		b *= 2
		i--
	}
	return NewNumber(LMNumber(int(n)))
}

func Hex(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	n, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return nil
	}
	return NewNumber(LMNumber(int(n)))
}

func Num(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	n, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return nil
	}
	return NewNumber(LMNumber(int(n)))
}

// func Quoted(s *Stream, x MachineElement) MachineElement {
//	t := x.ToVal().ToString()
//	return NewQuote(s.Nsy.Unique(NewSym(t)))
// }

func Usym(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	return s.Usy().UniqueE(NewSym(t))
}

func Ulsym(s *Stream, x MachineElement) MachineElement {
	t := strings.ToLower(x.ToVal().ToString())
	return s.Usy().UniqueE(NewSym(t))
}

func Uusym(s *Stream, x MachineElement) MachineElement {
	t := strings.ToUpper(x.ToVal().ToString())
	return s.Usy().UniqueE(NewSym(t))
}

func Ssym(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	return s.Nsy().UniqueE(NewSym(t))
}

func Slsym(s *Stream, x MachineElement) MachineElement {
	t := strings.ToLower(x.ToVal().ToString())
	return s.Nsy().UniqueE(NewSym(t))
}

func Susym(s *Stream, x MachineElement) MachineElement {
	t := strings.ToUpper(x.ToVal().ToString())
	return s.Nsy().UniqueE(NewSym(t))
}

func Variable(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	return s.LM.Vsy.UniqueE(NewSym(t))
}

func Urn(s *Stream, x MachineElement) MachineElement {
	t := UrlEscape(x.ToVal().ToString())
	return NewSym(t)
}

func Urd(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	su := UrlUnescape(t)
	return NewSym(su)
}

func Lcase(s *Stream, x MachineElement) MachineElement {
	t := strings.ToLower(x.ToVal().ToString())
	return NewSym(t)
}

func Ucase(s *Stream, x MachineElement) MachineElement {
	t := strings.ToUpper(x.ToVal().ToString())
	return NewSym(t)
}

func Stripl(s *Stream, x MachineElement) MachineElement {
	t := strings.TrimLeft(x.ToVal().ToString(), " ")
	return NewSym(t)
}

func Stripr(s *Stream, x MachineElement) MachineElement {
	t := strings.TrimRight(x.ToVal().ToString(), " ")
	return NewSym(t)
}

func Strip(s *Stream, x MachineElement) MachineElement {
	t := strings.Trim(x.ToVal().ToString(), " ")
	return NewSym(t)
}

func Buffer(s *Stream) MachineElement {
	return NewLMBuffer()
}

func Include(s *Stream, x MachineElement) MachineElement {
	y := x.ToVal().ToString()
	if y == "-" {
		s.LM.AddInput(NewGramInputFromEngine(s.LM))
	} else {
		s.LM.AddInput(NewGramInputFile(s.LM, y))
	}
	return x
}

func TrOn(s *Stream, x MachineElement) MachineElement {
	y, err := x.ToVal().(*Number)
	if !err {
		return s.Ssy().ZLM
	}
	return NewNumber(LMNumber(s.LM.SetTraceFlag(y.ToUlong())))
}

func TrOff(s *Stream, x MachineElement) MachineElement {
	y, err := x.ToVal().(*Number)
	if !err {
		return s.Ssy().ZLM
	}
	return NewNumber(LMNumber(s.LM.UnsetTraceFlag(y.ToUlong())))
}

func Use(s *Stream, x MachineElement) MachineElement {
	s.LM.SetMachineElement(x.ToVal())
	return x
}

func ToChars(s *Stream, x MachineElement) MachineElement {
	t := x.ToVal().ToString()
	v := make([]MachineElement, len(t))
	n := 0
	for i := 0; i < len(t); {
		//+
		v[n] = s.LM.Tsy.UniqueR(rune(t[i]))
		n++
	}
	v = v[:n]
	return NewChrStr(v)
}

func VarSi(s *Stream, v *Var) MachineElement {
	return NewNumber(LMNumber(v.Si()))
}

func VarGsy(s *Stream, v *Var) MachineElement {
	return v.Gsy()
}

func VarLsy(s *Stream, v *Var) MachineElement {
	return v.Lsy()
}

func VarRsy(s *Stream, v *Var) MachineElement {
	return v.Rsy()
}

func VarIfn(s *Stream, v *Var) MachineElement {
	return NewQuote(s.Nsy().UniqueE(NewSym(v.Ifn())))
}

func VarCp(s *Stream, v *Var) MachineElement {
	return NewNumber(LMNumber(v.Cp()))
}

func VarLn(s *Stream, v *Var) MachineElement {
	return NewNumber(LMNumber(v.Ln()))
}

func VarCn(s *Stream, v *Var) MachineElement {
	return NewNumber(LMNumber(v.Cn()))
}

func LmVersion(s *Stream) MachineElement {
	return NewQuote(s.Nsy().UniqueE(NewSym(summary.VersionString)))
}

func LmDate(s *Stream) MachineElement {
	return NewQuote(s.Nsy().UniqueE(NewSym(summary.DateStamp)))
}
