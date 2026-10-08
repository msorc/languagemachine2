# lm2n2go: compiling rules to Go

`lm2n2go` does for Go what the original `lm2n2d` and `lmn2c` back ends did for D and C (see `lm/07-compilation-and-bytecode.md`). It builds a ruleset into a Go program or package, so the program needs no `.lm2` file at run time, and it binds the functions the rules call, such as `thisabout(...)` in `examples/basics/calc.lm2n`, to Go functions.

```sh
make lm2n2go                                      # -> bin/lm2n2go
bin/lm2n2go -o calc.go -stubs funcs.go calc.lm2n   # package main, func main, funcs.go to fill in
go build -o calc . && ./calc calc.input          # takes the same options as lm2
bin/lm2n2go -pkg calc -o calc_lm2.go calc.lm2n      # a library package instead
bin/lm2n2go -o calc.go calc.lm2                    # rules that are already compiled
```

In a Go project, `//go:generate lm2n2go -o calc_lm2.go calc.lm2n` keeps the generated file up to date. With `go get -tool github.com/msorc/languagemachine2/cmd/lm2n2go`, the line can be `//go:generate go tool lm2n2go -o calc_lm2.go calc.lm2n`, which pins the generator to the version in `go.mod`.

## How lm2n2d did it

`lm2n2dbe.lm2n` wrote each rule as the same text that `lm2n2mbe` writes to a `.lm2` file, as a string in a D module (`static char[] rulN = "…"`), and an `lmdInit` that defined those rules when the program started. For every function `F` that the rules called with `N` arguments, it declared `extern (C) element F(inout stream, ...)` and a wrapper `funI` that popped `N` arguments and called `F`. A call compiled to `F:I` (`uf`), an index into the function table, instead of a lookup by name. Linking failed if a function was not defined. The program was then linked with the engine and `lm2.application`, so it took the usual command line.

`lm2n4dbe`/`lm2n4cbe` went further and compiled each rule side to a D or C function. lm2n2go does not do that (see below).

## Design

```
file.lm2n ──lm2n compiler──► .lm2 bytecode ──lmgo.Generate──► file.go ──go build──► program
           (internal/lmgo,                (loads the rules,           │
            built from the                 finds the calls)           └─ imports github.com/msorc/languagemachine2/lm
            embedded sources)
```

| Part | Role |
| --- | --- |
| `lm/` | Public runtime for generated code and for programs that embed a ruleset: `Program` (rules and `Funcs`), `Func`, `Call`, `Value`, and `Main`, `Run`, `Translate`, `TranslateReader`. It is the only package that generated code imports, so the engine stays in `internal/`. |
| `lm2/lm2n` | `Compile` and `CompileFiles`: lm2n to bytecode for programs that build rules at run time, such as `lmhl -rules file.lm2n`. |
| `internal/lmgo/` | The generator. `Compiler` builds the lm2n compiler from `internal/lmnsrc` in two stages (about 0.3s), `Compile` runs it on `.lm2n` files in-process, and `Generate` writes the Go file. `cli.go` holds the command line (`ParseArgs`, `Options.Generate`) and `UpToDate`, which the freshness tests use to regenerate a file with the arguments of its own `//go:generate` line. |
| `cmd/lm2n2go/` | The command line. |
| `internal/lmnsrc/embed.go` | Embeds `lm2nbs.lm2`, `lm2n2xfe.lm2n` and `lm2n2mbe.lm2n`, so the binary carries its own compiler. |
| `machine.Calls` | Lists the functions that rules call, from the bytecode. |
| `application.Program` | Rules and functions that the command line loads before its options, so `-rules` replaces them and `-add` adds to them. |

### Decisions

