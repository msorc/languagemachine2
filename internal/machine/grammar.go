package machine

import (
	"cmp"
	"fmt"
	"io"
	"slices"
)

const (
	// Rule
	BRACKET = 0x4000000
	PRIMASK = 0x3fffffe
	CXTMASK = 0x3ffffff
)

// The circular buffer that provides input elements to the outermost level on the RHS
type RZBuffer struct {
	currentValue []Element
	max          int
	length       int
	charPosition int
	lineNumber   int
}

func NewRZBuffer(v []Element, m int) *RZBuffer {
	return &RZBuffer{
		currentValue: v,
		max:          m,
		length:       len(v),
		charPosition: 0,
		lineNumber:   1,
	}
}

func (r *RZBuffer) SetMax(m int) int {
	r.max = m
	return r.max
}

func (r *RZBuffer) GetChr(e *Engine, ci int) Element {
	if ci < r.charPosition {
		// the slot has been reused once we have read a full buffer past ci
		if r.charPosition-ci > len(r.currentValue) {
			fail("backtracking overflow: the input buffer (-buffer %d) is too small", r.max)
		}
		return r.currentValue[ci%len(r.currentValue)]
	}
	if ci == r.charPosition {
		if r.charPosition < len(r.currentValue) {
			r.currentValue[r.charPosition%len(r.currentValue)] = e.GetInput()
			r.charPosition++
			return r.currentValue[(r.charPosition-1)%len(r.currentValue)]
		}
		// grow while under the limit; the buffer has not wrapped yet, so a
		// plain copy keeps every position in place
		if len(r.currentValue)*2 <= r.max {
			newlen := len(r.currentValue) * 2
			temp := make([]Element, newlen)
			copy(temp, r.currentValue)
			r.currentValue = temp
		}
		r.currentValue[r.charPosition%len(r.currentValue)] = e.GetInput()
		r.charPosition++
		return r.currentValue[(r.charPosition-1)%len(r.currentValue)]
	}
	panic(fmt.Sprintf("backtrack wraparound: position %d is ahead of the input (%d)", ci, r.charPosition))
}

type Rule struct {
	next                      *Rule     // next in list of rules for same group
	grammarSymbol             Element   // grammar symbol
	priority                  int       // encoded priority value
	length                    int       // effective length to determine ordering within group
	offset                    int       // offset of start position in RHS - 0 or 1
	lhsEffectiveInitialSymbol Element   // effective initial symbol on lhs
	rhsEffectiveInitialSymbol Element   // effective initial symbol on rhs
	lhs                       []Element // left hand side - pattern to match
	rhs                       []Element // right hand side - pattern substitute
	text                      string    // explanatory or diagnostic text - not currently used
	number                    int       // number of rule in order of creation
}

func NewRuleFromElements(gs Element, pri, len, offset int, x, y Element, l, r []Element, t string, i int) *Rule {
	return &Rule{
		grammarSymbol:             gs,
		priority:                  pri,
		length:                    len,
		offset:                    offset,
		lhsEffectiveInitialSymbol: x,
		rhsEffectiveInitialSymbol: y,
		lhs:                       l,
		rhs:                       r,
		text:                      t,
		number:                    i,
	}
}

