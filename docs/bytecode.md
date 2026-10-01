# Language Machine Bytecode Specification

This document describes the textual bytecode (`.lm`, historically `.lmr`) that `internal/machine/loader.go` loads. The loader is a small stack machine: its opcodes build `Rule` objects, and the engine interprets those rules when it resolves mismatches. The spec is detailed enough to emit `.lm` files without reading the loader. For how the `lmn` compiler produces this format, see `lm/07-compilation-and-bytecode.md`.

Code references name functions and types rather than line numbers. Use `grep` or your editor to find them.

## 1. Token stream

* **Lexing.** `Loader.Load` tokenises the whole text with one regular expression:

  ```
  ([().reAtpbPgGVsawz])|(.:\S*)|#[^\n]*\n|(\S)|\s*
  ```

  Each token is one of:
  * a single-character opcode from `( ) . r e A t p b P g G V s a w z`,
  * a two-part opcode `X:value`, where `X` is one character and `value` runs up to the next whitespace,
  * a comment from `#` to the end of the line. This covers a `#!` shebang header, which lets compiled grammars run as scripts.

  Whitespace is skipped. Any other single character is caught by `(\S)` and rejected with a `bad load format` error, except `E` and bare `B` (see §3). `T`, which `lmn2mbe` would emit for `top`, is rejected with `unsupported opcode`; no lmn source produces it.
* **Arguments.** The text after the colon is URL-decoded (`url.PathUnescape`, so `+` stays `+`) and then C-unescaped (`Loader.MStr` → `conv.Unescape`). Literals can therefore contain `%XX` escapes (spaces must be written `%20`) as well as `\n`-style escapes.

## 2. Stack machine model

The loader keeps one operand stack and a `count` register, which holds how many values have been pushed since the last `(`. Most opcodes push an element or rewrite the top one. `r` pops five values and defines a rule.

### 2.1 Sequence construction

`(` (`Loader.O` → `BMark`) pushes the current `count` onto the stack and resets it to zero. `)` (`Loader.C`) takes the `count` most recent operands, restores the saved count (`EMark`), and pushes the operands wrapped in a `Str` element. Sequences can nest, so a sequence can hold a sub-sequence (for example the body of a `repeat`).

### 2.2 Rule layout

When `r` runs, the stack must hold these five values, from bottom to top:

1. Grammar symbol, from `m:<name>`.
2. Priority word, from `L`, `R`, `B` or `M`.
3. RHS offset, from `n:0` or `n:1`.
4. LHS body, a `Str` built with parentheses. This is the pattern the rule matches against the input.
5. RHS body, another `Str`. This is what the rule substitutes.

`Loader.r` pops the five values and calls `Engine.AddRule`. That selects the rule's grammar by name (`Selector.Select`); the first grammar defined becomes the initial grammar. It then calls `Grammar.DefineRule`, which:

1. unpacks the bodies with `ToBody()`;
2. computes the rule's length as the sum of the LHS elements' `Weight()`;
3. files the rule under the token pair (LHS initial, RHS initial), through the first LHS element's `AddRule` hook, which normally calls `Grammar.Add`.

`Grammar.Add` keeps each group as a linked list ordered by descending length. A newer rule goes before older rules of the same length, which is why `-add` can override earlier rules.

### 2.3 Priorities and offsets

| Opcode | Stored value | Meaning |
| --- | --- | --- |
| `L:x` | `2*x` | left-associative: can start only in a context of lower priority |
| `R:x` | `2*x+1` | right-associative: can also start at the same level |
| `B:x` | `2*x \| BRACKET` | bracket: can always start, and opens a new priority level |
| `M:x` | `PRIMASK \| BRACKET` | maximal: can always start, and its context priority is `PRIMASK`, at which `Engine.ResolveE` starts no further rule. The level `x` is ignored. |

`Rule.Allow` decides whether a rule can start in a context, and `Rule.Cxtpri` gives the priority of the context a rule starts. A priority word of `0` (a rule under a bare `.name()` selector) always starts and inherits the enclosing context's priority. The tracer decodes the word with `priValue` and `priAssoc` for display.

