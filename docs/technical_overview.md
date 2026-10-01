# Language Machine 2 Technical Overview

## Background

Language Machine 2 is a Go reimplementation of Peri Hankey's Language Machine, a toolkit for writing online grammars and translators. The original documentation (https://languagemachine.sourceforge.net, digested in `lm/`) describes the paradigm: rules are applied when what is expected and what is there fail to match, and recognition and substitution interleave on the incoming symbol stream. The Go port keeps this execution model, the lm-diagram, and the behaviour the published examples depend on. The examples in `examples/` produce the same output as the original engine, and `internal/machine/examples_test.go` checks this.

## Command-line application

The `lm` binary (`cmd/lm`) is a thin wrapper over `internal/application.Application`, which parses the flags and drives one `machine.Engine`. The options are:

- **Metadata:** `-version`, `-license`, `-shebang PATH` (prints a `#!` header for executable `.lm` scripts).
- **Rules:** `-rules file` loads bytecode, replacing any loaded rules. `-add file` adds rules to those already loaded.
- **I/O:** `-input string`, `-stdin`, positional input files, `-output file` and `-errout file` (where `err` writes).
- **Engine limits:** `-lexpri` (accepted but not applied, as in the original), `-buffer`, `-max-repeat` and `-max-depth`.
- **Tracing:** `-trace` takes short codes, comma-separated or repeated (`-trace m,s`), and `-dwidth` sets the diagram width. `-trace-out` writes a Go runtime trace, not an LM trace.

Every flag has a callback. After parsing, the callbacks run in the order the flags appeared on the command line (so `-dwidth` must come before `-trace D`, and `-add` after `-rules`), and then the positional files are queued.

## Package layout

| Package | Contents |
| --- | --- |
| `cmd/lm` | `main` |
| `internal/application` | flag parsing and the callback order |
| `cmd/lmn`, `cmd/lmn2go` | the lmn compiler built by lmn2go, and lmn2go itself (`lmn2go.md`) |
| `internal/lmgo` | compiles `.lmn` and generates the Go program for lmn2go |
| `lm` | the public runtime that generated programs import |
| `internal/machine` | the whole runtime: loader, grammar store, engine, elements, modes, contexts, variables, I/O, builtins, tracer and diagram |
| `internal/conv` | URI encoding and decoding, C-style unescaping, and C-compatible number parsing (`Strtod`, `Strtoi`, `ScanOctal`, `ScanBinary`) |
| `internal/version` | version and licence strings |

The module has no third-party dependencies.

`internal/machine` is one package because nearly every part refers to the `Engine`. Its files are split by concern:

| File | Contents |
| --- | --- |
| `engine.go` | `Engine`, `Match`, `ResolveE`, inputs and options |
| `loader.go` | the bytecode loader |
| `grammar.go`, `symbols.go` | rules, grammars and `Selector`; `Dict`, `Predef` and `defineSymbols` |
| `element.go` | the `Element` interface and `GenericElement` |
| `value.go`, `array.go`, `varref.go` | numbers, symbols, characters, strings, buffers; arrays and cells; variable references |
| `special.go`, `lex.go` | predefined symbols and builtins; lexical classes |
| `operator.go`, `control.go` | operators; control structures (`if`, loops, `break`/`continue`, `foreach`, `rule`), arrays, `each`/`all (expr)` and function calls |
| `mode.go`, `stream.go`, `context.go`, `variable.go` | generator modes, streams and operand stacks, contexts, variables and scopes |
| `input.go`, `convert.go`, `buffer.go` | input sources and `IOSymbol`; the `to…` conversions; `RZBuffer` |
| `builtin.go`, `extension.go`, `calls.go` | the functions rules call by name and `LMExternal`; `Calls` for lmn2go |
| `tracer.go`, `diagram.go` | tracing and the lm-diagram |
| `errors.go`, `self_pointer.go` | `Error` and `fail`/`catch`; `SelfPointing` |

## Engine architecture

`machine.Engine` owns all run-time state:

