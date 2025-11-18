# Language Machine 2 – Internal Execution Flow

This document dives into how the Go runtime ingests LM bytecode and executes it. Use it together with `docs/bytecode.md` when you need a mental model of what the virtual machine is doing.

## 1. Loading Bytecode

### 1.1 Tokenisation and Dispatch
`internal/machine/loader.go` drives ingestion via a single `Load` function. It tokenises rule text with `([().reAtpbPgGVsawz])|(.:\S*)|#[^\n]*\n|\s*`, filters whitespace, and optionally emits `LOAD` traces. Every token’s leading character selects a handler (e.g. `L` → `Loader.L`, `c` → `Loader.c`, `(` → `Loader.O`). Payloads in `X:value` tokens pass through URL decode + `utils.Unescape` so bytecode can contain C-style escapes.

### 1.2 Stack Machine Semantics
The loader keeps a single operand stack plus a counter that records how many values have been pushed since the last `(`. Opcodes push elements (`NewSym`, `NewChr`, `NewVarSym`, `NewLexFromEngine`, etc.), mutate the stack (`.` drops, `w` creates fresh variables), or mark list boundaries. `(` stores the previous count (via `BMark`), resets it, and `)` collects the pending operands into a `Str` before restoring the count (`EMark`).

### 1.3 Rule Assembly
`r` (or `e`) pops five operands: grammar symbol, priority word, RHS offset, LHS sequence, RHS sequence. `Loader.r` forwards them to `Engine.DefineElements`, which calls `Grammar.Define`. During this step the runtime:

1. Computes the LHS “weight” for rule ordering.
2. Derives the effective first symbols on both sides (tokens drive lookup in `Grammar.Selector`).
3. Inserts the new `Rule` into the grammar’s linked list sorted by length/priority.

At the end of `Engine.Load*` the engine has populated dictionaries, built `Rule` objects, and set `Engine.oneGrammar` to the initial grammar entry.

## 2. Engine Bootstrapping

1. `NewEngine` creates symbol dictionaries (`Dict`), the `RZBuffer`, left/right streams, contexts, and the predefined symbol table (`Predef`).
2. `defineSymbols` seeds canonical non-terminals (`start`, `eof`, `repeat`, `anything`, etc.) and predefined functions (`append`, `repeat`, `option`, numeric helpers, IO symbols).
3. Loading bytecode optionally resets the grammar selector (`LoadFromStringReset`). `Engine.loader` is cached so subsequent `-add` calls reuse the same stack machine.
4. External helper table (`LMExternal`) is initialised with numeric/base converters, casing helpers, include/tracing hooks, buffer creation, and metadata accessors.

## 3. Streams, Modes, and Contexts

- **Streams.** Each `Stream` has a `codeVector`, `codeIndex`, operand stack, active `GenMode`, and links to the owning `Engine`. The LHS stream produces the pattern the RHS must match, and uses `LHMode` subclasses; the RHS stream produces input symbols via `RHMode`.
- **Modes.** A `GenMode` encapsulates execution state: references to the stream, operands, current symbol/value, scope references, and a back-link (`stackMode`). `Advance` drives execution by asking the stream to `Act`, and `Ret` restores prior state.
- **Contexts.** `ContextHolder` instances capture state snapshots whenever the engine resolves mismatches. LHS contexts track the rule currently matching and expose variable scopes; RHS contexts synchronise substitutions. Each context records a priority (derived from rule priority bits), nesting depth, operands, and variable chains.

## 4. Runtime Matching Loop

1. `Engine.Start` guarantees an input source (stdin if none supplied), attaches `start` if that rule exists, and kicks off `Match`.
2. `Match` is a dual generator loop: it repeatedly advances LHS and RHS modes until both have emitted current symbols. Tracing mode invokes `Tracer.TraceShort` before each advancement and prints symbol pairs before comparisons.
3. Special symbol `nil` acts as epsilon; when either side yields it, the engine resets the symbol and continues without comparing.
4. If `lhsSymbol.Match(rhsSymbol)` returns true, both streams clear their `currentSymbol` and resume advancing. A mismatch triggers `ResolveE`.

