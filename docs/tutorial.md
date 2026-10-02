# Language Machine tutorial

This tutorial teaches the Language Machine from first principles. It first explains how the machine works, with pictures, before you write any code. It then starts with a one-rule program that copies its input, works up through tokenisers, parsers, calculators and translators, and ends by turning a ruleset into an ordinary Go binary or library.

Each example is a complete program. You can paste it into a file and run it, and the output shown is what this repository's engine produces. The diagrams with numbered brackets (`000001`) are **lm-diagrams**, drawn by the engine itself with `-trace D`. The other pictures are drawn by hand to explain an idea.

The reference material lives elsewhere:

- [`lm/`](lm/README.md) is a digest of the original website: the execution model, lmn syntax, builtins and the lm-diagram.
- [`lmn2go.md`](lmn2go.md) covers the Go code generator.
- [`bytecode.md`](bytecode.md) specifies the `.lm` format.

This tutorial links to those pages instead of repeating them.

## Contents

1. [How the Language Machine works](#1-how-the-language-machine-works)
   - [Rules rewrite a stream of symbols](#rules-rewrite-a-stream-of-symbols)
   - [Two streams and one comparison](#two-streams-and-one-comparison)
   - [The cycle](#the-cycle)
   - [A rule application, step by step](#a-rule-application-step-by-step)
   - [The lm-diagram: the machine draws its own picture](#the-lm-diagram-the-machine-draws-its-own-picture)
   - [Goals nest](#goals-nest)
   - [When a rule fails: backtracking](#when-a-rule-fails-backtracking)
   - [Which rule is tried first](#which-rule-is-tried-first)
   - [The three meanings of `-`](#the-three-meanings-of--)
   - [The whole model on one page](#the-whole-model-on-one-page)
2. [Setting up](#2-setting-up)
3. [Part I: basics](#part-i-basics)
   - [3. The smallest program](#3-the-smallest-program)
   - [4. Substitution](#4-substitution)
   - [5. Recognising something and saying so](#5-recognising-something-and-saying-so)
4. [Part II: grammars](#part-ii-grammars)
   - [6. A top-down sentence checker](#6-a-top-down-sentence-checker)
   - [7. Tokens: priorities, `%`, `repeat` and `option`](#7-tokens-priorities--repeat-and-option)
   - [8. Variables and binding](#8-variables-and-binding)
   - [9. A Polish calculator](#9-a-polish-calculator)
   - [10. An infix calculator: precedence from priorities](#10-an-infix-calculator-precedence-from-priorities)
5. [Part III: advanced techniques](#part-iii-advanced-techniques)
   - [11. Deferred sequences and left recursion](#11-deferred-sequences-and-left-recursion)
   - [12. Collecting repeated items with `each`](#12-collecting-repeated-items-with-each)
   - [13. Actions, tables and `foreach`](#13-actions-tables-and-foreach)
   - [14. Several grammars and `use()`](#14-several-grammars-and-use)
   - [15. Nested alternatives `{| … |}`](#15-nested-alternatives---)
   - [16. Programs that read no input](#16-programs-that-read-no-input)
   - [17. Error recovery and error messages](#17-error-recovery-and-error-messages)
   - [18. Specialising a ruleset with `-add`](#18-specialising-a-ruleset-with--add) (and running rules as scripts)
6. [Part IV: debugging](#part-iv-debugging)
   - [19. Traces and the lm-diagram](#19-traces-and-the-lm-diagram)
   - [20. Pitfalls checklist](#20-pitfalls-checklist)
7. [Part V: building Go binaries](#part-v-building-go-binaries)
   - [21. The pipeline](#21-the-pipeline)
   - [22. A standalone command](#22-a-standalone-command)
   - [23. Calling Go from the rules](#23-calling-go-from-the-rules)
   - [24. A library package](#24-a-library-package)
   - [25. Embedding a ruleset by hand](#25-embedding-a-ruleset-by-hand)
   - [26. Reference: `lmn2go` flags and the `lm` API](#26-reference-lmn2go-flags-and-the-lm-api)
8. [Where to go next](#where-to-go-next)

---

## 1. How the Language Machine works

This section has no exercises. It explains the machine's one mechanism, and every later section is an application of it. Section 2 installs the tools, and after that you can come back and run the three small rulesets used here.

### Rules rewrite a stream of symbols

A Language Machine program is a set of **rules**. Each rule says: *when you see this in the input, carry on as if you had seen that instead.*

```
 'hi'  <-  eof - "HELLO" ;
```

The part before `<-` is the **left side**, the pattern to recognise. The part after it is the **right side**, which names the goal the rule works for and what to substitute:

```
     'hi'      <-      eof    -    "HELLO" ;
     ────              ───    ─    ───────
      │                 │     │       └── substitute this: it goes in front of the input
      │                 │     └────────── "and the goal stays as it was" (explained below)
      │                 └──────────────── the goal this rule works for
      └────────────────────────────────── recognise this in the input
```

The arrow points left because rules are written from the text towards its meaning: `what you see <- what it means`. This is BNF turned round. BNF says "a greeting is `hi`", and lmn says "`hi` is a greeting".

There is no separate lexer, parser and tree walker. The same kind of rule recognises characters, builds tokens, parses phrases, computes values and produces output. What a rule substitutes is analysed again by other rules, exactly as if it had been typed in, and that is how rules build on one another.

### Two streams and one comparison

The machine looks at two things at a time:

- the **goal**: the symbol it wants next
- the **input**: the symbol that is actually next

```
        goal side                         input side
        (what is wanted)                  (what is there)

           eof          ◄─ compare ─►     'h'  'i'  '!'  eof
```

The outermost goal is always `eof`, the end of the input. The input system supplies an `eof` symbol after the last real symbol. The machine has only one ambition, which is to match that `eof`, and it applies rules only because other symbols are in the way.

- If the goal and the input are **the same symbol**, both move on.
- If they **differ**, that is a **mismatch**, and a mismatch is the only thing that ever starts a rule.

A rule resolves a mismatch by consuming the input that was in the way, by producing the symbol that was wanted, or both.

### The cycle

```
             ┌────────────────────────────────────────┐
     ┌──────►│ compare the goal with the input symbol │
     │       └───────────┬────────────────┬───────────┘
     │              same │                │ different: a MISMATCH
     │                   ▼                ▼
     │          both move on     ┌──────────────────────────────────┐
     │                   │       │ take the best untried rule filed │◄─┐
     │                   │       │ under (this input, this goal)    │  │
     │                   │       └────────────────┬─────────────────┘  │
     │                   │                        ▼                    │
     │                   │       ┌──────────────────────────────────┐  │
     │                   │       │ RECOGNISE: match its left side   ├──┘
     │                   │       │ against the input                │ failed: put
     │                   │       └────────────────┬─────────────────┘ everything back
     │                   │                        ▼ matched
     │                   │       ┌──────────────────────────────────┐
     │                   │       │ SUBSTITUTE: put its right side   │
     │                   │       │ in front of the remaining input  │
     │                   │       └────────────────┬─────────────────┘
     └───────────────────┴────────────────────────┘
```

Two details complete the picture:

- **Recognising is the same cycle, one level down.** While a rule's left side is being matched, the symbols of that left side are the goals. A mismatch there starts another rule inside the first one.
- **If no rule is left to try**, the mismatch is unresolved. The rule that was being recognised fails, and its own alternatives are tried. If the outermost goal fails, the run ends with exit status 1.

### A rule application, step by step

Here is a complete ruleset of two rules (`hi.lmn`):

```
 'hi'      <- eof - "HELLO" ;
 - out     <- eof - ;
```

The second rule is the copy rule. Its left side starts with `-`, which means "whatever the input is", followed by `out`, a built-in symbol that consumes one input symbol and prints it. Its right side substitutes nothing.

With the input `hi!` the machine goes through these steps:

| Step | Goal | Input, next symbol first | What happens |
| --- | --- | --- | --- |
| 1 | `eof` | `'h' 'i' '!' eof` | Mismatch. The rule `'hi'` starts with the input symbol `'h'`, so it is tried first. |
| 2 | `'i'` | `'i' '!' eof` | **Recognition.** The `'h'` is taken, and the rest of the left side, `'i'`, becomes the goal. It matches. |
| 3 | `eof` | `HELLO '!' eof` | **Substitution.** The right side puts the symbol `HELLO` in front of the input. The goal is still `eof`. |
| 4 | `eof` | `HELLO '!' eof` | Mismatch again. No rule starts with `HELLO`, so the copy rule runs: `out` consumes `HELLO` and prints it. |
| 5 | `eof` | `'!' eof` | Mismatch. The copy rule prints `!`. |
| 6 | `eof` | `eof` | The goal matches the input. The run ends successfully. |

```sh
$ bin/lm -rules hi.lm -input 'hi!'
HELLO!
```

Every rule application has these two phases: **recognition** of the left side, then **substitution** of the right side. Either phase can be empty. The copy rule has no substitution, so it only removes symbols (and prints them on the way).

### The lm-diagram: the machine draws its own picture

The engine can draw what it does. This is the same run, drawn with `-trace D`:

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

(The program's own output shares the terminal with the diagram. It has been removed from the diagrams in this tutorial.)

Read it from top to bottom. Each line is one step:

- **The two middle columns are the two streams**: the goal on the left and the input symbol on the right. `eof 'h'` means "the goal is `eof`, the input is `'h'`".
- **`?` marks a mismatch.**
- **Brackets on the left are recognition.** `┌───000001` opens the left side of a rule, and `└───000001` closes it. The lines between are that rule's left side being matched: here `'i'` meets `'i'`. The number is the serial number of the mismatch that started the rule.
- **Brackets on the right are substitution.** `└───000001---- ----000001───┐` closes the recognition and opens the right side of the same rule. While the bar `│` runs down the right edge, the input symbols come from that right side, not from the real input. `000001───┘` closes it when its symbols are used up. A `?` on that bar marks a mismatch on a substituted symbol.
- A goal with no bracket to its left is the outermost goal. An input symbol with no bar to its right is real input.

So the diagram says: mismatch 1 started the `'hi'` rule, which recognised `'i'` and substituted `HELLO`. Mismatch 2 started the copy rule on `HELLO`, inside that substitution. Mismatch 3 started the copy rule on the real `'!'`. Then `eof` met `eof`.

Recognition brackets nest inside recognition brackets, and substitution bars inside substitution bars. Those two nestings are all there is to the machine, and the rest of this section shows what they look like.

### Goals nest

A left side can contain **nonterminals**: names such as `greeting` that never occur in the real input. They can only be matched if some other rule substitutes them. This ruleset (`nest.lmn`) answers `ok` to a greeting followed by a name:

```
 - greeting name  <- eof - "ok" ;
 'hi '            <- greeting ;
 'bob'            <- name ;
 'tom'            <- name ;
 - out            <- eof - ;
```

```sh
$ bin/lm -rules nest.lm -dwidth 30 -trace D -input 'hi tom'
                     eof 'h'
?                    eof 'h'
┌───────000001
│               greeting 'h'
│?              greeting 'h'
│┌──────000002
││                   'i' 'i'
││                   ' ' ' '
│└──────000002---------- ----------000002────────┐
│               greeting greeting                │
│                                  000002────────┘
│                   name 't'
│?                  name 't'
│┌──────000003
││                   'o' 'o'
││                   'm' 'm'
│└──────000003---------- ----------000003────────┐
│                   name name                    │
│                                  000003────────┘
└───────000001---------- ----------000001────────┐
                     eof ok                      │
?                    eof ok                      ?
┌───────000004                                   │
│               greeting ok                      │
│?              greeting ok                      ?
│-              greeting ok                      │
│                    out ok                      │
└───────000004                                   │
│                                  000001────────┘
                     eof eof
```

Follow the brackets:

1. `eof` meets `'h'`. The rule `- greeting name <- eof - "ok"` starts (bracket 1). Its left side makes `greeting` the goal.
2. `greeting` meets `'h'`: a mismatch *inside* bracket 1. The rule `'hi ' <- greeting` starts (bracket 2), recognises `'i'` and `' '`, and substitutes the symbol `greeting`. Now `greeting` meets `greeting`, and the outer rule moves on to `name`.
3. `name` meets `'t'`. The rule `'tom' <- name` does the same (bracket 3).
4. Bracket 1 closes. Its right side substitutes `ok`, and the copy rule prints it (bracket 4).

Rule 1 behaves like a function that calls `greeting` and `name`, as in a recursive-descent parser. Rules 2 to 4 look like a lexer. The machine does not distinguish them. Both are rules started by a mismatch.

### When a rule fails: backtracking

The same ruleset with an input that does not fit:

```sh
$ bin/lm -rules nest.lm -dwidth 30 -trace D -input 'hi x'
                     eof 'h'
?                    eof 'h'
┌───────000001
│               greeting 'h'
│?              greeting 'h'
│┌──────000002
││                   'i' 'i'
││                   ' ' ' '
│└──────000002---------- ----------000002────────┐
│               greeting greeting                │
│                                  000002────────┘
│                   name 'x'
│?                  name 'x'
│-                  name 'x'
│                    out 'h'
└───────000001
                     eof 'i'
?                    eof 'i'
┌───────000003
│               greeting 'i'
│?              greeting 'i'
│-              greeting 'i'
│                    out 'i'
└───────000003
                     eof ' '
?                    eof ' '
┌───────000004
│               greeting ' '
│?              greeting ' '
│-              greeting ' '
│                    out ' '
└───────000004
                     eof 'x'
?                    eof 'x'
┌───────000005
│               greeting 'x'
│?              greeting 'x'
│-              greeting 'x'
│                    out 'x'
└───────000005
                     eof eof
```

- The `greeting` is recognised as before. Then `name` meets `'x'`, and no rule can resolve that.
- **`-` marks a backtrack.** The rule in bracket 1 fails, and the machine puts everything back as it was at mismatch 1: the input is at `'h'` again, as the line `out 'h'` shows.
- The next candidate for (goal `eof`, input `'h'`) is the copy rule, which prints `h`.
- For each of the remaining characters the longer rule is tried first, fails straight away (`-`), and the copy rule takes over. The output is `hi x`, unchanged.

Backtracking is thorough at each mismatch: every candidate is tried. But once a rule has succeeded, the machine never goes back into it to make it match differently. The original author called the strategy "greedy fastback".

Look again at the end of the previous diagram, bracket 4. The same thing happened there: for the input `ok` the longer rule `- greeting name` was tried first, failed (`-`), and the copy rule took over.

### Which rule is tried first

Rules are filed under two symbols:

- the **left-initial**, the first symbol of the left side, which has to match the **input**
- the **right-initial**, the first symbol of the right side, which names the **goal** the rule serves

Either can be `-`, "don't care". That gives four kinds of rule. For a mismatch between goal `g` and input `i` they are tried in this order:

| Order | Rule shape | Style | Example |
| --- | --- | --- | --- |
| 1 | `i … <- g …` | specific: this input, for this goal | `'hi ' <- greeting ;` |
| 2 | `i … <- - …` | bottom-up: this input, whatever the goal | `' ' <- - ;` (delete spaces anywhere) |
| 3 | `- … <- - …` | speculative (rare) | |
| 4 | `- … <- g …` | top-down: this goal, whatever the input | `- greeting name <- eof - ;` |

```
                              the right side starts with
                              the goal g           -  (any goal)
                            ┌────────────────────┬────────────────────┐
 the left side   input i    │ 1  specific        │ 2  bottom-up       │
 starts with                ├────────────────────┼────────────────────┤
                 - (any)    │ 4  top-down        │ 3  speculative     │
                            └────────────────────┴────────────────────┘
```

Within one kind, **longer left sides are tried before shorter ones, and among equals the newest rule (the one later in the file) goes first.** The "length" counts the matchable items at the outermost brace level, so a left side wrapped in `{ … }` has length 0 and is tried last. Section 17 uses that on purpose.

Bottom-up rules are what make the machine more than a recursive-descent parser. A rule such as `' ' <- - ;` applies under any goal, so one line deletes spaces everywhere, and a rule can begin with a nonterminal that an earlier rule produced (section 11).

The order explains the two runs above. For (goal `eof`, input `'h'`) the `'hi'` rule of `hi.lmn` is kind 1 and beats the copy rule, which is kind 4. In `nest.lmn` both rules for `eof` are kind 4, and `- greeting name` is longer than `- out`, so it is tried first.

### The three meanings of `-`

1. At the start of the **left** side: any input. The rule is top-down.
2. At the start of the **right** side: any goal. The rule is bottom-up.
3. **Directly after the right-initial**, as in `<- eof - "HELLO"`: the rule is filed under that goal, but it does not produce it. Only the symbols after the `-` are substituted, and the goal stays in force.

The third form is what makes loops. Compare:

```
 'hi '   <- greeting ;        substitutes the symbol greeting: the goal greeting is satisfied
 'hi'    <- eof - "HELLO" ;   substitutes only HELLO: the goal is still eof afterwards
 - out   <- eof - ;           substitutes nothing: the goal is still eof afterwards
```

A rule of the second or third shape can apply again and again under the same goal, which is how `- out <- eof - ;` copies a whole file.

### The whole model on one page

| Term | Meaning |
| --- | --- |
| symbol | A character such as `'a'` (a *terminal*, which can occur in real input) or a name such as `greeting` (a *nonterminal*, which only rules can produce). |
| goal | The symbol the machine wants next: `eof`, or the next symbol of a left side that is being recognised. |
| mismatch | The goal and the input symbol differ. Nothing else starts a rule. |
| recognition | Matching a rule's left side against the input. Drawn as brackets on the left of the lm-diagram. |
| substitution | Putting a rule's right side in front of the input. Drawn as bars on the right. |
| context | One rule application's recognition phase, with everything nested inside it. Variables (section 8), priorities (section 7) and the current grammar (section 14) belong to contexts. |
| backtracking | When a left side cannot be completed, the machine restores its state and tries the next candidate rule. |

Three things were left out, and each gets its own section:

- **Priorities** (section 7) can forbid a rule from starting inside certain contexts. That gives tokens, operator precedence and associativity.
- **Variables** (section 8) carry values from where they are recognised to where they are used.
- **Grammars** (section 14) are named sets of rules, of which one is current at a time.

The full model is in [`lm/02-execution-model.md`](lm/02-execution-model.md), and the legend of the diagram in [`lm/06-lm-diagram.md`](lm/06-lm-diagram.md).

---

## 2. Setting up

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

```
 hello.lmn ──bin/lmn──► hello.lm ──┐
 (rules, as text)       (bytecode) ├──bin/lm──► output
                        input ─────┘
```

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

To see any example in this tutorial as an lm-diagram, add `-dwidth 30 -trace D` before the input. Options take effect in the order given, so `-dwidth` must come before `-trace D`. Section 19 has more on tracing.

---

# Part I: basics

## 3. The smallest program

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

This is the copy rule from section 1, on its own. In `- out <- eof - ;`:

- **`eof` is the goal** the rule is filed under. It is the machine's outermost goal, so the rule is a candidate at every mismatch at the top level.
- **The leading `-` on the left** means "whatever the input symbol is". The rule does not care what it sees.
- **`out`** is a special symbol. As a goal, it consumes one input symbol and writes it to standard output.
- **`eof -` on the right** means that the rule substitutes *nothing*: whatever follows the second `-` is substituted, and here nothing follows. So the goal remains `eof`.

Here is the run, step by step:

1. The goal is `eof` and the input is `'h'`. They differ, which is a **mismatch**.
2. The machine looks for rules relevant to (goal `eof`, input `'h'`) and finds `- out <- eof - ;`.
3. The left side `out` consumes `'h'` and prints it. The right side substitutes nothing, so the goal is still `eof`.
4. Steps 1–3 repeat until the real `eof` arrives. It matches the goal, and the run stops with status 0.

The lm-diagram for the input `hi` shows one small bracket per character and nothing on the right, because nothing is ever substituted:

```sh
$ bin/lm -rules cat.lm -dwidth 30 -trace D -input 'hi'
                     eof 'h'
?                    eof 'h'
┌───────000001
│                    out 'h'
└───────000001
                     eof 'i'
?                    eof 'i'
┌───────000002
│                    out 'i'
└───────000002
                     eof eof
```

## 4. Substitution

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

```
 before:   goal eof     input  'c' 'a' 't' ' ' 's' 'a' 't' …
                               └────┬────┘
                           recognised by 'cat'
                               ┌────┴────┐
 after:    goal eof     input     dog      ' ' 's' 'a' 't' …
```

This is the central idea of the machine: **a rule's right side is treated as if it had appeared in the input**. The substituted material is analysed again, which is what lets rules build on one another.

### `'single'` versus `"double"` quotes

| Form | Meaning |
| --- | --- |
| `'abc'` | three terminal (character) symbols `'a' 'b' 'c'` |
| `"abc"` | one symbol whose text is `abc` |

Both print the same way. The difference matters when the substitution is analysed again. If the rule were `'cat' <- eof - 'cats' ;`, the substituted characters would start with `c a t` again, and the rule would fire forever. With `"cats"`, a single non-character symbol, the problem cannot arise.

### Why `cat` wins over `- out`, and what happens when it fails

Both rules are relevant when the goal is `eof` and the input is `'c'`. By the order of section 1, `'cat'` names the input symbol explicitly (kind 1), so it comes before `- out` (kind 4). If `'cat'` fails, the machine **backtracks** and tries `- out` instead. Here is the diagram for the input `cow`:

```sh
$ bin/lm -rules swap.lm -dwidth 30 -trace D -input 'cow'
                     eof 'c'
?                    eof 'c'
┌───────000001
│                    'a' 'o'
│?                   'a' 'o'
│-                   'a' 'o'
│                    out 'c'
└───────000001
                     eof 'o'
?                    eof 'o'
┌───────000002
│                    out 'o'
└───────000002
                     eof 'w'
?                    eof 'w'
┌───────000003
│                    out 'w'
└───────000003
                     eof eof
```

In bracket 1 the `'cat'` rule has taken `'c'` and wants `'a'`, but the input is `'o'`. The `?` is that mismatch, and no rule resolves it. The `-` is the backtrack: the input goes back to `'c'`, and `out` prints it. The `'o'` and `'w'` never start the `'cat'` rule at all, because it is only filed under `'c'`.

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

### The diagram: one substitution feeds two goals

The diagram for the input `bye` shows how the pieces interlock. (The symbol `See you.` ends with a newline, which is written `\n` here to keep the columns straight.)

```sh
$ bin/lm -rules greet.lm -dwidth 40 -trace D -input 'bye'
                          eof 'b'
?                         eof 'b'
┌────────────000001
│                    greeting 'b'
│?                   greeting 'b'
│┌───────────000002
││                        'y' 'y'
││                        'e' 'e'
│└───────────000002---------- ----------000002─────────────┐
│                    greeting greeting                     │
│                      output See you.\n                   │
│?                     output See you.\n                   ?
│┌───────────000003                                        │
││                        eom See you.\n                   │
││?                       eom See you.\n                   ?
││┌──────────000004                                        │
│││                       out See you.\n                   │
││└──────────000004                                        │
││                        eom eom                          │
││                                      000002─────────────┘
│└───────────000003---------- ----------000003─────────────┐
│                      output output                       │
└────────────000001                                        │
│                                       000003─────────────┘
                          eof eof
```

- Bracket 2 is the `'bye'` rule. Its right side (the bar labelled `000002`) holds three symbols: `greeting`, the text, and `eom`.
- The first of them, `greeting`, satisfies the goal inside bracket 1. The outer rule moves on to its next goal, `output`.
- `output` meets the text. Brackets 3 and 4 are `- eom <- output` and `- out <- eom -`, which print the text. They consume symbols from substitution 2, so the bar for 2 stays open until `eom` meets `eom`.
- Bracket 3 then substitutes `output`, which completes bracket 1.

A right side does not have to be consumed by the rule that wanted it first. Whatever is left over stays in the input for the goals that follow, and this is how a recognising rule hands text to the output rules.

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

The top-down rules spell out a tree of goals. For `the cat likes a dog .` the machine works through it from the top, left to right, and each goal is satisfied by a rule that substitutes it:

```
                          sentence
          ┌───────────┬──────┴─────┬───────────┐
       subject       verb        object       '.'
          │           │            │
      nounphrase   'likes '    nounphrase
       ┌──┴───┐                 ┌──┴───┐
    'the '   noun             'a '    noun
              │                        │
           'cat '                   'dog '
```

In the lm-diagram the same tree appears lying on its side, as recognition brackets nested five deep. [`lm/06-lm-diagram.md`](lm/06-lm-diagram.md) prints that diagram in full for a slightly simpler version of this grammar.

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

```
 priority 20      ┌─ word ─┐       ┌─ word ─┐      inside a word: ' ' (20L) may not start,
                  H e l l o   ' '  w o r l d       because 20 is not greater than 20
 priority 0    ───────────────────────────────     outside: 20 is greater than 0, so ' ' is deleted
```

The lm-diagram for the input `ab c` shows both cases. Inside the word rule (bracket 2), the class `[a-zA-Z]` meets the space. That mismatch (`?`) is not resolved, so the `repeat` stops (`-`), and the word is converted and bound. Later, at the top level, the same space meets the goal `eof` and is deleted without a trace beyond its `?`:

```
│┌──────000002
││                     % 'b'
││                repeat 'b'
││              [a-zA-Z] 'b'
││                     % ' '
││*
││              [a-zA-Z] ' '
││?             [a-zA-Z] ' '
││-             [a-zA-Z] ' '
││                 toSym ' '
││                     : ' '
│└──────000002---------- ----------000002────────┐
│                   word word                    │
│                      : :                       │
    …
                     eof ' '
?                    eof ' '
                     eof 'c'
```

`*` marks a turn of a `repeat`. Symbols such as `%`, `toSym` and `:` appear in the goal column because they are steps of the left side like any other. A rule whose whole left side is the one symbol that started it, and whose right side is empty, draws no brackets, which is why the deleted space shows only its `?`.

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

```
 ┌─ start var Total = 0; var Count = 0;        the context that owns Total and Count
 │
 │   ┌─ - number :N  Total = Total + N; …      starts inside it, so it sees Total
 │   └─
 │   ┌─ - number :N  Total = Total + N; …      and so does every later one
 │   └─
 │
 │   eof                                       the last symbol of the left side
 └─► generate "count: " Count ", total: " Total "\n" output eof
```

This is the scope rule read off the lm-diagram: a rule application sees the variables of every recognition bracket that encloses it.

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

The diagram for `+ 2 3` shows values nesting inside values. It is abridged here: `…` stands for the steps inside the number rule and for the output rules at the end.

```
││                     x '+'
││?                    x '+'
││┌─────000003                                         '+' x :A x :B starts
│││                    x ' '
│││?                   x ' '                           the space is deleted
│││                    x '2'
│││?                   x '2'
│││┌────000005                                         the number rule
    …
│││└────000005---------- ----------000005────────┐
│││                    x x                       │     it substitutes x :2,
│││                    : :                       │     and :A receives the value
│││                                000005────────┘
    …                                                  the same again for 3 and :B
││└─────000003---------- ----------000003────────┐
││                     x x                       │     the '+' rule substitutes x :(A + B)
││                     : :                       │
││                                 000003────────┘
```

An operand can itself be an operator expression, as in `* 100 / 1 3`. Then a whole bracket like 3 sits where bracket 5 is here.

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

```
                ┌──────────────────── round again ───────────────────┐
                ▼                                                    │
 goal expr:  opnd :A  then goal op ──┬── '+' expr :B  ──►  op  opnd :(A + B)
                                     │
                                     └── anything else ──►  op  expr :A      finished
```

The empty-left-side rule is easy to spot in an lm-diagram, because it has a substitution bar and no recognition bracket. This is the end of the sub-expression `2` in `1+2`:

```
│││││                 op '\n'
│││││?                op '\n'
│││││                              000008────────┐
│││││                 op op                      │
││││└───000006                                   │
││││                expr expr                    │
││││                   : :                       │
││││                               000008────────┘
```

### Precedence and associativity

The priorities decide which operator rule may start inside which, and that fixes the shape of the result:

```
   1 + 2 * 3            10 - 4 - 3           (1 + 2) * 3

       +                     -                    *
      / \                   / \                  / \
     1   *                 -   3                +   3
        / \               / \                  / \
       2   3            10   4                1   2
```

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

Left recursion has a characteristic shape in the lm-diagram: the substitution of one application overlaps the recognition of the next. Here is the middle of the run for `a b;`, with the steps inside the `item` rule left out:

```
│└───────────000002---------- ----------000002─────────────┐
│                      result list                         │     - item :X  substituted  list :X :X
│?                     result list                         ?     list is not result: a mismatch
│┌───────────000005                                        │     list :F :B item :X  starts
││                          : :                            │     F and B are bound
││                          : :                            │
││                                      000002─────────────┘
││                       item ' '
││?                      item ' '
││                       item 'b'
││?                      item 'b'
    …
│└───────────000005---------- ----------000005─────────────┐
│                      result list                         │     a longer list, substituted
│?                     result list                         ?     and the same thing happens again
│┌───────────000008                                        │
```

Each `list` is produced on the right and immediately becomes the first symbol recognised by the next bracket on the left. The brackets do not get deeper as the list grows, so a left-recursive rule runs as a loop.

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

```
 input:     Use *bold* and _italic_, but `a*b_c` stays.
 grammar:   text ─────────────────────── code ── text ─
                                         ▲     ▲
                         '`' use("code") ┘     └ '`' <- span closes the rule that switched
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

### The mismatch trace

Take `hi.lmn` from section 1:

```
 'hi'      <- eof - "HELLO" ;
 - out     <- eof - ;
```

```sh
$ bin/lm -rules hi.lm -trace m -input 'hi!'
   1   ??    0    0     0L    0    1    0      1      lm_         eof         'h'         ---
   1   ??    0    1     0L    0    1    0      2      lm_         eof       HELLO         'i'
   1   ??    0    0     0L    0    1    0      3      lm_         eof         '!'       HELLO
```

Each `??` line is a mismatch. Among other columns, it shows the current grammar, the goal (`eof`), the input symbol, and the previously matched symbol.

The trace and the program's output share standard output, so they interleave. Each trace line starts with a tab, and whatever the program has printed since the previous trace line appears before that tab. On a terminal the `HELLO` therefore shows up at the start of the third line, and the `!` after it. They have been removed here, as in the diagrams.

### The lm-diagram

Section 1 introduced the diagram, and most sections since have shown one. This is the complete list of its marks:

| Mark | Where | Meaning |
| --- | --- | --- |
| `goal input` | the two middle columns | one step: the goal symbol and the input symbol |
| `?` | left | a mismatch at this nesting level |
| `-` | left | a backtrack: the rule or the `repeat`/`option` being tried was given up |
| `*` | left | another turn of a `repeat` |
| `┌───NNNNNN` | left | recognition of a rule's left side starts. `NNNNNN` is the number of the mismatch that started it. |
| `└───NNNNNN` | left | recognition ends, either successfully or after a backtrack |
| `----- -----NNNNNN───┐` | right | the same rule's right side is substituted |
| `│` and `?` | right edge | the input symbol comes from that substitution; `?` is a mismatch on it |
| `NNNNNN───┘` | right | the substitution is used up |

And these are the shapes worth learning to recognise:

| Shape | What it is | Example |
| --- | --- | --- |
| a left bracket with nothing on the right | a rule that only consumes ("something to nothing") | the copy rule, section 3 |
| a right bar with no left bracket | a rule with an empty left side ("something from nothing") | `- <- op expr :A ;`, section 10 |
| a `?` with no bracket at all | a one-symbol rule with an empty right side | whitespace deletion, section 7 |
| left brackets nested deeper and deeper | top-down analysis, or right recursion | the sentence checker, section 6 |
| a right bar that overlaps the next left bracket, at constant depth | left recursion | the list builder, section 11 |
| `?` followed by `-` | an alternative that failed | `cow`, section 4 |

Practical points:

- `-dwidth N` sets the width of each half. Deeply nested grammars need more than the 30 used in this tutorial, and the default is 80.
- The program's output is written at the start of the diagram lines, before a tab. When the output contains no tabs or newlines, `… | cut -f2-` removes it.
- A symbol that contains a newline breaks the line it is drawn on. That is harmless, but it is easier to study a grammar on input that produces short symbols.
- Start with the smallest input that shows the problem. A diagram has one line per step.

[`lm/06-lm-diagram.md`](lm/06-lm-diagram.md) explains the idea behind the diagram and walks through the `cats` grammar.

### Guard rails

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

Without `-stubs`, `go build` fails until `lmPow` and `lmSqrt` exist. That is deliberate: a function you forgot to write becomes a compile error, not a wrong answer at run time. The stubs do compile, and they return null, so a calculator built with them unchanged prints `= null` for `^` and `sqrt`.

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
