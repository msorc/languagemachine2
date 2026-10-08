package machine

import (
	"bufio"
	"errors"
	"io"
	"os"
	"unicode/utf8"
)

const (
	eofRune = ^int(0)
)

// grammarSystem handles the symbols of an ioSymbol (the input sources and
// the to... conversions). Implementations embed gramSystem for the parts
// they share.
type grammarSystem interface {
	setSymbol(Element) Element
	get() Element
	put(Element)
	match(*Engine, Element, Element) bool
	action()
	finish()
}

type gramSystem struct {
	engine *Engine
	symbol Element
}

func newGramSystem(e *Engine) *gramSystem {
	return &gramSystem{engine: e}
}

func (gs *gramSystem) setSymbol(x Element) Element {
	gs.symbol = x
	return gs.symbol
}

func (gs *gramSystem) get() Element {
	panic("not implemented")
}

func (gs *gramSystem) put(x Element) {
}

// Match of an input source: its own symbol acts, eof finishes it and any
// other symbol is put to it.
func (g *stdinInput) match(e *Engine, l, r Element) bool {
	e.matched(l, r)
	switch {
	case r.token() == g.symbol:
		g.action()
	case r == e.predefinedSymbols.eof:
		g.finish()
	default:
		g.put(r)
	}
	return true
}

func (gs *gramSystem) action() {
}

func (gs *gramSystem) finish() {
}

type Input interface {
	grammarSystem
	getElement(int) Element
	Filename() string
	lineNo() int
	charNo() int
	charPos() int
}

type stdinInput struct {
	gramSystem
	reader     *bufio.Reader
	filename   string
	position   int
	lineNumber int
	charNumber int
	buffer     string
}

// NewStdinInput reads the process's standard input.
func NewStdinInput(e *Engine) Input {
	return newStdinInput(e)
}

func newStdinInput(e *Engine) *stdinInput {
	return &stdinInput{
		gramSystem: *newGramSystem(e),
		filename:   "stdin",
		lineNumber: 1,
	}
}

func (g *stdinInput) getElement(c int) Element {
	if c == eofRune {
		return g.engine.predefinedSymbols.eof
	}
	g.position++
	if c == '\n' {
		g.lineNumber++
		g.charNumber = 0
	} else {
		g.charNumber++
	}
	return g.engine.terminalSymbols.uniqueR(rune(c))
}

func (g *stdinInput) Filename() string {
	return g.filename
}

func (g *stdinInput) lineNo() int {
	return g.lineNumber
}

func (g *stdinInput) charNo() int {
	return g.charNumber
}

func (g *stdinInput) charPos() int {
	return g.position
}

func (g *stdinInput) get() Element {
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
		return g.getElement(eofRune)
	}
	if err != nil {
		fail("cannot read stdin: %v", err)
	}
	return g.getElement(int(c))
}

func (g *stdinInput) put(x Element) {
	_, _ = g.engine.out.WriteString(x.ToString())
}

// NewFileInput reads the whole file, which is then input like a string.
func NewFileInput(e *Engine, filename string) (Input, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	g := newStringInput(e, string(content))
	g.filename = filename
	return g, nil
}

// stringInput is input from a string; it uses the embedded stdinInput
// fields so that Filename, charPos and friends report on this input.
type stringInput struct {
	stdinInput
	offset int // byte offset into buffer
}

// NewStringInput reads the string buffer.
func NewStringInput(e *Engine, buffer string) Input {
	return newStringInput(e, buffer)
}

func newStringInput(e *Engine, buffer string) *stringInput {
	g := &stringInput{stdinInput: *newStdinInput(e)}
	g.filename = "input"
	g.buffer = buffer
	return g
}

func (g *stringInput) get() Element {
	if g.offset < len(g.buffer) {
		c, size := utf8.DecodeRuneInString(g.buffer[g.offset:])
		g.offset += size
		return g.getElement(int(c))
	}
	return g.getElement(eofRune)
}

type ioSymbol struct {
	symbol
	h grammarSystem
}

func newIOSymbol(x string, handler grammarSystem) *ioSymbol {
	iosymbol := reSelf(&ioSymbol{symbol: *newSymbol(x)})
	iosymbol.h = handler
	handler.setSymbol(iosymbol)
	return iosymbol
}

func (i *ioSymbol) match(e *Engine, r Element) bool {
	return i.h.match(e, i, r)
}
