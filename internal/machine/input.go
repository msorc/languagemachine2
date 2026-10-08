package machine

import (
	"bufio"
	"errors"
	"io"
	"os"
	"unicode/utf8"
)

const (
	EOF = ^int(0)
)

// GrammarSystem handles the symbols of an IOSymbol (the input sources and
// the to... conversions). Implementations embed GramSystem for the parts
// they share.
type GrammarSystem interface {
	SetSymbol(Element) Element
	Get() Element
	Put(Element)
	Match(*Engine, Element, Element) bool
	Action()
	Finish()
}

type GramSystem struct {
	engine *Engine
	symbol Element
}

func NewGramSystemFromEngine(e *Engine) *GramSystem {
	return &GramSystem{engine: e}
}

func (gs *GramSystem) SetSymbol(x Element) Element {
	gs.symbol = x
	return gs.symbol
}

func (gs *GramSystem) Get() Element {
	panic("not implemented")
}

func (gs *GramSystem) Put(x Element) {
}

// Match of an input source: its own symbol acts, eof finishes it and any
// other symbol is put to it.
func (g *GramStdio) Match(e *Engine, l, r Element) bool {
	e.Matched2E(l, r)
	switch {
	case r.Token() == g.symbol:
		g.Action()
	case r == e.predefinedSymbols.eof:
		g.Finish()
	default:
		g.Put(r)
	}
	return true
}

func (gs *GramSystem) Action() {
}

func (gs *GramSystem) Finish() {
}

type GrammarIO interface {
	GrammarSystem
	GetElement(int) Element
	Filename() string
	LineNo() int
	CharNo() int
	CharPos() int
}

type GramStdio struct {
	GramSystem
	reader     *bufio.Reader
	filename   string
	position   int
	lineNumber int
	charNumber int
	buffer     string
}

func NewGramStdioFromEngine(e *Engine) *GramStdio {
	return &GramStdio{
		GramSystem: *NewGramSystemFromEngine(e),
		filename:   "stdin",
		lineNumber: 1,
	}
}

func (g *GramStdio) GetElement(c int) Element {
	if c == EOF {
		return g.engine.predefinedSymbols.eof
	}
	g.position++
	if c == '\n' {
		g.lineNumber++
		g.charNumber = 0
	} else {
		g.charNumber++
	}
	return g.engine.terminalSymbols.UniqueR(rune(c))
}

func (g *GramStdio) Filename() string {
	return g.filename
}

func (g *GramStdio) LineNo() int {
	return g.lineNumber
}

func (g *GramStdio) CharNo() int {
	return g.charNumber
}

func (g *GramStdio) CharPos() int {
	return g.position
}

func (g *GramStdio) Get() Element {
	// the reader must persist between calls, or buffered input is lost
	if g.reader == nil {
		g.reader = bufio.NewReader(os.Stdin)
	}
	if g.reader.Buffered() == 0 {
		// about to block: show pending output first (interactive use)
		_ = g.engine.Flush()
	}
	c, _, err := g.reader.ReadRune()
	if errors.Is(err, io.EOF) {
		return g.GetElement(EOF)
	}
	if err != nil {
		fail("cannot read stdin: %v", err)
	}
	return g.GetElement(int(c))
}

func (g *GramStdio) Put(x Element) {
	_, _ = g.engine.out.WriteString(x.ToString())
}

// NewGramInputFile reads the whole file, which is then input like a string.
func NewGramInputFile(e *Engine, filename string) (*GramInputBuffer, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	g := NewGramInputBuffer(e, string(content))
	g.filename = filename
	return g, nil
}

// GramInputBuffer is input from a string; it uses the embedded GramStdio
// fields so that Filename, CharPos and friends report on this input.
type GramInputBuffer struct {
	GramStdio
	offset int // byte offset into buffer
}

func NewGramInputBuffer(e *Engine, buffer string) *GramInputBuffer {
	g := &GramInputBuffer{GramStdio: *NewGramStdioFromEngine(e)}
	g.filename = "input"
	g.buffer = buffer
	return g
}

func (g *GramInputBuffer) Get() Element {
	if g.offset < len(g.buffer) {
		c, size := utf8.DecodeRuneInString(g.buffer[g.offset:])
		g.offset += size
		return g.GetElement(int(c))
	}
	return g.GetElement(EOF)
}

type IOSymbol struct {
	Symbol
	H GrammarSystem
}

func NewIOSymbol(x string, handler GrammarSystem) *IOSymbol {
	iosymbol := ReSelf(&IOSymbol{Symbol: *NewSymbol(x)})
	iosymbol.H = handler
	handler.SetSymbol(iosymbol)
	return iosymbol
}

func (i *IOSymbol) Match(e *Engine, r Element) bool {
	return i.H.Match(e, i, r)
}
