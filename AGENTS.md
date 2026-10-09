# AGENTS.md

This file provides guidance to coding agents when working with code in this repository.

## What this is

Language Machine 2 is a Go port of Peri Hankey's Language Machine (https://languagemachine.sourceforge.net). It is a grammar/translator engine in which a rule applies when the expected (LHS) symbol stream and the incoming (RHS) symbol stream fail to match. It has to keep the legacy execution semantics and the lm-diagram trace output, so results can be checked against the diagrams in the published guides.

In-depth docs live in `docs/`: `technical_overview.md`, `internal_machine.md` (the execution flow) and `bytecode.md` (the `.lm` bytecode spec). `docs/lm/` is a structured digest of the original Language Machine website, and it is the reference for the intended semantics:

- the execution model and rule selection order
- lmn syntax
- special symbols and builtins
- the CLI and trace flags
- the lm-diagram, including a diagram as published on the site (drawn by an earlier engine; the golden traces in `internal/machine/testdata/trace` are the released engine's output)
- how lmn compiles to bytecode

Check it before you change engine behaviour. `docs/lm/README.md` also lists the places where the Go loader is known to differ from the original compiler's output. When you change the loader, bytecode or runtime semantics, update `docs/bytecode.md` too.

## Commands

```sh
make build        # -> bin/lm   (go build -o bin/lm ./cmd/lm); VERSION=x.y.z stamps the version
make lmn2go       # -> bin/lmn2go, compiles rules to a Go program (docs/lmn2go.md)
make lmn          # -> bin/lmn, the lmn compiler built by lmn2go (cmd/lmn/lmn.go is generated)
make generate     # go generate ./... (rebuild cmd/lmn/lmn.go after changing internal/lmnsrc sources)
make test         # go test ./...   (engine: internal/machine/*_test.go, CLI: internal/application/application_test.go)
make check        # fmt-check, vet, lint, test, race, generate-check (what CI runs: make ci)
make lint         # golangci-lint run ./... (.golangci.yml)
make race         # go test -race ./... (engines are independent; tests run them in parallel)
make bench        # engine benchmarks (internal/machine/bench_test.go) at -cpu 1,4,16
make fix          # go fix ./... (apply the toolchain's modernizers)
make release      # uses the release constant in internal/version/version.go: runs make check,
                  # tags vX.Y.Z, pushes it, creates the GitHub release with notes from the commits
                  # since the last tag, then asks proxy.golang.org to fetch it (needs gh and a clean
                  # master with the release constant bumped and committed; VERSION=x.y.z overrides;
                  # make release-notes previews the notes; YES=1 skips the prompt)
make vet / make fmt / make tidy
go test ./internal/machine -run TestName   # single test
```

To run a grammar:

```sh
bin/lm -rules calc.lm calc.input          # positional args are input files
bin/lm -rules calc.lm -input '+ 2 3'      # input given as a string
bin/lm -rules calc.lm -trace D calc.input # lm-diagram (set -dwidth before -trace D)
```

Flags are recorded as they are parsed and applied in the order they are given on the command line, every occurrence of them, and the positional input files are queued last (`optionDefs` in `internal/application/application.go`). `-trace` takes short codes that can be combined as CSV or repeated (e.g. `-trace m,s`); the code→flag table is `traceCodes` in the same file. `-trace-out` writes a Go runtime trace, not the LM trace.

Wrap ad-hoc runs in `timeout`, because a grammar that does not reach its end state can loop forever and flood output (e.g. endless `eof`).

## Files at the repo root

`*.lmn` files are grammars in LM notation (the source language). `*.lm` files are the compiled bytecode that `-rules` loads. `lmnbs.lm` is the original lmn bootstrap compiler, which compiles `.lmn` into `.lm` (`bin/lm -rules lmnbs.lm -output foo.lm foo.lmn`). `*.input` files are sample inputs, and `*.dia` files are saved diagram output. None of these are tracked in git; they are scratch/example files.

The tracked material from the original release lives in two places. `examples/` holds the original grammars, inputs and reference outputs, and `internal/lmnsrc/` holds the lmn compiler sources (moved from the original `lmn/` directory, so the Go tools can embed them) plus the original `lmnbs.lm`; its README covers the layout and the two-stage compiler build. `docs/original/` holds the website's `.wiki` sources and images. `internal/machine/examples_test.go` runs regression tests against them: the bootstrap fixpoint, the original `test-inc`, compiling every example, the golden sample outputs, and `TestTraceGolden` (lm-diagram and trace output against the original engine's, in `testdata/trace`).

## Architecture

