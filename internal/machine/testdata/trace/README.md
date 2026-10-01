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