- Symbol dictionaries (`Dict`) for terminal, non-terminal, variable, user and function symbols. Symbols are unique within their dictionary and are compared by identity. The predefined symbols (`start`, `eof`, `repeat`, `option`, `out`, the `to…` conversions, the operators, …) are created once by `defineSymbols`.
- The grammar table (`Selector`), which maps names to `Grammar`s. Each grammar files its rules by the pair (LHS initial token, RHS initial token).
- Two `Stream`s (LHS for what is expected, RHS for what is there), each with a stack of generator modes (`GenMode`).
- Two context chains (`ContextHolder`): the rule being recognised and the rule whose substitution is being read.
- The input stack (`GrammarIO`), which feeds characters through the backtracking buffer `RZBuffer`.
- The external function table (`LMExternal`), plus the `Tracer` and `Diagram`.

`Engine.Match` is a loop over two generators. It takes a symbol from each stream and asks the LHS symbol to match the RHS symbol. On a mismatch, `ResolveE` looks up candidate rules in a fixed order, saves the modes, and recurses into `Match` for each candidate's pattern. It restores the saved state when a candidate fails. `internal_machine.md` describes this step by step.

## Grammar loading and bytecode

Grammars are compiled from `lmn` notation by the `lmn` compiler, which is itself an LM grammar (`examples/lmn`), into a textual stack-machine bytecode. `Loader` builds `Rule`s from it: `m:`, `c:`, `v:` and similar opcodes push symbols, parentheses build sequences, and `r` defines a rule from five operands. `bytecode.md` is the full specification. `lm/07-compilation-and-bytecode.md` lists where the Go loader differs from what the compiler can emit.

## Elements, modes and variables

Everything the machine handles is an `Element`: terminal characters, symbols, numbers, strings, arrays, variables, lexical classes, and the builtins (take `%`, bind `:`, output symbols, `repeat`, arithmetic, control statements, …). Each element type defines its own `Match` and `Act` behaviour.

`GenMode`s generate symbols. There are root modes, modes that step through a rule's pattern or substitution, and modes for loop bodies, in-place substitutions and variable references. Their `Save`/`Restore` snapshots, together with the persistent operand stack (`OpStack`), make backtracking cheap.

Variables (`Var`) are linked lists that record where each binding was made (state, grammar, input position). They are linked into both a scope chain and an all-variables chain, which `each` and `all` walk.

Go has no virtual dispatch through embedding, so the element, mode, context, variable and I/O handler types embed `SelfPointing[T]` and call overridable methods through `Self()`. Constructors must set the pointer with `MakeSelf`/`ReSelf`.

## I/O

Inputs implement `GrammarIO`: `GramStdio` (stdin), `GramInputFile` (a whole file) and `GramInputBuffer` (a string). Command-line inputs are queued in order. The `include` builtin pushes a source that is read to its end before reading returns to the previous one. The output symbols (`out`, `uri`, `urd`) write through the engine's buffered output writer, and `err` writes to its error writer. The `ToConvert` handlers (`toNum`, `toSym`, `toUstr`, …) turn grabbed material into numbers, symbols and strings.

## Tracing and the diagram

`-trace` sets `Tracer` flags for mismatches, symbol comparisons, bindings, references, loops, loading and grammar dumps. `-trace D` (Unicode) or `-trace d` also enables the categories the lm-diagram needs, and draws it with `Diagram` at the `-dwidth` width. The diagram, its text form and the mismatch and symbol traces match the original engine byte for byte, apart from Unicode box drawing; `TestTraceGolden` checks them against reference output in `internal/machine/testdata/trace`.

## Embedding and extending

To embed the machine in another Go program: create an engine with `machine.NewEngine()`, load bytecode with `LoadFromString` (or `LoadFromStringReset(text, false)` to add to it), queue inputs with `AppendInput(machine.NewGramInputBuffer(e, text))`, and call `Start`. `LoadFromString` returns an error for malformed bytecode. `Start` returns the exit status and an error for failures such as an exceeded limit or a missing include file. Output goes to stdout unless you pass another writer to `SetOutput` (and `SetErrOutput` for `err`).

To add a builtin that grammars can call, register a function with `LMExternal.Set` in `NewLMExternal` (`internal/machine/extension.go`).

When you change the loader, the bytecode or the runtime semantics, update `bytecode.md` and check the change against `lm/`. Keep the trace output consistent with the lm-diagram.
