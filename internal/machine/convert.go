package machine

import (
	"strings"

	"github.com/msorc/languagemachine2/internal/conv"
)

type ToConvert struct {
	GramSystem
}

func NewToConvertFromEngine(e *Engine) *ToConvert {
	return ReSelf(&ToConvert{GramSystem: *NewGramSystemFromEngine(e)})
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
	return NewNumber(LMNumber(conv.ScanOctal(tc.ToString())))
}

func (tc *ToConvert) BinaryNumber() Element {
	return NewNumber(LMNumber(conv.ScanBinary(tc.ToString())))
}

// HexNumber converts the grabbed hex digits, which come without their 0x.
func (tc *ToConvert) HexNumber() Element {
	return NewNumber(LMNumber(conv.Strtod("0x" + tc.ToString())))
}

func (tc *ToConvert) ToNumber() Element {
	s := tc.ToString()
	n := conv.Strtod(s)
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
	return ReSelf(&ToQuote{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToQuote) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewQuote(t.engine.nonTerminalSymbols.UniqueE(NewSym(t.ToString()))))
	t.Finish()
}

type ToSym struct {
	ToConvert
}

func NewToSymFromEngine(e *Engine) *ToSym {
	return ReSelf(&ToSym{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToSym) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewSym(t.ToString()))
	t.Finish()
}

type ToLsym struct {
	ToConvert
}

func NewToLsymFromEngine(e *Engine) *ToLsym {
	return ReSelf(&ToLsym{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToLsym) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewSym(strings.ToLower(t.ToString())))
	t.Finish()
}

type ToUsym struct {
	ToConvert
}

func NewToUsymFromEngine(e *Engine) *ToUsym {
	return ReSelf(&ToUsym{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToUsym) Action() {
	t.engine.rsLastMatchElement = t.engine.userSymbols.UniqueE(NewSym(strings.ToUpper(t.ToString())))
	t.Finish()
}

type ToSys struct {
	ToConvert
}

func NewToSysFromEngine(e *Engine) *ToSys {
	return ReSelf(&ToSys{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToSys) Action() {
	t.engine.rsLastMatchElement = t.engine.nonTerminalSymbols.UniqueE(NewSym(t.ToString()))
	t.Finish()
}

type ToLsys struct {
	ToConvert
}

func NewToLsysFromEngine(e *Engine) *ToLsys {
	return ReSelf(&ToLsys{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToLsys) Action() {
	t.engine.rsLastMatchElement = t.engine.nonTerminalSymbols.UniqueE(NewSym(strings.ToLower(t.ToString())))
	t.Finish()
}

type ToUsys struct {
	ToConvert
}

func NewToUsysFromEngine(e *Engine) *ToUsys {
	return ReSelf(&ToUsys{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToUsys) Action() {
	t.engine.rsLastMatchElement = t.engine.nonTerminalSymbols.UniqueE(NewSym(strings.ToUpper(t.ToString())))
	t.Finish()
}

type ToStr struct {
	ToConvert
}

func NewToStrFromEngine(e *Engine) *ToStr {
	return ReSelf(&ToStr{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToStr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRow())
	t.Finish()
}

type ToNum struct {
	ToConvert
}

func NewToNumFromEngine(e *Engine) *ToNum {
	return ReSelf(&ToNum{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToNum) Action() {
	t.engine.rsLastMatchElement = t.ToNumber()
	t.Finish()
}

type ToHex struct {
	ToConvert
}

func NewToHexFromEngine(e *Engine) *ToHex {
	return ReSelf(&ToHex{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToHex) Action() {
	t.engine.rsLastMatchElement = t.HexNumber()
	t.Finish()
}

type ToOct struct {
	ToConvert
}

func NewToOctFromEngine(e *Engine) *ToOct {
	return ReSelf(&ToOct{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToOct) Action() {
	t.engine.rsLastMatchElement = t.OctalNumber()
	t.Finish()
}

type ToBin struct {
	ToConvert
}

func NewToBinFromEngine(e *Engine) *ToBin {
	return ReSelf(&ToBin{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToBin) Action() {
	t.engine.rsLastMatchElement = t.BinaryNumber()
	t.Finish()
}

type ToVar struct {
	ToConvert
}

func NewToVarFromEngine(e *Engine) *ToVar {
	return ReSelf(&ToVar{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToVar) Action() {
	t.engine.rsLastMatchElement = t.engine.varSymbols.UniqueE(NewSym(t.ToString()))
	t.Finish()
}

type ToLstr struct {
	ToConvert
}

func NewToLstrFromEngine(e *Engine) *ToLstr {
	return ReSelf(&ToLstr{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToLstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowF(strings.ToLower))
	t.Finish()
}

type ToUstr struct {
	ToConvert
}

func NewToUstrFromEngine(e *Engine) *ToUstr {
	return ReSelf(&ToUstr{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToUstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowF(strings.ToUpper))
	t.Finish()
}

type ToUrNstr struct {
	ToConvert
}

func NewToUrNstrFromEngine(e *Engine) *ToUrNstr {
	return ReSelf(&ToUrNstr{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToUrNstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowR(conv.EncodeComponent))
	t.Finish()
}

type ToUrDstr struct {
	ToConvert
}

func NewToUrDstrFromEngine(e *Engine) *ToUrDstr {
	return ReSelf(&ToUrDstr{ToConvert: *NewToConvertFromEngine(e)})
}

func (t *ToUrDstr) Action() {
	t.engine.rsLastMatchElement = NewChrStr(t.ToRowR(decodeURI))
	t.Finish()
}
