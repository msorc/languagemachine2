# Language Machine 2 – Internal Execution Flow

This document explains how the Go runtime loads LM bytecode and runs it. Read it together with `bytecode.md`, which covers the format, and `lm/02-execution-model.md`, which covers the intended semantics.

Naming convention: the engine's **LHS** stream produces what is *expected* (the goal and the patterns of the rules being matched). Its **RHS** stream produces what is *there* (the input, and the substitutions of rules that have matched). Inside a `Rule`, `lhs` is the pattern to match and `rhs` is the substitution.

## 1. Loading bytecode

### 1.1 Tokenisation and dispatch
`Loader.Load` tokenises the whole text with one regex (see `bytecode.md` §1). It skips whitespace and `#` comments, and prints each token when the `LOAD` trace flag (`-trace b`) is set. The first character of a token selects the handler: `L` → `Loader.L`, `c` → `Loader.c`, `(` → `Loader.O`, and so on. Payloads of `X:value` tokens go through `Loader.MStr` (URL decode, then C unescape).

### 1.2 Stack machine
The loader keeps one operand stack and a counter of the values pushed since the last `(`. Opcodes push elements (`NewSym`, `NewChr`, `NewVarSym`, `NewLexFromEngine`, …) or wrap the top element (`p`, `G`, `V`, `e`, `A`). `(` saves the counter (`BMark`), and `)` wraps the pending operands in a `Str` and restores the counter (`EMark`).

### 1.3 Rule assembly
`r` pops five operands: grammar symbol, priority word, RHS offset, LHS body and RHS body. It then calls `Engine.AddRule` → `Grammar.DefineRule` → `Element.AddRule` → `Grammar.Add`. Along the way the runtime:

1. selects or creates the grammar by name (`Selector.Select`). The first grammar defined becomes `Engine.initGrammar`;
2. computes the rule's length as the sum of the LHS elements' weights;
3. files the rule under the pair (LHS initial token, RHS initial token), in a list ordered by descending length, with newer rules first among equals.

## 2. Engine bootstrapping

1. `NewEngine` creates the symbol dictionaries (`Dict`), the `RZBuffer`, the LHS and RHS streams with their root modes (`LZMode`, `RZMode`), the root contexts, and the external function table (`LMExternal`).
2. The first load calls `defineSymbols`, which creates the predefined symbols once: `start`, `eof`, `repeat`, `option`, `anything`, the output symbols (`out`, `err`, `uri`, `urd`), the conversion symbols (`toNum`, `toSym`, …) and the operators. Loaded rules hold these objects and are compared by identity, so a later `-add` reuses them.
3. `LoadFromString` replaces the grammar table. `LoadFromStringReset(text, false)` adds to it. The same `Loader` is reused for every load.

## 3. Streams, modes and contexts

- **Streams.** A `Stream` is a register set: the current mode, the current symbol, the code vector and index, a persistent operand stack (`OpStack`), and the head of the variable chain.
- **Modes.** A `GenMode` is a generator of symbols with a link to the mode it was pushed over (`stackMode`). `Advance` produces the next symbol, usually by `Stream.Act`, which acts on the next code element. `Save` snapshots the stream registers into a new mode. `Restore` puts them back, including the operands, and `Return` puts them back without the operands. The mode types are:
  - `LZMode` and `RZMode`, the roots. `LZMode` produces the goal `eof`, and `RZMode` reads input through `RZBuffer`.
  - `LHMode`, for the pattern of a rule being matched, and `RHMode`, for the substitution of a rule that has matched.
  - `RPMode` for `repeat` bodies, `RFMode` for variable references, and `STMode` for single-element substitutions made by bind.
- **Contexts.** A `Context` records a rule application: its `State` (the grammar, the goal and input symbols at the mismatch, and the input position), its rule, its priority, its nesting depth and its variable chain. `lhsContext` is the rule currently being recognised, and `rhsContext` is the one whose substitution is being read.

Only one type is usually used for each role, but Go has no virtual dispatch through embedding. Elements, modes, contexts and variables therefore embed `SelfPointing[T]`, and they call overridable methods through `Self()`. Constructors must set the self pointer with `MakeSelf`/`ReSelf`. The I/O handlers use the same idea through `GramSystem.SetSelf`.

## 4. The matching loop

1. `Engine.Start` falls back to stdin if no input was queued. It selects the initial grammar, dumps it when `-trace G` is set, and makes `start` the first RHS symbol if a rule for it exists. It then runs `Match`. The exit status is `0` if the match succeeded and no `flagError` was produced.
2. `Match` advances both streams until each has a current symbol (`Stream.ModeAdvance`). It stops with success when either stream runs out of modes.
3. The nil symbol `-` is skipped on either side.
4. Otherwise it calls `lhsSymbol.Match(engine, rhsSymbol)`. Most elements compare by identity and call `Matched3E`, which clears both current symbols. Builtins (take, bind, output symbols, conversions, `repeat`, …) implement their own `Match`. If the symbols do not match, `GenericElement.Match` calls `ResolveE`. If that fails too, `Match` returns false, and the caller backtracks.

