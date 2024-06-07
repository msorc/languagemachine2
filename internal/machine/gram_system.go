package machine

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	EOF = ^uint(0)
)

func read(filename string) string {
	content, err := ioutil.ReadFile(filename)
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
	SetSymbol(MachineElement) MachineElement
	Get() MachineElement
	Put(MachineElement)
	Match(*Engine, MachineElement, MachineElement) bool
	Action()
	Finish()
}

type GramSystem struct {
	lmEngine *Engine
	lmSymbol MachineElement
}

func NewGramSystem() *GramSystem {
	return &GramSystem{}
}

func NewGramSystemFromEngine(e *Engine) *GramSystem {
	return &GramSystem{lmEngine: e}
}

func (gs *GramSystem) SetSymbol(x MachineElement) MachineElement {
	gs.lmSymbol = x
	return gs.lmSymbol
}

func (gs *GramSystem) Get() MachineElement {
	panic("not implemented")
	return nil
}

func (gs *GramSystem) Put(x MachineElement) {
}

func (gs *GramSystem) Match(e *Engine, l, r MachineElement) bool {
	e.Matched2E(l, r)
	if r.Token() == gs.lmSymbol {
		gs.Action()
	} else if r == e.predefinedSymbols.EOF {
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

type GrammarStdio interface {
	GrammarSystem
	GetElement(uint) MachineElement
	Filename() string
	LineNo() uint
	CharNo() uint
	CharPos() uint
	Buffer() string
}

type GramStdio struct {
	GramSystem
	theFile io.Writer
	fname   string
	pos     uint
	lno     uint
	cno     uint
	buf     string
}

func NewGramStdio() *GramStdio {
	gs := &GramStdio{
		fname: "stdin",
	}

	return gs
}

func NewGramStdioFromEngine(e *Engine) *GramStdio {
	return &GramStdio{
		GramSystem: *NewGramSystemFromEngine(e),
		//+
		fname: "stdin",
		lno:   1,
	}
}

func (g *GramStdio) GetElement(c uint) MachineElement {
	if c == EOF {
		return g.lmEngine.predefinedSymbols.EOF
	}
	if c == '\n' {
		g.lno++
		g.cno = 0
	} else {
		g.cno++
	}
	return g.lmEngine.terminalSymbols.UniqueR(rune(c))
}

func (g *GramStdio) SetSymbol(x MachineElement) MachineElement {
	g.lmSymbol = x
	return g.lmSymbol
}

func (g *GramStdio) Filename() string {
	return g.fname
}

func (g *GramStdio) LineNo() uint {
	return g.lno
}

func (g *GramStdio) CharNo() uint {
	return g.cno
}

func (g *GramStdio) CharPos() uint {
	return g.pos
}

func (g *GramStdio) Buffer() string {
	return g.buf
}

func (g *GramStdio) Get() MachineElement {
	panic("not implemented")
}

func (g *GramStdio) Put(x MachineElement) {
	fmt.Fprintf(g.theFile, "%s", x.ToString())
}

func (g *GramStdio) Match(e *Engine, l, r MachineElement) bool {
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

func (g *GramInput) Get() MachineElement {
	r := bufio.NewReader(os.Stdin)
	c, _, err := r.ReadRune()
	if err != nil {
		panic("error")
	}
	return g.GetElement(uint(c))
}

type GramInputFile struct {
	GramStdio
	fname string
	buf   string
	pos   uint
}

func NewGramInputFile(e *Engine, filename string) *GramInputFile {
	return &GramInputFile{
		GramStdio: *NewGramStdioFromEngine(e),
		fname:     filename,
		buf:       read(filename),
	}
}

func (g *GramInputFile) Get() MachineElement {
	if g.buf == "" {
		g.buf = read(g.fname)
	}
	if g.pos < uint(len(g.buf)) {
		element := g.GetElement(uint(g.buf[g.pos]))
		g.pos++
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

func (g *GramInputBuffer) Get() MachineElement {
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
	gof.theFile = file
	return gof
}

func (g *GramOutputFile) Put(x MachineElement) {
	fmt.Fprintf(g.theFile, "%s", x.ToString())
}

func (g *GramOutputFile) Match(e *Engine, l, r MachineElement) bool {
	g.Put(r)
	e.Matched2E(l, r)
	return true
}

func (g *GramOutputFile) Finish() {
	// close(g.thefile)
}

type GramOutputBuffer struct {
	GramStdio
	buf string
}

func NewGramOutputBuffer(e *Engine) *GramOutputBuffer {
	return &GramOutputBuffer{
		GramStdio: *NewGramStdioFromEngine(e),
		buf:       "",
	}
}

func (g *GramOutputBuffer) Put(x MachineElement) {
	g.buf += x.ToString()
}

func (g *GramOutputBuffer) Match(e *Engine, l, r MachineElement) bool {
	e.Matched2E(l, r)
	if r.Token() == g.lmSymbol {
		g.Action()
	} else if r == g.lmEngine.predefinedSymbols.EOF {
		g.Finish()
	} else {
		g.Put(r)
	}
	return true
}

func (g *GramOutputBuffer) Finish() {
	g.buf = ""
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

func (tc *ToConvert) Match(e *Engine, l, r MachineElement) bool {
	e.Matched2E(l, nil)
	tc.Action()
	tc.lmEngine.lhsStream.XS = nil
	return true
}

func (tc *ToConvert) ToRow() []MachineElement {
	v := make([]MachineElement, tc.Count())
	x := tc.lmEngine.lhsStream.XS
	for i := len(v); i > 0; x = x.S {
		i--
		v[i] = x.V
	}
	return v
}

func (tc *ToConvert) ToRowF(f func(string) string) []MachineElement {
	x := tc.lmEngine.lhsStream.XS
	s := ""
	for x != nil {
		s += f(x.V.ToString())
		x = x.S
	}
	v := make([]MachineElement, len(s))
	n := 0
	for i := 0; i < len(s); {
		//+
		v[n] = tc.lmEngine.terminalSymbols.UniqueR(rune(s[i]))
		n++
	}
	w := make([]MachineElement, n)
	for i := 0; i < len(w); i++ {
		n--
		w[i] = v[n]
	}
	return w
}

func (tc *ToConvert) ToRowR(f func(string) string) []MachineElement {
	x := tc.lmEngine.lhsStream.XS
	var y *Opnd
	s := ""
	for x != nil {
		y = NewOpnd(y, x.V)
		x = x.S
	}
	for y != nil {
		s += y.V.ToString()
		y = y.S
	}
	s = f(s)
	v := make([]MachineElement, len(s))
	n := 0
	for i := 0; i < len(s); {
		//+
		v[n] = tc.lmEngine.terminalSymbols.UniqueR(rune(s[i]))
		n++
	}
	return v
}

func (tc *ToConvert) ToString() string {
	s := ""
	for x := tc.lmEngine.lhsStream.XS; x != nil; x = x.S {
		s = x.V.ToString() + s
	}
	return s
}

func (tc *ToConvert) OctalNumber() MachineElement {
	s := tc.ToString()
	var n uint
	fmt.Sscanf(s, "%o", &n)
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) BinaryNumber() MachineElement {
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

func (tc *ToConvert) HexNumber() MachineElement {
	s := "0x" + tc.ToString()
	n := strtod(s)
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) ToNumber() MachineElement {
	s := tc.ToString()
	n := strtod(s)
	return NewNumber(LMNumber(n))
}

func (tc *ToConvert) Count() uint64 {
	var n uint64
	for x := tc.lmEngine.lhsStream.XS; x != nil; x = x.S {
		n++
	}
	return n
}

func (tc *ToConvert) Dump() {
	for x := tc.lmEngine.lhsStream.XS; x != nil; x = x.S {
		fmt.Printf("x: %s\n", x.V.ToString())
	}
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.userSymbols.UniqueE(NewQuote(t.lmEngine.nonTerminalSymbols.UniqueE(NewSym(t.ToString()))))
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.userSymbols.UniqueE(NewSym(t.ToString()))
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.userSymbols.UniqueE(NewSym(strings.ToLower(t.ToString())))
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.userSymbols.UniqueE(NewSym(strings.ToUpper(t.ToString())))
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.nonTerminalSymbols.UniqueE(NewSym(t.ToString()))
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.nonTerminalSymbols.UniqueE(NewSym(strings.ToLower(t.ToString())))
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.nonTerminalSymbols.UniqueE(NewSym(strings.ToUpper(t.ToString())))
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
	t.lmEngine.rsLastMatchElement = NewChrStr(t.ToRow())
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
	t.lmEngine.rsLastMatchElement = t.ToNumber()
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
	t.lmEngine.rsLastMatchElement = t.HexNumber()
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
	t.lmEngine.rsLastMatchElement = t.OctalNumber()
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
	t.lmEngine.rsLastMatchElement = t.BinaryNumber()
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
	t.lmEngine.rsLastMatchElement = t.lmEngine.varSymbols.UniqueE(NewSym(t.ToString()))
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
	t.lmEngine.rsLastMatchElement = NewChrStr(t.ToRowF(strings.ToLower))
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
	t.lmEngine.rsLastMatchElement = NewChrStr(t.ToRowF(strings.ToUpper))
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
	t.lmEngine.rsLastMatchElement = NewChrStr(t.ToRowR(UrlEscape))
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
	t.lmEngine.rsLastMatchElement = NewChrStr(t.ToRowR(UrlUnescape))
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
	t.lmEngine.rsLastMatchElement = nil
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
	t.lmEngine.rsLastMatchElement = nil
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
	t.lmEngine.rsLastMatchElement = nil
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
	t.lmEngine.rsLastMatchElement = nil
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
	t.lmEngine.rsLastMatchElement = nil
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
	t.lmEngine.rsLastMatchElement = nil
	t.Finish()
}
