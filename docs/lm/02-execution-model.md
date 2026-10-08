# Execution Model

Sources: `guide.html`, `glossary.html`, `picturebook.html`, `special_symbols.html`, `lmn2xfe.html`, `bottles.html`.

## Symbols, patterns, rules

- A **symbol** is a token that is either the same as or different from another symbol.
  - **Terminal symbols** can appear in the external input. They are usually characters, written `'a'`.
  - **Nonterminal symbols** appear only in rules, as names such as `noun`, `x` or `eof`.
  - The input could also be a stream of tokens rather than characters. The distinction that matters is whether a symbol can occur in the external input.
- A **pattern** is a sequence of symbols. Two patterns match if their symbols match pairwise, in order.
- A **rule** has a left side (the pattern to recognise) and a right side (what to substitute). Applying a rule means matching its left side at the current point in the input, then carrying on *as if the right side had appeared in the input* in place of what was matched. Either side may be empty. Either side may contain actions.
- A **grammar** is a named collection of rules. A ruleset may contain several grammars, and only one grammar is current at a time.

## The obsessive machine: the goal `eof`

The machine has one objective: to **match the symbol `eof`**, which the input system appends after all external input.

- If the input is empty, it matches `eof` straight away and stops with success.
- `eof` is the ultimate goal, not a starting point.
- A ruleset must therefore contain at least one rule that helps match `eof`, for example `- result output <- eof - ;`.

### `start`

Before reading any input, the machine behaves as if there were a mismatch between the goal `eof` and the input symbol `start`. It looks in the grammar that owns the **first rule of the loaded ruleset**, and any rules relevant to (`eof`, `start`) are tried first. This is the standard way to set up an outermost context, with `var` declarations, tables and so on, before any input is consumed:

```
start var N = 99; <- eof - first :N :"s" thing :N song eof;
```

If there is no such rule, the analysis starts normally with `eof` as the goal.

## Mismatch is the central event

The machine keeps comparing a **goal symbol** with an **input symbol**:

- The goal symbol, also called the *left symbol* or *context symbol*, is `eof` or a symbol produced by the left side of a rule that is being matched.
- The input symbol, also called the *right symbol*, comes from the external input or from a right side that has been substituted into the input.

If the two match, both advance. If they differ, that is a **mismatch**, and **rules are only ever started in response to a mismatch**, and only by mismatches they are relevant to. A mismatch is **resolved** when a rule, directly or indirectly, consumes the offending input symbol, produces the wanted goal symbol, or both.

## Which rules are relevant

Every rule is tied to a class of mismatch events by its **left-initial symbol** L and its **right-initial symbol** R. Either one can be "don't care", which is written `-`. When the goal `g` meets the input `i`, relevant rules are tried in the following category order:

| Order | L | R | Style | Example | Relevant when |
| --- | --- | --- | --- | --- | --- |
| 1 | `i` | `g` | specific / direct | `'cat' <- noun;` | input `'c'`, goal `noun` |
| 2 | `i` | `-` | bottom-up (input-driven, LR-like) | `'\n' <- - ;` | input `'\n'`, any goal |
| 3 | `-` | `-` | speculative (rare) | `- word :W y <- - W;` | always |
| 4 | `-` | `g` | top-down (goal-driven, recursive descent) | `- line <- error;` | goal `error`, any input |

**Within each category, longer rules are tried before shorter ones, and newer rules before older ones.** Consequences:

- A later rule overrides an earlier rule of the same effective length in the same context. This is how you specialise a general ruleset by adding rules with `-add`.
- The **effective length** of a left side is the number of matchable items at the outermost brace level. Matchable items are terminals, nonterminals, `:` and `%`. So `- { error :X restOfLine } <- something;` has length 0 and is tried after every rule of non-zero length in its group.

### The three uses of "that hyphen"

1. **At the start of the left side:** don't care about the input symbol. The rule is top-down and relevant only to the goal.
2. **At the start of the right side:** don't care about the goal symbol. The rule is bottom-up and can apply in any context that priorities allow, for example deleting whitespace with `.[ \t\n] <- - ;`.
3. **Directly after the right-initial symbol:** the rule is filed as relevant to that goal symbol, but it substitutes only what follows the `-`. It does not produce the goal. After `"exit" <- code - 'return';` has been applied, `'return'` is in the input and the goal is *still* `code`.

   This "sideways" form ties a whole family of rules to a context. Output filters are the standard example: `- out <- eom - ;` keeps consuming and printing symbols while the goal stays `eom`.

## Priorities