## 5. Mismatch Resolution

`ResolveE` implements the “grammar substitution” described on the original LM site:

1. Compute the candidate priority (`Element.Priority` combined with current context priority). If the context priority is already the sentinel (`PRIMASK`), backtracking stops.
2. Query the active grammar via `Grammar.Get(r.Token(), l.Token())` to find rules whose RHS initial token matches the incoming symbol and whose LHS initial token matches the expected symbol.
3. For each candidate rule:
   - Snapshot stream modes (`Save`), contexts, and state (input position, line/char counters) so the engine can roll back.
   - Push an RHS context via `PushRhx0`/`PushRhx1`, which also prepares `take` variables if operands are present and emits diagram “scope” updates.
   - Re-run `Match` recursively. If it succeeds, restoration occurs and the engine continues with the substituted RHS.
   - Otherwise, restore snapshots and try the next rule in the list.
4. If no rules match, `ResolveContext` eventually bubbles the failure up, allowing outer contexts to attempt their own replacements or ultimately fail the engine run.

## 6. Substitution and Execution Helpers

- **Variable binding.** `BindCvar`, `BindLvar`, and `BindXvarE` attach actual elements to the symbolic variable nodes pushed from bytecode (e.g., `v:name`). Bindings live in the current context and are visible according to the scope chain.
- **Stack/operand transfer.** `PushX`, `PushR`, `PushXElem` move elements between streams when RHS code needs to feed LHS operands or vice versa. RHS instructions can reference operands through `Stream.operands`.
- **RHS offsets.** Rules with `offset > 0` skip the first few RHS elements when initialising `RHMode`, which mimics the classic LM ability to start substitutions partway into the RHS.
- **Input buffer.** `RZBuffer` backs RHS `GetChr` calls so repeated scans over the most recent characters do not re-read the external source. It grows up to `SetBuffer`/`bufferLength` and enforces backtrack limits.

## 7. Control Flow Primitives

- **Repeat/Repeatx.** These helpers implement LM’s `repeat`/`repeatN` operators. They clone LHS modes, rerun `Match` up to `maxRepeat` times, and take care to restore RHS mode snapshots between iterations.
- **Include and external calls.** The `include` builtin (exposed both in bytecode and via CLI flags) pushes additional `GrammarIO` sources on the input stack, letting grammars load other files or strings mid-run. External calls dispatch through `LMExternal.Call`, so custom Go functions can be registered.

## 8. Tracing and Diagramming Hooks

- `SetTraceFlag` lazily creates a `Tracer`, enables requested bitmasks, and auto-enables dependent flags for diagram drawing (`DIAGRAM` requires mismatch, symbols, and context scope tracing).
- During matching/resolution the tracer logs symbol comparisons, mismatches (`Back`), rule scope replacements (`RuleScope`), operand dumps (`Dumpx`), and diagram glyphs (`Diagram.Trace`, `Diagram.Replace`, etc.). This reproduces the lm-diagram touted in the legacy documentation.

## 9. Putting It All Together

1. **Load phase:** call `LoadFromString`/`LoadFromStringReset` (or `LoadFromLMEString`) to decode bytecode into in-memory `Rule`s.
2. **Prepare inputs:** push files, buffers, or stdin readers via CLI flags or `Engine.AddInput`.
3. **Execute:** `Engine.Start` launches the dual-stream matcher. It continually expands rules on demand via `ResolveE`, binds variables, and emits symbols to the RHS output buffer or downstream `GrammarIO` targets.
4. **Trace/visualise:** enable tracing flags to inspect symbol comparisons, variable scopes, replacement scopes, or view the Unicode lm-diagram in real time.

Armed with this flow you can read a bytecode sequence, predict what structures the loader will build, and reason about how the engine rewrites the input stream as it runs.
