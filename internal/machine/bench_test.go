package machine

import "testing"

// Benchmarks on the lmn compiler compiling its own sources: the largest
// ruleset and input in the repository.

// Loading the compiler's rules into a new engine.
func BenchmarkLoad(b *testing.B) {
	rules := compiler(b)
	b.SetBytes(int64(len(rules)))
	b.ReportAllocs()
	for b.Loop() {
		if err := NewEngine().LoadFromString(rules); err != nil {
			b.Fatal(err)
		}
	}
}

// One engine at a time: load the compiler and compile its sources.
func BenchmarkCompile(b *testing.B) {
	rules := compiler(b)
	b.ReportAllocs()
	for b.Loop() {
		if runFiles(b, rules, lmnSources...) != rules {
			b.Fatal("the compiler did not reproduce itself")
		}
	}
}

// The same with an engine per goroutine; compare the ns/op at -cpu 1,4,16
// to see how independent engines scale on one heap.
func BenchmarkCompileParallel(b *testing.B) {
	rules := compiler(b)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// no Fatal here: it must not be called from a RunParallel goroutine
			out, err := runFilesErr(rules, lmnSources...)
			if err != nil {
				b.Error(err)
				return
			}
			if out != rules {
				b.Error("the compiler did not reproduce itself")
				return
			}
		}
	})
}
