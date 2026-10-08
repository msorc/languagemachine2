# Reference traces

Output of the original Language Machine engine (release 0.2.5) for three
grammars from `examples/`, compiled with the current lm2n compiler. Each
file is the complete stdout of a run like

    lm2 -r cats.lm2 -W 40 -t D -i 'the cat likes the dog .\n'

| File | Grammar | Width | Trace | Input |
| --- | --- | --- | --- | --- |
| `cats.diagram.txt` | `web/cats.lm2n` | 40 | `D` | `the cat likes the dog .\n` |
| `cats.diagram-text.txt` | `web/cats.lm2n` | 40 | `d` | same |
| `cats.mismatch-symbols.txt` | `web/cats.lm2n` | — | `ms` | same |
| `fpCalc.diagram.txt` | `basics/fpCalc.lm2n` | 50 | `D` | `/ 307 241\n` |
| `fpCalc.mismatch-symbols.txt` | `basics/fpCalc.lm2n` | — | `ms` | same |
| `rpCalc.diagram.txt` | `basics/rpCalc.lm2n` | 40 | `D` | `0 5 N + 2 * =\n` |

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
| `cats.grammar.txt` | `web/cats.lm2n` | `G` | `the cat likes the dog .\n` |
| `cats.vars.txt` | `web/cats.lm2n` | all variable and scope flags | same |
| `fpCalc.arith-assign.txt` | `basics/fpCalc.lm2n` | ARITHMETIC, ASSIGN | `/ 307 241\n` |
| `expr.apply.txt` | `exprRules` (machine_test.go) | APPLY | `a` |
| `control.arith-relation.txt` | `testdata/control.lm2n` | ARITHMETIC, RELATION | `w2:` |
| `control.assign-loop.txt` | `testdata/control.lm2n` | ASSIGN, LOOP | `f3:` |
| `control.index.txt` | `testdata/control.lm2n` | INDEX | `h2:` |
| `control.vars-each.txt` | `testdata/control.lm2n` | all variable and scope flags | `e1:abc;` |
| `control.vars-foreach.txt` | `testdata/control.lm2n` | all variable and scope flags | `h2:` |

`testdata/symbols.txt` (TestPredefinedSymbols) lists the predefined symbols
with the Go type behind each, and is updated the same way.
