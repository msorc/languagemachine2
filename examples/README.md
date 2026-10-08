# Examples from the original Language Machine

Grammars, inputs and reference outputs copied from the original Language Machine 0.2.5 release by Peri Hankey (GNU GPL v2, see `../COPYING.languagemachine`). The engine sources, autotools files and the C/D build wrappers were left out. `internal/machine/examples_test.go` uses these files for regression tests.

| Directory | Contents |
| --- | --- |
| `samples/` | `flatten`, `reorder` and `lmcat`, plus the reference outputs `flatten.flat` (from `flatten.input`) and `reorder.reorder` (from `reorder.lmn` itself). |
| `basics/` | calculators (`calc`, `fpCalc`, `rpCalc`, `lmnCalc`, `calc2tcc`), `copy`, `leftRecursion`, `reorder`. |
| `testing/` | The original test grammars and inputs (`t2`/`w2` take `testinput`). `t2` includes `testinclude` and `lmn2minc.lmn` includes `../lmn/*.lmn`, and both paths are relative to the working directory, so run them from this directory. |
| `lambda/` | Lambda-calculus experiment: `*.lam` sources, the translators `*.lmn`, and the published outputs `*.out.lmn`. |
| `web/` | Grammars from the website's worked examples (bottles, cats, grok*, stemming, lcm/lct, aibjaibj). |
| `wiki/` | The mediawiki-to-HTML site generator and `wiki2make`. |
| `golang/` | `golog.lmn` (not from the release) adds a `name_log` twin that traces the call and its origin to every function of a Go source file. `sample.go.txt` is the input and `sample.golog.txt` the output that `TestGologGolden` checks. |
| `highlight/` | A syntax highlighter (not from the release): the rules `go.lmn` and `lmn.lmn`, the Go package that turns their output into spans, and the command `lmhl`. They call `hl`, which `lmhl` provides, so they do not run under plain `lm`. See `highlight/README.md`. |
| `translators/` | The large D and Java front ends and back ends. `d2xfe-j2d.lmn` is the j2d copy of `d2xfe.lmn`, which differs slightly. |

## Compiling and running

The current compiler sources are newer than `lmnbs.lm`: for example, the `{| |}` alternatives used by `testing/alt.lmn` are only understood by a compiler rebuilt from them. Build it in two stages. The second stage is a fixpoint, which is what `TestLmnBootstrapFixpoint` checks.

```sh
bin/lm -rules internal/lmnsrc/lmnbs.lm internal/lmnsrc/lmn2xfe.lmn internal/lmnsrc/lmn2mbe.lmn > stage1.lm
bin/lm -rules stage1.lm internal/lmnsrc/lmn2xfe.lmn internal/lmnsrc/lmn2mbe.lmn > lmn.lm

bin/lm -rules lmn.lm -output flatten.lm examples/samples/flatten.lmn
bin/lmn -output flatten.lm examples/samples/flatten.lmn   # the same, with the built-in compiler (make lmn)
bin/lm -rules flatten.lm examples/samples/flatten.input | diff - examples/samples/flatten.flat
```

## Checking against the original

Every example gives the same output as the original engine. The lm-diagrams and traces match too, apart from the port's Unicode box drawing; `TestTraceGolden` checks this. The diagrams published on the website were made by an earlier engine (see `docs/lm/README.md`). Some examples fail with the original as well, and exit with status 1: `testing/dlex` (a work in progress) and the gcc back end `d2gccbe`, which needs a gcc front end that is not included here.

## The lmn compiler

The compiler sources from the original `lmn/` directory now live in [`internal/lmnsrc`](../internal/lmnsrc), where the Go tools embed them: front end `lmn2xfe.lmn` and back ends `lmn2mbe` (bytecode), `lmn2dbe`/`lmn4dbe` (D) and `lmn2cbe`/`lmn4cbe` (C), with `lmnbs.lm`, the bootstrap compiler in bytecode. `testing/lmn2minc.lmn` includes them from there; that path is the only change made to the original files.
