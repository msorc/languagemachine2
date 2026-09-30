# The Language Machine: Reference Notes

These notes are a structured digest of the original Language Machine website by Peri Hankey (https://languagemachine.sourceforge.net, © 2005–2006, documentation under the GNU FDL). The site describes the original D-language implementation (`liblm`, `lm`, and the `lmn2*` metalanguage compilers). Language Machine 2 (this repository) is a Go port of that engine, so these pages define the behaviour the port is aiming for.

The site was crawled on 2026-09-30 (80 HTML pages). Each section here names the original pages it draws on, so details can be checked against the source.

## Contents

| File | Covers | Main source pages |
| --- | --- | --- |
| [01-overview.md](01-overview.md) | What the machine is, the analytic-vs-generative idea, history, why not BNF/yacc | `index`, `grammar`, `unrestricted`, `paradigm_shift`, `lex_and_yacc_and_all_that`, `faq`, `project` |
| [02-execution-model.md](02-execution-model.md) | The goal `eof`, mismatch, how rules are selected, priorities, contexts, backtracking, variable scope | `guide`, `glossary`, `picturebook`, `special_symbols` |
| [03-lmn-language.md](03-lmn-language.md) | The `lmn` metalanguage: syntax, binding, acquisition, actions, buffers, `each`/`all`, nested rules, lexical conventions | `lmn2xfe`, `guide`, `output_buffers`, `what_does_that_hyphen_mean_`, `glossary` |
| [04-special-symbols-and-builtins.md](04-special-symbols-and-builtins.md) | Predefined symbols, conversions, the builtin and external function interface | `special_symbols`, `builtin` |
| [05-cli-and-tracing.md](05-cli-and-tracing.md) | The original `lm` command line, trace flags, mapping to the Go flags | `lm`, `bottles`, `lambda` |
| [06-lm-diagram.md](06-lm-diagram.md) | How to read the lm-diagram and its trace output | `picturebook`, `lm-diagram`, `leftrecursive`, `fpcalcdiagram` |
| [07-compilation-and-bytecode.md](07-compilation-and-bytecode.md) | The `lmn` compilers, the bootstrap, and how `lmn` maps to `.lm` bytecode, including where the Go loader differs | `lmn2mbe`, `metalanguage_bootstrap`, `minimal_rule_set_-_lmcat`, `code_produced_by_different_backends` |
| [08-examples-and-recipes.md](08-examples-and-recipes.md) | A catalogue of the worked examples, with their key rules | `documentation` and the example pages |
| [09-glossary.md](09-glossary.md) | Short definitions of the terms | `glossary` |

See also the Go-specific docs one level up: `../technical_overview.md`, `../internal_machine.md` and `../bytecode.md`.

## Differences found between the Go port and the original

These are recorded here because they matter when you load the `.lm` files the original compilers produce, including the repository's `lmnbs.lm`. The details are in [07-compilation-and-bytecode.md](07-compilation-and-bytecode.md#go-loader-compatibility).

- **`e`** (fixed): `lmn2mbe` emits bare `e` for `each Name`. The loader used to treat `e` as `r` (define rule), which made `lmnbs.lm` fail to load. It now builds an `EachRef`.
- **`M:`** (fixed): `lmn2mbe` emits `M:<n>` for maximal priority. The loader now encodes it as `PRIMASK|BRACKET` (see `../bytecode.md` §2.3).
- **`E`, bare `B`, and `T`:** `lmn2mbe` emits these for `each(expr)`, `all(expr)` and `top`. The runtime does not implement them, and the loader rejects them with `unsupported opcode` (it used to skip them silently).
- **`#` comment lines** (fixed): the original loader splits on `#…\n`, so compiled files may start with a `#!` shebang and comment header, as `examples/lmn/lmnbs.lm` does. The Go loader used to panic with `bad load format` on these lines. It now skips them.
- **Sample outputs:** the grammars shipped with lm-0.2.5 are in `examples/` (see `examples/README.md`). `samples/flatten` and `samples/reorder` do not reproduce their reference outputs (`TestSamplesGolden` skips them for now). `testing/lexical` and `testing/lexicalbuffer` panic in `GenericElement.ToNumber` (`not implemented`).
