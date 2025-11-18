# Language Machine 2 Technical Overview

## Background and Philosophy

Language Machine 2 is a Go reimplementation of Peri Hankey's Language Machine, a toolkit for writing online grammars and translators. The legacy documentation on https://languagemachine.sourceforge.net describes the paradigm: rules are applied by interleaving recognition and substitution phases, operating directly on the incoming symbol stream. The Go port preserves this execution model, the lm-diagram visualization, and the idioms used in the published guides, tutorials, and recipes.

## Command-Line Application

The `lm` binary is a thin wrapper over `internal/application.Application`. The same constructor can load grammars from files (`-rules`), merge additional rules (`-add`), or accept prebuilt engines/external tables. Options include:

- Metadata: `-version`, `-license`, `-shebang`, `-gomain`.
- I/O: `-input`, `-stdin`, positional file arguments, `-output`, `-errout` (placeholder), plus multi-input stacking via `include`.
- Engine tuning: `-lexpri`, `-buffer`, `-max-repeat`, `-max-depth`, `-dwidth`.
- Tracing: `-trace` accepts CSV/accumulated short codes to toggle categories.

Every flag is mapped to a callback; after parsing, callbacks are executed in the order flags were specified, then the positional file handler runs.

## Engine Architecture

`internal/machine.Engine` owns the parser state:

- Symbol dictionaries (`Dict`) for terminal, non-terminal, variable, user, and function symbols, with predefined entries (`start`, `eof`, `repeat`, `option`, `true/false`, IO helpers, etc.).
- Two `Stream` instances (LHS/RHS) with corresponding `GenMode`s to produce elements during matching.
- Context stacks (`ContextHolder`) capturing snapshots for mismatch resolution, depth limits, and tracing.
- Input stack (`bidlist.List[GrammarIO]`) feeding characters from stdin, files, or string buffers into the RHS circular buffer (`RZBuffer`).
- External function table (`LMExternal`) exposing helpers such as casing, numeric parsing, trace toggles, file inclusion, and symbol access.
- Tracing/diagramming infrastructure (`Tracer`, `Diagram`) that mirrors the original lm-diagram depiction.

`Engine.Start` injects `start` if defined, forces an input if none supplied, dumps the grammar when tracing, and runs `Match`; exit status `0` indicates success with no flagged errors.

## Grammar Loading and Bytecode

Grammars are expressed in the LM bytecode format loaded by `internal/machine/Loader`. Instructions form a stack machine that constructs `Rule` objects (see `docs/bytecode.md` for a detailed specification). Key opcodes:

- `m:<name>` push non-terminal symbols, `f:<name>` push function/operator symbols, `c:<text>` emits terminal characters.
- `L/R/B:<n>` encode priorities, `n:<n>` sets the RHS offset.
- Parentheses build `Str` sequences used as LHS/RHS bodies, `r` finalizes a rule (grammar symbol, priority, offset, LHS, RHS).

The loader finally calls `Engine.DefineElements`, which weighs rules, sets effective initial symbols, and stores the linked list under the owning grammar via `Grammar.Add`.

## Modes, Streams, and Variables

`Stream` is a register set consisting of an operand stack, current symbol/value, and code vector. `GenMode` implementations (LHS and RHS modes) orchestrate stepping through rule bodies while respecting nested scopes and allowing backtracking via `Save/Restore` semantics.

Variables (`VarElement`) carry bindings, scope chains, source-file metadata (line/column/file), and provide arithmetic/boolean conversion fallbacks. Contexts (`Context`) snapshot state, priority, operands, and variable chains at mismatch points, enabling recursive descent plus rewrites to interleave.

## I/O Abstractions

All input/output moves through the `GrammarIO` interface:

- `GramInput` reads from stdin, `GramInputFile` from files, `GramInputBuffer` from literal strings.
- Output sinks include `GramOutputFile` and `GramOutputBuffer`.
- Helpers like `ToConvert` build convenience transformations used by some legacy grammars.

RHS matching consumes characters via `RZBuffer`, which grows exponentially (subject to `SetBuffer`) and supports backtracking constraints.

## Tracing and Visualization

Trace flags map to specific concerns: mismatches, symbol matching, variable binding, reference scopes, arithmetic, loop execution, loader activity, grammar dumps, and diagramming. Setting `-trace D` or `-trace d` automatically enables the dependent categories required for the lm-diagram.

`Tracer` prints categorized events and dumps operand stacks; `Diagram` renders the Unicode visualization of LHS/RHS timelines, making it possible to compare Go output directly with the diagrams shown on the historic website.

## Extending or Embedding

To embed LM in another Go program, instantiate `machine.NewEngine`, load bytecode via any `LoadFrom*` helper, add inputs, and call `Start`. Extend `LMExternal` with custom functions when a grammar needs additional primitives.

When modifying or extending the bytecode or runtime, update both `docs/bytecode.md` and the related loader/engine code so future grammar authors can rely on the documentation. Keep the tracing output aligned with the lm-diagram semantics to ensure consistency with the published guides.

