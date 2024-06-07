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
	currentValue []MachineElement
	max          uint
	length       uint
	charPosition uint
	lineNumber   uint
}

func NewRZBuffer(v []MachineElement, m uint) *RZBuffer {
	return &RZBuffer{
		currentValue: v,
		max:          uint(m),
		length:       uint(len(v)),
		charPosition: 0,
		lineNumber:   1,
	}
}

func (r *RZBuffer) SetMax(m uint) uint {
	r.max = m
	return r.max
}

func (r *RZBuffer) GetChr(e *Engine, ci uint) MachineElement {
	if ci < r.charPosition {
		return r.currentValue[ci%uint(len(r.currentValue))]
	}
	if ci == r.charPosition {
		if r.charPosition < uint(len(r.currentValue)) {
			r.currentValue[r.charPosition%uint(len(r.currentValue))] = e.GetInput()
			r.charPosition++
			return r.currentValue[(r.charPosition-1)%uint(len(r.currentValue))]
		}
		if uint(len(r.currentValue)) < r.max && uint(len(r.currentValue))*2 < r.max {
			newlen := uint(len(r.currentValue)) * 2
			temp := make([]MachineElement, newlen)
			copy(temp, r.currentValue)
			r.currentValue = temp
		}
		if !((r.charPosition - ci) < uint(len(r.currentValue))) {
			panic("BackTrackOverflow")
		}
		r.currentValue[(r.charPosition-1)%uint(len(r.currentValue))] = e.GetInput()
		r.charPosition++
		return r.currentValue[(r.charPosition-1)%uint(len(r.currentValue))]
	}
	panic("backTrackWraparound")
	return nil
}

// Stackable input sources
type IStack struct {
	next  *IStack
	input GrammarIO
}

func NewIStack(a *IStack, b GrammarIO) *IStack {
	return &IStack{
		next:  a,
		input: b,
	}
}

type Rule struct {
	next                      *Rule            // next in list of rules for same group
	grammarSymbol             MachineElement   // grammar symbol
	priority                  uint             // encoded priority value
	length                    uint             // effective length to determine ordering within group
	offset                    uint             // offset of start position in RHS - 0 or 1
	lhsEffectiveInitialSymbol MachineElement   // effective initial symbol on lhs
	rhsEffectiveInitialSymbol MachineElement   // effective initial symbol on rhs
	lhs                       []MachineElement // left hand side - pattern to match
	rhs                       []MachineElement // right hand side - pattern substitute
	text                      string           // explanatory or diagnostic text - not currently used
	number                    uint             // number of rule in order of creation
}

func NewRule() *Rule {
	return &Rule{}
}

func NewRuleFromElements(g MachineElement, p, n, k uint, x, y MachineElement, l, r []MachineElement, t string, i uint) *Rule {
	return &Rule{
		grammarSymbol:             g,
		priority:                  p,
		length:                    n,
		offset:                    k,
		lhsEffectiveInitialSymbol: x,
		rhsEffectiveInitialSymbol: y,
		lhs:                       l,
		rhs:                       r,
		text:                      t,
		number:                    i,
	}
}

func NewRuleFromRule(x *Rule, l MachineElement) *Rule {
	return &Rule{
		grammarSymbol:             x.grammarSymbol,
		priority:                  x.priority,
		length:                    x.length,
		offset:                    x.offset,
		lhsEffectiveInitialSymbol: l,
		rhsEffectiveInitialSymbol: x.rhsEffectiveInitialSymbol,
		lhs:                       x.lhs,
		rhs:                       x.rhs,
		text:                      x.text,
		number:                    x.number,
	}
}

func (r *Rule) Additional(l MachineElement) *Rule {
	return NewRuleFromRule(r, l)
}

func (r *Rule) Lhlength() uint {
	return uint(len(r.lhs))
}

func (r *Rule) Rhlength() uint {
	return uint(len(r.rhs))
}

func (r *Rule) Newlhs(m GenMode, c EngineStateContext) GenMode {
	return newLHModeFromElement(m, r.lhs, 1, c)
}

func (r *Rule) Newrhs(m GenMode, c EngineStateContext, x LMScope) GenMode {
	return NewRHModeFromParamsAndScope(m, r.rhs, r.offset, c, x)
}

func (r *Rule) Match(e *Engine) bool {
	return e.Match()
}

func (r *Rule) Allow(p uint) bool {
	return r.priority == 0 || r.priority > (p&PRIMASK)
}

func (r *Rule) Cxtpri(p uint) uint {
	if r.priority > 0 {
		return r.priority & CXTMASK
	}
	return p
}

func (r *Rule) Trapri(p uint) {
	fmt.Printf("P: %6d %6d %6d %6d\n", r.priority, r.priority&PRIMASK, p, r.Cxtpri(p))
}

