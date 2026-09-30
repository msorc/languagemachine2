# The lm-diagram

Sources: `picturebook.html`, `lm-diagram.html`, `leftrecursive.html`, `fpcalcdiagram.html`, `rpcalcdiagram.html`, `fact2diagram.html`.

Peri Hankey devised the lm-diagram in the 1970s to show what happens when unrestricted substitution rules are applied to a stream of symbols. The site claims that almost everything about how rule applications combine can be understood from it. The Go port reproduces its text form with `-trace D` (`internal/machine/diagram.go`). Keep the Go output consistent with the reference output below.

## The idea

Every rule application has two phases:

- a **recognition** (match) phase, for the left side
- a **substitution** (replacement) phase, for the right side

Each phase can involve any number of symbols, including none, and phases nest within phases of the same kind. There are therefore **two independent nesting structures**: recognition nesting and substitution nesting. They interlock, and any drawing in which each side nests properly is a possible sequence of rule applications.

- **Horizontal form:** recognition is drawn below the centre line and substitution above it.
- **Vertical form:** goals, which come from left sides, are in the left column. Input symbols, whether actual or substituted, are in the right column. Recognition nesting is drawn to the left and substitution nesting to the right. The built-in generator uses this form, one step per line.

The shapes in the diagram correspond to kinds of analysis:

- left recursion
- right recursion
- central recursion, as in brackets
- **something to nothing**: rules with an empty right side, which delete what they match
- **something from nothing**: rules with an empty left side, which inject material

The diagram also describes evaluation. `f(x)` is a rule application in which a function and its argument are replaced by the result. For recursive functions such as factorial, the recursive calls appear as nesting above the line.

Variable scope can be read from the diagram. A reference sees the most recent instance created in the current rule application, or in the recognition phase of any application that **encloses** it. X encloses Y if X's recognition phase enclosed Y's.

## Legend for the text output

- The two central columns are the **goal symbol** on the left and the **input symbol** on the right.
- `?` on a left nesting level means a mismatch happened there.
- `-` on a left nesting level means an alternative analysis was tried (a backtrack).
- `*` on a left nesting level means repetition (`repeat`).
- `.------NNNNNN` opens a left (recognition) nesting. The numeric label identifies the mismatch event that caused visible nesting.
- `'------NNNNNN---- ----NNNNNN---.` closes the recognition and opens the matching substitution nesting on the right.
- `NNNNNN---'` on the right closes a substitution level.
- A left symbol with nothing covering it on the left is an ultimate goal. A right symbol with nothing covering it on the right is actual input.
- No nesting is drawn when a left side has length 1 or a right side is empty.
- A complete rule application appears as nesting first on the left and then on the right. Some applications never complete, and an alternative is tried instead.
- `??` in the right-hand columns appears where a substituted symbol mismatched. See `leftrecursive.html`.

## Reference example: `cats`

```
  .cats()
  - sentence                <- eof - ;
  - subject verb object '.' <- sentence ;
  - nounphrase              <- subject ;
  - nounphrase              <- object  ;
  'the ' noun               <- nounphrase ;
  'a '   noun               <- nounphrase ;
  'cat '                    <- noun;
  'dog '                    <- noun;
  'bit '                    <- noun;
  'bit '                    <- verb;
  'ate '                    <- verb;
  'likes '                  <- verb;
  ' '                       <- - ;
  '\n'                      <- - ;
```

Input `the cat likes the dog .`, output of `-t D`, verbatim from the original site:

```
                                 eof 't'
       ?                         eof 't'
       .------------000001
       |                    sentence 't'
       |?                   sentence 't'
       |.-----------000002
       ||                    subject 't'
       ||?                   subject 't'
       ||.----------000003
       |||                nounphrase 't'
       |||?               nounphrase 't'
       |||.---------000004
       ||||                      'h' 'h'
       ||||                      'e' 'e'
       ||||                      ' ' ' '
       ||||                     noun 'c'
       ||||?                    noun 'c'
       ||||.--------000005
       |||||                     'a' 'a'
       |||||                     't' 't'
       |||||                     ' ' ' '
       ||||'--------000005---------- ----------000005-------------.
       ||||                     noun noun                         |
       |||'---------000004---------- ----------000004------------.|
       |||                nounphrase nounphrase                  ||
       ||'----------000003---------- ----------000003-----------.||
       ||                    subject subject                    |||
       ||                                      000003-----------'||
       ||                                      000004------------'|
       ||                                      000005-------------'
       ||                       verb 'l'
       ||?                      verb 'l'
       ||.----------000006
       |||                       'i' 'i'
       |||                       'k' 'k'
       |||                       'e' 'e'
       |||                       's' 's'
       |||                       ' ' ' '
       ||'----------000006---------- ----------000006-------------.
       ||                       verb verb                         |
       ||                                      000006-------------'
       ||                     object 't'
       ||?                    object 't'
       ||.----------000007
       |||                nounphrase 't'
       |||?               nounphrase 't'
       |||.---------000008
       ||||                      'h' 'h'
       ||||                      'e' 'e'
       ||||                      ' ' ' '
       ||||                     noun 'd'
       ||||?                    noun 'd'
       ||||.--------000009
       |||||                     'o' 'o'
       |||||                     'g' 'g'
       |||||                     ' ' ' '
       ||||'--------000009---------- ----------000009-------------.
       ||||                     noun noun                         |
       |||'---------000008---------- ----------000008------------.|
       |||                nounphrase nounphrase                  ||
       ||'----------000007---------- ----------000007-----------.||
       ||                     object object                     |||
       ||                                      000007-----------'||
       ||                                      000008------------'|
       ||                                      000009-------------'
       ||                        '.' '.'
       |'-----------000002---------- ----------000002-------------.
       |                    sentence sentence                     |
       '------------000001                                        |
                                               000002-------------'
                                 eof '\n'
       ?                         eof '\n'
                                 eof eof
```

## Where to find more diagrams

- `lm-diagram.html`: the forward Polish calculator `fpCalc`, run with `lm -r fpCalc.lmr -W 50 -t D` and input `/ 307 241`. It shows `%`, `repeat` and lexical classes.
- `leftrecursive.html`: a left-recursive list with backtracking. It shows `-` markers, and `??` in the substitution columns when the newer `number` rule is tried before the `word` rule.
- `fpcalcdiagram.html` and `rpcalcdiagram.html`: forward and reverse Polish calculators.
- `fact2diagram.html`: lambda-calculus evaluation of factorial 2. It is very long.