The offset `n:<k>` is stored as `Rule.offset`. It becomes the initial `codeIndex` of the RHS mode (`Rule.Newrhs` → `NewRHModeFromParamsAndScope`). With `n:1` the first RHS element (the right initial) only files the rule and is not substituted, as in `<- eof - …`. A rule whose offset reaches the end of its RHS substitutes nothing.

## 3. Instruction reference

| Opcode | Form | Stack effect | Description |
| --- | --- | --- | --- |
| `m` | `m:<name>` | push | Non-terminal symbol, unique in `nonTerminalSymbols`. Predefined names such as `eof`, `out`, `repeat` or `toNum` resolve to their builtin element. |
| `f` | `f:<name>` | push | Function or operator symbol from `functionSymbols` (`+`, `==`, `sel`, `apply`, …). |
| `c` | `c:<text>` | push × n | One terminal `Chr` per rune of the decoded text. `c:` with no text pushes nothing. |
| `d` | `d:<name>` | push | Quoted non-terminal (`Quote`), a symbol used as a value rather than matched. |
| `l` | `l:<class>` | push | Lexical class such as `[0-9]` or `[^\n]`, compiled by `NewLexFromEngine`. |
| `v` | `v:<name>` | push | Variable symbol (`VarSym`), unique in `varSymbols`. |
| `L` `R` `B` `M` | `X:<n>` | push | Priority word; see §2.3. |
| `n` | `n:<n>` | push | Numeric literal, integer or real (`n:2.5`). Also used for the RHS offset. |
| `(` `)` | literal | restructure | Begin and end a sequence; see §2.1. |
| `z` | literal | push | The predefined nil symbol `-`. It pads the start of either side and is skipped by the matcher. |
| `.` | literal | push | The drop primitive (`DropF`), which clears the operand stack. |
| `r` | literal | pop 5 | Define a rule; see §2.2. |
| `e` | literal | rewrite | `each Name`: replaces the top element with `EachRef(top)`, which substitutes every value bound to that name in the current context. |
| `A` | literal | rewrite | `all Name`: replaces the top element with `AllRef(top)`, which substitutes every value bound to that name along the whole variable chain. |
| `E` | literal | push | `each (expr)`: `EachX` pops a value at run time and acts as `each` for the variable that the value names (`Engine.varKey`). A value that names no variable substitutes nothing. |
| `B` | literal | push | `all (expr)`: `AllX`, the same for `all`. `B:<n>` is a priority word. |
| `t` | literal | push | The take primitive `%` (`TakeF`), which grabs the matched symbol onto the operand stack. |
| `b` | literal | push | The bind primitive `:` (`BindF`). |
| `p` / `P` | literal | rewrite + push | Replaces the top element with `GetXF(top)` and pushes `bind`. `v:X p` binds what was matched to `X`. The two opcodes are identical. |
| `g` | literal | push | `GetF`: at run time, pushes the next element of the code vector onto the operand stack as a literal. |
| `G` | literal | rewrite | Replaces the top element with `GetXF(top)`, which pushes it onto the operand stack at run time (no bind). Used for call arguments. |
| `V` | literal | rewrite | Replaces the top element with `GetVF(top)`, which pushes a reference (`LMRef`) to the variable at run time. |
| `s` | literal | push | The `str` primitive (`StrF`). It has no effect at run time. |
| `a` | literal | push | The `act` primitive (`ActF`). Acting on it stops the run with an error; no known compiler output relies on it. |
| `w` | literal | push | `NewVar`: at run time pops a value and a name and creates a variable in the current scope. |

## 4. Runtime-visible constructs

