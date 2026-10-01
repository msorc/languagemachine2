# Compilation: from lmn to `.lm` Bytecode

Sources: `lmn2mbe.html`, `lmn2xfe.html`, `metalanguage_bootstrap.html`, `minimal_rule_set_-_lmcat.html`, `code_produced_by_different_backends.html`, `lmn2dbe.html`, `lmn4dbe.html`. For how the Go loader decodes the format, see `../bytecode.md` and `../internal_machine.md`.

## The compiler family

Every lmn compiler is itself an lmn ruleset. Each one is a **shared frontend** (`lmn2xfe.lmn`) combined with a **backend**. The frontend turns rules into an internal representation that the backend's rules then consume.

| Compiler | Backend | Output |
| --- | --- | --- |
| `lmn2m` | `lmn2mbe.lmn` | compact text format `.lm`/`.lmr`, loaded at run time with `lm -r`. Used with `-s` to make shebang scripts. |
| `lmn2d` | `lmn2dbe.lmn` | the text format wrapped as a D module or program |
| `lmn2c` | `lmn2cbe.lmn` | the text format wrapped as a C module or program |
| `lmn2D` (later `lmn4d`) | `lmn2Dbe.lmn` / `lmn4dbe.lmn` | rules compiled to D functions, one generator function per rule side |
| `lmn2C` (later `lmn4c`) | `lmn2Cbe.lmn` | rules compiled to C functions |

The interface between frontend and backend:

- The backend drives recognition of `outer` by wrapping `units`.
- Each compilation unit produces `unit … lmn`.
- A rule definition reaches the backend as `rdef :Line :Grammar :Prio :Dash :LInit :LBody :RInit :RBody`.
- The default grammar is `lm_` and the default priority is `L:0`, set by `start var G = "lm_"; var P = { lp :0 }; …`.

## The bootstrap

`lmnbs.lm` is a hand-seeded compiled ruleset for the lmn compiler. It is "the descendant of a ruleset that originally had to be constructed by hand". The repository contains a copy (grammar `lmn2x`). The bootstrap sequence was:

```sh
lm -r lmnbs.lm   -o lmn2m.lmr -s /usr/bin/lm lmn2xfe.lmn lmn2mbe.lmn   # bootstrap -> lmn2m.lm
lm -r lmn2m.lm   -o lmn2m.a1  -s /usr/bin/lm lmn2xfe.lmn lmn2mbe.lmn   # self-compile
lm -r lmn2m.lm   -o lmn2m.a2  -s /usr/bin/lm lmn2xfe.lmn lmn2mbe.lmn   # self-compile again (fixpoint)
lm -r lmn2m.lm   -o lmn2d.lmr -s /usr/bin/lm lmn2xfe.lmn lmn2dbe.lmn   # other compilers …
```

Each compiled variant was then rebuilt with gdc or gcc, run on its own sources, and `diff`ed against the interpreted output.

To compile an `.lmn` file with the Go port, use the same approach, loading the bootstrap with `-rules` and redirecting with `-output`:

```sh
bin/lm -rules lmnbs.lm -output foo.lm foo.lmn
```

## Output format

- The output is text and line-oriented, with one rule per line ending in `r`.
- `#` starts a comment that runs to the end of the line, so shebang headers are allowed:

  ```
  #! /usr/bin/lm -r
  # Language Machine (C) 2005 Peri Hankey ...
  m:lm_ L:0 n:1 ( z m:out ) ( m:eof ) r
  ```

  That is the whole of `lmcat`: `- out <- eof - ;`.
- Symbol text is **URI-encoded**, because the backend writes through `uri`. For example `m:extern%20(C)%20mode…` and `c:%5Cn`. The Go loader URL-decodes the text and then unescapes C escapes (`Loader.MStr`).

### Rule layout

```
m:<grammar> <prio> n:<dash> ( <LHS-initial> <LHS-body…> ) ( <RHS-initial> <RHS-body…> ) r
```

- `<prio>` is `L:n`, `R:n`, `B:n` or `M:n`.
- `n:` holds the frontend's **dash** value, which the Go port stores as `Rule.offset`, the RHS start index. It is `1` when the right initial is followed by `-`, as in `<- eof - …`, or when the right initial is itself `-`, as in `<- - ;`. In both cases the right initial only files the rule and is **not** substituted. Otherwise it is `0`.
- An initial written as `-` compiles to `z`, the nil symbol.

### How lmn constructs map to bytecode (from `lmn2mbe`)

