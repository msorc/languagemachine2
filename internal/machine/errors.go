package machine

import "fmt"

// Error is a failure caused by the rules or the input rather than by a bug
// in the machine: a malformed rule file, a missing input file, an exceeded
// limit. The engine raises it with fail, deep inside matching, and
// LoadFromString and Start return it as an ordinary error.
type Error struct {
	Msg string
}

func (e *Error) Error() string { return e.Msg }

func fail(format string, args ...any) {
	panic(&Error{Msg: fmt.Sprintf(format, args...)})
}

// catch turns a panic raised by fail into *err; other panics are bugs and
// are passed on. It must be deferred directly (defer catch(&err)), because
// recover only stops a panic when the deferred function calls it.
func catch(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(*Error); ok {
			*err = e
			return
		}
		panic(r)
	}
}
