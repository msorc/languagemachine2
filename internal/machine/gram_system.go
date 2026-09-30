package machine

import (
	"bufio"
	"fmt"
	"io"
	"languagemachine2/internal/utils"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	EOF = ^int(0)
)

func read(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("error reading file: `%s`", filename))
	}
	return string(content)
}

type GrammarSystem interface {
	SetSelf(GrammarSystem)
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
	self   GrammarSystem // outermost value, so Action/Put/Finish dispatch to the embedding type
}

func NewGramSystemFromEngine(e *Engine) *GramSystem {
	return &GramSystem{engine: e}
}

func (gs *GramSystem) SetSelf(x GrammarSystem) {
	gs.self = x
}

// Self returns the embedding handler; Go embedding gives no virtual dispatch,
// so calls to overridable methods must go through it.
func (gs *GramSystem) Self() GrammarSystem {
	if gs.self != nil {
		return gs.self
	}
	return gs
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

func (gs *GramSystem) Match(e *Engine, l, r Element) bool {
	e.Matched2E(l, r)
	if r.Token() == gs.symbol {
		gs.Self().Action()
	} else if r == e.predefinedSymbols.eof {
		gs.Self().Finish()
	} else {
		gs.Self().Put(r)
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
	Buffer() string
}

type GramStdio struct {
	GramSystem
	reader     *bufio.Reader
	writer     io.Writer
	filename   string
	position   int
	lineNumber int
	charNumber int
	buffer     string
}

func NewGramStdio() *GramStdio {
	gs := &GramStdio{
		filename:   "stdin",
		lineNumber: 1,
	}

	return gs
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

func (g *GramStdio) SetSymbol(x Element) Element {
	g.symbol = x
	return g.symbol
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

func (g *GramStdio) Buffer() string {
	return g.buffer
}

func (g *GramStdio) Get() Element {
	// the reader must persist between calls, or buffered input is lost
	if g.reader == nil {
		g.reader = bufio.NewReader(os.Stdin)
	}
	c, _, err := g.reader.ReadRune()
	if err == io.EOF {
		return g.GetElement(EOF)
	}
	if err != nil {
		panic(fmt.Sprintf("error reading stdin: %v", err))
	}
	return g.GetElement(int(c))
}

func (g *GramStdio) Put(x Element) {
	w := g.writer
	if w == nil {
		w = os.Stdout
	}
	if _, err := fmt.Fprintf(w, "%s", x.ToString()); err != nil {
		panic(err)
	}
}

func (g *GramStdio) Match(e *Engine, l, r Element) bool {
	panic("not implemented")
}

// GramInputFile reads a whole file; it uses the embedded GramStdio fields so
// that Filename, CharPos and friends report on this input.
type GramInputFile struct {
	GramStdio
	offset int // byte offset into buffer
}

func NewGramInputFile(e *Engine, filename string) *GramInputFile {
	g := &GramInputFile{GramStdio: *NewGramStdioFromEngine(e)}
	g.filename = filename
	g.buffer = read(filename)
	return g
}

func (g *GramInputFile) Get() Element {
	if g.offset < len(g.buffer) {
		c, size := utf8.DecodeRuneInString(g.buffer[g.offset:])
		g.offset += size
		return g.GetElement(int(c))
	}
	return g.GetElement(EOF)
}

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

type ToConvert struct {
	GramSystem
}

func NewToConvertFromEngine(e *Engine) *ToConvert {
	return &ToConvert{GramSystem: *NewGramSystemFromEngine(e)}
}

func (tc *ToConvert) Match(e *Engine, l, r Element) bool {
	e.Matched2E(l, nil)
	tc.Self().Action()
	tc.engine.lhsStream.ClearX()
	return true
}

func (tc *ToConvert) ToRow() []Element {
	return tc.engine.lhsStream.Operands().ToSlice()
}

// ToRowF converts the grabbed material to a string, applies f, and returns
// the result as character symbols.
func (tc *ToConvert) ToRowF(f func(string) string) []Element {
	s := f(tc.ToString())
	v := make([]Element, 0, len(s))
	for _, se := range s {
		v = append(v, tc.engine.terminalSymbols.UniqueR(se))
	}
	return v
}

func (tc *ToConvert) ToRowR(f func(string) string) []Element {
	return tc.ToRowF(f)
}

func (tc *ToConvert) ToString() string {
	var b strings.Builder
	for _, e := range tc.engine.lhsStream.Operands().ToSlice() {
		b.WriteString(e.ToString())
	}
	return b.String()
}

func (tc *ToConvert) OctalNumber() Element {
	s := tc.ToString()
	var n int
	if _, err := fmt.Sscanf(s, "%o", &n); err != nil {
		return NewErrSym(err.Error())
	}
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) BinaryNumber() Element {
	s := tc.ToString()
	var n int64
	for i, b := len(s), int64(1); i > 0; {
		i--
		if s[i] == '1' {
			n += b
		}
		b *= 2
	}
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) HexNumber() Element {
	s := strings.TrimPrefix(strings.TrimPrefix(tc.ToString(), "0x"), "0X")
	n, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return NewErrSym(err.Error())
	}
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) ToNumber() Element {
	s := tc.ToString()
	n := utils.Strtod(s)
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) Count() int {
	return tc.engine.lhsStream.Countx()
}

func (tc *ToConvert) Dump() {
	tc.engine.lhsStream.DumpXPlain()
}

func (tc *ToConvert) Action() {
	tc.Finish()
}

type ToQuote struct {
	ToConvert
}

func NewToQuoteFromEngine(e *Engine) *ToQuote {
	return &ToQuote{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToQuote) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewQuote(t.engine.nonTerminalSymbols.UniqueE(NewSym(t.ToString()))))
	t.Finish()
}

type ToSym struct {
	ToConvert
}

func NewToSymFromEngine(e *Engine) *ToSym {
	return &ToSym{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToSym) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewSym(t.ToString()))
	t.Finish()
}

type ToLsym struct {
	ToConvert
}

func NewToLsymFromEngine(e *Engine) *ToLsym {
	return &ToLsym{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToLsym) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewSym(strings.ToLower(t.ToString())))
	t.Finish()
}