## 5. Mismatch resolution

`ResolveE(l, r)` is the core of the machine. `l` is the goal and `r` is the input symbol.

1. It computes the priority of the mismatch from the context. If the context is at maximal priority (`PRIMASK`), nothing can nest, and resolution fails.
2. It tries four groups of rules, in this order. Each lookup is `Grammar.Get(rule LHS initial, rule RHS initial)`:
   1. `Get(r, l)`: rules that start with the input symbol and produce the goal;
   2. `Get(r, -)`: bottom-up rules that start with the input symbol, whatever the goal;
   3. `Get(-, -)`: rules that do not care about either;
   4. `Get(-, l)`: top-down rules for the goal, whatever the input.
3. When it reaches the first group that has candidates, it creates a `State` for the new context and saves both stream modes (`Save`). The state and snapshots are shared by the later groups. For each group, `ResolveState` walks the list (longest rules first):
   - It skips rules that `Rule.Allow` rejects at this priority.
   - A one-element LHS matches at once. `PushRhx0` pushes its substitution, if it has one.
   - Otherwise it pushes an LHS context (`NewLHContextFromRule`) and an `LHMode` for the rest of the pattern, and recurses into `Match`. If that fails, it restores the saved contexts and modes and tries the next rule.
   - When a rule matches, `PushRhx1` stores any grabbed operands as the `%` variable, pushes an RHS context, and pushes an `RHMode` for the substitution, starting at `Rule.offset`. The LHS mode is restored, so the goal is matched again, now against the substituted symbols.
4. If no rule in any group applies, `ResolveE` returns false, and the enclosing `Match` fails. Its caller then restores its own snapshot and tries its next candidate.

`-max-depth` limits the context nesting (`Context.CheckDepth`), and `-max-repeat` limits `repeat` iterations. Exceeding either panics.

## 6. Substitution helpers

- **Binding.** `BindF.Match` handles `:` in its three forms: bind to the last matched element (`BindXvarE`), bind a take (`BindTvar`), and bind to an explicit RHS value (`BindUvar`). Variables are created with `ScopeHolder.MakeVar`, and each is linked into both the scope chain and the all-variables chain.
- **Take.** `TakeF.Match` pushes the matched element (or the `%` row) onto the LHS operands (`PushX`, `PushR`, `TakeTvar`).
- **References.** On the RHS, a variable symbol is resolved with `Engine.Deref`/`TheRef` into an `RFMode` over the value. `EachRef`/`AllRef` push one `RFMode` for each matching binding.
- **Repeat and option.** `repeat`, `repeatN` and `option` in a pattern all call `Engine.Repeat` with a limit of 0 (none), N or 1. It reruns the body with `Match` until an iteration fails. A failed iteration gives back its input and drops what it grabbed or bound (`markLhs`/`releaseLhs`). (`Engine.Repeatx` and the `RepxSym`/`OptxSym` elements are not reachable from loaded rules.)
- **Input buffer.** `RZBuffer.GetChr` gives every RHS position a stable index, so backtracking can re-read input. The buffer doubles in size up to `-buffer` (`SetBuffer`). After that it is circular, and backtracking too far panics with `BackTrackOverflow`.

## 7. Inputs, outputs and externals

- Inputs are `GrammarIO` values on a stack: `GramStdio` (stdin), `GramInputFile` and `GramInputBuffer` (`-input`). `AppendInput` queues command-line sources in order. `AddInput`, used by the `include` builtin, pushes a source that is read to its `eof`, after which reading returns to the previous source.
- Output (`out`, `uri`, `urd`), traces and diagrams go through the engine's buffered writer (`SetOutput`, default stdout), and `err` goes to `SetErrOutput` (default stderr). `Start` flushes the buffer when it returns, `err` flushes it before writing, and stdin input flushes it before it blocks, so interactive grammars answer at once.
- Builtins that grammars call as functions live in `LMExternal` (`extension.go`, with helpers in `builtin.go`). Register new ones with `LMExternal.Set`.

## 8. Tracing and the diagram

- `SetTraceFlag` creates the `Tracer` on first use and sets the requested bits. The diagram flags (`-trace D`/`d`) also turn on `MISMATCH`, `SYMBOLS` and `CXSCOPE` and create the `Diagram`, which uses the width from `-dwidth`, so `-dwidth` must come first.
- During matching the tracer reports symbol comparisons, mismatches (`Resolve`, `Back`), context changes (`RuleScope`), bindings and repeat iterations. `Diagram` draws them as the Unicode lm-diagram (see `lm/06-lm-diagram.md`).

## 9. Putting it together

1. **Load:** `LoadFromString` or `LoadFromStringReset` turns bytecode into `Rule`s.
2. **Queue inputs:** `AppendInput` (or the CLI's `-input`, `-stdin` and positional files).
3. **Run:** `Engine.Start` runs the matcher. It resolves mismatches with `ResolveE`, binds variables and writes output as symbols reach the output symbols.
4. **Inspect:** enable trace flags, or `-trace D`, to watch the process.