- `cmd/lm/main.go` is a thin wrapper around `internal/application.Application` (`application.New`), which parses flags from a table (`optionDefs`), applies every occurrence in command-line order, and drives a `machine.Engine`.
- `internal/machine` holds the whole runtime, as a single package. Only its API is exported (the `Engine` and its setters, `Element`, `Input`, `ExtFn`, `External`, `Error`, the `Trace…` flags, `Calls`, `Builtin`); everything else is unexported. `golangci-lint` (`.golangci.yml`) reports what nothing uses.
  - `loader.go` tokenises the bytecode with one regex and dispatches on the opcode's first character. It builds `str` lists on an operand stack, and `r` checks and pops 5 operands (grammar, priority, offset, LHS, RHS) and calls `Engine.addRule` → `grammar.defineRule` → `grammar.add`. Faults are `*Error`s with line:col, and a failed load leaves the rules as they were. `priority.go` holds the priority encoding (`L:`/`R:`/`B:`/`M:`).
  - `grammar.go` stores rules per grammar, filed under the pair (LHS initial, RHS initial) and found with `grammar.get` as a slice. Each group is ordered by descending length (the sum of LHS weights), newest first among equals. `selector` maps grammar names to grammars.
  - `engine.go`: `Engine.Start` → `match` is a dual-generator loop that advances the LHS and RHS `Stream`s. When they mismatch, `resolve` looks up candidate rules, takes a `checkpoint` (both streams' `modeSnap`s and the left context), recurses into `match` for each candidate, and rolls back to the checkpoint when one fails. This is the core algorithm.
  - `symbols.go`: the symbol dictionaries (`dict`), the `predef` table and `Engine.defineSymbols`, which creates every predefined symbol once.
  - `mode.go` (`GenMode`, LHS/RHS modes), `stream.go`, `context.go` (`contextHolder`: state, rule, priority, depth), `variable.go` (`scopeHolder`, variable binding and scope chains).
  - Elements: `element.go` holds the `Element` interface (grouped: engine, values, text, and the `operand` arithmetic, which takes the stream so invalid operations reach the engine's error output) and `genericElement`, the defaults every element embeds. Elements use `selfPointing` (`self_pointer.go`) for overridable defaults; nothing else does. The element types are spread over `value.go` (`number`, `symbol`, `chr`, `str`, …), `array.go`, `varref.go` (`varSym`, `each`/`all` references), `special.go` (take/bind/drop, output symbols, `repeat`/`option`, …), `lex.go` (lexical classes), `operator.go` (one `operator` type and the table of operators) and `control.go` (`if`, loops, `foreach`, `rule`, `fun`, …). Each type defines its own `match`/`act` behaviour.
  - `builtin.go` / `extension.go`: predefined functions and the `builtins` table that each engine's `External` copies (numeric/casing helpers, include, trace toggles). Add new primitives to that table.
  - `tracer.go` / `diagram.go`: categorised tracing and the Unicode lm-diagram renderer. The trace output must stay byte-identical to the original's (after mapping box drawing to ASCII); `TestTraceGolden` checks it, and pins the port's own output for the other trace categories (regenerate deliberately with `LM_UPDATE_GOLDEN=1`). The tracer may be nil; its methods then do nothing.
- Input goes through the `Input` interface (`input.go`: `NewStdinInput`, `NewReaderInput`, `NewStringInput`, `NewFileInput`; `ioSymbol`), which sits on an input stack; `Engine.SetStdin` sets standard input; `convert.go` holds the table of `to…` converters; output symbols, traces and diagrams write to the engine's buffered `out` writer (`SetOutput`) and `err` to `errOut`. RHS characters are read through the growable backtracking buffer `rzBuffer` (`buffer.go`).
- `internal/conv` holds the text conversions that must match the original's C behaviour (URI encoding, C escapes, `strtod`); `internal/version` holds the version and license strings.
- The module has no third-party dependencies; keep it that way.
- lmn2go (`docs/lmn2go.md`), the Go counterpart of the original `lmn2d`: `cmd/lmn2go` is the CLI; `internal/lmgo` compiles `.lmn` with the lmn compiler built from `internal/lmnsrc` (embedded by `internal/lmnsrc/embed.go`) and generates a Go file holding the bytecode and a table of the Go functions the rules call (found by `machine.Calls`); `lm/` is the public runtime that generated code imports (`Program.Run`, `Translate`, `TranslateReader`), and `lm/lmn` compiles lmn for programs outside the module. `examples/highlight` is an importable example built only on them. Keep `lm/` small and free of engine types. The `//go:generate` lines are the single source of lmn2go arguments: `lmgo.UpToDate` regenerates with them in the freshness tests.
