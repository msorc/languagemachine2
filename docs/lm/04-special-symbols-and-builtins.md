# Special Symbols and Builtin Functions

Sources: `special_symbols.html`, `builtin.html`, `extendcalc.html`, `lexicalbuffer.html`.

Special symbols look like ordinary symbols, but they have built-in behaviour when they are matched as goals.

## Start and end

| Symbol | Behaviour |
| --- | --- |
| `eof` | An ordinary symbol, except that the input system supplies it once all external input has been consumed. It is the initial and ultimate goal. |
| `start` | At startup, rules relevant to (goal `eof`, input `start`) are tried before any input is read. See [02-execution-model.md](02-execution-model.md#start). |

## Catch-all symbols (as goals)

These succeed immediately if the input symbol is of the right kind. The value can be captured with `:` or `%`.

| Symbol | Matches |
| --- | --- |
| `anything` | any symbol |
| `terminal` | any character (terminal) symbol |
| `nonTerminal` | any symbol that is not a character |

## Output

Each of these consumes one symbol and writes its text. See [03-lmn-language.md](03-lmn-language.md#output-buffers-var) for buffers.

| Symbol | Writes to |
| --- | --- |
| `out` | stdout |
| `err` | stderr |
| `uri` | stdout, URI-encoded |
| `urd` | stdout, URI-decoded |
| `sp`, `nl` | an actual space or newline in URI output, instead of `%20` or `%0a` |

The standard output idiom is:

```
- eom <- output ;
- out <- eom - ;     // keep printing until `eom` is matched
```

## Repetition

| Symbol | Behaviour |
| --- | --- |
| `repeat` | On the left side, match the rest of the enclosing sequence zero or more times. The limit is set with `-N`/`--max-repeat`. |
| `option` | On the left side, match the rest of the enclosing sequence zero or one time. |

## Binding and acquisition

`:` is the binding symbol and `%` is the grab symbol. Both are described in [03-lmn-language.md](03-lmn-language.md#binding-with-).

## Conversions

Each conversion takes the material grabbed with `%` in the current left side and produces a value, which is usually bound straight away, as in `toNum :N`.

| Symbol | Result |
| --- | --- |
| `toStr`, `toLstr`, `toUstr` | pattern (a sequence of characters), unchanged, lower-cased or upper-cased |
| `toQuote` | quoted-symbol value |
| `toSym`, `toLsym`, `toUsym` | symbol, unchanged, lower-cased or upper-cased |
| `toSys`, `toLsys`, `toUsys` | system symbol, unchanged, lower-cased or upper-cased |
| `toVar` | variable name |
| `toNum`, `toOct`, `toHex`, `toBin` | number from decimal, octal, hex or binary digits |
| `toUrn`, `toUrd` | URI-encoded or URI-decoded pattern |

## Diagnostics and information

These consume no input. The result can be grabbed with `%` or `:`.

| Symbol | Result |
| --- | --- |
| `lineNo` | line number in the current input source |
| `fileName` | name of the current input source (a file name or `stdin`) |
| `flagError` | an error message value. Increments the error counter, and a non-zero count at exit means a failure status. |
| `warnError` | a warning message value. Increments the warning counter. |

## Builtin functions

Rules call these from actions, for example `var N = octal(T);` or `$(format("%s", X))`. In the original there were two interfaces:

- the **mapped** interface, a table from names to procedures, used when rules are interpreted from `.lm`
- the **direct** interface, used by compiled C or D rulesets

Both used the same set of builtins. The Go port's table is `External` in `internal/machine/extension.go`.

| Name | Behaviour |
| --- | --- |
| `octal(x)`, `binary(x)`, `hex(x)`, `num(x)` | number from the string value of `x` |
| `usym(x)`, `ulsym(x)`, `uusym(x)` | unique user symbol, unchanged or case-changed\* |
| `ssym(x)`, `slsym(x)`, `susym(x)` | unique system symbol, unchanged or case-changed\* |
| `variable(x)` | variable symbol named by `x` |
| `urn(x)`, `urd(x)` | URI-encoded or URI-decoded non-unique symbol |
| `lcase(x)`, `ucase(x)` | lower-cased or upper-cased non-unique symbol |
| `format(f, …)` | a formatted string, returned as a non-unique symbol. It is printf-style, and any number of arguments is allowed: arguments left over after the format are appended. |
| `include(file)` | push a new input source. Input returns to the previous source at the file's end. Input source levels have no connection to substitution nesting. |
| `use(g)` | select grammar `g` for the rest of the current left-side nesting level and for inner levels |
| `trOn(bits)`, `trOff(bits)` | switch trace flags on or off. `trOn` ORs the bits in, and `trOff` XORs them. |
| `toChars(x)` | convert `x` to a pattern of character symbols |
| `lmVersion()`, `lmDate()` | version string and build date |
| `reopen(file)` | redirect stdout. Documented as not yet implemented. |

\* The original sources disagree about which variant is which:

- The comments on the declarations in `builtin.html` say `ulsym`/`slsym` give **upper** case and `uusym`/`susym` give **lower** case.
- The `lexicalbuffer` example says the opposite: `ulsym` gives lowercase and `uusym` gives uppercase.

The Go port follows `lexicalbuffer` and the `toLsym`/`toUsym` naming: `ulsym` lower-cases and `uusym` upper-cases (`internal/machine/builtin.go`).

### Where a variable came from

Each of these takes a variable and describes the context that created it. That context began at a mismatch where at least one relevant rule existed.

| Name | Result |
| --- | --- |
| `varSi(V)` | state index of the originating context. A new state is created at every such mismatch. |
| `varGsy(V)` | grammar that was current at that mismatch |
| `varLsy(V)`, `varRsy(V)` | goal symbol and input symbol of that mismatch |
| `varIfn(V)` | input source name |
| `varCp(V)` | absolute input position |
| `varLn(V)`, `varCn(V)` | line number, and character position in the line |

Example from `extendcalc`:

```
- line var Z; <- eof - say "error: " $(format("line %s char %s", varLn(N), varCn(N))) nl output;
```

### User-defined externals

In the original, an extended machine registered extra functions in its external table, and the rules then called them like any builtin. In the Go port, register them with `External.Set` (`internal/machine/extension.go`) and install the table with `Engine.setExternal`.
