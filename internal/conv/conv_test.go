package conv

import (
	"math"
	"testing"
)

func TestStrtod(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"10", 10},
		{"  10.5xyz", 10.5},
		{"-.5e2;", -50},
		{"1e", 1},
		{"1e+", 1},
		{".", 0},
		{"abc", 0},
		{"", 0},
		{"0x10", 16},
		{"0x10.8", 16.5},
		{"0X.8", 0.5},
		{"0x1p4", 16},
		{"0x1p", 1},
		{"0x", 0},
		{"0xg", 0},
		{"+inf", math.Inf(1)},
		{"-Infinity", math.Inf(-1)},
		{"1e999", math.Inf(1)},
	}
	for _, c := range cases {
		if got := Strtod(c.in); got != c.want {
			t.Errorf("Strtod(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	if got := Strtod("nan"); !math.IsNaN(got) {
		t.Errorf("Strtod(\"nan\") = %v, want NaN", got)
	}
}

func TestEncodeComponent(t *testing.T) {
	if got, want := EncodeComponent("l:%5B a/b"), "l%3A%255B%20a%2Fb"; got != want {
		t.Errorf("EncodeComponent = %q, want %q", got, want)
	}
	if got, want := Encode("l:%5B a/b"), "l:%255B%20a/b"; got != want {
		t.Errorf("Encode = %q, want %q", got, want)
	}
}
