# Reference traces

Output of the original Language Machine engine (release 0.2.5) for three
grammars from `examples/`, compiled with the current lmn compiler. Each
file is the complete stdout of a run like

    lm -r cats.lm -W 40 -t D -i 'the cat likes the dog .\n'

| File | Grammar | Width | Trace | Input |
| --- | --- | --- | --- | --- |
| `cats.diagram.txt` | `web/cats.lmn` | 40 | `D` | `the cat likes the dog .\n` |
| `cats.diagram-text.txt` | `web/cats.lmn` | 40 | `d` | same |
| `cats.mismatch-symbols.txt` | `web/cats.lmn` | — | `ms` | same |
| `fpCalc.diagram.txt` | `basics/fpCalc.lmn` | 50 | `D` | `/ 307 241\n` |
| `fpCalc.mismatch-symbols.txt` | `basics/fpCalc.lmn` | — | `ms` | same |
| `rpCalc.diagram.txt` | `basics/rpCalc.lmn` | 40 | `D` | `0 5 N + 2 * =\n` |

The original draws the diagram in ASCII; `TestTraceGolden` maps the port's
Unicode box drawing to ASCII before comparing. The diagrams published on the
website (`docs/original/*.wiki`) were made by an earlier engine and differ.

## Traces pinned from the Go port

The remaining files are the Go port's own output, kept so that refactoring
the engine cannot change a trace unnoticed. Addresses in variable traces
are replaced by `ADDR` before comparison. They are regenerated, after a
deliberate change, with

    LM_UPDATE_GOLDEN=1 go test ./internal/machine -run TestTraceGolden

| File | Grammar | Trace | Input |
| --- | --- | --- | --- |
| `cats.grammar.txt` | `web/cats.lmn` | `G` | `the cat likes the dog .\n` |
| `cats.vars.txt` | `web/cats.lmn` | all variable and scope flags | same |
| `fpCalc.arith-assign.txt` | `basics/fpCalc.lmn` | ARITHMETIC, ASSIGN | `/ 307 241\n` |
| `expr.apply.txt` | `exprRules` (machine_test.go) | APPLY | `a` |
| `control.arith-relation.txt` | `testdata/control.lmn` | ARITHMETIC, RELATION | `w2:` |
| `control.assign-loop.txt` | `testdata/control.lmn` | ASSIGN, LOOP | `f3:` |
| `control.index.txt` | `testdata/control.lmn` | INDEX | `h2:` |
| `control.vars-each.txt` | `testdata/control.lmn` | all variable and scope flags | `e1:abc;` |
| `control.vars-foreach.txt` | `testdata/control.lmn` | all variable and scope flags | `h2:` |

`testdata/symbols.txt` (TestPredefinedSymbols) lists the predefined symbols
with the Go type behind each, and is updated the same way.
