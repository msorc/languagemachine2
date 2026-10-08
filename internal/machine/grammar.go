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
	number                    int       // number of rule in order of creation
}

func NewRuleFromElements(gs Element, pri, length, offset int, x, y Element, l, r []Element, i int) *Rule {
	return &Rule{
		grammarSymbol:             gs,
		priority:                  pri,
		length:                    length,
		offset:                    offset,
		lhsEffectiveInitialSymbol: x,
		rhsEffectiveInitialSymbol: y,
		lhs:                       l,
		rhs:                       r,
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
	return s.grammars[grammarKey]
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

// weight is the length of a rule for ordering: the sum of the weights of its
// left side.
func weight(lhs []Element) int {
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

func (g *Grammar) DefineRule(v []Element, n int) {
	grammarSymbol := v[0]
	priority := v[1].ToInt()
	offset := v[2].ToInt()
	left := v[3].ToBody()
	right := v[4].ToBody()
	x := left[0]

	x.AddRule(g,
		NewRuleFromElements(grammarSymbol,
			priority,
			weight(left),
			offset,
			left[0].Token(), right[0].Token(),
			left,
			right,
			n))
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
