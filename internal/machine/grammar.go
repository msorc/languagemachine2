package machine

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
)

type rule struct {
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

func newRule(gs Element, pri priority, length, offset int, x, y Element, l, r []Element, i int) *rule {
	return &rule{
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

func copyRule(x *rule, l Element) *rule {
	return &rule{
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

func (r *rule) additional(l Element) *rule {
	return copyRule(r, l)
}

func (r *rule) lhsLen() int {
	return len(r.lhs)
}

func (r *rule) rhsLen() int {
	return len(r.rhs)
}

func (r *rule) newLHS(m GenMode, c contextHolder) GenMode {
	return newLHMode(m, r.lhs, 1, c)
}

func (r *rule) newRHS(m GenMode, c contextHolder, x scopeHolder) GenMode {
	return newRHMode(m, r.rhs, r.offset, c, x)
}

func (r *rule) match(e *Engine) bool {
	return e.match()
}

func (r *rule) dump(w io.Writer) {
	lx := r.lhsEffectiveInitialSymbol.toTrace()
	rx := r.rhsEffectiveInitialSymbol.toTrace()
	_, _ = fmt.Fprintf(w, "line %4d: %16s %16s %16s %4d %4d %4d %4d\n", r.number, r.grammarSymbol.ToString(), rx, lx, r.length, r.offset, len(r.lhs), len(r.rhs))
}

func (r *rule) ToString() string {
	return fmt.Sprintf("pri: %d len: %d", r.priority, r.length)
}

// Each rule belongs to the grammar specified by its grammar symbol
type selector struct {
	grammars map[string]*grammar
}

func newSelector() *selector {
	return &selector{
		grammars: make(map[string]*grammar),
	}
}

func (s *selector) get(grammarKey string) *grammar {
	return s.grammars[grammarKey]
}

func (s *selector) selectGrammar(g Element) *grammar {
	key := g.ToString()
	grammar := s.get(key)
	if grammar != nil {
		return grammar
	}
	newGrammar := newGrammar(g)
	s.grammars[key] = newGrammar
	return newGrammar
}

// selectorState is a copy of a Selector's grammars and their rule groups.
// Groups are never changed in place (see Grammar.Add), so copying the maps
// is enough.
type selectorState struct {
	grammars map[string]*grammar
	rules    map[*grammar]map[rulePair][]*rule
}

func (s *selector) save() selectorState {
	st := selectorState{grammars: maps.Clone(s.grammars), rules: make(map[*grammar]map[rulePair][]*rule, len(s.grammars))}
	for _, g := range s.grammars {
		st.rules[g] = maps.Clone(g.rules)
	}
	return st
}

// restore returns s to the state saved, dropping the grammars and rules
// added since.
func (s *selector) restore(st selectorState) {
	s.grammars = st.grammars
	for g, rules := range st.rules {
		g.rules = rules
	}
}

// Grammar holds the rules of one grammar, filed in groups by the pair
// (initial symbol of the left side, goal on the right side).
type grammar struct {
	symbol Element
	rules  map[rulePair][]*rule
}

// rulePair is the key of a group of rules.
type rulePair struct {
	lhs, rhs Element
}

func newGrammar(g Element) *grammar {
	return &grammar{
		symbol: g,
		rules:  make(map[rulePair][]*rule),
	}
}

// weight is the length of a rule for ordering: the sum of the weights of its
// left side.
func weight(lhs []Element) int {
	var w int
	for _, x := range lhs {
		w += x.weight()
	}
	return w
}

// Add inserts x into its (lhs initial, rhs initial) group, which is kept
// ordered by descending length; a newer rule goes before older rules of the
// same length. The group is reallocated, never changed in place, so a copy
// of the map (see Selector.save) keeps the groups as they were.
func (g *grammar) add(x *rule) {
	k := rulePair{x.lhsEffectiveInitialSymbol, x.rhsEffectiveInitialSymbol}
	rules := g.rules[k]
	i := slices.IndexFunc(rules, func(r *rule) bool { return x.length >= r.length })
	if i < 0 {
		i = len(rules)
	}
	g.rules[k] = slices.Insert(slices.Clip(rules), i, x)
}

func (g *grammar) defineRule(v []Element, n int) {
	grammarSymbol := v[0]
	pri := priority(v[1].toInt())
	offset := v[2].toInt()
	left := v[3].toBody()
	right := v[4].toBody()
	x := left[0]

	x.addRule(g,
		newRule(grammarSymbol,
			pri,
			weight(left),
			offset,
			left[0].token(), right[0].token(),
			left,
			right,
			n))
}

// Get returns the group of rules for the pair (l, r), longest first.
func (g *grammar) get(l, r Element) []*rule {
	return g.rules[rulePair{l, r}]
}

// Dump lists the rules in the order they were defined; the copies a
// lexical class makes of a rule follow in order of their initial symbol.
func (g *grammar) dump(w io.Writer) {
	var rules []*rule
	for _, group := range g.rules {
		rules = append(rules, group...)
	}
	slices.SortFunc(rules, func(a, b *rule) int {
		return cmp.Or(
			cmp.Compare(a.number, b.number),
			cmp.Compare(a.lhsEffectiveInitialSymbol.toTrace(), b.lhsEffectiveInitialSymbol.toTrace()),
			cmp.Compare(a.rhsEffectiveInitialSymbol.toTrace(), b.rhsEffectiveInitialSymbol.toTrace()),
		)
	})
	for _, r := range rules {
		r.dump(w)
	}
}
