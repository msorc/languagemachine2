// Command lmn is the lmn compiler (lmn2xfe + lmn2mbe) built into a Go
// program by lmn2go, as the original built lmn2m with lmn2d and gdc. It
// compiles lmn sources to .lm bytecode without loading lmnbs.lm:
//
//	lmn -output foo.lm foo.lmn
//
// It takes the same options as lm. lmn.go is generated; TestUpToDate fails
// when it is out of date with the sources.
package main

//go:generate go run ../lmn2go -name lmn -o lmn.go ../../examples/lmn/lmn2xfe.lmn ../../examples/lmn/lmn2mbe.lmn
