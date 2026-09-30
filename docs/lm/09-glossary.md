# Glossary

Condensed from `glossary.html`. The links point to the fuller treatment elsewhere in these notes.

- **acquisition**: Nothing that is consumed is kept unless a rule explicitly acquires it, either with `%` or with an output buffer `(Var)`. See [03](03-lmn-language.md#acquisition-with-).
- **action**: A side effect embedded in a rule, written in a JavaScript subset. On the right side it must be enclosed in braces.
- **alternative**: One of several rules relevant to the same mismatch. The order is specific, then bottom-up, then speculative, then top-down. Within each group, longer and newer rules come first.
- **backtrack**: When an alternative fails, the machine resets itself as far as it can to the state at the mismatch and tries the next alternative. Side effects on enclosing contexts are not undone.
- **binding**: Attaching a value to a variable name within a scope, through `:`. Two values meeting at `:` are matched instead of bound. See [03](03-lmn-language.md#binding-with-).
- **bootstrap**: The hand-seeded compiled ruleset `lmnbs.lm`, which compiles the lmn compilers. See [07](07-compilation-and-bytecode.md#the-bootstrap).
- **bottom-up rule**: A rule relevant only to the input symbol. Its right side starts with `-`, as in `'+=' <- - "+=";`.
- **context**: A level of left-side (match-phase) nesting. A context opens at a mismatch for which at least one relevant rule exists. Its goal symbol is the *context symbol*. The nesting of contexts determines variable scope.
- **dash / hyphen**: `-` means "never mind". At the start of the left side it makes the rule top-down. At the start of the right side it makes the rule bottom-up. After the right initial it means the rule substitutes only what follows the dash.
- **direction**: The part of a priority that says how rules nest: `L` needs a priority greater than the context's, `R` needs greater or equal, `B` can always start and opens a new level, and `M` is maximal and allows no further nesting.
- **goal**: The symbol currently being matched. The initial and ultimate goal is `eof`.
- **grammar**: In the Language Machine, a named collection of rules, strictly a sub-grammar. Only one is current at a time, and `use()` switches between them.
- **input symbol**: The current symbol from the external input, or from a right side that has been substituted into the input.
- **inject**: `{ … <- pattern }` puts material straight into the input, as if a rule with an empty left side had produced it.
- **length**: The number of matchable items (symbols, `:` and `%`) at the outermost brace level of a left side. Longer rules are tried first.
- **left / right**: The left side is the pattern the rule recognises. Goals come from the left. The right side is what the rule substitutes. Input arrives from the right.
- **lm-diagram**: A picture of two interlocking nesting structures, recognition and substitution. See [06](06-lm-diagram.md).
- **lmn**: Language meta notation, or language machine notation: the rule language, which is also used to describe itself.
- **macro**: The analogy for rules: substitution macros whose patterns and replacements contain grammatical symbols.
- **mismatch**: The goal and input symbols differ. This is the only event that triggers rules.
- **nonterminal**: A symbol that occurs only in rules. `eof` is a special case, because the input system inserts it.
- **priority**: A number plus a direction attached to rules. It controls how match phases may nest. See [02](02-execution-model.md#priorities).
- **relevant**: Specific when L = input and R = goal. Bottom-up when L = input and R = `-`. Speculative when both are `-`. Top-down when L = `-` and R = goal.
- **resolve**: A mismatch is resolved by a rule that consumes the input symbol, produces the goal symbol, or does both, directly or indirectly.
- **scope**: On the left side, a reference sees variables in the current context and enclosing contexts, with the newest instance masking older ones. On the right side, it sees what was visible at the end of the matching left side. Deferred values keep the scope of the rule application that created them.
- **specific rule**: A rule with both a left initial and a right initial, as in `'cat' <- noun;`.
- **speculative rule**: A rule whose left and right sides both start with `-`. Rarely used, and possibly to be deprecated.
- **start**: The pseudo input symbol at startup. Rules for (`eof`, `start`) run before any input is read.
- **substitution**: After a left side matches, the machine behaves as if the right side had appeared in the input in place of what was matched.
- **symbol**: A terminal (can appear in the external input, usually a character `'a'`) or a nonterminal (a name). `"abc"` is one symbol, while `'abc'` is three.
- **terminal**: A symbol that can appear in the external input.
- **top-down rule**: A rule relevant only to the goal symbol. Its left side starts with `-`. It corresponds to recursive descent.
- **variable**: A name, starting with an upper-case letter, whose value depends on the scope of the reference. Variables are created by binding or by `var`.
