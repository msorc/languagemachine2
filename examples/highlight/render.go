package highlight

import (
	"encoding/json"
	"html"
	"io"
	"strings"
)

// ansiColors are the SGR parameters of the classes that the built-in rules
// use. A class that is not listed is not coloured.
var ansiColors = map[string]string{
	"keyword":  "35",
	"type":     "36",
	"builtin":  "36",
	"constant": "31",
	"function": "34",
	"variable": "33",
	"string":   "32",
	"regexp":   "32",
	"number":   "31",
	"operator": "1",
	"comment":  "90",
}

// WriteANSI writes src with the spans coloured for a terminal. The colour is
// set again at the start of each line of a span, because pagers reset it at
// a new line.
func WriteANSI(w io.Writer, src string, spans []Span) error {
	var b strings.Builder
	pos := 0
	for _, s := range spans {
		b.WriteString(src[pos:s.Start])
		pos = s.End
		color, ok := ansiColors[s.Class]
		if !ok {
			b.WriteString(src[s.Start:s.End])
			continue
		}
		for line := range strings.SplitAfterSeq(src[s.Start:s.End], "\n") {
			text := strings.TrimSuffix(line, "\n")
			if text != "" {
				b.WriteString("\x1b[" + color + "m" + text + "\x1b[0m")
			}
			b.WriteString(line[len(text):])
		}
	}
	b.WriteString(src[pos:])
	_, err := io.WriteString(w, b.String())
	return err
}

// WriteHTML writes src as a pre element in which each span is a span element
// of the class hl-<class>.
func WriteHTML(w io.Writer, src string, spans []Span) error {
	var b strings.Builder
	b.WriteString(`<pre class="hl">`)
	pos := 0
	for _, s := range spans {
		b.WriteString(html.EscapeString(src[pos:s.Start]))
		b.WriteString(`<span class="hl-` + html.EscapeString(s.Class) + `">`)
		b.WriteString(html.EscapeString(src[s.Start:s.End]))
		b.WriteString(`</span>`)
		pos = s.End
	}
	b.WriteString(html.EscapeString(src[pos:]))
	b.WriteString("</pre>\n")
	_, err := io.WriteString(w, b.String())
	return err
}

// WriteJSON writes the spans as a JSON array of {start, end, class}, with
// byte offsets.
func WriteJSON(w io.Writer, spans []Span) error {
	if spans == nil {
		spans = []Span{}
	}
	return json.NewEncoder(w).Encode(spans)
}
