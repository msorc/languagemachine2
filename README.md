# Language Machine 2

A Go port of Peri Hankey's [Language Machine](https://languagemachine.sourceforge.net), a toolkit for language and grammar. In a Language Machine grammar, a rule applies when what is expected and what is in the input fail to match. Recognition and substitution interleave on the input stream, so one notation (`lmn`) covers lexing, parsing, translation and evaluation.

The port keeps the original execution semantics and the lm-diagram trace. The grammars and reference outputs from the original release (in `examples/`) produce the same output, and the regression tests check this.

## Build

Requires Go 1.25 or later.

```sh
make build        # -> bin/lm
make test         # regression tests
```

## Quick start

Grammars are written in `lmn` and compiled to `.lm` bytecode by the `lmn` compiler, which is itself a grammar. First build the compiler from its sources. `lmnbs.lm` is the original bootstrap, and the second stage is a fixpoint:

```sh
bin/lm -rules examples/lmn/lmnbs.lm examples/lmn/lmn2xfe.lmn examples/lmn/lmn2mbe.lmn > stage1.lm
bin/lm -rules stage1.lm -output lmn.lm examples/lmn/lmn2xfe.lmn examples/lmn/lmn2mbe.lmn
```

Then compile a grammar and run it:

```sh
bin/lm -rules lmn.lm -output calc.lm examples/basics/calc.lmn
bin/lm -rules calc.lm -input '3*(4+5)='          # the answer is 27
bin/lm -rules calc.lm examples/basics/calc.input  # input files as arguments
bin/lm -rules calc.lm -trace D -input '1+2='      # draw the lm-diagram
```

Run `bin/lm -h` for all options. Flags take effect in the order given (for example, `-dwidth` must come before `-trace D`).

## Documentation

- [`docs/technical_overview.md`](docs/technical_overview.md): architecture and CLI
- [`docs/internal_machine.md`](docs/internal_machine.md): how the engine loads and runs rules
- [`docs/bytecode.md`](docs/bytecode.md): the `.lm` bytecode format
- [`docs/lm/`](docs/lm/README.md): a structured digest of the original documentation, covering the execution model, `lmn` syntax, builtins, tracing and the lm-diagram
- [`docs/original/`](docs/original): the original website's wiki sources
- [`examples/`](examples/README.md): the original grammars, inputs and reference outputs

## License

Language Machine 2 is distributed under the GNU GPL v3 (see `COPYING`). It is based on the Language Machine © 2005 Peri Hankey, GNU GPL v2 (see `COPYING.languagemachine`).
