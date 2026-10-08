# The Language Machine: Reference Notes

These notes are a structured digest of the original Language Machine website by Peri Hankey (https://languagemachine.sourceforge.net, © 2005–2006, documentation under the GNU FDL). The site describes the original D-language implementation (`liblm`, `lm2`, and the `lmn2*` metalanguage compilers). Language Machine 2 (this repository) is a Go port of that engine, so these pages define the behaviour the port is aiming for.

The site was crawled on 2026-09-30 (80 HTML pages). Each section here names the original pages it draws on, so details can be checked against the source.

## Contents

| File | Covers | Main source pages |
| --- | --- | --- |
| [01-overview.md](01-overview.md) | What the machine is, the analytic-vs-generative idea, history, why not BNF/yacc | `index`, `grammar`, `unrestricted`, `paradigm_shift`, `lex_and_yacc_and_all_that`, `faq`, `project` |
| [02-execution-model.md](02-execution-model.md) | The goal `eof`, mismatch, how rules are selected, priorities, contexts, backtracking, variable scope | `guide`, `glossary`, `picturebook`, `special_symbols` |
| [03-lm2n-language.md](03-lm2n-language.md) | The `lm2n` metalanguage: syntax, binding, acquisition, actions, buffers, `each`/`all`, nested rules, lexical conventions | `lm2n2xfe`, `guide`, `output_buffers`, `what_does_that_hyphen_mean_`, `glossary` |
| [04-special-symbols-and-builtins.md](04-special-symbols-and-builtins.md) | Predefined symbols, conversions, the builtin and external function interface | `special_symbols`, `builtin` |
| [05-cli-and-tracing.md](05-cli-and-tracing.md) | The original `lm2` command line, trace flags, mapping to the Go flags | `lm2`, `bottles`, `lambda` |
| [06-lm-diagram.md](06-lm-diagram.md) | How to read the lm-diagram and its trace output | `picturebook`, `lm-diagram`, `leftrecursive`, `fpcalcdiagram` |
| [07-compilation-and-bytecode.md](07-compilation-and-bytecode.md) | The `lm2n` compilers, the bootstrap, and how `lm2n` maps to `.lm2` bytecode, including where the Go loader differs | `lm2n2mbe`, `metalanguage_bootstrap`, `minimal_rule_set_-_lmcat`, `code_produced_by_different_backends` |
| [08-examples-and-recipes.md](08-examples-and-recipes.md) | A catalogue of the worked examples, with their key rules | `documentation` and the example pages |
| [09-glossary.md](09-glossary.md) | Short definitions of the terms | `glossary` |

See also the Go-specific docs one level up: `../technical_overview.md`, `../internal_machine.md` and `../bytecode.md`.

## Differences found between the Go port and the original

These are recorded here because they matter when you load the `.lm2` files the original compilers produce, including the repository's `lm2nbs.lm2`. The details are in [07-compilation-and-bytecode.md](07-compilation-and-bytecode.md#go-loader-compatibility).

- **`e`** (fixed): `lm2n2mbe` emits bare `e` for `each Name`. The loader used to treat `e` as `r` (define rule), which made `lm2nbs.lm2` fail to load. It now builds an `eachRef`.
- **`M:`** (fixed): `lm2n2mbe` emits `M:<n>` for maximal priority. The loader now encodes it as `maximal` (`priMask|bracketBit`, `internal/machine/priority.go`) (see `../bytecode.md` §2.3).
- **`E` and bare `B`** (added): `lm2n2mbe` emits these for `each (expr)` and `all (expr)`, but the original loader could not load them. The Go loader builds `eachX` and `allX`, which take the variable name from the value of the expression.
- **`foreach`** (added): the original runtime registered `f:foreach`, `f:ret`, `f:lamda` and `f:spec` as primitives that do nothing. No compiler emitted them, and neither the release nor the website says what they were for. `foreach` now has syntax, `foreach (K, V; E) B`, which iterates an array in the order its keys were added (see [03-lm2n-language.md](03-lm2n-language.md)). `ret`, `lamda` and `spec` stay as they were: acting on one prints `act: <name>`.
- **`T`:** `lm2n2mbe` has a rule that emits `T` for `top`, but `lm2n2xfe` never produces `top`. The loader rejects `T` with `unsupported opcode`.
- **Loops and rule values** (fixed): the original `lm2n2mbe` compiled `while` without its test, so a `while` loop never ended. `break`, `continue` and `rule (G, P) { … }` compiled to `f:break`, `f:continue` and `f:rule`, which the original runtime did not define, so a rule that used them failed without a message. `internal/lmnsrc/lm2n2mbe.lm2n` now tests the `while` condition and compiles `for` to `f:for`, so that `continue` runs the step. The runtime defines all four primitives (`internal/machine/testdata/control.lm2n` tests them). `internal/lmnsrc/lm2nbs.lm2` is the original compiler and still has the old behaviour, so compile with `bin/lm2n`.
- **`#` comment lines** (fixed): the original loader splits on `#…\n`, so compiled files may start with a `#!` shebang and comment header, as `internal/lmnsrc/lm2nbs.lm2` does. The Go loader used to panic with `bad load format` on these lines. It now skips them.
- **Sample outputs:** the grammars shipped with the original release are in `examples/` (see `examples/README.md`). They give the same output as the original engine. The reference outputs in `samples/`, the lambda translations and outputs, and `lexicalresults` are checked by `internal/machine/examples_test.go`.
- **Lexical priority:** the site says a terminal goal gets a high lexical priority (`-l`), but the original engine never applies it. The Go port follows the engine, so a rule such as whitespace deletion can start while the goal is a terminal (`rpCalc` depends on this).
- **Published diagrams:** the diagrams on the site (`cats`, `fpCalc`, `rpCalc`, `fact2diagram`) were made by earlier versions of the engine and differ in places from what the released engine prints. The Go port matches the released engine, which `TestTraceGolden` checks for the diagram, its text form and the `m`/`s` traces.
- **Faults reported, not fatal** (Go port): malformed bytecode, operands of the wrong kind (`5 | 'q'`, `varSi` of a non-variable, a block where none is given), an empty operand stack and `-dwidth` below 20 are reported as errors or tolerated (`ToInt` of a non-number is 0), where the original could crash. A load that fails leaves the rules as they were.
- **Repeated options** (Go port): every occurrence of an option takes effect, in order (`-input a -input b`); ARITHMETIC tracing is `-trace o`, since `v` is REF in the port.
- **Operator traces** (Go port): the ACT, APPLY, ARITHMETIC, RELATION, ASSIGN, INDEX and LOOP traces name the operator with its `toString` (`ARITHMETIC +`), as the original's `writefln("%s", x)` did.
