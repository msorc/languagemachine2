package machine

import "slices"

// assocArray is the table behind an array value. Keys keeps the keys in the
// order they were added, which is the order foreach visits them in.
type assocArray struct {
	a    map[Element]Element
	keys []Element
}

func newAssocArray() *assocArray {
	return &assocArray{a: make(map[Element]Element)}
}

func (a *assocArray) Set(k, v Element) {
	if _, ok := a.a[k]; !ok {
		a.keys = append(a.keys, k)
	}
	a.a[k] = v
}

type arrayValue struct {
	genericElement
	aa *assocArray
	sx scopeHolder
}

func newArrayValue(sr *Stream, s GenMode, z scopeHolder) *arrayValue {
	la := makeSelf[arrayValue]()
	la.aa = newAssocArray()
	la.sx = s

	if la.sx == nil {
		panic("sx cannot be nil")
	}

	var v Element
	var i int
	var items []Element // the items, last first

	// cells [k: v] are assigned by key; the other items are counted, then
	// numbered 0..i-1 in order as they are popped (last first)
	sr.Operands().each(func(o Element) bool {
		if o == sr.Engine.predefinedSymbols.mark {
			return false
		}
		if c, ok := o.(*cell); ok {
			la.assign(sr, c)
		} else {
			i++
		}
		return true
	})
	for !sr.emptyX() {
		v = sr.popX()
		if v == sr.Engine.predefinedSymbols.mark {
			break
		}
		items = append(items, v)
		if _, ok := v.(*cell); !ok {
			i--
			la.assignE(sr, i, v)
		}
	}

	// the keys in the order the items were written
	la.aa.keys = la.aa.keys[:0]
	seen := make(map[Element]bool)
	for _, item := range slices.Backward(items) {
		var k Element
		if c, ok := item.(*cell); ok {
			k = sr.Engine.userSymbols.uniqueE(c.k)
		} else {
			k = sr.Engine.userSymbols.uniqueE(newNumber(LMNumber(i)))
			i++
		}
		if !seen[k] {
			seen[k] = true
			la.aa.keys = append(la.aa.keys, k)
		}
	}

	return la
}

func (la *arrayValue) act(sr *Stream, s GenMode) GenMode {
	sr.pushX(la.self())
	return s
}

func (la *arrayValue) assign(e *Stream, c Element) Element {
	if c != nil {
		lm, ok := c.(*cell)
		if !ok {
			panic("not lmcell")
		}
		la.aa.Set(e.Engine.userSymbols.uniqueE(lm.k), lm.v)
		return lm.v
	}
	return nil
}

func (la *arrayValue) assignE(e *Stream, i int, v Element) Element {
	la.aa.Set(e.Engine.userSymbols.uniqueE(newNumber(LMNumber(i))), v)
	return v
}

func (la *arrayValue) idxf(sr *Stream, y Element) Element {
	return newArrayRef(la.aa, la.sx.scopeContextMode().State().engine.userSymbols.uniqueE(y.ToVal()), la.sx)
}

func (la *arrayValue) idtf(sr *Stream, y Element) Element {
	return newArrayRef(la.aa, la.sx.scopeContextMode().State().engine.userSymbols.uniqueE(y.ToVal()), la.sx)
}

type cell struct {
	genericElement
	k Element
	v Element
}

func newCell(y, z Element) *cell {
	el := makeSelf[cell]()
	el.k = y
	el.v = z
	return el
}

func (lc *cell) ToString() string {
	return "LMCell:" + lc.k.ToString() + lc.v.ToString()
}
