# Language Machine Bytecode Specification

This document summarizes the textual bytecode understood by `internal/machine/loader.go`.  The loader feeds the parsing `Engine` with stack-based instructions that assemble `Rule` objects which are later interpreted while resolving mismatches.  The goal of this specification is to make it possible to emit `.lm` or `.lmr` files without relying on the legacy Go code base.

## 1. Token stream

* **Lexing.** Source is tokenised with the regular expression `([().reAtpbPgGVsawz])|(.:\S*)|#[^\n]*\n|\s*` (`internal/machine/loader.go:219-299`).  Therefore every token is either
  * a single-character opcode from the set `() . r e A t p b P g G V s a w z`,
  * a two-part opcode written `X:value` where `X` is any other leading letter, or
  * a comment starting with `#` and running to the end of the line.
  Whitespace is skipped.
* **Arguments.** The text after the colon in `X:value` tokens is URL-decoded and then unescaped (`Loader.MStr`, `internal/machine/loader.go:215-217`).  This allows both `%` escape sequences and C-style backslash escapes inside bytecode literals.

## 2. Stack machine model

The loader maintains a single operand stack plus a `count` register that tracks how many values have been pushed since the last `(`.  Every opcode either pushes a new `Element` onto the stack or rewrites the top item.  The defining operation `r` pops the top five values and converts them into a `Rule` (`internal/machine/loader.go:74-150`).

### 2.1 Sequence construction

`(` invokes `BMark` which stashes the previous `count` on the stack and resets the counter (`internal/machine/loader.go:84-138`).  `)` collects the `count` most-recent operands, restores the previous count, and wraps the collected slice in a `Str` element (`internal/machine/element.go:997-1071`).  These lists represent either the left-hand side (LHS) or right-hand side (RHS) bodies of a rule.

### 2.2 Rule layout

When `r` (or its alias `e`) executes the stack must contain the following values, from bottom to top:

1. Grammar symbol – inserted via `m:<name>`.
2. Priority number – produced by `L`, `R` or `B`.
3. RHS offset – usually `n:0` or `n:1`.
4. LHS body – a `Str` built with parentheses.
5. RHS body – another `Str`.

`r` pops those five elements and forwards them to `Engine.DefineElements`, which eventually calls `Grammar.Define` (`internal/machine/engine.go:325-332`, `internal/machine/grammar.go:317-326`).  The `Str` bodies are unpacked with `ToBody()` and stored inside a `Rule`.  The third operand becomes `Rule.offset`, i.e. the initial instruction pointer for the RHS when the rule fires (`internal/machine/mode.go:236-286`).

### 2.3 Priorities and offsets

`L:x`, `R:x` and `B:x` encode the priority field that drives mismatch resolution.  Left- and right-associative priorities occupy the low bits, while `B` sets the `BRACKET` flag before storing the value as `2*x` (`internal/machine/loader.go:101-111`).  `Grammar.Priassoc` and `Privalue` decode these fields when rules are compared (`internal/machine/grammar.go:302-315`).

The offset (`n:<k>`) is simply stored as `Rule.offset`.  It becomes the initial `codeIndex` when the RHS is turned into a `RHMode` (`internal/machine/mode.go:236-286`) so that lexical rules can skip already-consumed symbols by starting part-way through the RHS.

## 3. Instruction reference

| Opcode | Form | Stack effect | Description |
| --- | --- | --- | --- |
| `m` | `m:<name>` | push | Non-terminal symbol. Uses `nonTerminalSymbols` and creates or reuses a `Sym` (`internal/machine/loader.go:200-213`). |
| `f` | `f:<name>` | push | Function/operator symbol looked up in `functionSymbols` (same file). |
| `c` | `c:<text>` | push* | Pushes each decoded rune as a terminal `Chr` (`internal/machine/loader.go:117-123`). Multiple characters yield multiple pushes. |
| `d` | `d:<name>` | push | Quoted non-terminal via `NewQuote`, so it compares by token rather than by identity (`internal/machine/loader.go:124-126`). |
| `l` | `l:<class>` | push | Lexical class compiled by `NewLexFromEngine` (`internal/machine/loader.go:207-209`, `internal/machine/element.go:2051-2140`). |
| `v` | `v:<name>` | push | Variable symbol (`VarSym`) that can later be bound (`internal/machine/element.go:1336-1359`). |
| `L` | `L:<n>` | push | Priority word for left-associative rules (`internal/machine/loader.go:101-104`). |
| `R` | `R:<n>` | push | Priority word for right-associative rules (`internal/machine/loader.go:105-107`). |
| `B` | `B:<n>` | push | Priority with the `BRACKET` flag set (`internal/machine/loader.go:109-111`). |
| `n` | `n:<n>` | push | Plain numeric literal, typically used for the RHS offset (`internal/machine/loader.go:113-115`). |
| `(` / `)` | literal | restructure | Begin/end of a list; see §2.1. `(` saves the current element count; `)` wraps collected operands in a `Str`. |
| `z` | literal | push | Pushes the predefined `nil` symbol `-` (`internal/machine/engine.go:138-176`). Used for epsilon matches and padding on either side of rules. |
| `.` | literal | push | Pushes the builtin drop function (clears operand stack) (`internal/machine/loader.go:262-264`, `internal/machine/element.go:2147-2156`). |
| `r`/`e` | literal | pop 5 | Defines a rule; see §2.2. |
| `A` | literal | rewrite | Pops the top element and replaces it with `AllRef(top)`; at runtime this scans all variables whose key matches `top` (`internal/machine/loader.go:151-153`, `internal/machine/engine.go:977-995`). |
| `t` | literal | push | Pushes the builtin `TakeF` symbol `%` (`internal/machine/loader.go:167-169`, `internal/machine/element.go:1378-1405`). |
| `b` | literal | push | Pushes the builtin bind function `:` (`internal/machine/loader.go:171-173`). |
| `p`/`P` | literal | rewrite+push | Pops the top element, wraps it in `GetXF`, then pushes `bind`.  This is the idiom `v:X p` used to bind matches to variables (`internal/machine/loader.go:155-165`, `internal/machine/element.go:1675-1697`, `1408-1446`). |
| `g` | literal | push | Pushes the builtin `GetF` instruction that reads the next literal embedded in the RHS (`internal/machine/engine.go:170-175`, `internal/machine/element.go:1599-1618`). |
| `G` | literal | rewrite | Pops the top element and wraps it in `GetXF` without adding a bind (`internal/machine/loader.go:183-185`). |
| `V` | literal | rewrite | Pops the top element and wraps it in `GetVF`, i.e. pushes a runtime `LMRef` to the named variable (`internal/machine/loader.go:187-189`, `internal/machine/element.go:1724-1746`, `internal/machine/variable.go:323-341`). |
| `s` | literal | push | Pushes the predefined string constructor (`str`) so captured material can be re-emitted (`internal/machine/loader.go:191-193`, `internal/machine/element.go:1830-1843`). |
| `a` | literal | push | Pushes the `act` primitive used for host callbacks (`internal/machine/loader.go:195-197`). |
| `w` | literal | push | Pushes the `newvar` constructor so that two top operands (name/value) become a scoped variable (`internal/machine/loader.go:199-205`, `internal/machine/element.go:1219-1235`). |
| `X` | literal | push | (Currently mapped from `.` in source) pushes the drop primitive that clears captured operands (`internal/machine/loader.go:179-181`). |

