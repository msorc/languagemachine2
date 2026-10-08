# Language Machine 2 Technical Overview

## Background

Language Machine 2 is a Go reimplementation of Peri Hankey's Language Machine, a toolkit for writing online grammars and translators. The original documentation (https://languagemachine.sourceforge.net, digested in `lm2/`) describes the paradigm: rules are applied when what is expected and what is there fail to match, and recognition and substitution interleave on the incoming symbol stream. The Go port keeps this execution model, the lm-diagram, and the behaviour the published examples depend on. The examples in `examples/` produce the same output as the original engine, and `internal/machine/examples_test.go` checks this.

## Command-line application

The `lm2` binary (`cmd/lm2`) is a thin wrapper over `internal/application.Application`, which parses the flags and drives one `machine.Engine`. The options are:

- **Metadata:** `-version`, `-license`, `-shebang PATH` (prints a `#! PATH -rules` header for executable `.lm2` scripts; put it after `-output`).
- **Rules:** `-rules file` loads bytecode, replacing any loaded rules. `-add file` adds rules to those already loaded.
- **I/O:** `-input string`, `-stdin`, positional input files, `-output file` and `-errout file` (where `err` writes).
- **Engine limits:** `-lexpri` (accepted but not applied, as in the original), `-buffer`, `-max-repeat` and `-max-depth`.
- **Tracing:** `-trace` takes short codes, comma-separated or repeated (`-trace m,s`), and `-dwidth` sets the diagram width. `-trace-out` writes a Go runtime trace, not an LM trace.

The options are a table (`optionDefs`). Every occurrence of a flag is recorded as it is parsed, and after parsing they are applied in command-line order, so `-input a -input b` reads both, `-dwidth` must come before `-trace D`, and `-add` after `-rules`. The positional files are queued last. The `-trace` codes are a second table (`traceCodes`), which also generates the help text.

## Package layout

| Package | Contents |
| --- | --- |
| `cmd/lm2` | `main` |
| `internal/application` | the lm2 command line: the option table, its order, and loading a built-in program |
| `cmd/lm2n`, `cmd/lm2n2go` | the lm2n compiler built by lm2n2go, and lm2n2go itself (`lm2n2go.md`) |
| `internal/lm2go` | compiles `.lm2n`, generates the Go program and holds lm2n2go's command line |
| `internal/lm2nsrc` | the lm2n compiler sources and the bootstrap `lm2nbs.lm2`, embedded |
| `lm2` | the public runtime that generated programs import |
| `lm2/lm2n` | the public way to compile lm2n to bytecode |
| `examples/highlight` | an example library and command (`lmhl`) built on `lm2` |
| `internal/machine` | the whole runtime: loader, grammar store, engine, elements, modes, contexts, variables, I/O, builtins, tracer and diagram |
| `internal/conv` | URI encoding and decoding, C-style unescaping, and C-compatible number parsing (`Strtod`, `Strtoi`, `ScanOctal`, `ScanBinary`) |
| `internal/version` | version and licence strings |

The module has no third-party dependencies.

`internal/machine` is one package because nearly every part refers to the `Engine`. Its files are split by concern:

| File | Contents |
| --- | --- |
| `engine.go` | `Engine`, `Match`, `resolve`, inputs and options |
| `loader.go` | the bytecode loader |
| `grammar.go`, `priority.go`, `symbols.go` | rules, grammars and the `selector`; the priority encoding; `dict`, `predef` and `defineSymbols` |
| `element.go` | the `Element` interface and `genericElement` |
| `value.go`, `array.go`, `varref.go` | numbers, symbols, characters, strings, buffers; arrays and cells; variable references |
| `special.go`, `lex.go` | predefined symbols and builtins; lexical classes |
| `operator.go`, `control.go` | the operator table; control structures (`if`, loops, `break`/`continue`, `foreach`, `rule`), arrays, `each`/`all (expr)` and function calls |
| `mode.go`, `stream.go`, `context.go`, `variable.go` | generator modes, streams and operand stacks, contexts, variables and scopes |
| `input.go`, `convert.go`, `buffer.go` | input sources and `ioSymbol`; the `to…` conversions; `rzBuffer` |
| `builtin.go`, `extension.go`, `calls.go` | the functions rules call by name and `External`; `Calls` for lm2n2go |
| `tracer.go`, `diagram.go` | tracing and the lm-diagram |
| `errors.go`, `self_pointer.go`, `api.go` | `Error` and `fail`/`catch`; `selfPointing`; `Symbol` and `Number` for lm2 |

## Engine architecture

`machine.Engine` owns all run-time state:

