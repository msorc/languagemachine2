package machine

import "fmt"

// rzBuffer is the circular buffer that provides input elements to the
// outermost level on the RHS. It grows up to max, and keeps the symbols read
// so that the engine can backtrack over them.
type rzBuffer struct {
	currentValue []Element
	max          int
	charPosition int
}

func newRZBuffer(v []Element, m int) *rzBuffer {
	return &rzBuffer{
		currentValue: v,
		max:          m,
	}
}

func (r *rzBuffer) setMax(m int) int {
	r.max = m
	return r.max
}

func (r *rzBuffer) getChr(e *Engine, ci int) Element {
	if ci < r.charPosition {
		// the slot has been reused once we have read a full buffer past ci
		if r.charPosition-ci > len(r.currentValue) {
			fail("backtracking overflow: the input buffer (-buffer %d) is too small", r.max)
		}
		return r.currentValue[ci%len(r.currentValue)]
	}
	if ci == r.charPosition {
		if r.charPosition < len(r.currentValue) {
			r.currentValue[r.charPosition%len(r.currentValue)] = e.getInput()
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
		r.currentValue[r.charPosition%len(r.currentValue)] = e.getInput()
		r.charPosition++
		return r.currentValue[(r.charPosition-1)%len(r.currentValue)]
	}
	panic(fmt.Sprintf("backtrack wraparound: position %d is ahead of the input (%d)", ci, r.charPosition))
}