type ToUsym struct {
	ToConvert
}

func NewToUsymFromEngine(e *Engine) *ToUsym {
	return &ToUsym{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToUsym) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewSym(strings.ToUpper(t.ToString())))
	t.Finish()
}

type ToSys struct {
	ToConvert
}

func NewToSysFromEngine(e *Engine) *ToSys {
	return &ToSys{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToSys) Action() {
	t.engine.rsLastMatchElement = t.engine.nonTerminalSymbols.UniqueE(NewSym(t.ToString()))
	t.Finish()
}

type ToLsys struct {
	ToConvert
}

func NewToLsysFromEngine(e *Engine) *ToLsys {
	return &ToLsys{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToLsys) Action() {
	t.engine.rsLastMatchElement = t.engine.nonTerminalSymbols.UniqueE(NewSym(strings.ToLower(t.ToString())))
	t.Finish()
}

type ToUsys struct {
	ToConvert
}

func NewToUsysFromEngine(e *Engine) *ToUsys {
	return &ToUsys{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToUsys) Action() {
	t.engine.rsLastMatchElement = t.engine.nonTerminalSymbols.UniqueE(NewSym(strings.ToUpper(t.ToString())))
	t.Finish()
}

type ToStr struct {
	ToConvert
}

func NewToStrFromEngine(e *Engine) *ToStr {
	return &ToStr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToStr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRow())
	t.Finish()
}

type ToNum struct {
	ToConvert
}

func NewToNumFromEngine(e *Engine) *ToNum {
	return &ToNum{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToNum) Action() {
	t.engine.rsLastMatchElement = t.ToNumber()
	t.Finish()
}

type ToHex struct {
	ToConvert
}

func NewToHexFromEngine(e *Engine) *ToHex {
	return &ToHex{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToHex) Action() {
	t.engine.rsLastMatchElement = t.HexNumber()
	t.Finish()
}

type ToOct struct {
	ToConvert
}

func NewToOctFromEngine(e *Engine) *ToOct {
	return &ToOct{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToOct) Action() {
	t.engine.rsLastMatchElement = t.OctalNumber()
	t.Finish()
}

type ToBin struct {
	ToConvert
}

func NewToBinFromEngine(e *Engine) *ToBin {
	return &ToBin{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToBin) Action() {
	t.engine.rsLastMatchElement = t.BinaryNumber()
	t.Finish()
}

type ToVar struct {
	ToConvert
}

func NewToVarFromEngine(e *Engine) *ToVar {
	return &ToVar{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToVar) Action() {
	t.engine.rsLastMatchElement = t.engine.varSymbols.UniqueE(NewSym(t.ToString()))
	t.Finish()
}

type ToLstr struct {
	ToConvert
}

func NewToLstrFromEngine(e *Engine) *ToLstr {
	return &ToLstr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToLstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowF(strings.ToLower))
	t.Finish()
}

type ToUstr struct {
	ToConvert
}

func NewToUstrFromEngine(e *Engine) *ToUstr {
	return &ToUstr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToUstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowF(strings.ToUpper))
	t.Finish()
}

type ToUrNstr struct {
	ToConvert
}

func NewToUrNstrFromEngine(e *Engine) *ToUrNstr {
	return &ToUrNstr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToUrNstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowR(utils.Encode))
	t.Finish()
}

type ToUrDstr struct {
	ToConvert
}

func NewToUrDstrFromEngine(e *Engine) *ToUrDstr {
	return &ToUrDstr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToUrDstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowR(utils.Decode))
	t.Finish()
}