- **The generator works on bytecode, not on the lm2n front end's output.** `lm2n2dbe` embedded exactly what `lm2n2mbe` writes, so `.lm2` is already the boundary between front end and back end. Working on it means one compiler (the tested `lm2n2mbe`), any `.lm2` can be wrapped, including ones without source, and the builtin list (`machine.Builtin`) and loader are the engine's own, not copies in an lm2n back end. The generated rules are byte for byte what `lm2 -rules` would load, apart from comment lines, so traces and lm-diagrams are unchanged.
- **The rules are loaded at generation time.** Rules that the engine rejects (for example `T`, which the loader does not implement, see `lm2/README.md`) fail in `lm2n2go`, not when the program starts.
- **Functions are bound by name.** The rules keep `v:f G f:args … f:fun` and the generated `Funcs` map registers each Go function in the engine's external table under its lm2n name. There is no `F:index` opcode: a map lookup costs little next to the engine's matching, and the same rules still run under `lm2 -rules`, where a missing function prints `external not found` and gives 0, as in the original.
- **A missing function is a compile error.** `Funcs` names a Go identifier for each called function that is not a builtin (`thisabout` → `lmThisabout`), so `go build` fails until the package defines it, as linking failed for lm2n2d. `-stubs file` writes stubs to start from, and never overwrites the file. Builtins are left out; defining one in `Funcs` by hand replaces it.
- **One variadic signature.** `func(c *lm2.Call) (lm2.Value, error)` takes any number of arguments, so there are no per-arity wrappers (`funI` in lm2n2d). An error stops the run and is reported; a panic in the function is reported the same way instead of ending the process. `Value` is opaque, so user code does not depend on the engine's `Element` interface.
- **Calls through an expression** (`T[i](x)`, a variable's value) cannot be resolved from the bytecode. They are counted and reported, and they are bound at run time by name like everything else.

### Finding the calls

A call `f(a, b)` compiles to `v:f G f:args <a> <b> f:fun`. An array literal `[a, b]` compiles to `f:args <a> <b> f:array`, so `machine.Calls` pairs each `f:args` with the `f:fun` or `f:array` that closes it, keeping a stack because calls nest. A call whose function is not `v:name G` is dynamic.

## Generated code

```go
// Code generated by lm2n2go from shout.lm2n. DO NOT EDIT.

package main

import "github.com/msorc/languagemachine2/lm"

// Program is the ruleset compiled from shout.lm2n.
var Program = &lm2.Program{
	Name:  "shout",
	Rules: programRules,
	Funcs: map[string]lm2.Func{
		"count": lmCount,
		"shout": lmShout,
	},
}

func main() { lm2.Main(Program) }

const programRules = `
m:shout L:0 n:1 ( c:! v:W G v:shout G f:args d:hey G f:fun w . ) ( m:eof v:W ) r
…
`
```

A function the rules call:

```go
func lmShout(c *lm2.Call) (lm2.Value, error) {
	return lm2.Sym(strings.ToUpper(c.Arg(0).String()) + "!"), nil
}
```

With `-pkg name` there is no `func main`. Run the program with `Program.Run(args, stdin, stdout, stderr)`, which takes `lm2`'s options, with `Program.Translate(input)`, which returns the output, or with `Program.TranslateReader(r, w)`, which streams it.

Both may be called from several goroutines at once. Each call loads the rules into a machine of its own, so calls share nothing except the `Funcs`, which must then be safe for concurrent use.

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-o file` | stdout | the Go file |
| `-pkg name` | `main` | package name; `main` also writes `func main` |
| `-var name` | `Program` | the `*lm2.Program` variable |
| `-name name` | first file's base name | program name in messages |
| `-stubs file` | | write stubs for the called functions, unless the file exists |
| `-prefix p` | `lm2` | prefix of the Go names of called functions |
| `-compiler file.lm2` | built in | lm2n compiler to use instead of the embedded one |
| `-emit-lm2 file` | | also write the compiled rules |
| `-import path` | `github.com/msorc/languagemachine2/lm` | import path of the runtime |

`.lm2n` inputs are compiled together as one input, as `lm2 -rules lm2n.lm2 a.lm2n b.lm2n` would; `.lm2` inputs are joined. `.include` paths are relative to the working directory, as with `lm2`.

## Using it outside this module

The module is `github.com/msorc/languagemachine2`, so a project outside this repository gets the runtime with `go get github.com/msorc/languagemachine2/lm` and the generator with `go install github.com/msorc/languagemachine2/cmd/lm2n2go@latest`. To use a local checkout instead, add `replace github.com/msorc/languagemachine2 => /path/to/languagemachine2` to its `go.mod`; `TestGeneratedProgram` builds generated programs that way.

## The native lm2n compiler

`cmd/lm2n` is the lm2n compiler (`lm2n2xfe` + `lm2n2mbe`) built by lm2n2go, the step the original took with `lm2n2d` and gdc. `lm2n.go` is generated by `go generate ./cmd/lm2n` (or `make generate`), and `TestUpToDate` fails when it is out of date with the sources. `TestFixpoint` checks that it compiles its own sources to itself.

```sh
make lm2n
bin/lm2n -output foo.lm2 foo.lm2n      # same as bin/lm2 -rules lm2n.lm2 -output foo.lm2 foo.lm2n
```

`lm2n2go` itself still builds its compiler from `lm2nbs.lm2` in two stages, so it never depends on a generated file.

## Not done: lmn4go

`lm2n4dbe` compiled each rule side to a D function. A Go version that only builds the rule elements in Go code, instead of parsing text, would save nothing: loading the 33 KB lm2n compiler takes about 10 ms, and matching dominates. A real compiled back end needs the engine to accept a rule side written in Go next to the interpreted `Str` bodies, while keeping the trace output byte-identical, which is a larger change to `engine.go` and `mode.go`.
