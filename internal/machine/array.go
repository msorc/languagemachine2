package machine

import "slices"

// AArray is the table behind an array value. Keys keeps the keys in the
// order they were added, which is the order foreach visits them in.
type AArray struct {
	A    map[Element]Element
	Keys []Element
}

func NewAArray() *AArray {
	return &AArray{A: make(map[Element]Element)}
}

func (a *AArray) Set(k, v Element) {
	if _, ok := a.A[k]; !ok {
		a.Keys = append(a.Keys, k)
	}
	a.A[k] = v
}

type LMArray struct {
	GenericElement
	aa *AArray
	sx ScopeHolder
}

func NewLMArray(sr *Stream, s GenMode, z ScopeHolder) *LMArray {
	la := MakeSelf[LMArray]()
	la.aa = NewAArray()
	la.sx = s

	if la.sx == nil {
		panic("sx cannot be nil")
	}

	var v Element
	var i int
	var items []Element // the items, last first

	// cells [k: v] are assigned by key; the other items are counted, then
	// numbered 0..i-1 in order as they are popped (last first)
	sr.Operands().Each(func(o Element) bool {
		if o == sr.Engine.predefinedSymbols.mark {
			return false
		}
		if c, ok := o.(*LMCell); ok {
			la.Assign(sr, c)
		} else {
			i++
		}
		return true
	})
	for !sr.EmptyX() {
		v = sr.Popx()
		if v == sr.Engine.predefinedSymbols.mark {
			break
		}
		items = append(items, v)
		if _, ok := v.(*LMCell); !ok {
			i--
			la.AssignE(sr, i, v)
		}
	}

	// the keys in the order the items were written
	la.aa.Keys = la.aa.Keys[:0]
	seen := make(map[Element]bool)
	for _, item := range slices.Backward(items) {
		var k Element
		if c, ok := item.(*LMCell); ok {
			k = sr.Engine.userSymbols.UniqueE(c.K)
		} else {
			k = sr.Engine.userSymbols.UniqueE(NewNumber(LMNumber(i)))
			i++
		}
		if !seen[k] {
			seen[k] = true
			la.aa.Keys = append(la.aa.Keys, k)
		}
	}

	return la
}

func (la *LMArray) ToVal() Element {
	return la.Self()
}

func (la *LMArray) Act(sr *Stream, s GenMode) GenMode {
	sr.Pushx(la.Self())
	return s
}

// func (la *LMArray) Assign(e *Stream, c *LMCell) MachineElement {
func (la *LMArray) Assign(e *Stream, c Element) Element {
	if c != nil {
		lm, ok := c.(*LMCell)
		if !ok {
			panic("not lmcell")
		}
		la.aa.Set(e.Engine.userSymbols.UniqueE(lm.K), lm.V)
		return lm.V
	}
	return nil
}

func (la *LMArray) AssignE(e *Stream, i int, v Element) Element {
	la.aa.Set(e.Engine.userSymbols.UniqueE(NewNumber(LMNumber(i))), v)
	return v
}

func (la *LMArray) Idxf(y Element) Element {
	return NewARef(la.aa, la.sx.ScopeContextMode().State().engine.userSymbols.UniqueE(y.ToVal()), la.sx)
}

func (la *LMArray) Idtf(y Element) Element {
	return NewARef(la.aa, la.sx.ScopeContextMode().State().engine.userSymbols.UniqueE(y.ToVal()), la.sx)
}

type LMCell struct {
	GenericElement
	K Element
	V Element
}

func NewLMCell(y, z Element) *LMCell {
	el := MakeSelf[LMCell]()
	el.K = y
	el.V = z
	return el
}

func (lc *LMCell) ToString() string {
	return "LMCell:" + lc.K.ToString() + lc.V.ToString()
}
