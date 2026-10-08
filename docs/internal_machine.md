# Language Machine 2 – Internal Execution Flow

This document explains how the Go runtime loads LM bytecode and runs it. Read it together with `bytecode.md`, which covers the format, and `lm/02-execution-model.md`, which covers the intended semantics.

Naming convention: the engine's **LHS** stream produces what is *expected* (the goal and the patterns of the rules being matched). Its **RHS** stream produces what is *there* (the input, and the substitutions of rules that have matched). Inside a `Rule`, `lhs` is the pattern to match and `rhs` is the substitution.

## 1. Loading bytecode

### 1.1 Tokenisation and dispatch
`Loader.Load` tokenises the whole text with one regex (see `bytecode.md` §1). It skips whitespace and `#` comments, and prints each token when the `LOAD` trace flag (`-trace b`) is set. The first character of a token selects the handler: `L` → `Loader.L`, `c` → `loader.pushChars`, `(` → `loader.openList`, and so on. Payloads of `X:value` tokens go through `Loader.decode` (URL decode, then C unescape).

### 1.2 Stack machine
The loader keeps one operand stack and a counter of the values pushed since the last `(`. Opcodes push elements (`newSym`, `newChr`, `newVarSym`, `newLex`, …) or wrap the top element (`p`, `G`, `V`, `e`, `A`). `(` saves the counter (`openMark`), and `)` wraps the pending operands in a `Str` and restores the counter (`closeMark`).

### 1.3 Rule assembly
`r` pops five operands: grammar symbol, priority word, RHS offset, LHS body and RHS body. It then calls `Engine.addRule` → `Grammar.defineRule` → `Element.addRule` → `Grammar.Add`. Along the way the runtime:

1. selects or creates the grammar by name (`Selector.Select`). The first grammar defined becomes `Engine.initGrammar`;
2. computes the rule's length as the sum of the LHS elements' weights;
3. files the rule under the pair (LHS initial token, RHS initial token), in a list ordered by descending length, with newer rules first among equals.

## 2. Engine bootstrapping

1. `NewEngine` creates the symbol dictionaries (`Dict`), the `rzBuffer`, the LHS and RHS streams with their root modes (`lzMode`, `rzMode`), the root contexts, and the external function table (`External`).
2. The first load calls `defineSymbols`, which creates the predefined symbols once: `start`, `eof`, `repeat`, `option`, `anything`, the output symbols (`out`, `err`, `uri`, `urd`), the conversion symbols (`toNum`, `toSym`, …) and the operators. Loaded rules hold these objects and are compared by identity, so a later `-add` reuses them.
3. `LoadFromString` replaces the grammar table. `LoadFromStringReset(text, false)` adds to it. The same `loader` is reused for every load; it starts each load with an empty stack. A load that fails (an `*Error` with the line and column) leaves the rules as they were: the grammar table is saved before loading (`selector.save`, cheap because rule groups are never changed in place) and restored on failure.

## 3. Streams, modes and contexts

- **Streams.** A `Stream` is a register set: the current mode, the current symbol, the code vector and index, a persistent operand stack (`opStack`), and the head of the variable chain.
- **Modes.** A `GenMode` is a generator of symbols with a link to the mode it was pushed over (`stackMode`). `advance` produces the next symbol, usually by `Stream.act`, which acts on the next code element. A new mode saves the stream registers it replaces (current symbol, code vector and index), and `exit` puts them back and returns the mode below; the operands stay, so a mode leaves its results on the stack. Backtracking uses value snapshots instead (`modeSnap`, below). The mode types are:
  - `lzMode` and `rzMode`, the roots. `lzMode` produces the goal `eof`, and `rzMode` reads input through `rzBuffer`.
  - `lhMode`, for the pattern of a rule being matched, and `rhMode`, for the substitution of a rule that has matched.
  - `rpMode` for loop bodies (`loop`, `for`, `foreach`), `rfMode` for variable references, and `stMode` for a sequence substituted in place (a bound value, a `Str` acted on, the chosen branch of `if` or `sel`).
