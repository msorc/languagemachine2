package machine

import "strings"

type Diagram struct {
	e *Engine

	lastLd int
	side   int

	q string // for mismatch events
	v string // vertical   bars
	h string // horizontal bars
	t string // horizontal bars of a replacement (==) line
}

// bars repeats s n times; a negative count, which a deep nesting in a narrow
// diagram produces, gives nothing (C's %.*s treats it as no limit, which
// could not print less).
func bars(s string, n int) string {
	return strings.Repeat(s, max(n, 0))
}

func NewDiagram(x *Engine, w int) *Diagram {
	return &Diagram{
		e:    x,
		side: w / 2,
		q:    "?",
		v:    "│",
		h:    "─",
		t:    strings.Repeat("-", 21), // printed as %10.10s
	}
}

func (d *Diagram) DoLhs(ld int, li int, x, l, r string) {
	if li > 0 {
		if ld > d.side-8 {
			d.e.printf("\t%s+%06d%10.10s %-10.10s", bars(d.v, d.side-8), li, l, r)
		} else {
			d.e.printf("\t%s%s%s%06d%10.10s %-10.10s", bars(d.v, ld-1), x, bars(d.h, d.side-ld-7), li, l, r)
		}
	} else {
		if ld > d.side-8 {
			d.e.printf("\t%s+      %10.10s %-10.10s", bars(d.v, d.side-8), l, r)
		} else {
			d.e.printf("\t%s%s%10.10s %-10.10s", bars(d.v, ld), bars(" ", d.side-ld-1), l, r)
		}
	}
}

func (d *Diagram) DoRhs(rd int, ri int, x string) {
	if ri > 0 {
		if rd > d.side-6 {
			d.e.printf("%06d+%s", ri, bars(d.v, d.side-7))
		} else {
			d.e.printf("%06d%s%s%s", ri, bars(d.h, d.side-rd-7), x, bars(d.v, rd))
		}
	} else {
		if rd > d.side-6 {
			d.e.printf("      +%s", bars(d.v, d.side-7))
		} else {
			d.e.printf("%s%s", bars(" ", d.side-rd), bars(d.v, rd))
		}
	}
	d.e.newline()
}

func (d *Diagram) DoLhq(ld int, li int, x, l, r string) {
	if ld > d.side-9 {
		d.e.printf("\t%s%s      %10.10s %-10.10s", bars(d.v, d.side-8), x, l, r)
	} else {
		d.e.printf("\t%s%s%s%10.10s %-10.10s", bars(d.v, ld), x, bars(" ", d.side-ld-2), l, r)
	}
}

func (d *Diagram) DoRhq(rd int, ri int, x string) {
	if rd > d.side-6 {
		d.e.printf("      +%s", bars(x, d.side-7))
	} else {
		d.e.printf("%s%s", bars(" ", d.side-rd), bars(x, rd))
	}
	d.e.newline()
}

func (d *Diagram) Trace(s string, li, ri int, ld, rd int, ls, rs, es string) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		d.e.printf("\t%4s%8d %8d %8d %8d %16.16s %16.16s %16.16s\n", s, li, ri, ld, rd, ls, rs, es)
	} else {
		switch s {
		case "--": // about to match symbols
			if ld > d.lastLd {
				d.DoLhs(ld, li, "┌", "", "")
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
}

func (d *Diagram) Repeat(i, li, ri int, ld, rd int) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		d.e.printf("\t%4s%8d %8d %8d %8d\n", "rr", li, ri, ld, rd)
	} else {
		d.DoLhq(ld, 0, "*", "", "")
		d.DoRhq(rd, 0, d.v)
	}
}

func (d *Diagram) EndLevel(s string, li, ri int, ld, rd int) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		d.e.printf("\t%4s%8d %8d %8d %8d\n", s, li, ri, ld, rd)
	} else {
		switch s {
		case "lx":
			d.DoLhs(ld, li, "└", "", "")
			d.DoRhs(rd, 0, "-")
			d.lastLd = ld - 1
		case "rx":
			d.DoLhs(ld, 0, "|", "", "")
			d.DoRhs(rd-1, ri, "┘")
		}
	}
}

func (d *Diagram) Replace(s string, li, ri int, ld, rd int) {
	if d.e.tracer.Flags&DIAGRAMT != 0 {
		d.e.printf("\t%4s%8d %8d %8d %8d\n", s, li, ri, ld, rd)
	} else {
		switch s {
		case "z=":
			d.DoLhs(ld, 0, "'", "", "")
			d.DoRhs(rd, li, "┐")
			d.lastLd = ld
		case "==":
			d.DoLhs(ld, li, "└", d.t, d.t)
			d.DoRhs(rd, li, "┐")
			d.lastLd = ld - 1
		}
	}
}