\* `c` pushes one `Chr` per rune; emitting a contiguous terminal string therefore requires repeating `c` with that text inside a pair of parentheses if you want the characters to live inside a `Str` body.

## 4. Runtime-visible constructs

Certain opcodes map to complex runtime behaviours:

* **Variables.** `v:<name>` emits a `VarSym`.  On the LHS, the usual idiom `v:A p` pushes a getter for the symbol and then the bind primitive `:`.  When the rule matches, the bind pulls the LHS variable from the operand stack and stores the captured RHS value (`internal/machine/element.go:1408-1446`).  On the RHS, `v:A V` emits a reference that reads the current value of `A` (`internal/machine/element.go:1724-1746`).
* **Capture / replay.** `t` pushes the `%` symbol (`TakeF`) which causes the engine to capture the text matched most recently before handing control to `bind` or `append` (`internal/machine/element.go:1378-1405`).
* **AllRef / EachRef.** `A` converts a variable symbol into `AllRef`, so when it runs the RHS iterates through every historical value stored under that key (`internal/machine/engine.go:977-995`).
* **New variables.** `w` allows the RHS to synthesize scoped variables by popping `<value, name>` and invoking `MakeVar` on the surrounding `ScopeHolder` (`internal/machine/element.go:1219-1235`).

## 5. Example

The first interesting rule in `calc.lm` illustrates several of the opcodes (`calc.lm:3-4`):

```
m:calc L:0 n:0 ( z m:x v:N p ) ( m:result c:result:%20 v:N c:%5Cn m:eom ) r
```

* `m:calc` selects the `calc` grammar; `L:0` gives it left-associative priority 0; `n:0` tells the RHS to start at element 0.
* The LHS sequence `( z m:x v:N p )` matches the symbol sequence `-, x, N` where `N` is bound by `p`.
* The RHS sequence emits the non-terminal `result`, literal text `result: ` (decoded from `%20`), the bound variable `N`, a newline, and the `eom` symbol.
* `r` finalises the definition.

When the loader reaches the end of the line, `r` pops the five operands and `Grammar.Define` builds a `Rule` whose `lhsEffectiveInitialSymbol` is the token of the first LHS element (`internal/machine/grammar.go:317-326`).  At runtime the engine looks up rules by this token pair and executes the RHS with the supplied offset, using `NewRHModeFromParamsAndScope` to set the starting `codeIndex` (`internal/machine/mode.go:236-286`).

## 6. Implementation checklist

* Emit tokens exactly as described in §1 so that the regex recognises them.
* Ensure every rule contributes five operands in the order described in §2.2 before emitting `r`.
* Choose priorities with `L`, `R` or `B` to match the intended associativity (`internal/machine/loader.go:101-111`, `internal/machine/grammar.go:302-315`).
* Remember that the LHS and RHS bodies are `Str` objects; keep related symbols between parentheses.
* Use idiomatic combos for variable handling: `v:X p` on the LHS, `v:X V` on the RHS, `v:X A` when you need all historical bindings (`internal/machine/element.go:1336-1746`, `internal/machine/engine.go:977-995`).
* For lexical sets encode the pattern inside a single `l:<class>` token and let `NewLexFromEngine` interpret ranges and escapes (`internal/machine/element.go:2051-2140`).

Following this specification yields bytecode that the existing Go engine can consume via `Engine.LoadFromString*` (`internal/machine/engine.go:334-366`), ensuring forward compatibility with the rest of the runtime.
