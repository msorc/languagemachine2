# Language Machine 2

A Go port of Peri Hankey's [Language Machine](https://languagemachine.sourceforge.net), a toolkit for language and grammar. In a Language Machine grammar, a rule applies when what is expected and what is in the input fail to match. Recognition and substitution interleave on the input stream, so one notation (`lm2n`) covers lexing, parsing, translation and evaluation.

The port keeps the original execution semantics and the lm-diagram trace. The grammars and reference outputs from the original release (in `examples/`) produce the same output, and the regression tests check this.

## Build

Requires Go 1.27 or later. There are no third-party dependencies.

```sh
make build        # -> bin/lm2
make test         # regression tests
make check        # vet + tests
make race         # tests under the race detector
make bench        # engine benchmarks at 1, 4 and 16 CPUs
```

## Quick start

Grammars are written in `lm2n` and compiled to `.lm2` bytecode by the `lm2n` compiler, which is itself a grammar. First build the compiler from its sources. `lm2nbs.lm2` is the original bootstrap, and the second stage is a fixpoint:

```sh
bin/lm2 -rules internal/lmnsrc/lm2nbs.lm2 internal/lmnsrc/lm2n2xfe.lm2n internal/lmnsrc/lm2n2mbe.lm2n > stage1.lm2
bin/lm2 -rules stage1.lm2 -output lm2n.lm2 internal/lmnsrc/lm2n2xfe.lm2n internal/lmnsrc/lm2n2mbe.lm2n
```

Then compile a grammar and run it:

```sh
bin/lm2 -rules lm2n.lm2 -output calc.lm2 examples/basics/calc.lm2n
bin/lm2 -rules calc.lm2 -input '3*(4+5)='          # the answer is 27
bin/lm2 -rules calc.lm2 examples/basics/calc.input  # input files as arguments
bin/lm2 -rules calc.lm2 -trace D -input '1+2='      # draw the lm-diagram
```

Run `bin/lm2-h` for all options. Flags take effect in the order given (for example, `-dwidth` must come before `-trace D`).

## Documentation

- [`docs/tutorial.md`](docs/tutorial.md): a hands-on tutorial, from a one-rule program to parsers, translators and Go binaries built with `lm2n2go`
- [`docs/technical_overview.md`](docs/technical_overview.md): architecture and CLI
- [`docs/internal_machine.md`](docs/internal_machine.md): how the engine loads and runs rules
- [`docs/bytecode.md`](docs/bytecode.md): the `.lm2` bytecode format
- [`docs/lm/`](docs/lm/README.md): a structured digest of the original documentation, covering the execution model, `lm2n` syntax, builtins, tracing and the lm-diagram
- [`docs/original/`](docs/original): the original website's wiki sources
- [`examples/`](examples/README.md): the original grammars, inputs and reference outputs

## License

Language Machine 2 is distributed under the GNU GPL v3 (see `COPYING`). It is based on the Language Machine © 2005 Peri Hankey, GNU GPL v2 (see `COPYING.languagemachine`).