- Symbol dictionaries (`dict`) for terminal, non-terminal, variable, user and function symbols. Symbols are unique within their dictionary and are compared by identity. The predefined symbols (`start`, `eof`, `repeat`, `option`, `out`, the `to…` conversions, the operators, …) are created once by `defineSymbols`.
- The grammar table (`selector`), which maps names to grammars. Each grammar files its rules by the pair (LHS initial token, RHS initial token), longest first.
- Two `Stream`s (LHS for what is expected, RHS for what is there), each with a stack of generator modes (`GenMode`).
- The left context chain (`contextHolder`): the rules being recognised. The context of a substitution being read belongs to the right stream's mode.
- The input stack (`Input`), which feeds characters through the backtracking buffer `rzBuffer`.
- The external function table (`External`), plus the tracer and the diagram. The tracer may be nil; all its methods then do nothing.

`Engine.match` is a loop over two generators. It takes a symbol from each stream and asks the LHS symbol to match the RHS symbol. On a mismatch, `resolve` looks up candidate rules in a fixed order, takes a `checkpoint` of both streams and the left context, and recurses into `match` for each candidate's pattern. It rolls back to the checkpoint when a candidate fails. `internal_machine.md` describes this step by step.

## Grammar loading and bytecode

Grammars are compiled from `lm2n` notation by the `lm2n` compiler, which is itself an LM grammar (`internal/lm2nsrc`), into a textual stack-machine bytecode. `Loader` builds `Rule`s from it: `m:`, `c:`, `v:` and similar opcodes push symbols, parentheses build sequences, and `r` defines a rule from five operands. `bytecode.md` is the full specification. `lm2/07-compilation-and-bytecode.md` lists where the Go loader differs from what the compiler can emit.

## Elements, modes and variables

Everything the machine handles is an `Element`: terminal characters, symbols, numbers, strings, arrays, variables, lexical classes, and the builtins (take `%`, bind `:`, output symbols, `repeat`, arithmetic, control statements, …). Each element type defines its own `match` and `act` behaviour. The interface is grouped: the grammar engine, values, text, and the arithmetic (`operand`), whose methods take the stream they act on so that an invalid operation is reported on the engine's error output. Elements that trace before they act implement `traceable`.

`GenMode`s generate symbols. There are root modes, modes that step through a rule's pattern or substitution, and modes for loop bodies, in-place substitutions and variable references. Value snapshots of the stream registers (`modeSnap`, gathered in a `checkpoint`), together with the persistent operand stack (`opStack`), make backtracking cheap.

Variables (`binding`) are linked lists that record where each binding was made (state, grammar, input position). They are linked into both a scope chain and an all-variables chain, which `each` and `all` walk.

Go has no virtual dispatch through embedding, so the element types embed `selfPointing[Element]` (through `genericElement`) and call overridable methods through `Self()`; their constructors set the pointer with `makeSelf`/`reSelf`. Modes, contexts and input handlers are plain structs.

## I/O

Inputs implement `Input`: `NewStdinInput` reads the engine's standard input (`SetStdin`), `NewReaderInput` any `io.Reader`, `NewStringInput` a string and `NewFileInput` a whole file. Command-line inputs are queued in order. The `include` builtin pushes a source that is read to its end before reading returns to the previous one. The output symbols (`out`, `uri`, `urd`) write through the engine's buffered output writer, and `err` writes to its error writer. The `to…` symbols (`toNum`, `toSym`, `toUstr`, …) are one `converter` type driven by a table; they turn grabbed material into numbers, symbols and strings.

## Tracing and the diagram

`-trace` sets `Tracer` flags for mismatches, symbol comparisons, bindings, references, loops, loading and grammar dumps. `-trace D` (Unicode) or `-trace d` also enables the categories the lm-diagram needs, and draws it with `Diagram` at the `-dwidth` width. The diagram, its text form and the mismatch and symbol traces match the original engine byte for byte, apart from Unicode box drawing; `TestTraceGolden` checks them against reference output in `internal/machine/testdata/trace`, and pins the port's own output for the other categories.

## Embedding and extending

Programs outside this module embed the machine through `lm2`: an `lm2.Program` holds the bytecode and the Go functions the rules call, and `Translate`, `TranslateReader` or `Run` run it, each in an engine of its own. `lm2/lm2n` compiles lm2n source to bytecode. `lm2n2go` generates the `Program` from lm2n sources (`lm2n2go.md`).

Inside the module, the engine is used directly: `machine.NewEngine()`, `LoadFromString` (or `LoadFromStringReset(text, false)` to add rules), `AppendInput(machine.NewStringInput(e, text))`, then `Start` or `Run`. A load that fails returns an error with the line and column and leaves the rules as they were. `Start` returns the exit status and an error for failures such as an exceeded limit or a missing include file; `Run` turns a failed analysis into `ErrNoMatch`. Output goes to stdout unless you pass another writer to `SetOutput` (and `SetErrOutput` for `err`).

To add a builtin that grammars can call, add it to the `builtins` table in `internal/machine/extension.go`.

When you change the loader, the bytecode or the runtime semantics, update `bytecode.md` and check the change against `lm2/`. Keep the trace output consistent with the lm-diagram.
