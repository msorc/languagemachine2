# lmn2go: compiling rules to Go

`lmn2go` does for Go what the original `lmn2d` and `lmn2c` back ends did for D and C (see `lm/07-compilation-and-bytecode.md`). It builds a ruleset into a Go program or package, so the program needs no `.lm` file at run time, and it binds the functions the rules call, such as `thisabout(...)` in `examples/basics/calc.lmn`, to Go functions.

```sh
make lmn2go                                      # -> bin/lmn2go
bin/lmn2go -o calc.go -stubs funcs.go calc.lmn   # package main, func main, funcs.go to fill in
go build -o calc . && ./calc calc.input          # takes the same options as lm
bin/lmn2go -pkg calc -o calc_lm.go calc.lmn      # a library package instead
bin/lmn2go -o calc.go calc.lm                    # rules that are already compiled
```

In a Go project, `//go:generate lmn2go -o calc_lm.go calc.lmn` keeps the generated file up to date.

## How lmn2d did it

`lmn2dbe.lmn` wrote each rule as the same text that `lmn2mbe` writes to a `.lm` file, as a string in a D module (`static char[] rulN = "…"`), and an `lmdInit` that defined those rules when the program started. For every function `F` that the rules called with `N` arguments, it declared `extern (C) element F(inout stream, ...)` and a wrapper `funI` that popped `N` arguments and called `F`. A call compiled to `F:I` (`uf`), an index into the function table, instead of a lookup by name. Linking failed if a function was not defined. The program was then linked with the engine and `lm.application`, so it took the usual command line.

`lmn4dbe`/`lmn4cbe` went further and compiled each rule side to a D or C function. lmn2go does not do that (see below).

## Design

```
file.lmn ──lmn compiler──► .lm bytecode ──lmgo.Generate──► file.go ──go build──► program
           (internal/lmgo,                (loads the rules,           │
            built from the                 finds the calls)           └─ imports github.com/msorc/languagemachine2/lm
            embedded sources)
```

| Part | Role |
| --- | --- |
| `lm/` | Public runtime for generated code and for programs that embed a ruleset: `Program` (rules and `Funcs`), `Func`, `Call`, `Value`, and `Main`, `Run`, `Translate`. It is the only package that generated code imports, so the engine stays in `internal/`. |
| `internal/lmgo/` | The generator. `Compiler` builds the lmn compiler from `examples/lmn` in two stages (about 0.3s), `Compile` runs it on `.lmn` files in-process, and `Generate` writes the Go file. |
| `cmd/lmn2go/` | The command line. |
| `examples/lmn/embed.go` | Embeds `lmnbs.lm`, `lmn2xfe.lmn` and `lmn2mbe.lmn`, so the binary carries its own compiler. |
| `machine.Calls` | Lists the functions that rules call, from the bytecode. |
| `application.Program` | Rules and functions that the command line loads before its options, so `-rules` replaces them and `-add` adds to them. |

### Decisions

