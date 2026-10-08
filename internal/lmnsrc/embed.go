// Package lmnsrc embeds the lm2n compiler: the bootstrap compiler lm2nbs.lm2 and
// the sources of the bytecode compiler, front end lm2n2xfe.lm2n and back end
// lm2n2mbe.lm2n. lm2nbs.lm2 predates the sources, so the current compiler is
// built from them in two stages (see README.md and internal/lmgo).
package lmnsrc

import _ "embed"

var (
	//go:embed lm2nbs.lm2
	Bootstrap string
	//go:embed lm2n2xfe.lm2n
	FrontEnd string
	//go:embed lm2n2mbe.lm2n
	BytecodeBackEnd string
)
