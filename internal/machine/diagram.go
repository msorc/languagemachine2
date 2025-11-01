package machine

import (
	"fmt"
	"strings"
)

type Diagram struct {
	e      *Engine

	lastLd uint
	lastRd uint

	width uint
	side  uint

	q string // for mismatch events
	v string // vertical   bars
	h string // horizontal bars
	t string // horizontal bars
	b string // horizontal bars
}

func multiS(s string, n uint) string {
	return strings.Repeat(s, int(n))
}

func NewDiagram(x *Engine, w uint) *Diagram {
	return &Diagram{
		e:     x,
		width: w,
		side:  w / 2,
		q:     "?",
		v:     "|",
		h:     "-",
		t:     "-",
		b:     ".",
	}
}

func (d *Diagram) Min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func (d *Diagram) DoLhs(ld uint, li uint, x, l, r string) {
	if li > 0 {
		if ld > d.side-8 {
			fmt.Printf("\t%-.*s+%06d%10.10s %-10.10s", d.side-8, multiS(d.v, d.side-8), li, l, r)
		} else {
			fmt.Printf("\t%-.*s%s%-.*s%06d%10.10s %-10.10s", ld-1, multiS(d.v, ld-1), x, d.side-ld-7, multiS(d.h, d.side-ld-7), li, l, r)
		}
	} else {
		if ld > d.side-8 {
			fmt.Printf("\t%-.*s+      %10.10s %-10.10s", d.side-8, multiS(d.v, d.side-8), l, r)
		} else {
			fmt.Printf("\t%-.*s%-.*s%10.10s %-10.10s", ld, multiS(d.v, ld), d.side-ld-1, multiS(" ", d.side-ld-1), l, r)
		}
	}
}

func (d *Diagram) DoRhs(rd uint, ri uint, x string) {
	if ri > 0 {
		if rd > d.side-6 {
			fmt.Printf("%06d+%-*s", ri, int(d.side-7), multiS(d.v, d.side-7))
		} else {
			fmt.Printf("%06d%-.*s%s%-.*s", ri, d.side-rd-7, multiS(d.h, d.side-rd-7), x, rd, multiS(d.v, rd))
		}
	} else {
		if rd > d.side-6 {
			fmt.Printf("      +%-.*s", d.side-7, multiS(d.v, d.side-7))
		} else {
			fmt.Printf("%-.*s%-.*s", d.side-rd, multiS(" ", d.side-rd), rd, multiS(d.v, rd))
		}
	}
	fmt.Println()
}

func (d *Diagram) DoLhq(ld uint, li uint, x, l, r string) {
	if ld > d.side-9 {
		fmt.Printf("\t%-.*s%s      %10.10s %-10.10s", d.side-8, multiS(d.v, d.side-8), x, l, r)
	} else {
		fmt.Printf("\t%-.*s%s%-.*s%10.10s %-10.10s", ld, multiS(d.v, ld), x, d.side-ld-2, multiS(" ", d.side-ld-2), l, r)
	}
}

func (d *Diagram) DoRhq(rd uint, ri uint, x string) {
	if rd > d.side-6 {
		fmt.Printf("      +%-.*s", d.side-7, multiS(x, d.side-7))
	} else {
		fmt.Printf("%-.*s%-.*s", d.side-rd, multiS(" ", d.side-rd), rd, multiS(x, rd))
	}
	fmt.Println()
}

func (d *Diagram) Trace(s string, li, ri uint, ld, rd uint, ls, rs, es string) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		fmt.Printf("\t%4s%8d %8d %8d %8d %16.16s %16.16s %16.16s\n", s, li, ri, ld, rd, ls, rs, es)
	} else {
		switch s {
		case "--": // about to match symbols
			if ld > d.lastLd {
				d.DoLhs(ld, li, ".", "", "")
				d.DoRhs(rd, 0, "|")
			}
			d.DoLhs(ld, 0, "|", ls, rs)
			d.DoRhs(rd, 0, "|")
		case "**": // failed at current level
			d.DoLhq(ld, 0, "-", ls, rs)
			d.DoRhq(rd, 0, d.v)
		case "??": // starting to resolve a mismatch
			d.DoLhq(ld, 0, "?", ls, rs)
			d.DoRhq(rd, 0, d.q)
		}
	}
	d.lastLd = ld
	d.lastRd = rd
}

func (d *Diagram) Repeat(i, li, ri uint, ld, rd uint) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		fmt.Printf("\t%4s%8d %8d %8d %8d\n", "rr", li, ri, ld, rd)
	} else {
		d.DoLhq(ld, 0, "*", "", "")
		d.DoRhq(rd, 0, d.v)
	}
}

func (d *Diagram) EndLevel(s string, li, ri uint, ld, rd uint) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		fmt.Printf("\t%4s%8d %8d %8d %8d\n", s, li, ri, ld, rd)
	} else {
		switch s {
		case "lx":
			d.DoLhs(ld, li, "'", "", "")
			d.DoRhs(rd, 0, "|")
			d.lastLd = ld - 1
		case "rx":
			d.DoLhs(ld, 0, "|", "", "")
			d.DoRhs(rd-1, ri, "'")
			d.lastRd = rd - 1
		}
	}
}

func (d *Diagram) Replace(s string, li, ri uint, ld, rd uint) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		fmt.Printf("\t%4s%8d %8d %8d %8d\n", s, li, ri, ld, rd)
	} else {
		switch s {
		case "z=":
			d.DoLhs(ld, 0, "'", "", "")
			d.DoRhs(rd, li, ".")
			d.lastLd = ld
		case "==":
			d.DoLhs(ld, li, "'", d.t, d.t)
			d.DoRhs(rd, li, ".")
			d.lastLd = ld - 1
		}
		d.lastRd = rd + 1
	}
}