func NewRuleFromRule(x *Rule, l Element) *Rule {
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

func (r *Rule) Additional(l Element) *Rule {
	return NewRuleFromRule(r, l)
}

func (r *Rule) Lhlength() int {
	return len(r.lhs)
}

func (r *Rule) Rhlength() int {
	return len(r.rhs)
}

func (r *Rule) Newlhs(m GenMode, c ContextHolder) GenMode {
	return NewLHModeFromElement(m, r.lhs, 1, c)
}

func (r *Rule) Newrhs(m GenMode, c ContextHolder, x ScopeHolder) GenMode {
	return NewRHModeFromParamsAndScope(m, r.rhs, r.offset, c, x)
}

func (r *Rule) Match(e *Engine) bool {
	return e.Match()
}

func (r *Rule) Allow(p int) bool {
	return r.priority == 0 || r.priority > (p&PRIMASK)
}

func (r *Rule) Cxtpri(p int) int {
	if r.priority > 0 {
		return r.priority & CXTMASK
	}
	return p
}

func (r *Rule) Dump(w io.Writer) {
	lx := r.lhsEffectiveInitialSymbol.ToTrace()
	rx := r.rhsEffectiveInitialSymbol.ToTrace()
	_, _ = fmt.Fprintf(w, "line %4d: %16s %16s %16s %4d %4d %4d %4d\n", r.number, r.grammarSymbol.ToString(), rx, lx, r.length, r.offset, len(r.lhs), len(r.rhs))
}

func (r *Rule) ToString() string {
	return fmt.Sprintf("pri: %d len: %d", r.priority, r.length)
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

func (s *Selector) Get(grammarKey string) *Grammar {
	if val, exists := s.grammars[grammarKey]; exists {
		return val
	}
	return nil
}

func (s *Selector) Select(g Element) *Grammar {
	key := g.ToString()
	grammar := s.Get(key)
	if grammar != nil {
		return grammar
	}
	newGrammar := NewGrammar(g)
	s.grammars[key] = newGrammar
	return newGrammar
}

// --- grammar;
type Dict struct {
	ascii      [256]Element
	characters map[rune]Element
	symbols    map[string]Element
	integers   map[int]Element
}

func NewDict() *Dict {
	return &Dict{
		characters: make(map[rune]Element),
		symbols:    make(map[string]Element),
		integers:   make(map[int]Element),
	}
}

func (d *Dict) GetByString(x string) Element {
	if val, exists := d.symbols[x]; exists {
		return val
	}
	return nil
}

func (d *Dict) UniqueR(x rune) Element {
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

func (d *Dict) UniqueE(x Element) Element {
	c := x.ToString()
	if val, exists := d.symbols[c]; exists {
		return val
	}
	d.symbols[c] = x
	return x
}

type Grammar struct {
	symbol Element
	rules  map[Element]map[Element]*Rule
}

func NewGrammar(g Element) *Grammar {
	return &Grammar{
		symbol: g,
		rules:  make(map[Element]map[Element]*Rule),
	}
}

func (g *Grammar) Weight(lhs []Element) int {
	var w int
	for _, x := range lhs {
		w += x.Weight()
	}
	return w
}

// Add inserts x into its (lhs initial, rhs initial) group, which is kept
// ordered by descending length; a newer rule goes before older rules of the
// same length.
func (g *Grammar) Add(x *Rule) {
	l := x.lhsEffectiveInitialSymbol
	r := x.rhsEffectiveInitialSymbol
	head := g.Get(l, r)

	if head == nil || x.length >= head.length {
		x.next = head
		head = x
	} else {
		p := head
		for p.next != nil && x.length < p.next.length {
			p = p.next
		}
		x.next = p.next
		p.next = x
	}

	if _, ok := g.rules[l]; !ok {
		g.rules[l] = make(map[Element]*Rule)
	}
	g.rules[l][r] = head
}

func (g *Grammar) DefineRule(v []Element, t string, n int) {
	grammarSymbol := v[0]
	priority := v[1].ToInt()
	offset := v[2].ToInt()
	left := v[3].ToBody()
	right := v[4].ToBody()
	x := left[0]

	x.AddRule(g,
		NewRuleFromElements(grammarSymbol,
			priority,
			g.Weight(left),
			offset,
			left[0].Token(), right[0].Token(),
			left,
			right,
			t, n))
}

func (g *Grammar) Get(l, r Element) *Rule {
	if rulesForL, ok := g.rules[l]; ok {
		if rule, ok := rulesForL[r]; ok {
			return rule
		}
	}
	return nil
}

// Dump lists the rules in the order they were defined; the copies a
// lexical class makes of a rule follow in order of their initial symbol.
func (g *Grammar) Dump(w io.Writer) {
	var rules []*Rule
	for _, x := range g.rules {
		for _, r := range x {
			for ; r != nil; r = r.next {
				rules = append(rules, r)
			}
		}
	}
	slices.SortFunc(rules, func(a, b *Rule) int {
		return cmp.Or(
			cmp.Compare(a.number, b.number),
			cmp.Compare(a.lhsEffectiveInitialSymbol.ToTrace(), b.lhsEffectiveInitialSymbol.ToTrace()),
			cmp.Compare(a.rhsEffectiveInitialSymbol.ToTrace(), b.rhsEffectiveInitialSymbol.ToTrace()),
		)
	})
	for _, r := range rules {
		r.Dump(w)
	}
}