- **Contexts.** A `context` records a rule application: its `state` (the grammar, the goal and input symbols at the mismatch, and the input position), its rule, its `priority`, its nesting depth and its variable chain. `lhsContext` is the rule currently being recognised; the context of a substitution being read is the one of the right stream's `rhMode`.

Go has no virtual dispatch through embedding. Elements therefore embed `selfPointing[Element]` (through `genericElement`) and call overridable methods through `Self()`; their constructors set the self pointer with `makeSelf`/`reSelf`. Modes, contexts, variables' scopes and I/O handlers are plain structs.

## 4. The matching loop

1. `Engine.Start` falls back to stdin if no input was queued. It selects the initial grammar, dumps it when `-trace G` is set, and makes `start` the first RHS symbol if a rule for it exists. It then runs `match`. The exit status is `0` if the match succeeded and no `flagError` was produced.
2. `match` advances both streams until each has a current symbol (`Stream.modeAdvance`). It stops with success when either stream runs out of modes.
3. The nil symbol `-` is skipped on either side.
4. Otherwise it calls `lhsSymbol.match(engine, rhsSymbol)`. Most elements compare by identity and call `matchedWith`, which clears both current symbols. Builtins (take, bind, output symbols, conversions, `repeat`, …) implement their own `match`. If the symbols do not match, `genericElement.match` calls `resolve`. If that fails too, `match` returns false, and the caller backtracks.

## 5. Mismatch resolution

`resolve(l, r)` is the core of the machine. `l` is the goal and `r` is the input symbol.

1. It takes the priority of the mismatch from the context. If the context is closed (`priority.closed`, the context of an `M:` rule), nothing can nest, and resolution fails.
2. It tries four groups of rules, in this order. Each lookup is `Grammar.Get(rule LHS initial, rule RHS initial)`:
   1. `Get(r, l)`: rules that start with the input symbol and produce the goal;
   2. `Get(r, -)`: bottom-up rules that start with the input symbol, whatever the goal;
   3. `Get(-, -)`: rules that do not care about either;
   4. `Get(-, l)`: top-down rules for the goal, whatever the input.
3. When it reaches the first group that has candidates, it creates a `state` for the new context and takes a `checkpoint`: snapshots of both streams (`modeSnap`) and the left context. The state and the checkpoint are shared by the later groups. For each group, `resolveGroup` walks the rules (longest first):
   - It skips rules whose priority does not allow them in this context (`priority.allows`).
   - A one-element LHS matches at once. `pushReplacement` pushes its substitution, if it has one.
   - Otherwise it pushes an LHS context (`newLHContext`) and an `lhMode` for the rest of the pattern, and recurses into `match`. If that fails, it rolls back to the checkpoint (`rollback`) and tries the next rule.
   - When a rule matches, `pushMatched` stores any grabbed operands as the `%` variable, pushes an RHS context, and pushes an `rhMode` for the substitution, starting at `Rule.offset`. `commit` restores the LHS mode, so the goal is matched again, now against the substituted symbols. The last match and the right stream's current symbol are deliberately not part of the checkpoint (see its comment).
4. If no rule in any group applies, `resolve` returns false, and the enclosing `match` fails. Its caller then rolls back to its own checkpoint and tries its next candidate.

`-max-depth` limits the context nesting (checked in `resolveGroup`), and `-max-repeat` limits `repeat` iterations. Exceeding either stops the run with an error.

Errors caused by the rules or the input (bad bytecode, exceeded limits, unreadable includes, buffer overflow) are raised deep inside matching with `fail`, which panics with a `*machine.Error`. `loader.load` and `Engine.Start` recover it with `defer catch(&err)` and return it as an error; `Start` prefixes the input position. Any other panic is a bug in the machine and is not recovered.

## 6. Substitution helpers

