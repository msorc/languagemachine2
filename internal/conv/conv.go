// Package conv holds the text conversions that must behave as the original
// machine's did: URI encoding like JavaScript's encodeURI, C-style
// escapes, and C library number parsing (strtod, sscanf "%o").
package conv

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Encode URI-encodes u the way the original machine did (like JavaScript's
// encodeURI): letters, digits and ;/?:@&=+$,#-_.!~*'() are
// left alone and every other byte becomes %XX.
func Encode(u string) string {
	return encode(u, ";/?:@&=+$,#-_.!~*'()")
}

// EncodeComponent is the component form of Encode (like JavaScript's
// encodeURIComponent): only letters, digits and -_.!~*'() are left alone.
func EncodeComponent(u string) string {
	return encode(u, "-_.!~*'()")
}

func encode(u, keep string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := range len(u) {
		c := u[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' ||
			strings.IndexByte(keep, c) >= 0 {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&15])
		}
	}
	return b.String()
}

// Decode undoes %XX escapes; + is left alone.
func Decode(u string) (string, error) {
	return url.PathUnescape(u)
}

// Unescape resolves C-style backslash escapes.
func Unescape(s string) (string, error) {
	var r strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			if i < len(s) {
				switch s[i] {
				case 'a':
					r.WriteRune('\a')
				case 'b':
					r.WriteRune('\b')
				case '"':
					r.WriteRune('"')
				case '\'':
					r.WriteRune('\'')
				case '\\':
					r.WriteRune('\\')
				case 'n':
					r.WriteRune('\n')
				case 'r':
					r.WriteRune('\r')
				case 't':
					r.WriteRune('\t')
				case 'f':
					r.WriteRune('\f')
				case 'v':
					r.WriteRune('\v')
				default:
					r.WriteByte(s[i])
				}
			} else {
				return "", fmt.Errorf("unfinished escape at the end of %q", s)
			}
		} else {
			r.WriteByte(s[i])
		}
	}
	return r.String(), nil
}

// Strtoi parses s as a decimal integer.
func Strtoi(s string) (int, error) {
	value, err := strconv.ParseInt(s, 10, 64)
	return int(value), err
}

// Strtod behaves like C's strtod: it converts the longest leading prefix of s
// that is a number (decimal, hexadecimal such as 0x10.8, inf or nan), and
// returns 0 if there is none.
func Strtod(s string) float64 {
	s = strings.TrimLeft(s, " \t\n\v\f\r")
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	if n := hexFloatPrefix(s[i:]); n > 0 {
		if e := exponent(s[i+n:], "pP"); e > 0 {
			return parseFloat(s[:i+n+e])
		}
		return parseFloat(s[:i+n] + "p0")
	}
	if n := decFloatPrefix(s[i:]); n > 0 {
		return parseFloat(s[:i+n])
	}
	for _, w := range []string{"infinity", "inf", "nan"} {
		if len(s)-i >= len(w) && strings.EqualFold(s[i:i+len(w)], w) {
			return parseFloat(s[:i+len(w)])
		}
	}
	return 0
}

// ScanBinary reads s as binary digits, as the original did: each '1' adds
// its place value and every other character counts as 0.
func ScanBinary(s string) int64 {
	var n int64
	b := int64(1)
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '1' {
			n += b
		}
		b *= 2
	}
	return n
}

// ScanOctal behaves like C's sscanf(s, "%o"): it reads an optionally signed
// octal integer after leading space, and returns 0 if there is none.
func ScanOctal(s string) int64 {
	s = strings.TrimLeft(s, " \t\n\v\f\r")
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	n := digits(s[i:], func(c byte) bool { return '0' <= c && c <= '7' })
	v, _ := strconv.ParseInt(s[:i+n], 8, 64)
	return v
}

// parseFloat ignores range errors: like strtod it gives ±Inf or 0 then.
func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func digits(s string, ok func(byte) bool) int {
	n := 0
	for n < len(s) && ok(s[n]) {
		n++
	}
	return n
}

func isDec(c byte) bool { return '0' <= c && c <= '9' }
func isHex(c byte) bool { return isDec(c) || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' }

// mantissa returns the length of digits[.digits] at the start of s, or 0
// if there is no digit.
func mantissa(s string, ok func(byte) bool) int {
	n := digits(s, ok)
	d := n
	if n < len(s) && s[n] == '.' {
		f := digits(s[n+1:], ok)
		d += f
		n += 1 + f
	}
	if d == 0 {
		return 0
	}
	return n
}

// exponent returns the length of an exponent (marker, sign, digits) at the
// start of s, or 0 if there is no complete one.
func exponent(s string, marker string) int {
	if len(s) == 0 || strings.IndexByte(marker, s[0]) < 0 {
		return 0
	}
	n := 1
	if n < len(s) && (s[n] == '+' || s[n] == '-') {
		n++
	}
	d := digits(s[n:], isDec)
	if d == 0 {
		return 0
	}
	return n + d
}

func decFloatPrefix(s string) int {
	n := mantissa(s, isDec)
	if n == 0 {
		return 0
	}
	return n + exponent(s[n:], "eE")
}

// hexFloatPrefix returns the length of a hexadecimal number without its
// binary exponent, which ParseFloat requires and Strtod adds if missing.
func hexFloatPrefix(s string) int {
	if len(s) < 2 || s[0] != '0' || (s[1] != 'x' && s[1] != 'X') {
		return 0
	}
	n := mantissa(s[2:], isHex)
	if n == 0 {
		return 0
	}
	return 2 + n
}
