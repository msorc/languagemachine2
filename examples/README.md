# Examples from lm-0.2.5

Grammars, inputs and reference outputs copied from the original Language Machine 0.2.5 release by Peri Hankey (GNU GPL v2, see `../COPYING.languagemachine`). The D sources, autotools files and the C/D build wrappers were left out. `internal/machine/examples_test.go` uses these files for regression tests.

| Directory | Source in lm-0.2.5 | Contents |
| --- | --- | --- |
| `lmn/` | `src/lmn`, `src/lmnBootstrap` | The lmn compiler sources: front end `lmn2xfe.lmn` and back ends `lmn2mbe` (bytecode), `lmn2dbe`/`lmn4dbe` (D), `lmn2cbe`/`lmn4cbe` (C). Also `lmnbs.lm`, the bootstrap compiler in bytecode, with its original `#!` header. |
| `samples/` | `src/samples` | `flatten`, `reorder` and `lmcat`, plus the reference outputs `flatten.flat` (from `flatten.input`) and `reorder.reorder` (from `reorder.lmn` itself). |
| `basics/` | `src/examples` | calculators (`calc`, `fpCalc`, `rpCalc`, `lmnCalc`, `calc2tcc`), `copy`, `leftRecursion`, `reorder`. |
| `testing/` | `src/testing` | The original test grammars and inputs (`t2`/`w2` take `testinput`). `t2` includes `testinclude` and `lmn2minc.lmn` includes `../lmn/*.lmn`, and both paths are relative to the working directory, so run them from this directory. |
| `lambda/` | `src/web` | Lambda-calculus experiment: `*.lam` sources, the translators `*.lmn`, and the published outputs `*.out.lmn`. |
| `web/` | `src/web` | Grammars from the website's worked examples (bottles, cats, grok*, stemming, lcm/lct, aibjaibj). |
| `wiki/` | `src/wiki`, `src/make` | The mediawiki-to-HTML site generator and `wiki2make`. |
| `translators/` | `src/d2d`, `src/gcc`, `src/j2d` | The large D and Java front ends and back ends. `d2xfe-j2d.lmn` is the j2d copy of `d2xfe.lmn`, which differs slightly. |

## Compiling and running

The current compiler sources are newer than `lmnbs.lm`: for example, the `{| |}` alternatives used by `testing/alt.lmn` are only understood by a compiler rebuilt from them. Build it in two stages. The second stage is a fixpoint, which is what `TestLmnBootstrapFixpoint` checks.

```sh
bin/lm -rules examples/lmn/lmnbs.lm examples/lmn/lmn2xfe.lmn examples/lmn/lmn2mbe.lmn > stage1.lm
bin/lm -rules stage1.lm examples/lmn/lmn2xfe.lmn examples/lmn/lmn2mbe.lmn > lmn.lm

bin/lm -rules lmn.lm -output flatten.lm examples/samples/flatten.lmn
bin/lm -rules flatten.lm examples/samples/flatten.input | diff - examples/samples/flatten.flat
```

## Checking against the original

Every example gives the same output as the original lm-0.2.5 engine, including the lm-diagrams, apart from the port's Unicode box drawing. The published `fact2diagram` was made by an earlier engine (see `docs/lm/README.md`). Some examples fail with the original as well, and exit with status 1: `testing/dlex` (a work in progress) and the gcc back end `d2gccbe`, which needs the gcc front end from `lm-0.2.5/src/gcc`.

The original builds with the D1 compiler DMD 1.076 (32-bit; needs the multilib toolchain), which makes it usable as a reference when the port's behaviour is in doubt:

```sh
cd lm-0.2.5/src && dmd -m32 -O -release -L-no-pie -oflmd -Ilm lm/lmdxMain.d lm/application.d \
  lm/builtin.d lm/element.d lm/engine.d lm/extension.d lm/licenseGnuGPLv2.d lm/loader.d \
  lm/variadic.d lm/options.d lm/tracer.d lm/versionInfo.d
```

Its options are the short ones listed in `docs/lm/05-cli-and-tracing.md` (`-r`, `-i`, `-t`, `-W`, `-o`).
