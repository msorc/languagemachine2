package conv

import (
	"math"
	"testing"
)

func TestStrtod(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	if got, want := EncodeComponent("l:%5B a/b"), "l%3A%255B%20a%2Fb"; got != want {
		t.Errorf("EncodeComponent = %q, want %q", got, want)
	}
	if got, want := Encode("l:%5B a/b"), "l:%255B%20a/b"; got != want {
		t.Errorf("Encode = %q, want %q", got, want)
	}
}

func TestUnescape(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{`plain`, "plain"},
		{`a\nb\tc`, "a\nb\tc"},
		{`\a\b\f\v\r`, "\a\b\f\v\r"},
		{`\"\'\\`, `"'\`},
		{`\q`, "q"}, // an unknown escape is the character itself
		{`é\n`, "é\n"},
	}
	for _, c := range cases {
		if got, err := Unescape(c.in); err != nil || got != c.want {
			t.Errorf("Unescape(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	if _, err := Unescape(`ends with \`); err == nil {
		t.Error("an unfinished escape gives no error")
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{"%41b": "Ab", "a+b": "a+b", "%20": " ", "%5Cn": `\n`} {
		if got, err := Decode(in); err != nil || got != want {
			t.Errorf("Decode(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := Decode("%zz"); err == nil {
		t.Error("Decode(%zz): no error")
	}
}

func TestStrtoi(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]int{"20": 20, "-3": -3, "+7": 7, "0": 0} {
		if got, err := Strtoi(in); err != nil || got != want {
			t.Errorf("Strtoi(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"", "x", "1x", "1.5", "99999999999999999999"} {
		if _, err := Strtoi(in); err == nil {
			t.Errorf("Strtoi(%q): no error", in)
		}
	}
}

func TestScanBinary(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]int64{"": 0, "1": 1, "101": 5, "1111": 15, "1x1": 5, "0b11": 3, "abc": 0} {
		if got := ScanBinary(in); got != want {
			t.Errorf("ScanBinary(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestScanOctal(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]int64{"17": 15, "  -10": -8, "+7": 7, "019": 1, "8": 0, "": 0, "x": 0, "-": 0} {
		if got := ScanOctal(in); got != want {
			t.Errorf("ScanOctal(%q) = %d, want %d", in, got, want)
		}
	}
}
