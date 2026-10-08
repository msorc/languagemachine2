package machine

import (
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
)

// converter handles a to... symbol (toStr, toNum, toSym and so on): when the
// symbol is matched, it turns the material grabbed on the left side into a
// value, which becomes the last match for the binding that follows.
type converter struct {
	GramSystem
	convert func(c *converter) Element
}

func newConverter(e *Engine, convert func(c *converter) Element) *converter {
	return &converter{GramSystem: *NewGramSystemFromEngine(e), convert: convert}
}

func (c *converter) Match(e *Engine, l, _ Element) bool {
	e.Matched2E(l, nil)
	e.rsLastMatchElement = c.convert(c)
	e.lhsStream.ClearX()
	return true
}

// row is the grabbed material, oldest first.
func (c *converter) row() []Element {
	return c.engine.lhsStream.Operands().ToSlice()
}

// text is the grabbed material as a string.
func (c *converter) text() string {
	var b strings.Builder
	for _, e := range c.row() {
		b.WriteString(e.ToString())
	}
	return b.String()
}

// chars converts the grabbed material to a string, applies f, and returns the
// result as a string of character symbols.
func (c *converter) chars(f func(string) string) Element {
	s := f(c.text())
	v := make([]Element, 0, len(s))
	for _, r := range s {
		v = append(v, c.engine.terminalSymbols.UniqueR(r))
	}
	return NewChrStr(v)
}

func (c *converter) symbol(d *Dict, f func(string) string) Element {
	return d.UniqueE(NewSym(f(c.text())))
}

func same(s string) string { return s }

// converters are the to... symbols, by name.
var converters = []struct {
	name    string
	convert func(c *converter) Element
}{
	{"toStr", func(c *converter) Element { return NewChrStr(c.row()) }},
	{"toLstr", func(c *converter) Element { return c.chars(strings.ToLower) }},
	{"toUstr", func(c *converter) Element { return c.chars(strings.ToUpper) }},
	{"toQuote", func(c *converter) Element {
		return c.engine.userSymbols.UniqueE(NewQuote(c.engine.nonTerminalSymbols.UniqueE(NewSym(c.text()))))
	}},
	{"toSym", func(c *converter) Element { return c.symbol(c.engine.userSymbols, same) }},
	{"toLsym", func(c *converter) Element { return c.symbol(c.engine.userSymbols, strings.ToLower) }},
	{"toUsym", func(c *converter) Element { return c.symbol(c.engine.userSymbols, strings.ToUpper) }},
	{"toSys", func(c *converter) Element { return c.symbol(c.engine.nonTerminalSymbols, same) }},
	{"toLsys", func(c *converter) Element { return c.symbol(c.engine.nonTerminalSymbols, strings.ToLower) }},
	{"toUsys", func(c *converter) Element { return c.symbol(c.engine.nonTerminalSymbols, strings.ToUpper) }},
	{"toVar", func(c *converter) Element { return c.symbol(c.engine.varSymbols, same) }},
	{"toNum", func(c *converter) Element { return NewNumber(LMNumber(conv.Strtod(c.text()))) }},
	{"toOct", func(c *converter) Element { return NewNumber(LMNumber(conv.ScanOctal(c.text()))) }},
	// the grabbed hex digits come without their 0x
	{"toHex", func(c *converter) Element { return NewNumber(LMNumber(conv.Strtod("0x" + c.text()))) }},
	{"toBin", func(c *converter) Element { return NewNumber(LMNumber(conv.ScanBinary(c.text()))) }},
	{"toUrn", func(c *converter) Element { return c.chars(conv.EncodeComponent) }},
	{"toUrd", func(c *converter) Element { return c.chars(decodeURI) }},
}
