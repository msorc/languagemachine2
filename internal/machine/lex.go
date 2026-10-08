package machine

import "github.com/msorc/languagemachine2/internal/conv"

// The states of the lexical class parser.
const (
	IN = iota
	C1
	E1
	C2
	RN
)

type Lex struct {
	Symbol
	Table     map[Element]Element
	Inclusive bool
}

func NewLex(x string) *Lex {
	lex := ReSelf(&Lex{Symbol: *NewSymbol(x)})
	lex.Table = make(map[Element]Element)
	lex.Inclusive = true

	return lex
}

func NewLexFromEngine(s string, e *Engine) *Lex {
	l := NewLex(s)

	state := IN
	var prevc int

	// s is "[...]": walk the characters between the brackets
	body := []rune(s)
	if len(body) >= 2 {
		body = body[1 : len(body)-1]
	} else {
		body = nil
	}
	for _, r := range body {
		var x Element
		c := int(r)

		if state == IN {
			state = C1
			if c == '^' {
				l.Inclusive = false
				continue
			}
		}

		switch state {
		case C1:
			if c == '\\' {
				state = E1
			} else {
				x = e.terminalSymbols.UniqueR(rune(c))
				l.Table[x] = x
				prevc = c
				state = C2
			}

		case E1:
			switch c {
			case 'b':
				c = ' '
			case 'n':
				c = '\n'
			case 'r':
				c = '\r'
			case 't':
				c = '\t'
			case 'f':
				c = '\f'
			}
			x = e.terminalSymbols.UniqueR(rune(c))
			l.Table[x] = x
			prevc = c
			state = C2

		case C2:
			switch c {
			case '\\':
				state = E1
			case '-':
				state = RN
			default:
				x = e.terminalSymbols.UniqueR(rune(c))
				l.Table[x] = x
				prevc = c
			}

		case RN:
			for prevc < c {
				x = e.terminalSymbols.UniqueR(rune(c))
				l.Table[x] = x
				c--
			}
			state = C1
		}
	}
	return l
}

func (l *Lex) ToTrace() string {
	return "[" + conv.Encode(l.V[1:len(l.V)-1]) + "]"
}

func (l *Lex) AddRule(g *Grammar, x *Rule) {
	if l.Inclusive {
		for k := range l.Table {
			g.Add(x.Additional(k))
		}
	} else {
		fail("a rule cannot start with the negated lexical class %s", l.ToTrace())
	}
}

func (l *Lex) Match(e *Engine, r Element) bool {
	if r.Token() == l.Self() {
		return e.Matched3E(l.Self(), r, nil)
	}
	if chr, ok := r.(*Chr); ok && (l.Inclusive != (l.Table[chr] == nil)) {
		return e.Matched3E(l.Self(), r, r)
	}
	return e.ResolveE(l.Self(), r)
}
