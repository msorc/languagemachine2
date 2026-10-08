# Command Line and Tracing

Sources: `lm2.html`, `bottles.html`, `lambda.html`, `metalanguage_bootstrap.html`. The Go flag definitions are in `internal/application/application.go`.

## Options are processed in order

Arguments are evaluated from left to right, and **the order matters**:

- `-t b` (trace loading) must come **before** `-r rules`, or the rules have already loaded by the time tracing starts.
- `-W width` must come before `-t D`.
- `-s`, `-c` and `-d` write text to the current output, so they must come **after** `-o file`.

The Go port keeps this behaviour: every flag registers a callback, and the callbacks run in the order the flags were given. Positional input files are handled last.

## Options

| Original | Go flag | Meaning |
| --- | --- | --- |
| `-v`, `--version` | `-version` | show the version |
| `-h`, `--help`, `-H`, `--detail` | (Go `flag` usage) | show usage |
| `-L`, `--license` | `-license` | show the license |
| `-s path`, `--shebang` | `-shebang` | write a `#! path -r` script header, so compiled rules can run as an executable script. Must follow `-o`. The Go port writes `#! path -rules`, because Go has no `-r` flag. |
| `-c`, `--cmain` | — | write a C main program. Must follow `-o`. |
| `-d`, `--dmain` | — | write a main program for rules compiled to D. Not in the Go port. |
| `-r file`, `--rules` | `-rules` | load rules in `.lm2`/`.lmr` format. `#` starts a comment, so shebang scripts load directly. |
| `-a file`, `--add` | `-add` | add more rules after the first set. Later rules win over earlier ones of the same effective length in the same contexts, which lets a general ruleset be specialised. |
| `-o file`, `--output` | `-output` | redirect output |
| `-e file`, `--errout` | `-errout` | redirect error output |
| `-i string` | `-input` | use the string as input. Used as a dummy input for rulesets that never read input, for example `./lists.lm2 -i z`. |
| `-` | `-stdin` | take input from the console until Ctrl-D |
| `-l n`, `--lexpri` | `-lexpri` | the lexical priority given to terminal symbols (accepted but not applied, as in the original engine) |
| `-b n`, `--buffer` | `-buffer` | how large the input symbol buffer grows before it becomes circular |
| `-N n`, `--max-repeat` | `-max-repeat` | limit on `repeat` iterations. Guards against loops such as `{ repeat nothing }` where `- <- nothing;` exists. |
| `-D n`, `--max-depth` | `-max-depth` | limit on nesting depth. Guards against rules such as `- nest <- nest;`. |
| `-W n`, `--dwidth` | `-dwidth` | width of the diagram. Must come before `-t D`. |
| `-t flags`, `--trace` | `-trace` | trace flags (see below) |
| — | `-trace-out file` | Go only: write a Go runtime execution trace (`runtime/trace`) |
| `files…` | positional | input files |

Typical invocations in the original:

```sh
lm2 -r fpCalc.lmr                        # interactive calculator
lm2 -r fpCalc.lmr -W 50 -t D             # with the lm-diagram
lm2 -r lm2nbs.lm2 -o lm2n2m.lmr -s /usr/bin/lm2 lm2n2xfe.lm2n lm2n2mbe.lm2n   # bootstrap compile
./fact2.lm2 -i z -W 40 -t D              # shebang ruleset, dummy input, diagram
```

## Trace flags

`-t m` is the shortest trace. It shows mismatch events, marked `??`, and fallback events, marked `**`. `-t D` is the easiest to understand. In the Go port, flags can be combined as CSV or by repeating the option: `-trace m,s` or `-trace m -trace s`.

| Flag | Name | Traces |
| --- | --- | --- |
| `a` | all | everything |
| `z` | none | nothing (turns tracing off) |
| `b` | LOAD | the rules as they are loaded, for example `lm2 -t b -r xyz.lm2` |
| `c` | CVAR | obsolete |
| `d` | DIAGRAM text | the text form behind the diagram, intended as a basis for graphical output |
| `D` | DIAGRAM | the textual lm-diagram. Put `-W` before it. |
| `e` | EACH | applications of `each` and `all` |
| `f` | EACHREFVAR | the same, in more detail |
| `E` | EACHSCOPE | the scope used by `each` |
| `G` | GRAMMAR | a summary of the ruleset after loading |
| `I` | INDEX | table and array index operators |
| `l` | RELATION | relational operators |
| `L` | LOOP | loops |
| `m` | MISMATCH | mismatch events |
| `q` | APPLY | the `apply` primitive, as in `$(Table[X])` |
| `r` | RVAR | variables created from the right side |
| `R` | RVAR_VAR | more detail for RVAR |
| `X` | RVARSCOPE | the scope for RVAR |
| `s` | SYMBOLS | symbols matched |
| `S` | ASSIGN | assignment operators |
| `U` | LVAR | variables created on the left side |
| `v` | ARITHMETIC | arithmetic operators |
| `w` | REFVAR | variable references |
| `V` | REFSCOPE | the scope of variable references |
| `x` | CXSCOPE | variable scope at the end of a left side |
| `y` | DEBUG | detailed tracing |
| `A` | ACT | obsolete |

Notes on the Go port's trace flags:

- The Go port maps `v` to `REF`, where the original table uses `v` for ARITHMETIC; ARITHMETIC is `o` in the Go port. The codes are listed in `traceCodes` (`internal/application/application.go`), and `-h` prints them.
- In the Go port every occurrence of a flag takes effect, in command-line order: `-input a -input b` reads both, `-trace m -trace z` ends with tracing off, and each `-rules` replaces the rules loaded before it.
- `-dwidth` must be at least 20; a narrower diagram cannot be drawn.
- In the Go port, `a` sets every category except the two diagram flags.
- In the Go port, setting `D` or `d` also turns on the categories the diagram depends on: mismatch, symbols and context scope.

Rules can also switch tracing at run time with `trOn(bits)` and `trOff(bits)`.
