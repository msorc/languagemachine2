package machine

// Symbol returns the symbol s as a value, for the Go functions that rules
// call (lm.Sym).
func Symbol(s string) Element { return newSym(s) }

// Number returns the number x as a value (lm.Num).
func Number(x float64) Element { return newNumber(LMNumber(x)) }