A grammar selector can carry a priority, for example `.calc(20L)`. Rules that follow it get that priority, and a new context takes on the priority of the rule that started it. A rule with no priority, under `.name()`, lets the new context **inherit** the priority of the enclosing context.

| Direction | Name | Can start when |
| --- | --- | --- |
| `L` | left-associative | the rule's priority is **greater** than the context's. Equivalently, it cannot start in a context of equal or higher priority. |
| `R` | right-associative | the rule's priority is **greater than or equal to** the context's. This allows nesting at the same level. |
| `B` | bracket | always. It starts a new priority level. |
| `M` | maximal | the guide calls it "left associative, maximal priority - no further nesting allowed". The glossary says "can always start, but they prevent further nesting". |
| (lexical) | — | the docs say terminal symbols carry a special high left-associative priority, which stops the machine from starting new rules when the goal in a mismatch is a terminal other than `eof`, and that `-l`/`--lexpri` changes it. The original engine stores the `-l` value but never applies it: a terminal goal is resolved at the priority of its context. The Go port follows the code (see `README.md`). |

The canonical use is keeping whitespace significant inside tokens:

```
.calc(20L)
  .[0-9] % { repeat .[0-9] % } { option '.' % repeat .[0-9] % } toNum :N <- x :N;
  .[ \t\n]                                                               <- -   ;
```

Once the number rule has started, its context has priority 20L. The space-deletion rule is also 20L, so it cannot start inside the number, and a space ends the number. Everywhere else, spaces are deleted. Output rules are usually given a *higher* right-associative priority, such as 30R, so that whitespace reaches the output instead of being deleted.

## Contexts, alternatives and backtracking

- When a relevant rule with a non-empty left side is tried, it opens a new level of **left-side nesting**, called a context.
- If the left side matches, the context closes and the right side opens a new **right-side (substitution) level**, provided the right side is non-empty.
- A mismatch inside the context starts an inner level.
- If a mismatch cannot be resolved, because there are no relevant rules or they all fail, **the rule fails**. The machine resets itself as far as possible to its state at the mismatch that opened the context, and tries the **next alternative**. If there are no alternatives left, the enclosing context fails, and so on outwards.
- **Backtracking is fairly complete, with one exception:** it does not undo side-effect actions that changed the state of enclosing contexts.
- **It is tenacious, but not exhaustive.** Every alternative is tried at each mismatch. However, once a rule has succeeded there is no way back into its context. Unlike Prolog, it does not unpick every step. If backtracking reapplies the same rule to the same material, the rule behaves the same way unless side effects have changed something.
- The guide describes the strategy as "greedy fastback": the longest and newest rule is tried first, and alternatives are tried only on failure. An ambiguous grammar will therefore not behave the way an exhaustive parser would.
- Left recursion works naturally. The lm-diagram shows the left and right nestings overlapping at each iteration. Deep backtracking should be avoided when performance matters (see `leftrecursive.html`).

### Collapsing right-side levels (0.2.1+)

Since version 0.2.1, completed right-side recursion levels are forced to collapse before new ones start. A side effect is that rulesets which never consume external input still tend to try to read some input when they return to the outermost level. The fix is to supply a dummy input string with `-i dummyInput` (Go: `-input`) that is never consumed. Before 0.2.1, a "something from nothing" rule `- <- sync;` was matched by each iteration to force the collapse (see `bottles.html`).

## Grammars

- `.name(prio)` makes the following rules belong to grammar `name`.
- Only rules in the **current** grammar are candidates for resolving a mismatch.
- The builtin `use(G)` selects grammar `G`. It becomes current for the rest of the current left-side nesting level and for all inner levels.

## Variables and scope

- Variables are created by binding (`:Name`) during matching, or by `var` declarations.
- **Left side:** a reference sees variables created so far in the current context or in any enclosing context. Only the most recent instance of a name is visible, because newer instances mask older ones.
- **Right side:** a reference sees the variables that were visible at the *end of the match phase* of the corresponding left side.
- A value bound from a right-side pattern, such as `:{ ... }`, is a **deferred sequence**, which works like a closure. It is not evaluated until the variable is evaluated, which may be never, and then it is evaluated in the scope of the rule application that provided it. Binding cost is constant however long the sequence is, so concatenating lists is cheap: `append list :A list :B <- list :{ A B };`.
- "The ghost of a parse tree": if you think of the nesting of substitutions as a tree, a reference can see back along its branch towards the root. Fragments of the transformed representation build up as bindings as the analysis proceeds. They cost almost nothing to throw away if backtracking chooses another analysis.

## Errors and exit status

The machine keeps an error counter and a warning counter, which `flagError` and `warnError` increment. If the error count is non-zero at the end of a run, the program returns an error status to the host environment.