* **Variables.** On the LHS, `v:A p` pushes the variable and then `:`. When the pattern matches, `BindF.Match` binds the matched value to `A` (`Engine.BindUvar`, `BindXvarE`, `BindTvar`). On the RHS, `v:A` substitutes the value of the nearest binding of `A` that is in scope, and `v:A V` pushes a reference to it as an operand.
* **Take.** `t` grabs the element just matched onto the LHS operand stack. When a rule whose LHS grabbed operands starts its RHS, `Engine.PushRhx1` stores the grabbed row as the `%` variable, so the RHS can hand it on.
* **Each / all.** `e` and `A` turn a variable symbol into `EachRef` or `AllRef`, which substitute one `RFMode` per matching binding (`Engine.EachRef`, `Engine.AllRef`).
* **Loops.** `( <body> ) G f:loop` repeats the body in an `RPMode` until an `f:test` in it pops a false value. `for` compiles to `I ( <E> f:test B ) G ( N ) G f:for`, which joins body and step into one loop body and records where the step starts. `f:break` ends the innermost loop. `f:continue` resumes it at the step of a `for` or at the test of a `while`. Both return out of any `if` blocks inside the loop, and stop the run with an error outside a loop (`loopMode` does not look past the start of a rule side).
* **foreach.** `<K> <V> <E> ( <body> ) G f:foreach` loops over the array `E`. `<K>` and `<V>` are variable references (`v:X V`), and `<K>` is `v:null G` when there is no key variable. `Foreachf` copies the array's keys (`AArray.Keys`, in the order they were added) and puts a `ForeachStep` in front of the body. The step assigns the next key and value, or ends the loop after the last key, so `continue` goes to the next key. `AArray.Set` records a key's position the first time the key is set. An array literal adds its items in the order they are written.
* **Rule values.** `rule (G, P) { lhs <- rhs }` compiles to `G P n:N ( lhs ) G ( rhs ) G f:rule`. At run time `f:rule` pops the five operands of `r` and defines the rule. `P` is a priority word encoded as in §2.3 (so `R:2000` is `4001`). The value is the grammar symbol.
* **Calls.** A builtin call such as `format("%d", 3)` compiles to `v:format G f:args d:%25d G n:3 G f:fun`. The name is pushed first, `f:args` pushes a mark, and the arguments are pushed with `G` or `V`. `f:fun` then collects the name and arguments (`Stream.ToArgv`), calls the Go function through `LMExternal.Call`, and pushes the result. `f:apply` substitutes the value on top of the stack. Operators such as `f:+` or `f:==` work directly on the operand stack.

## 5. Example

A rule from the calculator used in `internal/machine/machine_test.go`:

```
m:calc L:0 n:0 ( z m:x v:N p ) ( m:result c:result:%20 v:N c:%5Cn m:eom ) r
```

* `m:calc` selects the `calc` grammar. `L:0` makes it left-associative at level 0, and `n:0` starts the RHS at its first element.
* The LHS `( z m:x v:N p )` matches an `x` and binds the element produced with it to `N`. The leading `z` pads the left side.
* The RHS substitutes the non-terminal `result`, the text `result: ` (the `%20` is a space), the value of `N`, a newline (`%5Cn` decodes to `\n` and is then unescaped) and `eom`.
* `r` defines the rule. It is filed under the pair (`-`, `result`) and is found when the goal is `result` and the input is something no other rule handles.

## 6. Checklist for emitting bytecode

* Keep every `X:value` token free of whitespace. URL-encode spaces, newlines and `%`.
* Push the five rule operands in the order of §2.2 before each `r`.
* Choose `L`, `R`, `B` or `M` for the intended associativity.
* Put LHS and RHS bodies in parentheses. Nested parentheses make sub-sequences.
* Use the usual idioms: `v:X p` to bind on the LHS, `v:X` or `v:X V` on the RHS, and `( v:X e )` / `v:X A` for each / all.
* Put a lexical class in a single `l:` token, and let `NewLexFromEngine` handle ranges, negation and escapes.

Load the result with `Engine.LoadFromString` (replacing any loaded grammars) or `Engine.LoadFromStringReset(text, false)` (adding to them, as `-add` does).
