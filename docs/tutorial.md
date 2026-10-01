# Language Machine tutorial

This tutorial teaches the Language Machine from first principles. It starts with a one-rule program that copies its input, works up through tokenisers, parsers, calculators and translators, and ends by turning a ruleset into an ordinary Go binary or library.

Each example is a complete program. You can paste it into a file and run it, and the output shown is what this repository's engine produces. The reference material lives elsewhere:

- [`lm/`](lm/README.md) is a digest of the original website: the execution model, lmn syntax, builtins and the lm-diagram.
- [`lmn2go.md`](lmn2go.md) covers the Go code generator.
- [`bytecode.md`](bytecode.md) specifies the `.lm` format.

This tutorial links to those pages instead of repeating them.

## Contents

1. [Setting up](#1-setting-up)
2. [Part I: basics](#part-i-basics)
   - [2. The smallest program](#2-the-smallest-program)
   - [3. Substitution](#3-substitution)
   - [4. How the machine thinks](#4-how-the-machine-thinks)
   - [5. Recognising something and saying so](#5-recognising-something-and-saying-so)
3. [Part II: grammars](#part-ii-grammars)
   - [6. A top-down sentence checker](#6-a-top-down-sentence-checker)
   - [7. Tokens: priorities, `%`, `repeat` and `option`](#7-tokens-priorities--repeat-and-option)
   - [8. Variables and binding](#8-variables-and-binding)
   - [9. A Polish calculator](#9-a-polish-calculator)
   - [10. An infix calculator: precedence from priorities](#10-an-infix-calculator-precedence-from-priorities)
4. [Part III: advanced techniques](#part-iii-advanced-techniques)
   - [11. Deferred sequences and left recursion](#11-deferred-sequences-and-left-recursion)
   - [12. Collecting repeated items with `each`](#12-collecting-repeated-items-with-each)
   - [13. Actions, tables and `foreach`](#13-actions-tables-and-foreach)
   - [14. Several grammars and `use()`](#14-several-grammars-and-use)
   - [15. Nested alternatives `{| … |}`](#15-nested-alternatives---)
   - [16. Programs that read no input](#16-programs-that-read-no-input)
   - [17. Error recovery and error messages](#17-error-recovery-and-error-messages)
   - [18. Specialising a ruleset with `-add`](#18-specialising-a-ruleset-with--add) (and running rules as scripts)
5. [Part IV: debugging](#part-iv-debugging)
   - [19. Traces and the lm-diagram](#19-traces-and-the-lm-diagram)
   - [20. Pitfalls checklist](#20-pitfalls-checklist)
6. [Part V: building Go binaries](#part-v-building-go-binaries)
   - [21. The pipeline](#21-the-pipeline)
   - [22. A standalone command](#22-a-standalone-command)
   - [23. Calling Go from the rules](#23-calling-go-from-the-rules)
   - [24. A library package](#24-a-library-package)
   - [25. Embedding a ruleset by hand](#25-embedding-a-ruleset-by-hand)
   - [26. Reference: `lmn2go` flags and the `lm` API](#26-reference-lmn2go-flags-and-the-lm-api)
7. [Where to go next](#where-to-go-next)

---

## 1. Setting up

You need Go (the version named in `go.mod`). The module has no third-party dependencies. From the repository root:

```sh
make build lmn lmn2go     # -> bin/lm, bin/lmn, bin/lmn2go
make install              # optional: go install all three into $GOBIN
```

You get three programs:

| Program | What it does |
| --- | --- |
| `bin/lm` | The engine. It loads compiled rules (`.lm` bytecode) with `-rules` and runs them over the input. |
| `bin/lmn` | The lmn compiler. It turns `.lmn` source into `.lm` bytecode. |
| `bin/lmn2go` | The Go generator. It turns `.lmn` (or `.lm`) into a Go source file that you build with `go build`. |

The everyday workflow has two steps, compile and then run:

```sh
bin/lmn -output hello.lm hello.lmn          # compile
bin/lm  -rules hello.lm  input.txt          # run on a file
bin/lm  -rules hello.lm  -input 'some text' # run on a string
```

`bin/lmn` is itself a Language Machine ruleset, built into a Go binary by `lmn2go` (Part V shows how). The same compiler also runs on the engine as `bin/lm -rules lmn.lm …` (see `examples/README.md`). Use `bin/lmn` rather than the original bootstrap compiler `examples/lmn/lmnbs.lm`, because the bootstrap compiler does not understand `foreach` and compiles `while` loops wrongly.

> **Tip.** A grammar that never reaches its end state can loop forever and print without end. While you experiment, wrap runs in `timeout`, for example `timeout 5 bin/lm -rules x.lm -input '…'`.

A small helper script saves typing during the tutorial:

```sh
#!/bin/sh
# lmrun file.lmn [lm options...]: compile and run in one step
f=$1; shift
bin/lmn -output "${f%.lmn}.lm" "$f" && timeout 5 bin/lm -rules "${f%.lmn}.lm" "$@"
```

---

# Part I: basics

## 2. The smallest program

Save this as `cat.lmn`, including the leading space:

```
A ruleset that copies its input to its output.

 - out <- eof - ;
```

```sh
$ bin/lmn -output cat.lm cat.lmn
$ bin/lm -rules cat.lm -input 'hello world'
hello world
```

### The source format

**In an `.lmn` file, only lines that start with a space are code.** Every other line is commentary. The original author wrote his `.lmn` files as wiki pages, with prose in between the code. Inside code, `//` and `/* … */` comments also work, and `/* */` comments nest.

Remember the leading space. A rule written in column 0 is silently treated as a comment.

### What the rule means

A rule has the form

```
left-side  <-  right-side ;
```

The left side is a **pattern to recognise**. The right side says **what to substitute** for the recognised pattern. This is the reverse of BNF: lmn rules are written from the sentence towards the grammar, `what you see <- what it means`.

In `- out <- eof - ;`:

- **`eof` is the goal.** The machine has one ambition, which is to match the symbol `eof`. The input system supplies `eof` after the last real input symbol. When the goal `eof` meets the input `eof`, the program ends successfully.
- **The leading `-` on the left** means "whatever the input symbol is". The rule does not care what it sees.
- **`out`** is a special symbol. As a goal, it consumes one input symbol and writes it to standard output.
- **`eof -` on the right** means that the rule is filed under the goal `eof`, but it substitutes *nothing*: whatever follows the second `-` is substituted, and here nothing follows. So the goal remains `eof`.

Here is the run, step by step:

1. The goal is `eof` and the input is `'h'`. They differ, which is a **mismatch**.
2. The machine looks for rules relevant to (goal `eof`, input `'h'`) and finds `- out <- eof - ;`.
3. The left side `out` consumes `'h'` and prints it. The right side substitutes nothing, so the goal is still `eof`.
4. Steps 1–3 repeat until the real `eof` arrives. It matches the goal, and the run stops with status 0.

## 3. Substitution

```
Replace every "cat" with "dog" and copy everything else.

 'cat'      <- eof - "dog" ;
 - out      <- eof - ;
```

```sh
$ bin/lm -rules swap.lm -input 'the cat sat on the catalogue'
the dog sat on the dogalogue
```

The first rule's left side is `'cat'`. A single-quoted string is a *sequence of characters*, `'c' 'a' 't'`. The right side places the symbol `"dog"` back **into the input** in front of whatever comes next, and the second rule then prints it.

This is the central idea of the machine: **a rule's right side is treated as if it had appeared in the input**. The substituted material is analysed again, which is what lets rules build on one another.

### `'single'` versus `"double"` quotes

| Form | Meaning |
| --- | --- |
| `'abc'` | three terminal (character) symbols `'a' 'b' 'c'` |
| `"abc"` | one symbol whose text is `abc` |

Both print the same way. The difference matters when the substitution is analysed again. If the rule were `'cat' <- eof - 'cats' ;`, the substituted characters would start with `c a t` again, and the rule would fire forever. With `"cats"`, a single non-character symbol, the problem cannot arise.

### Why `cat` wins over `- out`

Both rules are relevant when the goal is `eof` and the input is `'c'`. The machine tries rules in a fixed order, explained in the next section, and `'cat'`, which names the input symbol explicitly, comes first. If `'cat'` fails, because the next characters are `'c' 'o' 'w'`, the machine **backtracks** and tries `- out` instead.

## 4. How the machine thinks

Everything in the machine happens in response to a **mismatch** between the current *goal* (the left symbol) and the current *input* (the right symbol). When they are equal, both advance. When they differ, the machine looks up rules by the pair (left-initial, right-initial):

- The left-initial is the first symbol of the rule's left side, and it must match the **input**.
- The right-initial is the first symbol of the right side, and it names the **goal** the rule serves.

Either one can be `-`, which means "don't care". For goal `g` and input `i`, the categories are tried in this order:

| Order | Rule shape | Style | Example |
| --- | --- | --- | --- |
| 1 | `i … <- g …` | specific | `'hello' <- greeting ;` |
| 2 | `i … <- - …` | bottom-up, driven by the input | `' ' <- - ;` (delete spaces anywhere) |
| 3 | `- … <- - …` | speculative (rare) | |
| 4 | `- … <- g …` | top-down, driven by the goal | `- subject verb object <- sentence ;` |

Within a category, **longer left sides are tried before shorter ones, and among equals the newest rule (the one later in the file) goes first.** The "length" counts the matchable items at the outermost brace level, so a left side wrapped in `{ … }` has length 0 and is tried last. Section 17 uses that on purpose.

If a rule's left side fails partway through, the machine restores its state and tries the next candidate. If none succeeds, the mismatch is unresolved and the *enclosing* rule fails in turn. If even the outermost goal fails, the run ends with exit status 1.

A summary of the three meanings of `-`:

1. At the start of the **left** side: any input.
2. At the start of the **right** side: any goal (a bottom-up rule).
3. **Directly after the right-initial**, as in `<- eof - …`: the rule is filed under that goal, but it does not produce it. Only the symbols after the `-` are substituted, and the goal stays in force. Copy loops such as `- out <- eof - ;` rely on this.

The full model, including priorities and scope, is in [`lm/02-execution-model.md`](lm/02-execution-model.md).

## 5. Recognising something and saying so

Printing is done by goals. The idiom that every ruleset in `examples/` uses is:

```
 - eom      <- output ;      // the goal "output" becomes "eom"
 - out      <- eom - ;       // print symbols until the input is the symbol eom
```

A rule that has recognised something substitutes `<the text to print> eom`. The `output` and `eom` goals print everything up to the `eom` marker.

```
 .greet()
 - greeting output <- eof - ;
 'hello'           <- greeting "Hi there!\n" eom ;
 'bye'             <- greeting "See you.\n"  eom ;
 '\n'              <- eof - ;
 - eom             <- output ;
 - out             <- eom - ;
```

```sh
$ bin/lm -rules greet.lm -input 'hello
bye
hello
'
Hi there!
See you.
Hi there!
$ bin/lm -rules greet.lm -input 'what'; echo "status $?"
status 1
```

Here is how it reads:

- `.greet()` starts a **grammar** named `greet`. Rules belong to the most recent grammar selector, and the grammar of the first rule in the file is the one that starts. Without a selector, rules go into the grammar `lm_`.
- `- greeting output <- eof - ;` says that while the goal is `eof`, the machine should expect a `greeting` followed by an `output`.
- `'hello' <- greeting "Hi there!\n" eom ;`: when the input `'h'` meets the goal `greeting`, the rule matches `hello` and substitutes `greeting "Hi there!\n" eom`. The `greeting` satisfies the goal, and the goal moves on to `output`, which prints `Hi there!` and a newline up to `eom`.
- `'\n' <- eof - ;` is a "something to nothing" rule: it deletes newlines between greetings.
- `what` matches nothing, so the outermost goal fails, nothing is printed, and the exit status is 1. Section 17 shows how to recover from errors instead.

Double-quoted strings understand C escapes such as `\n`, `\t` and `\"`.

---

# Part II: grammars

## 6. A top-down sentence checker

Goal-driven rules (`- … <- goal`) behave like recursive-descent parser functions. Here is a tiny English grammar that answers `yes` or `no` for each line:

```
 .cats()
 - bad  output              <- eof - ;
 - good output              <- eof - ;
 - sentence '\n'            <- good "yes\n" eom ;

 - subject verb object '.'  <- sentence ;
 - nounphrase               <- subject ;
 - nounphrase               <- object  ;
 'the ' noun                <- nounphrase ;
 'a '   noun                <- nounphrase ;
 'cat '                     <- noun ;
 'dog '                     <- noun ;
 'bit '                     <- verb ;
 'ate '                     <- verb ;
 'likes '                   <- verb ;

 .cats(30R)
 - anything                 <- line - ;
 '\n'                       <- line ;
 eof                        <- line eof ;
 - line                     <- bad "no\n" eom ;
 - eom                      <- output ;
 - out                      <- eom - ;
```

```sh
$ bin/lm -rules cats.lm -input 'the cat likes a dog .
a dog bit the cat .
the cat dog the dog .
a cat ate a dog .
'
yes
yes
no
yes
```

Points to notice:

- **Alternatives are separate rules.** `subject` and `object` are both a `nounphrase`, and `nounphrase` is either `'the ' noun` or `'a ' noun`.
- **Order matters for fallbacks.** `- good output` and `- bad output` have the same length, so the newer one, `good`, is tried first. `bad` only runs if `good` fails. If the order were swapped, every line would say `no`.
- **`line` skips a line.** `- anything <- line - ;` consumes any symbol while the goal is `line`. `'\n' <- line ;` ends the line, and `eof <- line eof ;` treats the end of input as the end of a line but puts `eof` back for the outer goal.
- The `.cats(30R)` priority on the error and output rules follows the convention of the original examples. This grammar has no whitespace-deletion rules, so it would work without it. In grammars that do have them, the high right-associative priority stops low-priority deletion rules from starting inside the output context, so spaces reach the output instead of being deleted. Priorities are explained in the next section.

## 7. Tokens: priorities, `%`, `repeat` and `option`

Real input has whitespace and multi-character tokens. This program prints each word on its own line:

```
Print every word in brackets, one per line.

 .words()
 - word :W                 <- eof - "[" W "]\n" ;
 - out                     <- eof - ;

 .words(20L)
 [a-zA-Z] % { repeat [a-zA-Z] % } toSym :W  <- word :W ;
 [ \t\n.,;!?]                                <- - ;
```

```sh
$ bin/lm -rules words.lm -input 'Hello, wide world!
second line'
[Hello]
[wide]
[world]
[second]
[line]
```

Several new things appear here.

**Lexical classes.** `[a-zA-Z]` (or `.[a-zA-Z]`) is a character class. As the first symbol of a rule it stands for one rule per member, so `[0-9] …` files ten rules. Elsewhere in a left side it matches any member. A negated class such as `[^\n]` matches fine inside a left side, but it cannot be a rule's first symbol, because the engine cannot file a rule under "everything except".

**`repeat`.** `repeat X` on a left side matches the *rest of the enclosing sequence* zero or more times. That is why the braces matter: `{ repeat [a-zA-Z] % }` repeats only the letter and its `%`. `option` works the same way but matches zero or one time. Neither ever fails, so to require at least one item, write the item once before the `repeat`, as this rule does.

**Acquisition with `%`.** Nothing that is matched is kept unless you ask. `%` grabs the symbol just matched. A **conversion symbol** then turns the grabbed text into a value:

| Conversion | Gives |
| --- | --- |
| `toSym`, `toLsym`, `toUsym` | a symbol, unchanged, lower-cased or upper-cased |
| `toNum`, `toHex`, `toOct`, `toBin` | a number |
| `toStr`, `toLstr`, `toUstr` | a sequence of characters |

The full list is in [`lm/04-special-symbols-and-builtins.md`](lm/04-special-symbols-and-builtins.md).

**Binding with `:`.** `toSym :W` binds the converted value to the variable `W`. Variables start with an upper-case letter, and nonterminals with a lower-case letter or `_`. On the right side, `word :W` hands the value on, and the rule that expected `word :W` receives it in its own `W`.

**Priorities make whitespace work.** This is the subtle part. `.words(20L)` gives the following rules priority 20, left-associative. A rule started in a context runs *at its own priority*, and an `L` rule may only start inside a context whose priority is *lower* than its own. So:

- At the top level (priority 0) the deletion rule `[ \t\n.,;!?] <- - ;` (20L) can start, and separators vanish.
- Inside a word (priority 20L) the deletion rule, also 20L, **cannot** start. A space is therefore not silently skipped in the middle of a word: it ends the word.

Without the priority, `Hello world` would come out as one word, `Helloworld`. The directions are:

| Suffix | Can start inside a context of priority P when… | Typical use |
| --- | --- | --- |
| `L` | its priority is **greater than** P | left-associative operators, tokens |
| `R` | its priority is **greater than or equal to** P | right-associative operators, output rules |
| `B` | always (it begins a new bracket level) | parentheses |
| none, `.g()` | it inherits the enclosing priority | structure rules |

## 8. Variables and binding

This program counts and sums all the numbers in its input:

```
Add up all the numbers in the input.

 .total()
 start var Total = 0; var Count = 0; eof
                                  <- eof - generate "count: " Count ", total: " Total "\n" output eof ;
 - number :N  Total = Total + N; Count++;
                                  <- eof - ;
 generate output                  <- eof - ;
 - out                            <- output - ;

 .total(20L)
 [0-9] % { repeat [0-9] % } { option '.' % repeat [0-9] % } toNum :N   <- number :N ;
 [ \t\n]                                                                <- - ;
```

```sh
$ bin/lm -rules total.lm -input '3 4
 35 1.5'
count: 4, total: 43.5
$ bin/lm -rules total.lm -input ''
count: 0, total: 0
```

### `start`

Before any input is read, the machine acts as if the goal `eof` had met an input symbol `start`. A rule whose left side begins with `start` therefore runs first. It is the place to set things up.

### Actions

Left sides can contain **actions**, written in a subset of JavaScript: `var Total = 0;`, `Total = Total + N;`, `Count++;`, and also `if`, `while`, `for`, `foreach`, tables and function calls. On a right side, actions must go inside braces, `{ … }` (see section 13).

### Scope: why the input is matched *inside* the `start` rule

A variable is visible to rules that start **inside** the context that created it. The `start` rule here goes on to match `eof`: its left side is `start var …; eof`. So the whole input is consumed *inside* that rule's context, and every `- number :N … Total = Total + N;` application can see `Total`. Only when `eof` arrives does the rule finish. Its right side then prints the result: `generate output` turns into an output goal, and `eof` is put back so the outer goal can finish.

If you write `start var Total = 0; <- eof - ;` instead, the context closes straight away. `Total` is then gone before the first number arrives, and the engine reports `BAD = Total (undefined)`.

The rule on the right side follows the same principle. A reference there sees the variables that were visible at the end of the corresponding left side.

### The binding table

`:` works in both directions:

| Left side | Right side supplies | Effect |
| --- | --- | --- |
| `:Name` | a value | bind `Name` |
| `:thing` (a constant) | a value | succeed only if the value equals `thing` |
| `:(expr)` | a value | succeed only if it equals the expression |
| | `:Name`, `:thing`, `:(expr)` | supply a value |
| | `:{ x y }` | supply an **unevaluated** sequence (section 11) |

Matching against constants gives you case analysis, which the next section uses.

## 9. A Polish calculator

This is the classic example from the original guide. Operators come first: `+ 2 3`.

```
 .calc()
 - error  output <- eof - ;
 - result output <- eof - ;
 - x :N          <- result 'result: ' N '\n' eom ;

 '-' x :A        <- x :(-A);
 '+' x :A x :B   <- x :(A + B);
 '-' x :A x :B   <- x :(A - B);
 '*' x :A x :B   <- x :(A * B);
 '/' x :A x :B   <- x :(A / B);
 'f' x :N        <- x - '*' x :(N) 'f' x :(N - 1) ;
 'f' x :1        <- x :1;
 'f' x :0        <- x :1;

 .calc(20L)
 [0-9] % { repeat [0-9] % } { option '.' % repeat [0-9] % } toNum :N <- x :N;
 [ \t\n]                                                             <- - ;

 .calc(30R)
 - anything      <- line - ;
 eof             <- line eof;
 '\n'            <- line;
 - line          <- error '--- not understood - skipping one line\n' eom;
 - eom           <- output ;
 - out           <- eom - ;
```

```sh
$ bin/lm -rules fp.lm -input '+ 2 3
* 100 / 1 3
f 6
z
* 1 + / 1 55 * 3 22
- 7
'
result: 5
result: 33.3333
result: 720
--- not understood - skipping one line
result: 66.0182
result: -7
```

How it works:

- **Every value is an `x :N`.** A number substitutes `x :N`. An operator rule matches its operator and then one or two `x` values, and substitutes a new `x` carrying the result. `:(A + B)` evaluates an expression.
- **Unary versus binary minus.** `'-' x :A x :B` is longer than `'-' x :A`, so it is tried first. On `- 7` it fails to find a second `x`, and the unary rule takes over.
- **Factorial by rewriting.** `f 6` matches `'f' x :N` and substitutes `'*' x :(6) 'f' x :(5)` *back into the input*. That is ordinary input for the `*` rule, which needs a second operand and finds it by expanding `f 5`, and so on. The rules `'f' x :1` and `'f' x :0` are newer and the same length, so they are tried first, and they only match when the bound value is 1 or 0. That is how the recursion ends.
- **Error recovery.** `result` is newer than `error`, so it is tried first. A line such as `z` fails as a `result`, and the `error` path consumes the line through `line` and prints the message.

`x - '*' …` uses the third meaning of `-`: the rule is filed under the goal `x`, but it substitutes only `'*' x … 'f' x …`. The goal `x` is still waiting when the `*` rule produces its result.

## 10. An infix calculator: precedence from priorities

Infix notation needs precedence and associativity. In the Language Machine both come from the priorities of section 7:

```
 .calc()
 - result output         <- eof - ;
 - expr :N '\n'          <- result "= " N "\n" eom ;

 .calc(0B)
 '(' expr :N ')'         <- opnd :N ;

 .calc()
 - opnd :A op            <- expr - ;
 -                       <- op expr :A ;

 .calc(10L)
 '+' expr :B             <- op opnd :(A + B) ;
 '-' expr :B             <- op opnd :(A - B) ;

 .calc(12L)
 '*' expr :B             <- op opnd :(A * B) ;
 '/' expr :B             <- op opnd :(A / B) ;

 .calc(18L)
 '-' opnd :A             <- opnd :(-A) ;

 .calc(20L)
 ' '                                                         <- - ;
 [0-9] % { repeat [0-9] % } { option '.' % repeat [0-9] % } toNum :N <- opnd :N ;

 .calc(30R)
 - eom                   <- output ;
 - out                   <- eom - ;
```

```sh
$ bin/lm -rules infix.lm -input '1 + 2 * 3
(1 + 2) * 3
10 - 4 - 3
-2 * 3
2 * (3 + 4) / 7
'
= 7
= 9
= 3
= -6
= 2
```

### The loop that folds an expression

- `- opnd :A op <- expr - ;`: while the goal is `expr`, read an operand into `A`, and then expect an `op`. The right side is `expr -` with nothing after it, so the goal stays `expr` and the rule can apply again.
- An operator rule such as `'+' expr :B <- op opnd :(A + B) ;` matches `+` and a whole sub-expression `B`. It substitutes `op`, which satisfies the waiting `op` goal, followed by a new operand `opnd :(A + B)`. That operand goes round the loop again as the next `A`. This is left recursion, written as iteration.
- When no operator follows, the goal `op` meets something else, such as `'\n'` or `')'`. The empty-left-side rule `- <- op expr :A ;` then fires. It recognises nothing and produces `op expr :A`: the `op` closes the loop, and `expr :A` satisfies the goal `expr` with the accumulated value. A rule with an empty left side is a "something from nothing" rule.

### Precedence and associativity

| Input | What happens |
| --- | --- |
| `1 + 2 * 3` | The `+` rule (10L) parses its `expr :B` at priority 10. Inside it, `*` (12L, which is greater than 10) may start, so `2 * 3` binds first. |
| `10 - 4 - 3` | Inside the first `-` (10L), the second `-` is also 10L. An `L` rule needs a *greater* priority, so it cannot start there. `B` stops at `4`, and the second `-` applies to the result: `(10 - 4) - 3`. |
| `(1 + 2) * 3` | `'('` is `0B`, a bracket. It can always start, and it opens a fresh priority level inside the parentheses. |
| `-2 * 3` | Unary minus at 18L binds tighter than `*`. |

To make an operator right-associative, give it an `R` priority. Section 22 adds `^` at `14R`, so that `2 ^ 3 ^ 2` is `2 ^ (3 ^ 2) = 512`.

The original `examples/basics/calc.lmn` extends this calculator with hex, octal and binary literals, and with reports that use the `var*` builtins.

---

# Part III: advanced techniques

## 11. Deferred sequences and left recursion

A right-side binding `:{ … }` passes a **sequence that has not been evaluated yet**, a closure over the current variables. Binding it costs the same however long it is, which makes list building cheap. This program prints a list forwards and backwards:

```
Read "a b c ;" and print the items forwards and backwards.

 .rev()
 - result output                 <- eof - ;
 - item :X                       <- result - list :X :X ;
 list :F :B ';'                  <- result "forwards:  " F "\n" "backwards: " B "\n" eom ;
 list :F :B item :X              <- result - list :{ F ", " X } :{ X ", " B } ;

 .rev(20L)
 [a-z0-9] % { repeat [a-z0-9] % } toSym :X   <- item :X ;
 [ \t\n]                                      <- - ;

 .rev(30R)
 - eom                           <- output ;
 - out                           <- eom - ;
```

```sh
$ bin/lm -rules rev.lm -input 'apple 1 2 cherry ;'
forwards:  apple, 1, 2, cherry
backwards: cherry, 2, 1, apple
```

- The first item becomes `list :X :X`, a list that is the same both ways round.
- Each further item meets a `list :F :B` that has already been substituted into the input, and produces a longer `list`. The forward list appends `X` and the backward list prepends it. Neither copies anything, because `{ F ", " X }` only refers to `F`.
- At `';'` the two lists are printed. Only then are the deferred sequences evaluated.

The rule `list :F :B item :X` has a *nonterminal* as its left-initial: it fires when a substituted `list` symbol meets the goal `result`. This is how the Language Machine does bottom-up, LR-style analysis, and why left recursion needs no special treatment.

## 12. Collecting repeated items with `each`

When `repeat` binds the same variable many times, `each Name` produces every instance, oldest first:

```
Collect the items of a bracketed list with repeat, then emit them with each.

 .colours()
 - list :L '\n'                   <- eof - "(" L ")\n" ;
 - out                            <- eof - ;
 '[' { repeat item :X } ']'       <- list :{ each X } ;

 .colours(20L)
 [a-z] % { repeat [a-z] % } toUsym :W  <- item :{ " " W " " } ;
 ' '                                    <- - ;
```

```sh
$ bin/lm -rules each.lm -input '[ red green blue ]
[ ]
'
( RED  GREEN  BLUE )
()
```

Notice the braces in `{ repeat item :X } ']'`. Without them, `repeat` would repeat `item :X ']'` as a unit. Also notice that the grammar is not called `each`: `each`, `all`, `rule`, `option`, `repeat`, `var`, `if`, `for`, `foreach`, `while` and the other keywords are reserved, and cannot be used as grammar or symbol names.

## 13. Actions, tables and `foreach`

Tables are associative arrays: `var T = [];`, `T[key]`, `T.field`. This word-frequency counter fills a table on the left side and reports it with a right-side action block:

```
Count how often each word occurs.

 .freq()
 start var Count = []; var K; var V; eof
        <- eof - generate
           { foreach (K, V; Count) { $(format("%-8s %d\n", K, V)) } }
           output eof ;

 - word :W  if (Count[W] == "null") Count[W] = 0; Count[W]++;
                                   <- eof - ;
 generate output                   <- eof - ;
 - out                             <- output - ;

 .freq(20L)
 [a-zA-Z] % { repeat [a-zA-Z] % } toLsym :W   <- word :W ;
 [ \t\n.,;:!?]                                <- - ;
```

```sh
$ bin/lm -rules freq.lm -input 'The cat saw the dog. The dog saw a Cat!'
the      3
cat      2
saw      2
dog      2
a        1
```

- `toLsym` lower-cases each word, so `The` and `the` are counted together.
- An unset cell compares equal to `"null"`.
- On a right side, `{ … }` holds actions. `foreach (K, V; Count)` visits the keys in the order they were added. `$(expr)` evaluates an expression and inserts its value as a symbol. `format` works like printf.
- The `start …; eof <- … generate { … } output eof` shape is the one from section 8. It keeps `Count` in scope for the whole input and prints only at the end.

Other useful builtins include `lcase`, `ucase`, `num`, `hex`, `toChars`, `include(file)` (push another input file) and `trOn`/`trOff` (switch tracing at run time). See [`lm/04-special-symbols-and-builtins.md`](lm/04-special-symbols-and-builtins.md).

**Output buffers.** Any unset variable or table cell can collect text: `- (Text) <- toX - ;` appends each consumed symbol to `Text`. The `examples/samples/reorder.lmn` and `flatten.lmn` examples use this to sort lines into groups and to pull nested blocks apart. See [`lm/03-lmn-language.md`](lm/03-lmn-language.md#output-buffers-var).

## 14. Several grammars and `use()`

Only the rules of the **current grammar** take part in resolving a mismatch. `use("g")` switches grammar for the rest of the current left side and everything nested inside it. This is how you write modes, such as "inside a string", "inside a comment" or "inside code":

```
Markdown-ish to plain text: drop emphasis markers, except inside `code`.

 .text()
 - out                       <- eof - ;
 '*'                         <- eof - ;
 '_'                         <- eof - ;
 '`' use("code"); span       <- eof - ;

 .code()
 '`'                         <- span ;
 - out                       <- span - ;
```

```sh
$ bin/lm -rules plain.lm -input 'Use *bold* and _italic_, but `a*b_c` stays.'
Use bold and italic, but a*b_c stays.
```

Inside the backticks only the `code` grammar's rules apply. That grammar has no rules that delete `*` or `_`, so they are copied as they are. When `` ` `` closes the `span`, control returns to the `text` grammar.

## 15. Nested alternatives `{| … |}`

`{| a <- x || b <- y || … |}` defines a small set of alternative rules right where they are used. It is handy for keyword tables:

```
 .alt()
 - out                                                   <- eof - ;
 '#' {| 'red' <- :"FF0000" || 'green' <- :"00FF00" || 'blue' <- :"0000FF" |} :C
                                                         <- eof - "0x" C ;
```

```sh
$ bin/lm -rules alt.lm -input 'colours: #red #blue #green #grey'
colours: 0xFF0000 0x0000FF 0x00FF00 #grey
```

Each alternative supplies a value with `:"…"`, which the `:C` after the group receives. `#grey` matches no alternative, so the `'#'` rule fails and `- out` copies the text instead.

## 16. Programs that read no input

A ruleset can generate output purely by rewriting. This one counts down:

```
Count down from 5, without reading any input.

 .count()
 start                  <- eof - countdown :5 eof ;
 countdown :N           <- eof - $(N) "...\n" countdown :(N - 1) ;
 countdown :0           <- eof - "lift-off!\n" ;
 - out                  <- eof - ;
```

```sh
$ bin/lm -rules count.lm -input z
5...
4...
3...
2...
1...
lift-off!
```

Two lessons:

1. **The special case must be newer.** `countdown :N` and `countdown :0` have the same length, so the one written *later* is tried first. If you put `countdown :0` first, the general rule always wins, and the countdown runs on into negative numbers for ever. The original `bottles.lmn` orders its rules this way for the same reason.
2. **Give a dummy input.** With no input files and no `-input`, `lm` reads standard input, so the program sits waiting for you. A dummy string such as `-input z` is never consumed, but it stops the wait. The original's `-i z` idiom is the same thing.

`examples/web/bottles.lmn` (99 bottles of beer) and the lambda-calculus experiments in `examples/lambda/` take this style much further.

## 17. Error recovery and error messages

Larger programs should report errors with a position instead of failing silently. This converter turns `key = value` lines into JSON-style pairs:

```
Turn "key = value" lines into JSON-style pairs; report bad lines on stderr.

 .conf()
 - entry output                    <- eof - ;
 - pair :K :V '\n'                 <- entry "\"" K "\": \"" V "\",\n" eom ;
 '\n'                              <- eof - ;
 - { flagError :F bad message }    <- eof - ;

 - key :K '=' value :V             <- pair :K :V ;

 .conf(20L)
 [a-z] % { repeat [a-z_0-9] % } toSym :K             <- key :K ;
 [a-zA-Z0-9/._] % { repeat [^\n] % } toSym :V         <- value :V ;
 [ \t]                                                <- - ;

 .conf(30R)
 - anything                        <- line - ;
 '\n'                              <- line ;
 eof                               <- line eof ;
 - line                            <- bad F "cannot parse this line\n" message ;
 - err                             <- message - ;
 - eom                             <- output ;
 - out                             <- eom - ;
```

```sh
$ bin/lm -rules conf.lm -input 'name = demo
port = 8080

= oops
path = /tmp/x y
'; echo "status $?"
"name": "demo",
"port": "8080",
input:4: cannot parse this line
"path": "/tmp/x y",
status 1
```

- **`flagError :F`** gives a `file:line: ` prefix for the current position and adds one to the error count. A non-zero count makes `lm` exit with status 1, so scripts and CI can detect bad input. `warnError` counts warnings instead and leaves the status alone.
- **`err`** works like `out` but writes to standard error. You can redirect it with `-errout file`.
- **The braces matter.** `- { flagError :F bad message }` has effective length 0, because its items are inside braces, so it is tried *after* every real rule for the goal `eof`. It is a last resort. The lmn compiler's own front end (`examples/lmn/lmn2xfe.lmn`) uses the same trick.
- **No catch-all copy rule.** A `- out <- eof - ;` here would be longer than the braced error rule and would win, silently copying bad lines. Use either a copy rule or an error rule as the fallback, not both.
- With several input files (`bin/lm -rules conf.lm a.txt b.txt`) the files are read one after another as a single stream, and messages name the right file, such as `b.txt:2:`.
- `lineNo` gives just the line number, for example `zzz lineNo :A <- …`. The `var*` builtins (`varLn(N)`, `varCn(N)`, …) tell you where the value bound to a variable came from.

## 18. Specialising a ruleset with `-add`

Rules loaded later win over earlier rules of the same length in the same context. So a general ruleset can be tailored without editing it:

```
base.lmn:
 - out          <- eof - ;
 'colour'       <- eof - "color" ;

extra.lmn:
 'colours'      <- eof - "hues" ;
 'colour'       <- eof - "COLOR" ;
```

```sh
$ bin/lm -rules base.lm -input 'my colour, your colours'
my color, your colors
$ bin/lm -rules base.lm -add extra.lm -input 'my colour, your colours'
my COLOR, your hues
```

`'colours'` is longer than `'colour'`, so it is tried first. The new `'colour'` rule overrides the old one because it is newer. Since neither file names a grammar, both use the default grammar `lm_`.

### Rulesets as executable scripts

`-shebang PATH` writes a `#! PATH -rules` header to the output. Compiled rules with that header can be run directly as a script. Put `-shebang` after `-output`, so that the header goes into the file:

```sh
$ bin/lmn -output swapper -shebang "$PWD/bin/lm" swap.lmn
$ chmod +x swapper
$ ./swapper -input 'a cat'
a dog
$ echo 'my cat' | ./swapper -stdin
my dog
```

The engine reads `#` lines as comments, so the script is also an ordinary `.lm` file for `-rules` and `-add`. A script still needs `lm` installed at `PATH`. For a program that runs without the engine installed, build a Go binary (Part V).

---

# Part IV: debugging

## 19. Traces and the lm-diagram

`-trace` takes one-letter codes, which can be combined as `-trace m,s` or given more than once:

| Code | Shows |
| --- | --- |
| `m` | mismatch events (`??`) and fallbacks or backtracks (`**`). This is the most useful short trace. |
| `s` | symbols as they match |
| `D` | the lm-diagram (set `-dwidth` *before* `-trace D`) |
| `b` | rules as they load (put it before `-rules`) |
| `G` | a summary of the grammar after loading |
| `a` | everything except the diagrams |

`bin/lm -h` lists them all. **Options take effect in the order given**, so `-dwidth 30 -trace D` works but `-trace D -dwidth 30` draws at the default width.

Take this ruleset, `hi.lmn`:

```
 'hi'      <- eof - "HELLO" ;
 - out     <- eof - ;
```

```sh
$ bin/lm -rules hi.lm -trace m -input 'hi!'
   1   ??    0    0     0L    0    1    0      1      lm_         eof         'h'         ---
   1   ??    0    1     0L    0    1    0      2      lm_         eof       HELLO         'i'
   1   ??    0    0     0L    0    1    0      3      lm_         eof         '!'       HELLO
HELLO!
```

Each `??` line is a mismatch. Among other columns, it shows the current grammar, the goal (`eof`), the input symbol, and the previously matched symbol. The trace and the program's output share standard output, so they interleave.

The **lm-diagram** draws the same run with the goals on the left and the inputs on the right:

```sh
$ bin/lm -rules hi.lm -dwidth 30 -trace D -input 'hi!'
                     eof 'h'
?                    eof 'h'
┌───────000001
│                    'i' 'i'
└───────000001---------- ----------000001────────┐
                     eof HELLO                   │
?                    eof HELLO                   ?
┌───────000002                                   │
│                    out HELLO                   │
└───────000002                                   │
│                                  000001────────┘
                     eof '!'
?                    eof '!'
┌───────000003
│                    out '!'
└───────000003
                     eof eof
```

(The program's own output, `HELLO` and `!`, also appears at the start of two of these lines, and has been removed here.)

How to read it:

- `?` marks a mismatch. `┌──000001` opens the **recognition** (left side) of rule application number 1, here `'hi'`, which matches `'h'` and then `'i'`.
- `└──000001---------- ----------000001──┐` closes the recognition and opens the **substitution** on the right. `HELLO` is then the input, and the bar on the right shows that it came from application 1, not from the real input.
- Application 2 is `- out`, which consumes `HELLO`. Application 3 prints `'!'`. Finally `eof` meets `eof`.

For larger runs the nesting shows left recursion, right recursion and bracketing at a glance. [`lm/06-lm-diagram.md`](lm/06-lm-diagram.md) explains the full legend and walks through the `cats` grammar.

Other guard rails:

- `-max-depth N` limits nesting depth. It catches rules such as `- nest <- nest ;`.
- `-max-repeat N` limits `repeat`, for example `{ repeat nothing }` where `- <- nothing ;` exists.
- `timeout 5 …` is the bluntest and most reliable guard.

## 20. Pitfalls checklist

Every one of these came up while writing this tutorial:

| Symptom | Likely cause |
| --- | --- |
| A rule seems to be ignored | The line does not start with a space, so it is a comment. |
| A specific rule never fires | A newer rule of the same length, or a longer one, in the same category always succeeds first. Typically a catch-all `- out` was written after it. Move the special case *after* the general one. |
| A recursion never stops | The base case (`thing :0`) was written *before* the general case, so the general case is tried first. |
| `repeat` swallows the closing token | `repeat` covers the rest of its sequence. Wrap it: `{ repeat item :X } ']'`. |
| `a rule cannot start with the negated lexical class` | Rules are filed by their first symbol, so a rule cannot begin with `[^…]`. List the class explicitly, or begin with a positive class. |
| `file.lmn:3: ERROR` when compiling | Often a reserved word (`each`, `all`, `rule`, …) used as a name. |
| `BAD = X (undefined)` | The variable's context has already closed. Keep the work inside the declaring rule's left side (the `start var …; eof <- …` shape). |
| Nothing is printed, status 1 | The outermost goal failed. Add an error path (section 17), or trace with `-trace m`. |
| Substituted text is never printed | Nothing consumes it. Right-side symbols go back into the input, so some goal (`out`, `output`/`eom`) must print them. |
| A program with no input hangs | `lm` is reading standard input. Pass `-input z`. |
| Words run together | Whitespace deletion is running inside tokens. Give the token and deletion rules the same `L` priority. |
| `external not found: f …` | The rules call a function the engine does not have. See section 23. |

---

# Part V: building Go binaries

## 21. The pipeline

`lmn2go` turns a ruleset into Go source, which `go build` compiles into a self-contained program. The program needs no `.lm` file at run time.

```
calc.lmn ──lmn2go──► calc_lm.go ──┐
                                  ├──go build──► ./calc
funcs.go (your Go code) ──────────┘
```

The generated file holds:

- the compiled rules, as a string constant
- an `lm.Program` value, which names the rules and maps every function the rules call to a Go function
- in package `main`, a `func main()` that runs the program with the same command-line options as `lm`

The generator embeds its own lmn compiler, so `lmn2go` is the only tool you need. The runtime package that generated code imports is `github.com/msorc/languagemachine2/lm`. It is deliberately small and exposes no engine internals.

## 22. A standalone command

We will build the infix calculator from section 10 as a binary, with two additions: a right-associative `^` and a `sqrt` prefix operator, both implemented in Go.

### Step 1: a module

```sh
mkdir calcapp && cd calcapp
go mod init example.com/calcapp
```

The module path of this repository is `github.com/msorc/languagemachine2`. To build against a local checkout, add a requirement and a `replace` directive:

```sh
go mod edit -require github.com/msorc/languagemachine2@v0.0.0 \
            -replace github.com/msorc/languagemachine2=/path/to/languagemachine2
```

If you fetch the module from its published location instead, `go get github.com/msorc/languagemachine2/lm` is enough, and `go install github.com/msorc/languagemachine2/cmd/lmn2go@latest` installs the generator.

### Step 2: the rules

Save this as `calc.lmn`. It is section 10's calculator plus the two new rules:

```
 .calc()
 - result output         <- eof - ;
 - expr :N '\n'          <- result "= " N "\n" eom ;

 .calc(0B)
 '(' expr :N ')'         <- opnd :N ;

 .calc()
 - opnd :A op            <- expr - ;
 -                       <- op expr :A ;

 .calc(10L)
 '+' expr :B             <- op opnd :(A + B) ;
 '-' expr :B             <- op opnd :(A - B) ;

 .calc(12L)
 '*' expr :B             <- op opnd :(A * B) ;
 '/' expr :B             <- op opnd :(A / B) ;

 .calc(14R)
 '^' expr :B             <- op opnd :(pow(A, B)) ;

 .calc(18L)
 '-' opnd :A             <- opnd :(-A) ;
 'sqrt' opnd :A          <- opnd :(sqrt(A)) ;

 .calc(20L)
 ' '                                                         <- - ;
 [0-9] % { repeat [0-9] % } { option '.' % repeat [0-9] % } toNum :N <- opnd :N ;

 .calc(30R)
 - eom                   <- output ;
 - out                   <- eom - ;
```

`pow(A, B)` and `sqrt(A)` are calls to functions that the engine does not provide.

### Step 3: generate

```sh
lmn2go -o calc_lm.go -stubs funcs.go calc.lmn
```

`calc_lm.go` begins like this:

```go
// Code generated by lmn2go from calc.lmn. DO NOT EDIT.

package main

import "github.com/msorc/languagemachine2/lm"

// Program is the ruleset compiled from calc.lmn.
var Program = &lm.Program{
	Name:  "calc",
	Rules: programRules,
	Funcs: map[string]lm.Func{
		"pow":  lmPow,
		"sqrt": lmSqrt,
	},
}

func main() { lm.Main(Program) }

const programRules = `
m:calc L:0 n:1 ( z m:result m:output ) ( m:eof ) r
…
```

`-stubs funcs.go` writes a starting point for the functions the rules call. It never overwrites an existing file:

```go
// lmPow implements pow(...) for the rules.
func lmPow(c *lm.Call) lm.Value {
	// TODO: implement pow
	return lm.Null()
}
```

Until `lmPow` and `lmSqrt` exist, `go build` fails. That is deliberate: a function you forgot to write becomes a compile error, not a wrong answer at run time.

### Step 4: implement the functions

Replace the stubs in `funcs.go`:

```go
package main

import (
	"math"

	"github.com/msorc/languagemachine2/lm"
)

// lmPow implements pow(a, b) for the rules.
func lmPow(c *lm.Call) lm.Value {
	return lm.Num(math.Pow(c.Arg(0).Number(), c.Arg(1).Number()))
}

// lmSqrt implements sqrt(x) for the rules.
func lmSqrt(c *lm.Call) lm.Value {
	return lm.Num(math.Sqrt(c.Arg(0).Number()))
}
```

### Step 5: build and run

```sh
$ go build -o calc .
$ printf '2 ^ 3 ^ 2\n(2 ^ 3) ^ 2\nsqrt 16 + 1\n1 + 2 * 3\n' > calc.input
$ ./calc calc.input
= 512
= 64
= 5
= 7
$ ./calc -input '2^10
'
= 1024
```

`^` at `14R` is right-associative, so `2 ^ 3 ^ 2` is `2 ^ 9`. The binary accepts every `lm` option: input files, `-input`, `-stdin`, `-output`, `-trace m`, `-trace D`, `-add more.lm` and so on. `-rules other.lm` *replaces* the built-in rules, and `-add` extends them.

## 23. Calling Go from the rules

A call `f(a, b)` in an action or expression looks `f` up by name: first among the functions in `Program.Funcs`, then among the builtins. A function in `Funcs` with the same name as a builtin replaces the builtin.

The Go side has a single signature:

```go
type Func func(c *lm.Call) lm.Value
```

| API | Meaning |
| --- | --- |
| `c.Name` | the name the rules used |
| `c.Args`, `c.Arg(i)` | the arguments. `Arg` returns null when `i` is out of range. |
| `v.String()`, `v.Number()`, `v.Bool()`, `v.IsNumber()` | read an argument |
| `lm.Sym(s)`, `lm.Num(x)`, `lm.Null()` | build a result |

For example, a function that turns its argument into a shout:

```go
func lmShout(c *lm.Call) lm.Value {
	return lm.Sym(strings.ToUpper(c.Arg(0).String()) + "!")
}
```

It would be called as `$(shout(W))` on a right side, or as `X = shout(W);` in an action.

The same rules also run under plain `lm`. There, a function that does not exist prints a message and yields 0:

```sh
$ bin/lmn -output calc.lm calc.lmn
$ bin/lm -rules calc.lm -input '2 ^ 3
'
external not found: pow A B
= 0
```

`lmn2go` reports calls it cannot resolve from the bytecode, such as a function taken from a table, `T[i](x)`. Those calls are still bound by name at run time.

## 24. A library package

To use rules inside a larger program, generate a package without `main`:

```
libapp/
├── go.mod
├── main.go
└── calc/
    ├── calc.lmn
    ├── gen.go
    ├── funcs.go        (package calc, the functions from section 22)
    └── calc_test.go
```

`calc/gen.go` keeps the generated code up to date with `go generate`:

```go
// Package calc evaluates infix arithmetic with Language Machine rules.
package calc

//go:generate lmn2go -pkg calc -o calc_lm.go calc.lmn
```

```sh
$ go generate ./...        # needs lmn2go on PATH
note: package calc must define func(*lm.Call) lm.Value: lmPow (pow), lmSqrt (sqrt)
```

`main.go` calls the ruleset with `Translate`, which takes an input string and returns what the rules print:

```go
package main

import (
	"fmt"
	"log"

	"example.com/libapp/calc"
)

func main() {
	out, err := calc.Program.Translate("1 + 2 * 3\n2 ^ 0.5\n")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(out)
}
```

```sh
$ go run .
= 7
= 1.41421
```

`Translate` returns an error when the run exits with a non-zero status, for example when the outermost goal fails or `flagError` was raised. The error includes the exit status and whatever the rules wrote to standard error. Extra `lm` options go before the input: `Translate(text, "-trace", "m")`.

That makes rulesets easy to test with ordinary Go tests:

```go
package calc

import "testing"

func TestCalc(t *testing.T) {
	for in, want := range map[string]string{
		"1 + 2 * 3\n":   "= 7\n",
		"(1 + 2) * 3\n": "= 9\n",
		"2 ^ 3 ^ 2\n":   "= 512\n",
		"sqrt 81\n":     "= 9\n",
	} {
		got, err := Program.Translate(in)
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
```

For full control over arguments and streams, use `Run`, which behaves exactly like the `lm` command and returns its exit status:

```go
status := calc.Program.Run([]string{"calc", "-trace", "m", "in.txt"}, os.Stdout, os.Stderr)
```

## 25. Embedding a ruleset by hand

`lmn2go` is a convenience. An `lm.Program` is just the bytecode plus a function map, so you can also compile with `bin/lmn` and embed the `.lm` file yourself:

```go
package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/msorc/languagemachine2/lm"
)

//go:embed freq.lm
var rules string

var freq = &lm.Program{Name: "freq", Rules: rules}

func main() {
	text, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	out, err := freq.Translate(string(text))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(strings.ToUpper(out))
}
```

```sh
$ bin/lmn -output freq.lm freq.lmn          # the word counter from section 13
$ echo 'one two two three three three' > in.txt
$ go run . in.txt
ONE      1
TWO      2
THREE    3
```

The disadvantage is that nothing checks at build time that every function the rules call is present in `Funcs`, which is the check `lmn2go` gives you.

## 26. Reference: `lmn2go` flags and the `lm` API

```sh
lmn2go -o calc.go -stubs funcs.go calc.lmn     # package main with func main
lmn2go -pkg calc -o calc_lm.go calc.lmn        # library package
lmn2go -o calc.go calc.lm                      # wrap rules that are already compiled
lmn2go -o all.go a.lmn b.lmn                   # several sources compiled as one
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `-o file` | stdout | the Go file to write |
| `-pkg name` | `main` | package name; `main` also writes `func main` |
| `-var name` | `Program` | name of the `*lm.Program` variable |
| `-name name` | first file's base name | program name used in messages |
| `-stubs file` | | write stubs for called functions, unless the file exists |
| `-prefix p` | `lm` | prefix for the Go function names (`pow` → `lmPow`) |
| `-emit-lm file` | | also write the compiled `.lm` |
| `-compiler file.lm` | built in | use another lmn compiler |
| `-import path` | `github.com/msorc/languagemachine2/lm` | import path of the runtime |

The `lm` package:

| Identifier | Purpose |
| --- | --- |
| `Program{Name, Rules, Funcs}` | a compiled ruleset |
| `lm.Main(p)` | run `p` on `os.Args` and exit |
| `p.Run(args, stdout, stderr) int` | run with `lm`'s command line, where `args[0]` is the program name |
| `p.Translate(input, options...) (string, error)` | run on a string and return standard output |
| `Func`, `Call`, `Value`, `Sym`, `Num`, `Null` | the function interface (section 23) |

The design notes, including how the generator finds calls in the bytecode, are in [`lmn2go.md`](lmn2go.md). The lmn compiler `bin/lmn` is itself built this way: `cmd/lmn/lmn.go` is generated by `go generate ./cmd/lmn` from the compiler's own lmn sources.

---

## Where to go next

- **Read real grammars.** `examples/` holds the original release's grammars, from `basics/calc.lmn` to complete D and Java front ends in `translators/`. `examples/lmn/lmn2xfe.lmn` is the lmn compiler's front end, written in lmn, and it is the definitive description of the notation.
- **Reference.** Start at [`lm/README.md`](lm/README.md). [`lm/03-lmn-language.md`](lm/03-lmn-language.md) covers the whole notation, and [`lm/08-examples-and-recipes.md`](lm/08-examples-and-recipes.md) catalogues techniques: output buffers, flattening nested structures, context-sensitive languages such as aⁿbⁿcⁿ, and the lambda calculus.
- **Inside the engine.** [`technical_overview.md`](technical_overview.md), [`internal_machine.md`](internal_machine.md) and [`bytecode.md`](bytecode.md) explain how this Go port implements the machine.
