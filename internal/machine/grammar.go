package machine

import (
	"fmt"
)

const (
	// Rule
	BRACKET = 0x4000000
	PRIMASK = 0x3fffffe
	CXTMASK = 0x3ffffff
)

// Static functions for priority calculations
func BPri(x uint) uint { return x*2 | BRACKET }
func LPri(x uint) uint { return x * 2 }
func RPri(x uint) uint { return x*2 + 1 }
func MPri(x uint) uint { return PRIMASK }

// The circular buffer that provides input elements to the outermost level on the RHS
type RZBuffer struct {
	Cv  []GrammarElement
	Max uint
	Len uint
	Cp  uint
	Ln  uint
}

func NewRZBuffer(v []GrammarElement, m uint) *RZBuffer {
	return &RZBuffer{
		Cv:  v,
		Max: uint(m),
		Len: uint(len(v)),
		Cp:  0,
		Ln:  1,
	}
}

func (r *RZBuffer) SetMax(m uint) uint {
	r.Max = m
	return r.Max
}

func (r *RZBuffer) GetChr(e *Engine, ci uint) GrammarElement {
	if ci < r.Cp {
		return r.Cv[ci%uint(len(r.Cv))]
	}
	if ci == r.Cp {
		if r.Cp < uint(len(r.Cv)) {
			r.Cv[r.Cp%uint(len(r.Cv))] = e.GetInput()
			r.Cp++
			return r.Cv[(r.Cp-1)%uint(len(r.Cv))]
		}
		if uint(len(r.Cv)) < r.Max && uint(len(r.Cv))*2 < r.Max {
			newlen := uint(len(r.Cv)) * 2
			temp := make([]GrammarElement, newlen)
			copy(temp, r.Cv)
			r.Cv = temp
		}
		if !((r.Cp - ci) < uint(len(r.Cv))) {
			panic("BackTrackOverflow")
		}
		r.Cv[(r.Cp-1)%uint(len(r.Cv))] = e.GetInput()
		r.Cp++
		return r.Cv[(r.Cp-1)%uint(len(r.Cv))]
	}
	panic("backTrackWraparound")
	return nil
}

// Stackable input sources
type IStack struct {
	Next  *IStack
	Input GrammarStdio
}

func NewIStack(a *IStack, b GrammarStdio) *IStack {
	return &IStack{
		Next:  a,
		Input: b,
	}
}

type Rule struct {
	Nxt *Rule            // next in list of rules for same group
	Gra GrammarElement   // grammar symbol
	Pri uint             // encoded priority value
	Len uint             // effective length to determine ordering within group
	Off uint             // offset of start position in RHS - 0 or 1
	Lsy GrammarElement   // effective initial symbol on lhs
	Rsy GrammarElement   // effective initial symbol on rhs
	Lhs []GrammarElement // left hand side - pattern to match
	Rhs []GrammarElement // right hand side - pattern substitute
	Txt string           // explanatory or diagnostic text - not currently used
	Num uint             // number of rule in order of creation
}

func NewRule() *Rule {
	return &Rule{}
}

func NewRuleFromElements(g GrammarElement, p, n, k uint, x, y GrammarElement, l, r []GrammarElement, t string, i uint) *Rule {
	return &Rule{
		Gra: g,
		Pri: p,
		Len: n,
		Off: k,
		Lsy: x,
		Rsy: y,
		Lhs: l,
		Rhs: r,
		Txt: t,
		Num: i,
	}
}

func NewRuleFromRule(x *Rule, l GrammarElement) *Rule {
	return &Rule{
		Gra: x.Gra,
		Pri: x.Pri,
		Len: x.Len,
		Off: x.Off,
		Lsy: l,
		Rsy: x.Rsy,
		Lhs: x.Lhs,
		Rhs: x.Rhs,
		Txt: x.Txt,
		Num: x.Num,
	}
}

func (r *Rule) Additional(l GrammarElement) *Rule {
	return NewRuleFromRule(r, l)
}

func (r *Rule) Lhlength() uint {
	return uint(len(r.Lhs))
}

func (r *Rule) Rhlength() uint {
	return uint(len(r.Rhs))
}

func (r *Rule) Newlhs(m GenMode, c EngineStateContext) GenMode {
	return newLHModeFromElement(m, r.Lhs, 1, c)
}

func (r *Rule) Newrhs(m GenMode, c EngineStateContext, x LMScope) GenMode {
	return NewRHModeFromParamsAndScope(m, r.Rhs, r.Off, c, x)
}

func (r *Rule) Match(e *Engine) bool {
	return e.Match()
}

func (r *Rule) Allow(p uint) bool {
	return r.Pri == 0 || r.Pri > (p&PRIMASK)
}

func (r *Rule) Cxtpri(p uint) uint {
	if r.Pri > 0 {
		return r.Pri & CXTMASK
	}
	return p
}

func (r *Rule) Trapri(p uint) {
	fmt.Printf("P: %6d %6d %6d %6d\n", r.Pri, r.Pri&PRIMASK, p, r.Cxtpri(p))
}