| lmn construct | Internal form | Bytecode emitted |
| --- | --- | --- |
| `-` (void initial) | `void` | `z` |
| `'x'` | `c :X` | `c:x` |
| `"sym"` | `d :X` | `m:sym` (a plain nonterminal) |
| `name` | `m :X` | `m:name` |
| number | `n :X` | `n:123` |
| `.[class]` | `x :X` | `l:[class]` |
| `Var` in a pattern | `v :X` | `v:Var` |
| `{ … }` | `lm :X` | `( … )` |
| `%` | `t` | `t` |
| `:Name` / `:sym` / `:"s"` / `:3` | `pp :X` | `v:Name p`, `m:sym p`, … |
| `:{ … }` or `:'chars'` | `pq :I :B` | `( … ) p` |
| `:(expr)` | `pb :X` | `<expr> b` |
| `option …` / `repeat …` | `r1` / `rz` | `m:option …` / `m:repeat …` |
| `each Name` | `ea :X` | `v:Name e` |
| `each (expr)` | `ex :X` | `<expr> E` |
| `all Name` | `al :X` | `v:Name A` |
| `all (expr)` | `ax :X` | `<expr> B` |
| `$(expr)` | `ap :X` | `<expr> f:apply` |
| `(Buffer)` | `ou :X` | `<var> f:append` |
| `{ … <- pattern }` | `ij` | `( … ) G f:inj` |
| `!` | `op :done` | `f:done` |
| statement `expr;` | `eox` | `<expr> .` (drop) |
| variable in an expression | `va :X` | `v:X V` (value) |
| symbol in an expression | `sy :X` | `v:X G` |
| `"str"` in an expression | `dq :X` | `d:str G` |
| `var A = e` | `initv` | `v:A G <e> w` |
| `var A` | `initz` | `v:A G v:null G w` |
| binary operator `A op B` | `fx`, `fc`, … | `<A> <B> f:op`, for example `f:+` and `f:%3C` |
| `\|\|` / `&&` | `fl` | `<A> ( <B> ) f:\|\|` (right side deferred) |
| `c ? a : b` | `"?"` | `<c> ( <a> ) G ( <b> ) G f:sel` |
| unary `-`, `!`, `~` | `ux`, `ub` | `<A> f:neg`, `f:not`, `f:inv` |
| `++x`, `x++`, … | `pre`, `post` | `f:preinc`, `f:postinc`, `f:predec`, `f:postdec` |
| `f(args)` | `fn` | `<f> f:args <args…> f:fun` |
| `a[i]` / `a.b` | `idx` / `dot` | `<a> <i> f:idx` / `<a> d:b G f:idt` |
| `[x, k: v]` | `arrayinit` | `f:args … f:array` (cells: `<k> <v> f:cell`) |
| `if (e) A else B` | `xif` | `<e> ( A ) G ( B ) G f:if` |
| `while (e) B` | `xwhile` | `( <e> f:test B ) G f:loop` |
| `for (I; E; N) B` | `xfor` | `I ( <E> f:test B ) G ( N ) G f:for` |
| `foreach (K, V; E) B` | `xforeachkv` | `v:K V v:V V <E> ( B ) G f:foreach` |
| `foreach (V; E) B` | `xforeachv` | `v:null G v:V V <E> ( B ) G f:foreach` |
| `break;` / `continue;` | `xbreak0` / `xcont0` | `f:break` / `f:continue` |
| `rule(G,P){…}` | `rval` | `G P n:N ( … ) G ( … ) G f:rule` |

The `while` and `for` rows are what this repository's `lmn2mbe.lmn` emits. The original emitted `( <e> B ) G f:loop` for `while`, with no `f:test`, so a `while` loop never ended. It emitted `I ( <E> f:test B N ) G f:loop` for `for`, where a `continue` could not reach the step `N`. The original runtime had no `f:break`, `f:continue` or `f:rule` either. A `.lm` file compiled with the old shapes still loads, and `f:loop` still runs them. `foreach` is not in the original `lmn2xfe` at all. Its C and D backends had an `xforeach` rule that only reported "feature not implemented", and the runtime registered `f:foreach` without doing anything. This port added the syntax, the bytecode and the runtime.

A worked example from `code_produced_by_different_backends`:

```
 .onerule()
   - var I = 0; <- postlude "extern (C) mode lmdInit(inout stream s){" indent nl
       table :"mt" :Mi {for(I = 0; I < Mi; I++) { defq :mt :I :(Mn[I]) eoc }} nl … lmn ;
```

This compiles to (with the original compiler; the `for` loop is in the old shape):

```
m:onerule L:0 n:0 ( z v:I G n:0 G w . ) ( m:postlude m:extern%20(C)%20mode%20lmdInit(inout%20stream%20s)%7B m:indent m:nl
m:table m:mt p v:Mi p ( v:I V n:0 G f:= . ( v:I V v:Mi V f:%3C f:test
m:defq m:mt p v:I p v:Mn V v:I V f:idx b m:eoc v:I V f:postinc ) G f:loop
) m:nl … m:lmn ) r
```

## Go loader compatibility

The Go loader (`internal/machine/loader.go`) tokenises with `([().reAtpbPgGVsawz])|(.:\S*)|#[^\n]*\n|(\S)|\s*` and dispatches on the first character of each token. Compared with what `lmn2mbe` emits:

| Emitted by lmn2mbe | Original meaning | Go loader |
| --- | --- | --- |
| bare `e` | `each Name` | handled: `NewEachRef` |
| bare `E` | `each (expr)` | handled: `NewEachX` (the original loader could not load it) |
| bare `B` | `all (expr)` | handled: `NewAllX` (the original loader could not load it). `B:n` is a bracket priority. |
| bare `T` | `top` | rejected with `unsupported opcode`. `lmn2xfe` never produces `top`, and the original loader could not load it either. |
| `M:n` | maximal priority | handled: encoded as `PRIMASK\|BRACKET` |
| `A` | `all Name` | handled: `NewAllRef` |

`lmnbs.lm` in this repository uses `e` in `each` position seven times, so it only loads now that `e` is handled. For example:

```
m:lmn2x L:0 n:0 ( z m:repeat m:element v:X p ) ( m:elements ( v:X e ) p ) r
```

`calc.lm` does not use any of these opcodes.
