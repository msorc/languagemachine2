package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func Tz(s string) { fmt.Printf("\ttz: %s\n", s) }
func Star()       { Tz("*") }

func Encode(u string) string {
	return url.PathEscape(u)
}

func Decode(u string) string {
	decoded, err := url.QueryUnescape(u)
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
