package machine

import (
	"bufio"
	"fmt"
	"io"
	"languagemachine2/internal/utils"
	"os"
	"strconv"
	"strings"
)

const (
	EOF = ^uint(0)
)

func read(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("error reading file: `%s`", filename))
	}
	return string(content)
}

func strtod(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("failed to convert string to float64")
	}
	return value
}

func strtoui(s string) uint {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic("failed to convert string to uint")
	}
	return uint(value)
}

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

func NewGramSystem() *GramSystem {
	return &GramSystem{}
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

func (gs *GramSystem) Match(e *Engine, l, r Element) bool {
	e.Matched2E(l, r)
	if r.Token() == gs.symbol {
		gs.Action()
	} else if r == e.predefinedSymbols.eof {
		gs.Finish()
	} else {
		gs.Put(r)
	}
	return true
}

func (gs *GramSystem) Action() {
}

func (gs *GramSystem) Finish() {
}

type GrammarIO interface {
	GrammarSystem
	GetElement(uint) Element
	Filename() string
	LineNo() uint
	CharNo() uint
	CharPos() uint
	Buffer() string
}

type GramStdio struct {
	GramSystem
	writer     io.Writer
	filename   string
	position   uint
	lineNumber uint
	charNumber uint
	buffer     string
}

func NewGramStdio() *GramStdio {
	gs := &GramStdio{
		filename: "stdin",
	}

	return gs
}

func NewGramStdioFromEngine(e *Engine) *GramStdio {
	return &GramStdio{
		GramSystem: *NewGramSystemFromEngine(e),
		//+
		filename:   "stdin",
		lineNumber: 1,
	}
}

func (g *GramStdio) GetElement(c uint) Element {
	if c == EOF {
		return g.engine.predefinedSymbols.eof
	}
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

func (g *GramStdio) LineNo() uint {
	return g.lineNumber
}

func (g *GramStdio) CharNo() uint {
	return g.charNumber
}

func (g *GramStdio) CharPos() uint {
	return g.position
}

func (g *GramStdio) Buffer() string {
	return g.buffer
}

func (g *GramStdio) Get() Element {
	panic("not implemented")
}

func (g *GramStdio) Put(x Element) {
	if _, err := fmt.Fprintf(g.writer, "%s", x.ToString()); err != nil {
		panic(err)
	}
}

func (g *GramStdio) Match(e *Engine, l, r Element) bool {
	panic("not implemented")
}

type GramInput struct {
	GramStdio
}

func NewGramInputFromEngine(e *Engine) *GramInput {
	return &GramInput{
		GramStdio: *NewGramStdioFromEngine(e),
	}
}

func (g *GramInput) Get() Element {
	r := bufio.NewReader(os.Stdin)
	c, _, err := r.ReadRune()
	if err != nil {
		panic("error")
	}
	return g.GetElement(uint(c))
}

type GramInputFile struct {
	GramStdio
	filename string
	buffer   string
	position uint
}

func NewGramInputFile(e *Engine, filename string) *GramInputFile {
	return &GramInputFile{
		GramStdio: *NewGramStdioFromEngine(e),
		filename:  filename,
		buffer:    read(filename),
	}
}

func (g *GramInputFile) Get() Element {
	if g.buffer == "" {
		g.buffer = read(g.filename)
	}
	if g.position < uint(len(g.buffer)) {
		element := g.GetElement(uint(g.buffer[g.position]))
		g.position++
		return element
	}
	return g.GetElement(EOF)
}

type GramInputBuffer struct {
	GramStdio
	buf string
	pos int
}

func NewGramInputBuffer(e *Engine, buffer string) *GramInputBuffer {
	return &GramInputBuffer{
		GramStdio: *NewGramStdioFromEngine(e),
		buf:       buffer,
	}
}

func (g *GramInputBuffer) Get() Element {
	if g.pos < len(g.buf) {
		element := g.GetElement(uint(g.buf[g.pos]))
		g.pos++
		return element
	}
	return g.GetElement(EOF)
}

type GramOutputFile struct {
	GramStdio
}

func NewGramOutputFile(e *Engine, gramname string, file io.Writer) *GramOutputFile {
	gof := &GramOutputFile{
		GramStdio: *NewGramStdioFromEngine(e),
	}
	gof.writer = file
	return gof
}

func (g *GramOutputFile) Put(x Element) {
	if _, err := fmt.Fprintf(g.writer, "%s", x.ToString()); err != nil {
		panic(err)
	}
}

func (g *GramOutputFile) Match(e *Engine, l, r Element) bool {
	g.Put(r)
	e.Matched2E(l, r)
	return true
}

func (g *GramOutputFile) Finish() {
	// close(g.thefile)
}

type GramOutputBuffer struct {
	GramStdio
	buffer string
}

func NewGramOutputBuffer(e *Engine) *GramOutputBuffer {
	return &GramOutputBuffer{
		GramStdio: *NewGramStdioFromEngine(e),
		buffer:    "",
	}
}

func (g *GramOutputBuffer) Put(x Element) {
	g.buffer += x.ToString()
}

func (g *GramOutputBuffer) Match(e *Engine, l, r Element) bool {
	e.Matched2E(l, r)
	if r.Token() == g.symbol {
		g.Action()
	} else if r == g.engine.predefinedSymbols.eof {
		g.Finish()
	} else {
		g.Put(r)
	}
	return true
}

func (g *GramOutputBuffer) Finish() {
	g.buffer = ""
}

type ToConvert struct {
	GramSystem
}

func NewToConvert() *ToConvert {
	return &ToConvert{}
}

func NewToConvertFromEngine(e *Engine) *ToConvert {
	return &ToConvert{GramSystem: *NewGramSystemFromEngine(e)}
}

func (tc *ToConvert) Match(e *Engine, l, r Element) bool {
	e.Matched2E(l, nil)
	tc.Action()
	tc.engine.lhsStream.ClearX()
	return true
}

func (tc *ToConvert) ToRow() []Element {
	v := make([]Element, tc.Count())
	operands := tc.engine.lhsStream.Operands()
	for i := operands.BackNode(); i != nil; i = i.Prev() {
		v = append(v, i.Value)
	}
	return v
}

func (tc *ToConvert) ToRowF(f func(string) string) []Element {
	s := ""

	operands := tc.engine.lhsStream.Operands()
	operands.Traversal(func(e Element) bool {
		s += f(e.ToString())
		return true
	})

	v := make([]Element, len(s))
	n := 0
	for _, se := range s {
		v[n] = tc.engine.terminalSymbols.UniqueR(se)
		n++
	}
	w := make([]Element, n)
	for i := 0; i < len(w); i++ {
		n--
		w[i] = v[n]
	}
	return w
}

func (tc *ToConvert) ToRowR(f func(string) string) []Element {
	s := ""
	operands := tc.engine.lhsStream.Operands()
	for i := operands.BackNode(); i != nil; i = i.Prev() {
		s += i.Value.ToString()
	}

	s = f(s)
	v := make([]Element, len(s))
	n := 0
	for _, se := range s {
		v[n] = tc.engine.terminalSymbols.UniqueR(se)
		n++
	}
	return v
}

func (tc *ToConvert) ToString() string {
	s := ""
	operands := tc.engine.lhsStream.Operands()
	operands.Traversal(func(e Element) bool {
		s = e.ToString() + s
		return true
	})
	return s
}

func (tc *ToConvert) OctalNumber() Element {
	s := tc.ToString()
	var n uint
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
	s := "0x" + tc.ToString()
	n := strtod(s)
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) ToNumber() Element {
	s := tc.ToString()
	n := strtod(s)
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) Count() uint {
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

func NewToQuote() *ToQuote {
	return &ToQuote{}
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

func NewToSym() *ToSym {
	return &ToSym{}
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

func NewToLsym() *ToLsym {
	return &ToLsym{}
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

func NewToUsym() *ToUsym {
	return &ToUsym{}
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

func NewToSys() *ToSys {
	return &ToSys{}
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

func NewToLsys() *ToLsys {
	return &ToLsys{}
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

func NewToUsys() *ToUsys {
	return &ToUsys{}
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

func NewToStr() *ToStr {
	return &ToStr{}
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

func NewToNum() *ToNum {
	return &ToNum{}
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

func NewToHex() *ToHex {
	return &ToHex{}
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

func NewToOct() *ToOct {
	return &ToOct{}
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

func NewToBin() *ToBin {
	return &ToBin{}
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

func NewToVar() *ToVar {
	return &ToVar{}
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

func NewToLstr() *ToLstr {
	return &ToLstr{}
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

func NewToUstr() *ToUstr {
	return &ToUstr{}
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

func NewToUrNstr() *ToUrNstr {
	return &ToUrNstr{}
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

func NewToUrDstr() *ToUrDstr {
	return &ToUrDstr{}
}

func NewToUrDstrFromEngine(e *Engine) *ToUrDstr {
	return &ToUrDstr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToUrDstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowR(utils.Decode))
	t.Finish()
}

type ToCsym struct {
	ToConvert
}

func NewToCsym() *ToCsym {
	return &ToCsym{}
}

func NewToCsymFromEngine(e *Engine) *ToCsym {
	return &ToCsym{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToCsym) Action() {
	t.engine.rsLastMatchElement = nil
	t.Finish()
}

type ToNsym struct {
	ToConvert
}

func NewToNsym() *ToNsym {
	return &ToNsym{}
}

func NewToNsymFromEngine(e *Engine) *ToNsym {
	return &ToNsym{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToNsym) Action() {
	t.engine.rsLastMatchElement = nil
	t.Finish()
}

type ToCsys struct {
	ToConvert
}

func NewToCsys() *ToCsys {
	return &ToCsys{}
}

func NewToCsysFromEngine(e *Engine) *ToCsys {
	return &ToCsys{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToCsys) Action() {
	t.engine.rsLastMatchElement = nil
	t.Finish()
}

type ToNsys struct {
	ToConvert
}

func NewToNsys() *ToNsys {
	return &ToNsys{}
}

func NewToNsysFromEngine(e *Engine) *ToNsys {
	return &ToNsys{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToNsys) Action() {
	t.engine.rsLastMatchElement = nil
	t.Finish()
}

type ToCstr struct {
	ToConvert
}

func NewToCstr() *ToCstr {
	return &ToCstr{}
}

func NewToCstrFromEngine(e *Engine) *ToCstr {
	return &ToCstr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToCstr) Action() {
	t.engine.rsLastMatchElement = nil
	t.Finish()
}

type ToNstr struct {
	ToConvert
}

func NewToNstr() *ToNstr {
	return &ToNstr{}
}

func NewToNstrFromEngine(e *Engine) *ToNstr {
	return &ToNstr{ToConvert: *NewToConvertFromEngine(e)}
}

func (t *ToNstr) Action() {
	t.engine.rsLastMatchElement = nil
	t.Finish()
}
