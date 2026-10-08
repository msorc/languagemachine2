// Package lmgo turns compiled Language Machine rules into Go source, as the
// original lmn2d back end turned them into D: the rules in .lm bytecode are
// embedded in a Go file that runs them with package lm, together with a
// table that binds the functions the rules call to Go functions.
package lmgo

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/msorc/languagemachine2/internal/lmnsrc"
	"github.com/msorc/languagemachine2/internal/machine"
)

// Input is one input to a run of the machine: the file File, or Text.
type Input struct {
	File string // the file to read, if not ""
	Text string // the text, when File is ""
}

// Run runs rules on the inputs, in order, and returns what the rules write
// to standard output.
func Run(rules string, inputs ...Input) (string, error) {
	e := machine.NewEngine()
	var out, errOut bytes.Buffer
	e.SetOutput(&out)
	e.SetErrOutput(&errOut)
	if err := e.LoadFromString(rules); err != nil {
		return "", err
	}
	for _, in := range inputs {
		if in.File == "" {
			e.AppendInput(machine.NewStringInput(e, in.Text))
			continue
		}
		g, err := machine.NewFileInput(e, in.File)
		if err != nil {
			return "", err
		}
		e.AppendInput(g)
	}
	if err := e.Run(); err != nil {
		if msg := bytes.TrimSpace(errOut.Bytes()); len(msg) > 0 {
			err = fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}
	return out.String(), nil
}

// Compiler returns the lmn compiler in bytecode, built from the embedded
// sources in two stages: the bootstrap compiler compiles them, and the result
// compiles them again (examples/README.md).
func Compiler() (string, error) {
	return compiler()
}

var compiler = sync.OnceValues(func() (string, error) {
	src := []Input{
		{Text: lmnsrc.FrontEnd},
		{Text: lmnsrc.BytecodeBackEnd},
	}
	stage1, err := Run(lmnsrc.Bootstrap, src...)
	if err != nil {
		return "", fmt.Errorf("building the lmn compiler, stage 1: %w", err)
	}
	stage2, err := Run(stage1, src...)
	if err != nil {
		return "", fmt.Errorf("building the lmn compiler, stage 2: %w", err)
	}
	return stage2, nil
})

// Compile compiles lmn source files, read in order as one input, to
// bytecode with the given compiler, or with Compiler() if it is "".
func Compile(compiler string, files ...string) (string, error) {
	in := make([]Input, len(files))
	for i, f := range files {
		in[i] = Input{File: f}
	}
	lm, err := CompileInputs(compiler, in...)
	if err != nil {
		return "", fmt.Errorf("compiling %v: %w", files, err)
	}
	return lm, nil
}

// CompileInputs compiles lmn sources, read in order as one input, with the
// given compiler, or with Compiler() if it is "".
func CompileInputs(compiler string, in ...Input) (string, error) {
	if compiler == "" {
		var err error
		if compiler, err = Compiler(); err != nil {
			return "", err
		}
	}
	return Run(compiler, in...)
}
