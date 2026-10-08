// Command lm2n is the lm2n compiler (lm2n2xfe + lm2n2mbe) built into a Go
// program by lm2n2go, as the original built lm2n2m with lm2n2d and gdc. It
// compiles lm2n sources to .lm2 bytecode without loading lm2nbs.lm2:
//
//	lm2n -output foo.lm2 foo.lm2n
//
// It takes the same options as lm2. lm2n.go is generated; TestUpToDate fails
// when it is out of date with the sources.
package main

//go:generate go run ../lm2n2go -name lm2n -o lm2n.go ../../internal/lm2nsrc/lm2n2xfe.lm2n ../../internal/lm2nsrc/lm2n2mbe.lm2n