func (r *Rule) Dump() {
	lx := r.Lsy.ToTrace()
	rx := r.Rsy.ToTrace()
	fmt.Printf("line %4d: %16s %16s %16s %4d %4d %4d %4d\n", r.Num, r.Gra, rx, lx, r.Len, r.Off, len(r.Lhs), len(r.Rhs))
}

func (r *Rule) ToString() string {
	return "pri: " + string(r.Pri) + " len: " + string(r.Len)
}

// Each rule belongs to the grammar specified by its grammar symbol
type Selector struct {
	Grammars map[string]*Grammar
}

func NewSelector() *Selector {
	return &Selector{
		Grammars: make(map[string]*Grammar),
	}
}

func (s *Selector) Get(g GrammarElement) *Grammar {
	key := g.ToString()
	if val, exists := s.Grammars[key]; exists {
		return val
	}
	return nil
}

func (s *Selector) Select(g GrammarElement) *Grammar {
	key := g.ToString()
	if val, exists := s.Grammars[key]; exists {
		return val
	}
	newGrammar := NewGrammar(g)
	s.Grammars[key] = newGrammar
	return newGrammar
}

// --- grammar;
type Dict struct {
	Ascii      [256]GrammarElement
	Characters map[rune]GrammarElement
	Symbols    map[string]GrammarElement
	Integers   map[uint]GrammarElement
}

func NewDict() *Dict {
	return &Dict{
		Characters: make(map[rune]GrammarElement),
		Symbols:    make(map[string]GrammarElement),
		Integers:   make(map[uint]GrammarElement),
	}
}

func (d *Dict) GetByString(x string) GrammarElement {
	if val, exists := d.Symbols[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) GetByRune(x rune) GrammarElement {
	if int(x) < len(d.Ascii) {
		return d.Ascii[x]
	}
	if val, exists := d.Characters[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) GetByInt(x uint) GrammarElement {
	if val, exists := d.Integers[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) ToRune(x string) rune {
	//+
	return rune(x[0])
}

func (d *Dict) UniqueR(x rune) GrammarElement {
	if int(x) < len(d.Ascii) {
		if y := d.Ascii[x]; y != nil {
			return y
		}
		d.Ascii[x] = NewChr(x)
		return d.Ascii[x]
	}
	if val, exists := d.Characters[x]; exists {
		return val
	}
	d.Characters[x] = NewChr(x)
	return d.Characters[x]
}

func (d *Dict) UniqueE(x GrammarElement) GrammarElement {
	c := x.ToString()
	if val, exists := d.Symbols[c]; exists {
		return val
	}
	d.Symbols[c] = x
	return x
}

type Grammar struct {
	Sy        GrammarElement
	Counter   uint
	RuleTable map[uint]*Rule
	Rules     map[GrammarElement]map[GrammarElement]*Rule
	Ssy       Predef
	Dummy     *Rule
}

func NewGrammar(g GrammarElement) *Grammar {
	return &Grammar{
		Sy:        g,
		Dummy:     NewRule(),
		RuleTable: make(map[uint]*Rule),
		Rules:     make(map[GrammarElement]map[GrammarElement]*Rule),
	}
}

func (g *Grammar) Weight(lhs []GrammarElement) uint {
	var w uint = 0
	for _, x := range lhs {
		w += x.Weight()
	}
	return uint(w)
}

func (g *Grammar) Add(x *Rule) {
	za := make(map[GrammarElement]*Rule)
	l := x.Lsy
	r := x.Rsy
	t := g.Get(l, r)
	p := g.Dummy
	q := t
	s := x

	g.RuleTable[g.Counter] = x
	g.Counter++

	for q = t; q != nil && x.Len < q.Len; {
		s = t
		p = q
		q = q.Nxt
	}

	p.Nxt = x
	x.Nxt = q

	if _, ok := g.Rules[l]; ok {
		g.Rules[l][r] = s
	} else {
		g.Rules[l] = za
		g.Rules[l][r] = s
	}
}

func Priassoc(pri uint) string {
	if pri == 0 {
		return "L"
	} else if pri&BRACKET != 0 {
		return "B"
	} else if pri&1 != 0 {
		return "R"
	}
	return "L"
}

func Privalue(pri uint) uint {
	return (pri & PRIMASK) / 2
}

func (g *Grammar) Define(v []GrammarElement, t string, i uint) {
	ge := v[0]
	p := v[1].ToLong()
	k := v[2].ToLong()
	l := v[3].ToBody()
	r := v[4].ToBody()
	x := l[0]

	x.AddRule(g, NewRuleFromElements(ge, uint(p), g.Weight(l), uint(k), l[0].Token(), r[0].Token(), l, r, t, i))
	// g.Add(NewRule(g, p, g.Weight(l), k, l[0].Token(), r[0].Token(), l, r, t, i))
}

func (g *Grammar) Get(l, r GrammarElement) *Rule {
	if rulesForL, ok := g.Rules[l]; ok {
		if rule, ok := rulesForL[r]; ok {
			return rule
		}
	}
	return nil
}

func (g *Grammar) Dumplist(r *Rule) {
	if r.Nxt != nil {
		g.Dumplist(r.Nxt)
	}
	r.Dump()
}

func (g *Grammar) Dump() {
	for _, x := range g.Rules {
		for _, r := range x {
			g.Dumplist(r)
		}
	}
}
