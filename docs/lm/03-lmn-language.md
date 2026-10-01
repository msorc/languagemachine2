# The lmn Metalanguage

Sources: `lmn2xfe.html` (the lmn frontend, written in lmn, which is "the definitive statement of what the metalanguage looks like"), `guide.html`, `special_symbols.html`, `output_buffers.html`, `glossary.html`, `what_does_that_hyphen_mean_.html`.

## Source format

- **A line that starts with a space is code. Any other line is a comment.** Comment lines are treated as MediaWiki text, so an lmn file is also a wiki page. This gives literate programming almost for free, and the original website was generated from `.lmn` sources this way.
- Inside code, `// ...` comments run to the end of the line, and `/* ... */` comments can be **nested**.
- Whitespace, including `\r`, is ignored between tokens.

## Units of compilation

```
 .include "other.lmn";          // include another source file
 .name(prio)                    // following rules belong to grammar `name` with priority `prio`
 .name(prio) { rules... }       // selector scoped to a braced group
 .name(prio) : rule             // selector applied to a single rule
 lhs <- rhs ;                   // rule definition, terminated by ';'
```

### Priority specifiers

A priority is a number whose suffix gives the direction:

| Written | Meaning |
| --- | --- |
| empty, as in `.calc()` | no priority: the context inherits the enclosing priority (encoded as `L:0`) |
| `20` or `20L` or `20l` | left-associative 20 |
| `30R` or `30r` | right-associative 30 |
| `4B` or `4b` | bracket, level 4 |
| `B` or `b` alone | bracket 0 |
| `M` | maximal |

## Rules

```
initial elements  <-  initial [-] pattern ;
```

