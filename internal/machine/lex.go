package machine

import "github.com/msorc/languagemachine2/internal/conv"

// The states of the lexical class parser.
const (
	lexIn = iota
	lexC1
	lexE1
	lexC2
	lexRN
)

type lex struct {
	symbol
	table     map[Element]Element
	inclusive bool
}

func allocLex(x string) *lex {
	lex := reSelf(&lex{symbol: *newSymbol(x)})
	lex.table = make(map[Element]Element)
	lex.inclusive = true

	return lex
}

func newLex(s string, e *Engine) *lex {
	l := allocLex(s)

	state := lexIn
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

		if state == lexIn {
			state = lexC1
			if c == '^' {
				l.inclusive = false
				continue
			}
		}

		switch state {
		case lexC1:
			if c == '\\' {
				state = lexE1
			} else {
				x = e.terminalSymbols.uniqueR(rune(c))
				l.table[x] = x
				prevc = c
				state = lexC2
			}

		case lexE1:
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
			x = e.terminalSymbols.uniqueR(rune(c))
			l.table[x] = x
			prevc = c
			state = lexC2

		case lexC2:
			switch c {
			case '\\':
				state = lexE1
			case '-':
				state = lexRN
			default:
				x = e.terminalSymbols.uniqueR(rune(c))
				l.table[x] = x
				prevc = c
			}

		case lexRN:
			for prevc < c {
				x = e.terminalSymbols.uniqueR(rune(c))
				l.table[x] = x
				c--
			}
			state = lexC1
		}
	}
	return l
}

func (l *lex) toTrace() string {
	return "[" + conv.Encode(l.v[1:len(l.v)-1]) + "]"
}

func (l *lex) addRule(g *grammar, x *rule) {
	if l.inclusive {
		for k := range l.table {
			g.add(x.additional(k))
		}
	} else {
		fail("a rule cannot start with the negated lexical class %s", l.toTrace())
	}
}

func (l *lex) match(e *Engine, r Element) bool {
	if r.token() == l.self() {
		return e.matchedWith(l.self(), r, nil)
	}
	if chr, ok := r.(*chr); ok && (l.inclusive != (l.table[chr] == nil)) {
		return e.matchedWith(l.self(), r, r)
	}
	return e.resolve(l.self(), r)
}
