package machine

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
)

type Rule struct {
	grammarSymbol             Element   // grammar symbol
	priority                  priority  // encoded priority value
	length                    int       // effective length to determine ordering within group
	offset                    int       // offset of start position in RHS - 0 or 1
	lhsEffectiveInitialSymbol Element   // effective initial symbol on lhs
	rhsEffectiveInitialSymbol Element   // effective initial symbol on rhs
	lhs                       []Element // left hand side - pattern to match
	rhs                       []Element // right hand side - pattern substitute
	number                    int       // number of rule in order of creation
}

func NewRuleFromElements(gs Element, pri priority, length, offset int, x, y Element, l, r []Element, i int) *Rule {
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

// selectorState is a copy of a Selector's grammars and their rule groups.
// Groups are never changed in place (see Grammar.Add), so copying the maps
// is enough.
type selectorState struct {
	grammars map[string]*Grammar
	rules    map[*Grammar]map[rulePair][]*Rule
}

func (s *Selector) save() selectorState {
	st := selectorState{grammars: maps.Clone(s.grammars), rules: make(map[*Grammar]map[rulePair][]*Rule, len(s.grammars))}
	for _, g := range s.grammars {
		st.rules[g] = maps.Clone(g.rules)
	}
	return st
}

// restore returns s to the state saved, dropping the grammars and rules
// added since.
func (s *Selector) restore(st selectorState) {
	s.grammars = st.grammars
	for g, rules := range st.rules {
		g.rules = rules
	}
}

// Grammar holds the rules of one grammar, filed in groups by the pair
// (initial symbol of the left side, goal on the right side).
type Grammar struct {
	symbol Element
	rules  map[rulePair][]*Rule
}

// rulePair is the key of a group of rules.
type rulePair struct {
	lhs, rhs Element
}

func NewGrammar(g Element) *Grammar {
	return &Grammar{
		symbol: g,
		rules:  make(map[rulePair][]*Rule),
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
// same length. The group is reallocated, never changed in place, so a copy
// of the map (see Selector.save) keeps the groups as they were.
func (g *Grammar) Add(x *Rule) {
	k := rulePair{x.lhsEffectiveInitialSymbol, x.rhsEffectiveInitialSymbol}
	rules := g.rules[k]
	i := slices.IndexFunc(rules, func(r *Rule) bool { return x.length >= r.length })
	if i < 0 {
		i = len(rules)
	}
	g.rules[k] = slices.Insert(slices.Clip(rules), i, x)
}

func (g *Grammar) DefineRule(v []Element, n int) {
	grammarSymbol := v[0]
	pri := priority(v[1].ToInt())
	offset := v[2].ToInt()
	left := v[3].ToBody()
	right := v[4].ToBody()
	x := left[0]

	x.AddRule(g,
		NewRuleFromElements(grammarSymbol,
			pri,
			weight(left),
			offset,
			left[0].Token(), right[0].Token(),
			left,
			right,
			n))
}

// Get returns the group of rules for the pair (l, r), longest first.
func (g *Grammar) Get(l, r Element) []*Rule {
	return g.rules[rulePair{l, r}]
}

// Dump lists the rules in the order they were defined; the copies a
// lexical class makes of a rule follow in order of their initial symbol.
func (g *Grammar) Dump(w io.Writer) {
	var rules []*Rule
	for _, group := range g.rules {
		rules = append(rules, group...)
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