func (r *Rule) Dump() {
	lx := r.lhsEffectiveInitialSymbol.ToTrace()
	rx := r.rhsEffectiveInitialSymbol.ToTrace()
	fmt.Printf("line %4d: %16s %16s %16s %4d %4d %4d %4d\n", r.number, r.grammarSymbol, rx, lx, r.length, r.offset, len(r.lhs), len(r.rhs))
}

func (r *Rule) ToString() string {
	return "pri: " + string(r.priority) + " len: " + string(r.length)
}

// Each rule belongs to the grammar specified by its grammar symbol
type Selector struct {
	grammars map[string]*Grammar
}

func NewSelector() *Selector {
	return &Selector{
		grammars: make(map[string]*Grammar),
	}
}

func (s *Selector) Get(g MachineElement) *Grammar {
	key := g.ToString()
	if val, exists := s.grammars[key]; exists {
		return val
	}
	return nil
}

func (s *Selector) Select(g MachineElement) *Grammar {
	key := g.ToString()
	if val, exists := s.grammars[key]; exists {
		return val
	}
	newGrammar := NewGrammar(g)
	s.grammars[key] = newGrammar
	return newGrammar
}

// --- grammar;
type Dict struct {
	ascii      [256]MachineElement
	characters map[rune]MachineElement
	symbols    map[string]MachineElement
	integers   map[uint]MachineElement
}

func NewDict() *Dict {
	return &Dict{
		characters: make(map[rune]MachineElement),
		symbols:    make(map[string]MachineElement),
		integers:   make(map[uint]MachineElement),
	}
}

func (d *Dict) GetByString(x string) MachineElement {
	if val, exists := d.symbols[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) GetByRune(x rune) MachineElement {
	if int(x) < len(d.ascii) {
		return d.ascii[x]
	}
	if val, exists := d.characters[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) GetByInt(x uint) MachineElement {
	if val, exists := d.integers[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) ToRune(x string) rune {
	//+
	return rune(x[0])
}

func (d *Dict) UniqueR(x rune) MachineElement {
	if int(x) < len(d.ascii) {
		if y := d.ascii[x]; y != nil {
			return y
		}
		d.ascii[x] = NewChr(x)
		return d.ascii[x]
	}
	if val, exists := d.characters[x]; exists {
		return val
	}
	d.characters[x] = NewChr(x)
	return d.characters[x]
}

func (d *Dict) UniqueE(x MachineElement) MachineElement {
	c := x.ToString()
	if val, exists := d.symbols[c]; exists {
		return val
	}
	d.symbols[c] = x
	return x
}

type Grammar struct {
	symbol            MachineElement
	counter           uint
	ruleTable         map[uint]*Rule
	rules             map[MachineElement]map[MachineElement]*Rule
	predefinedSymbols Predef
	dummyRule         *Rule
}

func NewGrammar(g MachineElement) *Grammar {
	return &Grammar{
		symbol:    g,
		dummyRule: NewRule(),
		ruleTable: make(map[uint]*Rule),
		rules:     make(map[MachineElement]map[MachineElement]*Rule),
	}
}

func (g *Grammar) Weight(lhs []MachineElement) uint {
	var w uint = 0
	for _, x := range lhs {
		w += x.Weight()
	}
	return uint(w)
}

func (g *Grammar) Add(x *Rule) {
	za := make(map[MachineElement]*Rule)
	l := x.lhsEffectiveInitialSymbol
	r := x.rhsEffectiveInitialSymbol
	t := g.Get(l, r)
	p := g.dummyRule
	q := t
	s := x

	g.ruleTable[g.counter] = x
	g.counter++

	for q = t; q != nil && x.length < q.length; {
		s = t
		p = q
		q = q.next
	}

	p.next = x
	x.next = q

	if _, ok := g.rules[l]; ok {
		g.rules[l][r] = s
	} else {
		g.rules[l] = za
		g.rules[l][r] = s
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

func (g *Grammar) Define(v []MachineElement, t string, i uint) {
	ge := v[0]
	p := v[1].ToLong()
	k := v[2].ToLong()
	l := v[3].ToBody()
	r := v[4].ToBody()
	x := l[0]

	x.AddRule(g, NewRuleFromElements(ge, uint(p), g.Weight(l), uint(k), l[0].Token(), r[0].Token(), l, r, t, i))
	// g.Add(NewRule(g, p, g.Weight(l), k, l[0].Token(), r[0].Token(), l, r, t, i))
}

func (g *Grammar) Get(l, r MachineElement) *Rule {
	if rulesForL, ok := g.rules[l]; ok {
		if rule, ok := rulesForL[r]; ok {
			return rule
		}
	}
	return nil
}

func (g *Grammar) Dumplist(r *Rule) {
	if r.next != nil {
		g.Dumplist(r.next)
	}
	r.Dump()
}

func (g *Grammar) Dump() {
	for _, x := range g.rules {
		for _, r := range x {
			g.Dumplist(r)
		}
	}
}
