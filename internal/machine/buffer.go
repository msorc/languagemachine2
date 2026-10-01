package machine

import "fmt"

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
