package machine

// priority is the encoded priority of a rule or a context, as the loader
// builds it from the bytecode: the level times 2, plus 1 for a
// right-associative rule, or with the bracket bit for a bracketing one. 0 is
// no priority.
type priority int

const (
	bracketBit priority = 0x4000000 // the rule brackets: it can always start
	priMask    priority = 0x3fffffe // the level, times 2
	cxtMask    priority = 0x3ffffff // what a rule passes on to the context it starts

	// maximal is M:n. The rule can always start (bracket bit), and the
	// context it starts is closed: nothing can nest inside it.
	maximal = priMask | bracketBit
)

func leftPriority(level int) priority    { return priority(level * 2) }
func rightPriority(level int) priority   { return priority(level*2 + 1) }
func bracketPriority(level int) priority { return priority(level*2) | bracketBit }

// level is the level written in the bytecode (L:level).
func (p priority) level() int { return int(p&priMask) / 2 }

// assoc is the association letter that traces show: L, R or B.
func (p priority) assoc() string {
	switch {
	case p == 0:
		return "L"
	case p&bracketBit != 0:
		return "B"
	case p&1 != 0:
		return "R"
	default:
		return "L"
	}
}

// allows reports whether a rule of priority p can start in a context of
// priority ctx: it has no priority, or binds more tightly.
func (p priority) allows(ctx priority) bool {
	return p == 0 || p > ctx&priMask
}

// context is the priority of the context that a rule of priority p starts
// inside a context of priority outer.
func (p priority) context(outer priority) priority {
	if p > 0 {
		return p & cxtMask
	}
	return outer
}

// closed reports whether no rule can start in a context of priority p.
func (p priority) closed() bool { return p == priMask }
