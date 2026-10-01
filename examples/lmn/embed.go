// Package lmn embeds the lmn compiler: the bootstrap compiler lmnbs.lm and
// the sources of the bytecode compiler, front end lmn2xfe.lmn and back end
// lmn2mbe.lmn. lmnbs.lm predates the sources, so the current compiler is
// built from them in two stages (see README.md and internal/lmgo).
package lmn

import _ "embed"

var (
	//go:embed lmnbs.lm
	Bootstrap string
	//go:embed lmn2xfe.lmn
	FrontEnd string
	//go:embed lmn2mbe.lmn
	BytecodeBackEnd string
)