- **The generator works on bytecode, not on the lmn front end's output.** `lmn2dbe` embedded exactly what `lmn2mbe` writes, so `.lm` is already the boundary between front end and back end. Working on it means one compiler (the tested `lmn2mbe`), any `.lm` can be wrapped, including ones without source, and the builtin list (`machine.Builtin`) and loader are the engine's own, not copies in an lmn back end. The generated rules are byte for byte what `lm -rules` would load, apart from comment lines, so traces and lm-diagrams are unchanged.
- **The rules are loaded at generation time.** Rules that the engine rejects (for example `T`, which the loader does not implement, see `lm/README.md`) fail in `lmn2go`, not when the program starts.
- **Functions are bound by name.** The rules keep `v:f G f:args … f:fun` and the generated `Funcs` map registers each Go function in the engine's external table under its lmn name. There is no `F:index` opcode: a map lookup costs little next to the engine's matching, and the same rules still run under `lm -rules`, where a missing function prints `external not found` and gives 0, as in the original.
- **A missing function is a compile error.** `Funcs` names a Go identifier for each called function that is not a builtin (`thisabout` → `lmThisabout`), so `go build` fails until the package defines it, as linking failed for lmn2d. `-stubs file` writes stubs to start from, and never overwrites the file. Builtins are left out; defining one in `Funcs` by hand replaces it.
- **One variadic signature.** `func(c *lm.Call) lm.Value` takes any number of arguments, so there are no per-arity wrappers (`funI` in lmn2d). `Value` is opaque, so user code does not depend on the engine's `Element` interface.
- **Calls through an expression** (`T[i](x)`, a variable's value) cannot be resolved from the bytecode. They are counted and reported, and they are bound at run time by name like everything else.

### Finding the calls

A call `f(a, b)` compiles to `v:f G f:args <a> <b> f:fun`. An array literal `[a, b]` compiles to `f:args <a> <b> f:array`, so `machine.Calls` pairs each `f:args` with the `f:fun` or `f:array` that closes it, keeping a stack because calls nest. A call whose function is not `v:name G` is dynamic.

## Generated code

```go
// Code generated by lmn2go from shout.lmn. DO NOT EDIT.

package main

import "github.com/msorc/languagemachine2/lm"

// Program is the ruleset compiled from shout.lmn.
var Program = &lm.Program{
	Name:  "shout",
	Rules: programRules,
	Funcs: map[string]lm.Func{
		"count": lmCount,
		"shout": lmShout,
	},
}

func main() { lm.Main(Program) }

const programRules = `
m:shout L:0 n:1 ( c:! v:W G v:shout G f:args d:hey G f:fun w . ) ( m:eof v:W ) r
…
`
```

A function the rules call:

```go
func lmShout(c *lm.Call) lm.Value {
	return lm.Sym(strings.ToUpper(c.Arg(0).String()) + "!")
}
```

With `-pkg name` there is no `func main`. Run the program with `Program.Run(args, stdout, stderr)`, which takes `lm`'s options, or `Program.Translate(input, options...)`, which returns the output.

Both may be called from several goroutines at once. Each call loads the rules into a machine of its own, so calls share nothing except the `Funcs`, which must then be safe for concurrent use.

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-o file` | stdout | the Go file |
| `-pkg name` | `main` | package name; `main` also writes `func main` |
| `-var name` | `Program` | the `*lm.Program` variable |
| `-name name` | first file's base name | program name in messages |
| `-stubs file` | | write stubs for the called functions, unless the file exists |
| `-prefix p` | `lm` | prefix of the Go names of called functions |
| `-compiler file.lm` | built in | lmn compiler to use instead of the embedded one |
| `-emit-lm file` | | also write the compiled rules |
| `-import path` | `github.com/msorc/languagemachine2/lm` | import path of the runtime |

`.lmn` inputs are compiled together as one input, as `lm -rules lmn.lm a.lmn b.lmn` would; `.lm` inputs are joined. `.include` paths are relative to the working directory, as with `lm`.

## Using it outside this module

The module is `github.com/msorc/languagemachine2`, so a project outside this repository gets the runtime with `go get github.com/msorc/languagemachine2/lm` and the generator with `go install github.com/msorc/languagemachine2/cmd/lmn2go@latest`. To use a local checkout instead, add `replace github.com/msorc/languagemachine2 => /path/to/languagemachine2` to its `go.mod`; `TestGeneratedProgram` builds generated programs that way.

## The native lmn compiler

`cmd/lmn` is the lmn compiler (`lmn2xfe` + `lmn2mbe`) built by lmn2go, the step the original took with `lmn2d` and gdc. `lmn.go` is generated by `go generate ./cmd/lmn` (or `make generate`), and `TestUpToDate` fails when it is out of date with the sources. `TestFixpoint` checks that it compiles its own sources to itself.

```sh
make lmn
bin/lmn -output foo.lm foo.lmn      # same as bin/lm -rules lmn.lm -output foo.lm foo.lmn
```

`lmn2go` itself still builds its compiler from `lmnbs.lm` in two stages, so it never depends on a generated file.

## Not done: lmn4go

`lmn4dbe` compiled each rule side to a D function. A Go version that only builds the rule elements in Go code, instead of parsing text, would save nothing: loading the 33 KB lmn compiler takes about 10 ms, and matching dominates. A real compiled back end needs the engine to accept a rule side written in Go next to the interpreted `Str` bodies, while keeping the trace output byte-identical, which is a larger change to `engine.go` and `mode.go`.
