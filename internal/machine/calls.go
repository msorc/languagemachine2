package machine

import (
	"fmt"
	"slices"
	"strings"

	"github.com/msorc/languagemachine2/internal/utils"
)

// Calls lists, sorted and without duplicates, the functions that rules call
// by name: a call f(...) compiles to v:f G f:args ... f:fun. f:args also
// opens an array literal, closed by f:array, so each f:args is paired with
// the f:fun or f:array that closes it. A call through an expression, such
// as a variable's value or a table entry, cannot be resolved from the
// bytecode; those are counted in dynamic.
func Calls(rules string) (names []string, dynamic int, err error) {
	var toks []string
	for _, st := range tokenRE.FindAllString(rules, -1) {
		if len(strings.TrimSpace(st)) != 0 && st[0] != '#' {
			toks = append(toks, st)
		}
	}
	seen := map[string]bool{}
	var open []string // per open f:args, the function name or "" when dynamic
	for i, st := range toks {
		switch st {
		case "f:args":
			name := ""
			if i >= 2 && toks[i-1] == "G" && strings.HasPrefix(toks[i-2], "v:") {
				if name, err = utils.Decode(toks[i-2][2:]); err == nil {
					name, err = utils.Unescape(name)
				}
				if err != nil {
					return nil, 0, err
				}
			}
			open = append(open, name)
		case "f:fun", "f:array":
			if len(open) == 0 {
				return nil, 0, fmt.Errorf("token %d: %s without f:args", i, st)
			}
			name := open[len(open)-1]
			open = open[:len(open)-1]
			switch {
			case st == "f:array":
			case name == "":
				dynamic++
			case !seen[name]:
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	slices.Sort(names)
	return names, dynamic, nil
}

// Builtin reports whether rules can call name without it being registered:
// it is in the default external table.
func Builtin(name string) bool {
	_, ok := NewLMExternal().Table[name]
	return ok
}
