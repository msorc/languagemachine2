# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Language Machine 2 is a Go port of Peri Hankey's Language Machine (https://languagemachine.sourceforge.net). It is a grammar/translator engine in which a rule applies when the expected (LHS) symbol stream and the incoming (RHS) symbol stream fail to match. It has to keep the legacy execution semantics and the lm-diagram trace output, so results can be checked against the diagrams in the published guides.

In-depth docs live in `docs/`: `technical_overview.md`, `internal_machine.md` (the execution flow) and `bytecode.md` (the `.lm` bytecode spec). `docs/lm/` is a structured digest of the original Language Machine website, and it is the reference for the intended semantics:

- the execution model and rule selection order
- lmn syntax
- special symbols and builtins
- the CLI and trace flags
- the lm-diagram, including a verbatim reference diagram
- how lmn compiles to bytecode

Check it before you change engine behaviour. `docs/lm/README.md` also lists the places where the Go loader is known to differ from the original compiler's output. When you change the loader, bytecode or runtime semantics, update `docs/bytecode.md` too.

## Commands

```sh
make build        # -> bin/lm   (go build -o bin/lm ./cmd/lm)
make test         # go test ./...   (regression tests: internal/machine/machine_test.go)
make vet / make fmt / make tidy
go test ./internal/machine -run TestName   # single test
```

To run a grammar:

```sh
bin/lm -rules calc.lm calc.input          # positional args are input files
bin/lm -rules calc.lm -input '+ 2 3'      # input given as a string
bin/lm -rules calc.lm -trace D calc.input # lm-diagram (set -dwidth before -trace D)
```

Flags run as callbacks in the order they are given on the command line, and the positional-files handler runs last (`internal/application/application.go`). `-trace` takes short codes that can be combined as CSV or repeated (e.g. `-trace m,s`); the code→flag map is `traceMap` in the same file. `-trace-out` writes a Go runtime trace, not the LM trace.

Wrap ad-hoc runs in `timeout`, because a grammar that does not reach its end state can loop forever and flood output (e.g. endless `eof`).

## Files at the repo root

`*.lmn` files are grammars in LM notation (the source language). `*.lm` files are the compiled bytecode that `-rules` loads. `lmnbs.lm` is the original lmn bootstrap compiler, which compiles `.lmn` into `.lm` (`bin/lm -rules lmnbs.lm -output foo.lm foo.lmn`). `*.input` files are sample inputs, and `*.dia` files are saved diagram output. None of these are tracked in git; they are scratch/example files.

The tracked material from the original release lives in two places. `examples/` holds the original grammars, inputs and reference outputs, and `examples/lmn/` holds the lmn compiler sources plus the original `lmnbs.lm`; its README covers the layout and the two-stage compiler build. `docs/original/` holds the website's `.wiki` sources and images. `internal/machine/examples_test.go` runs regression tests against them: the bootstrap fixpoint, the original `test-inc`, compiling every example, and the golden sample outputs.

## Architecture

- `cmd/lm/main.go` is a thin wrapper around `internal/application.Application`, which parses flags and drives a `machine.Engine`.
- `internal/machine` holds the whole runtime, as a single package:
  - `loader.go` tokenises the bytecode with one regex and dispatches on the opcode's first character. It builds `Str` lists on an operand stack, and `r` pops 5 operands (grammar, priority, offset, LHS, RHS) and calls `Engine.AddRule` → `Grammar.DefineRule` → `Grammar.Add`.
  - `grammar.go` / `gram_system.go` store rules per grammar, filed under the pair (LHS initial, RHS initial) and found with `Grammar.Get`. Each group is ordered by descending length (the sum of LHS weights), newest first among equals. `Selector` maps grammar names to grammars.
  - `engine.go`: `Engine.Start` → `Match` is a dual-generator loop that advances the LHS and RHS `Stream`s. When they mismatch, `ResolveE` looks up candidate rules, saves the mode/context/input state, pushes an RHS context, recurses into `Match`, and restores the saved state if that fails. This is the core algorithm.
  - `mode.go` (`GenMode`, LHS/RHS modes), `stream.go`, `context.go` (`ContextHolder` snapshots, priorities, depth), `variable.go` / `scope.go` (variable binding and scope chains).
  - `element.go` is the largest file. It defines the element types (`Sym`, `Chr`, `Str`, `VarSym`, lexical classes, builtins such as take/bind/drop) and their `Match` behaviour.
  - `builtin.go` / `extension.go`: predefined functions and the `LMExternal` table of Go functions that grammars can call (numeric/casing helpers, include, trace toggles). Add new primitives here.
  - `tracer.go` / `diagram.go`: categorised tracing and the Unicode lm-diagram renderer. The trace output should stay consistent with the legacy lm-diagram.
- Input goes through the `GrammarIO` interface (stdin/file/buffer inputs), which sits on an input stack; output symbols write to `os.Stdout`/`os.Stderr`. RHS characters are read through the growable backtracking buffer `RZBuffer`.
- `internal/summary` holds the version and license strings.
