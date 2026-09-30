package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func Tz(s string) { fmt.Printf("\ttz: %s\n", s) }

// Encode URI-encodes u the way the original machine did (D's std.uri.encode,
// like JavaScript's encodeURI): letters, digits and ;/?:@&=+$,#-_.!~*'() are
// left alone and every other byte becomes %XX.
func Encode(u string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(u); i++ {
		c := u[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' ||
			strings.IndexByte(";/?:@&=+$,#-_.!~*'()", c) >= 0 {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&15])
		}
	}
	return b.String()
}

func Decode(u string) string {
	decoded, err := url.PathUnescape(u)
	if err != nil {
		panic("Error decoding")
	}
	return decoded
}

func Unescape(s string) string {
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
				fmt.Printf("%s\n", s)
				panic("bad unescape")
			}
		} else {
			r.WriteByte(s[i])
		}
	}
	return r.String()
}

func Strtoi(s string) int {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic("failed to convert string to int")
	}
	return int(value)
}

// Strtod behaves like C's strtod: it converts the longest leading prefix of s
// that is a number, and returns 0 if there is none.
func Strtod(s string) float64 {
	s = strings.TrimSpace(s)
	for n := len(s); n > 0; n-- {
		if value, err := strconv.ParseFloat(s[:n], 64); err == nil {
			return value
		}
	}
	return 0
}