- The **left side** is `initial elements` and the **right side** is `initial [-] pattern`. The initial of each side is what files the rule under a class of mismatch (see [02-execution-model.md](02-execution-model.md#which-rules-are-relevant)).
- An initial is `-` (don't care), a `'quoted'` char sequence, a `"quoted"` symbol, a symbol name, a number, a lexical class, or `true`/`false`.
- The optional `-` directly after the right initial means "filed under this goal, but substitute only what follows".
- The frontend records the source line number of every rule, using `lineNo`.
- **Left side vs right side syntax:** side-effect actions can appear directly among the left-side *elements*. On the right side, which is a *pattern*, actions must be wrapped in braces. Without that rule, the boundary between rules would be ambiguous.

### Tokens

| Form | Meaning |
| --- | --- |
| `name`, `_name` (lower case or `_` first) | nonterminal symbol |
| `Name` (upper case first) | variable |
| `$name` | reference to the value of variable `name` in a pattern |
| `'abc'` | the sequence of terminal characters `'a' 'b' 'c'` |
| `"abc"` | **one** symbol whose text is `abc` |
| `.[0-9]`, `[a-z]` | lexical class (see below) |
| `12`, `0x1F`, `017`, `0b101`, `1.5e3` | number. D-style suffixes are accepted: `L`, `u`, `f`, `i`. |
| `true`, `false`, `null` | truth values and null |

Reserved words: `rule`, `option`, `repeat`, `each`, `all`, `var`, `if`, `else`, `for`, `foreach`, `while`, `break`, `continue`, `in`, `null`, `true`, `false`.

Quoted strings accept C escapes (`\a \b \f \n \r \t \v \\ \' \"`), `\xHH`, `\x"hex…"`, octal `\N`, `\NN` and `\NNN`, `\uHHHH`, `\UHHHHHHHH`, and `\&entity;`.

### Lexical classes

`.[...]` (or `[...]`) is a character class written in regex style, such as `.[a-zA-Z_]` or `.[^\n]`. Where it appears changes what it does:

- **As a left-initial:** it is shorthand for one rule per member. `.[0-9] ...` creates 10 rule entries that share the same body.
- **Elsewhere on the left side:** it matches itself and any member of the class.
- **On the right side:** it is just a symbol.

## Patterns and elements

A pattern is a possibly empty list of items. Braces `{ ... }` group a nested element sequence.

| Item | Meaning |
| --- | --- |
| symbol, `'chars'`, `"sym"`, number, class, `$Var` | match or produce that symbol or value |
| `:bind` | binding (see below) |
| `%` | acquisition (see below) |
| `repeat rest…` | on the left side, match the rest of the enclosing sequence **zero or more** times |
| `option rest…` | on the left side, match the rest of the enclosing sequence **zero or one** time |
| `{ elements }` | nested sequence. On the right side, this is how actions are embedded. |
| `{ elements <- pattern }` | **inject**: the rest is substituted directly into the input, as if by a rule with an empty left side |
| `{| alt || alt |}` | **nested rules** tied to the current context (see below) |
| `(Expr)` | **output buffer**: append to the variable or table cell `Expr` (see below) |
| `$(expr)` | evaluate the expression and use the value as a symbol or sequence, for example `$(Table[I])` or `$(format(...))` |
| `each Name`, `each (expr)` | every instance of variable `Name` visible in scope. Used after `repeat` has bound the same name many times: `- repeat item :X <- list :{ each X };`. In `each (expr)` the value of the expression names the variable, so `S = "X"; … each (S)` is `each X`. |
| `all Name`, `all (expr)` | all variable instances with the given name, across contexts. lmn2xfe used this for flattening nested structures, and the author wanted to phase it out. |
| `!` | prune: discard variables created since this rule started. lmn2xfe uses it after each compilation unit. |
| `;` | empty element (separator) |

Both `repeat` and `option` can succeed without consuming anything. To require at least one item, write `- item repeat item <- atLeastOneItem;`.

## Binding with `:`

`:` is an anonymous symbol: it must appear on both sides or a mismatch results. Its behaviour depends on which side it is on and what follows it:

| Side | Form | Meaning |
| --- | --- | --- |
| left | `:Name` | bind `Name` to the value supplied by the right side |
| left | `:thing` | the supplied value must match the constant `thing` |
| left | `:(X)` | the supplied value must match the value of `(X)` |
| left | `:{ x y }` | the supplied value must match the pattern `{x y}` |
| right | `:Name` | supply the value of `Name` |
| right | `:thing` | supply the constant `thing` |
| right | `:(X)` | supply the value of the expression, for example `x :(A + B)` |
| right | `:{ x y }` | supply an unevaluated element sequence, which works like a closure |

When the two sides meet:

| Left | Right | Effect |
| --- | --- | --- |
| variable | variable | the left variable takes the right variable's value |
| variable | value | the left variable is bound to the value |
| value | variable | the left value is matched against the variable's value |
| value | value | the two values are matched |

Rules can therefore select on argument values, like case analysis in ML:

```
'f' x :N  <- x - '*' x :(N) 'f' x :(N - 1) ;
'f' x :1  <- x :1;
'f' x :0  <- x :1;
```

## Acquisition with `%`

Nothing that is consumed is kept unless a rule explicitly acquires it. On the left side, `%` pushes the most recently matched symbol onto the context's grab stack. Conversion symbols such as `toNum` and `toSym` then turn the grabbed material into a value.

| Left | Right | Effect |
| --- | --- | --- |
| `%` | `%` | take all material grabbed by the originating rule |
| `%` | `:` | take a bindable value from the right side |
| `:` | `%` | bind all material grabbed by the originating rule |
| `%` | (none) | the value last matched by this left side |

The *originating rule* is the rule application whose right side produced the elements. A right-side `%` passes grabbed text up to the next acquirer. It is **not** substituted back into the input. This is how the chain of lexical sub-rules in `lmn2xfe` builds numbers:

```
[1-9] % { repeat [0-9] % } dpoint % type:T <- - number % :T ;
```

## Output buffers `(Var)`

Any uninitialised variable or table cell can serve as a text buffer. `()` evaluates to a fresh empty buffer.

- `- (Text) <- thing - ;` consumes one symbol and appends its text to `Text`, without producing `thing`. It is the buffered counterpart of `- out <- eom - ;`.
- If the rule has used `%` directly, `(Var)` consumes **no** input and instead appends everything on the grab stack.
- A buffer lives in its enclosing context and is not reset until that context is reset. By contrast, `%` material is reset when an alternative is tried.
- Elsewhere, a buffer behaves like a single "double-quoted" symbol that prints as its whole text.

Use `%` for material that has matched. Use `(Var)` to collect text up to a delimiter that other rules will match.

## Actions and expressions

The action language is a non-strict subset of JavaScript.

- **Statements:** `var A = 1, B;`, `if (…) … else if (…) … else …`, `while (…) …`, `for (init; test; next) …`, `foreach (V; E) …`, `foreach (K, V; E) …`, `break;`, `continue;`, and blocks `{ … }`. A statement expression ends with `;`.
- **Tables:** associative arrays such as `var T = [];`, `[1, 2, key: v]`, `T[k]` and `T.field`. There is also an `in` operator, whose semantics the site does not document. The `lexicalbuffer` example tests an unset cell with `Sy[V] == "null"`.
- **Calls:** `f(a, b)` calls builtin or external functions (see [04-special-symbols-and-builtins.md](04-special-symbols-and-builtins.md)).
- **foreach** (added by this port): `foreach (K, V; E) B` runs `B` once for each key of the array `E`, in the order the keys were added, with the key assigned to `K` and the value to `V`; `foreach (V; E)` assigns only the value. `K` and `V` are existing variables, as in `for`. A null `E` gives no passes, and any other non-array value is reported as `BAD`. `break` and `continue` work as in the other loops.
- **Rule values:** `rule(G, P) { lhs <- rhs }` defines a rule in grammar `G` while the rules run. `P` is the encoded priority word of the bytecode (`2n` for `nL`, `2n+1` for `nR`), and the value is the grammar symbol.

Operator precedence, from lowest to highest, as grammar priorities in `lmn2xfe`:

| Priority | Operators |
| --- | --- |
| 2L | `,` |
| 10R | `?:`, `=`, `+=`, `-=`, `*=`, `/=`, `%=` |
| 12L | `\|\|` |
| 14L | `&&` |
| 15L | `\|` |
| 16L | `^` |
| 18L | `&` |
| 20L | `in`, `===`, `!==`, `==`, `!=`, `<`, `>`, `<=`, `>=` |
| 22L | `+`, `-` |
| 24L | `*`, `/`, `%` |
| 26L | prefix and postfix `++`, `--` |
| 28R | unary `-`, `!`, `~` |
| 4B | call `f(…)`, index `a[i]`, field `a.b` |

Actions can appear on the left side, for example `start var N = 99; <- …` or `- if(Sy[V] == "null") Sy[V] = J++; <- identity:(Sy[V]);`. They can also appear on the right side inside braces, for example `{ for(var I = 0; I < 2; I++) { "\n=== group " I " ===\n" $(Table[I]) } }`.

## Nested rules

```
 {| lhs1 <- rhs1 || lhs2 || lhs3 <- rhs3 |}
```

This defines alternative rules that are tied to the context where the braces appear. The frontend generates a unique goal symbol for the group. A nested rule's right side is a plain pattern and may be omitted.

## A complete example: the forward Polish calculator

This is the example from the guide. The repository's `calc.lmn` is a cut-down version of it, with only `+` and integers.

```
.calc()
- error  output <- eof    - ;
- result output <- eof    - ;
- x :N          <- result 'result: ' N '\n' eom ;

'-' x :A        <- x :(-A);
'+' x :A x :B   <- x :(A + B);
'-' x :A x :B   <- x :(A - B);
'*' x :A x :B   <- x :(A * B);
'/' x :A x :B   <- x :(A / B);

.calc(20L)
.[0-9] % { repeat .[0-9] % } { option '.' % repeat .[0-9] % } toNum :N <- x :N;
.[ \t\n]                                                               <- -   ;

.calc(30R)
- anything      <- line -  ;
eof             <- line eof;
'\n'            <- line;
- line          <- error '--- not understood - skipping one line\n' output;
- eom           <- output ;
- out           <- eom -  ;
```

How it works:

- `error output` and `result output` have the same length, and `result` is newer, so `result` is tried first. `error` is the fallback.
- The `line` rules consume the rest of a bad line. They treat `eof` as the end of a line and pass `eof` back to the enclosing context.
- The output rules keep printing symbols with `out` until the `eom` marker is matched.
