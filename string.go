// Copyright 2019 Santhosh Kumar Tekuri
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package json

import (
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func (d *ByteDecoder) string() Token {
	mark := d.pos
	d.pos++
	for {
		if d.pos == len(d.buf) {
			return d.unexpectedEOF()
		}
		b := d.buf[d.pos]
		switch b {
		case '"':
			d.pos++
			return Token{Kind: Str, Data: d.buf[mark:d.pos]}
		case '\\':
			d.pos++
			if d.pos == len(d.buf) {
				return d.unexpectedEOF()
			}
			switch d.buf[d.pos] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				d.pos++
			case 'u':
				d.pos++
				for i := 0; i < 4; i++ {
					if d.pos == len(d.buf) {
						return d.unexpectedEOF()
					}
					b = d.buf[d.pos]
					switch {
					case '0' <= b && b <= '9', 'A' <= b && b <= 'F', 'a' <= b && b <= 'f':
						d.pos++
					default:
						return d.error("in \\u hexadecimal character escape")
					}
				}
			default:
				return d.error("in string escape code")
			}
		default:
			if b < 0x20 {
				return d.error("in string literal")
			}
			d.pos++
		}
	}
}

func (t Token) Eq(v string) bool {
	s := t.Data
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return false
	}
	s = s[1 : len(s)-1]
	i, j := 0, 0
	for i < len(s) && j < len(v) {
		c := s[i]
		if c == '\\' {
			i++
			c = s[i]
			switch c {
			case '"', '\\', '/', '\'':
				if v[j] != c {
					return false
				}
				i++
				j++
			case 'b':
				if v[j] != '\b' {
					return false
				}
				i++
				j++
			case 'f':
				if v[j] != '\f' {
					return false
				}
				i++
				j++
			case 'n':
				if v[j] != '\n' {
					return false
				}
				i++
				j++
			case 'r':
				if v[j] != '\r' {
					return false
				}
				i++
				j++
			case 't':
				if v[j] != '\t' {
					return false
				}
				i++
				j++
			case 'u':
				r := getu4(s[i-1:])
				i += 5
				if utf16.IsSurrogate(r) {
					r1 := getu4(s[r:])
					if dec := utf16.DecodeRune(r, r1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						i += 6
						b := make([]byte, utf8.UTFMax)
						n := utf8.EncodeRune(b, dec)
						if j+n > len(v) {
							return false
						}
						for x := 0; x < n; x++ {
							if b[x] != v[j] {
								return false
							}
							j++
						}
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					r = unicode.ReplacementChar
				}
				b := make([]byte, utf8.UTFMax)
				n := utf8.EncodeRune(b, r)
				if j+n > len(v) {
					return false
				}
				for x := 0; x < n; x++ {
					if b[x] != v[j] {
						return false
					}
					j++
				}
			}
			continue
		}
		if c < utf8.RuneSelf {
			if v[j] != c {
				return false
			}
			i++
			j++
			continue
		}
		r1, size1 := utf8.DecodeRune(s[i:])
		r2, size2 := utf8.DecodeRuneInString(v[j:])
		if r1 != r2 || size1 != size2 {
			return false
		}
		i += size1
		j += size2
	}
	return i == len(s) && j == len(v)
}

// from encoding/json/decode.go ---

func unquoteBytes(s []byte) (t []byte, ok bool) {
	//if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
	//	return
	//}
	s = s[1 : len(s)-1]

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room? Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}