- **Binding.** `bindF.match` handles `:` in its three forms: bind to the last matched element (`bindXvar`), bind a take (`bindTvar`), and bind to an explicit RHS value (`bindUvar`). Variables are created with `scopeHolder.makeVar`, and each is linked into both the scope chain and the all-variables chain.
- **Take.** `takeF.match` pushes the matched element (or the `%` row) onto the LHS operands (`pushX`, `pushR`, `takeTvar`).
- **References.** On the RHS, a variable symbol is resolved with `Engine.deref`/`theRef` into an `rfMode` over the value. `eachRef`/`allRef` push one `rfMode` for each matching binding.
- **Repeat and option.** `repeat`, `repeatN` and `option` in a pattern all call `Engine.repeat` with a limit of 0 (none), N or 1. It reruns the body in a fresh `lhMode` with `match` until an iteration fails. A failed iteration gives back its input and drops what it grabbed or bound (`markLhs`/`releaseLhs`).
- **Control statements** (`control.go`). `if` and `sel` push an `stMode` over the chosen branch. `f:loop`, `f:for` and `f:foreach` push an `rpMode` over the loop body, and `f:test` ends it (`Ends`) when it pops a false value. `f:for` appends the step to the body and records where it starts (`rpMode.next`), and `f:foreach` puts a `foreachStep` in front of the body. `f:break` and `f:continue` find the innermost `rpMode` with `loopMode`, which fails at the first `lhMode`, `rhMode` or root mode, so a loop in another rule is never found. `break` returns from that mode. `continue` returns from the modes inside it (the `if` blocks) and resumes it at `next`. `f:rule` (`rulef`) pops the five operands of `r` and calls `Engine.addRule`, so rules can be defined while the rules run.
- **Input buffer.** `rzBuffer.getChr` gives every RHS position a stable index, so backtracking can re-read input. The buffer doubles in size up to `-buffer` (`SetBuffer`). After that it is circular, and backtracking too far stops the run with a backtracking overflow error.

## 7. Inputs, outputs and externals

- Inputs are `Input` values on a stack: `readerInput` (standard input, `SetStdin`, or any reader) and `stringInput` (`-input` and whole files). `AppendInput` queues command-line sources in order. `addInput`, used by the `include` builtin, pushes a source that is read to its `eof`, after which reading returns to the previous source.
- Output (`out`, `uri`, `urd`), traces and diagrams go through the engine's buffered writer (`SetOutput`, default stdout), and `err` goes to `SetErrOutput` (default stderr). `Start` flushes the buffer when it returns, `err` flushes it before writing, and stdin input flushes it before it blocks, so interactive grammars answer at once.
- Builtins that grammars call as functions are the `builtins` table (`extension.go`, with helpers in `builtin.go`); each engine's `External` starts as a copy of it, and `External.Set` adds or replaces functions (lm2's Go functions).

## 8. Tracing and the diagram

- `SetTraceFlag` creates the `Tracer` on first use and sets the requested bits. The diagram flags (`-trace D`/`d`) also turn on `MISMATCH`, `SYMBOLS` and `CXSCOPE` and create the `Diagram`, which uses the width from `-dwidth`, so `-dwidth` must come first.
- During matching the tracer reports symbol comparisons, mismatches (`Resolve`, `Back`), context changes (`ruleScope`), bindings and repeat iterations. `Diagram` draws them as the Unicode lm-diagram (see `lm/06-lm-diagram.md`).

## 9. Putting it together

1. **Load:** `LoadFromString` or `LoadFromStringReset` turns bytecode into `Rule`s.
2. **Queue inputs:** `AppendInput` (or the CLI's `-input`, `-stdin` and positional files).
3. **Run:** `Engine.Start` runs the matcher. It resolves mismatches with `resolve`, binds variables and writes output as symbols reach the output symbols.
4. **Inspect:** enable trace flags, or `-trace D`, to watch the process.
