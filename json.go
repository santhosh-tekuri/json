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
	"fmt"
	"strconv"
)

type Decoder struct {
	buf   []byte
	pos   int
	stack []byte
	mark  int
	err   error
	empty Kind // tells action to take when stack is empty
	comma bool // if comma && stack.peek is '{' or '[' then read ','
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{buf: b, stack: make([]byte, 0, 50), empty: none, comma: true}
}

func (d *Decoder) Reset(b []byte) {
	d.buf, d.pos, d.stack, d.empty, d.comma = b, 0, d.stack[:0], none, true
}

func (d *Decoder) Token() Token {
	d.err, d.mark = nil, -1
	t := d.token()
	switch t {
	case String, Number, Boolean:
		return Token{t, d.buf[d.mark:d.pos], d.err}
	default:
		return Token{t, nil, d.err}
	}
}

func (d *Decoder) token() Kind {
	d.whitespace()
	if len(d.stack) == 0 {
		switch d.empty {
		case none:
			d.empty = EOD
		case EOD:
			d.empty = EOF
			return EOD
		case EOF:
			if !d.hasMore() {
				return EOF
			}
			d.empty = EOD
		}
	}
	if len(d.stack) > 0 {
		if d.comma {
			s := d.stack[len(d.stack)-1]
			if s == '{' || s == '[' {
				if d.buf[d.pos] == ',' {
					d.pos++
					d.whitespace()
					if s == '{' {
						t := d.string()
						d.stack = append(d.stack, ':')
						return t
					}
				} else {
					if s == '{' {
						if d.match('}', "after object key:value pair") == Error {
							return Error
						}
						d.stack = d.stack[:len(d.stack)-1]
						return ObjEnd
					} else if s == '[' {
						if d.match(']', "after array element") == Error {
							return Error
						}
						d.stack = d.stack[:len(d.stack)-1]
						return ArrEnd
					}
				}
			}
		} else {
			d.comma = true
		}

		s := d.stack[len(d.stack)-1]
		switch s {
		case ':':
			d.stack = d.stack[:len(d.stack)-1]
			if d.match(':', "after object key") == Error {
				return Error
			}
			d.whitespace()
		case '{':
			if !d.hasMore() {
				return d.unexpectedEOF()
			}
			switch d.buf[d.pos] {
			case '}':
				d.pos++
				d.stack = d.stack[:len(d.stack)-1]
				return ObjEnd
			default:
				t := d.string()
				d.stack = append(d.stack, ':')
				return t
			}
		case '[':
			if !d.hasMore() {
				return d.unexpectedEOF()
			}
			if d.buf[d.pos] == ']' {
				d.pos++
				d.stack = d.stack[:len(d.stack)-1]
				return ArrEnd
			}
		}
	}

	// read value ---
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	switch d.buf[d.pos] {
	case '{':
		d.stack = append(d.stack, '{')
		d.pos++
		d.comma = false
		return ObjBegin
	case '[':
		d.stack = append(d.stack, '[')
		d.pos++
		d.comma = false
		return ArrBegin
	case '"':
		return d.string()
	case 'n':
		if d.match('n', "in literal null") == Error {
			return Error
		}
		if d.match('u', "in literal null") == Error {
			return Error
		}
		if d.match('l', "in literal null") == Error {
			return Error
		}
		if d.match('l', "in literal null") == Error {
			return Error
		}
		return Null
	case 't':
		d.mark = d.pos
		if d.match('t', "in literal true") == Error {
			return Error
		}
		if d.match('r', "in literal true") == Error {
			return Error
		}
		if d.match('u', "in literal true") == Error {
			return Error
		}
		if d.match('e', "in literal true") == Error {
			return Error
		}
		return Boolean
	case 'f':
		d.mark = d.pos
		if d.match('f', "in literal false") == Error {
			return Error
		}
		if d.match('a', "in literal false") == Error {
			return Error
		}
		if d.match('l', "in literal false") == Error {
			return Error
		}
		if d.match('s', "in literal false") == Error {
			return Error
		}
		if d.match('e', "in literal false") == Error {
			return Error
		}
		return Boolean
	default:
		if !d.hasMore() {
			return d.unexpectedEOF()
		}
		p := d.buf[d.pos]
		if p == '-' || ('0' <= p && p <= '9') {
			return d.number()
		}
		return d.error(p, "looking for beginning of value")
	}
}

func (d *Decoder) hasMore() bool {
	return d.pos < len(d.buf)
}

func (d *Decoder) match(m byte, context string) Kind {
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	if b := d.buf[d.pos]; b != m {
		return d.error(b, context)
	}
	d.pos++
	return none
}

func (d *Decoder) whitespace() {
	for d.hasMore() {
		if p := d.buf[d.pos]; p == ' ' || p == '\t' || p == '\r' || p == '\n' {
			d.pos++
		} else {
			break
		}
	}
}

func (d *Decoder) Skip() error {
	c := 0
	for {
		t := d.Token()
		switch {
		case t.Error():
			return t.Err
		case t.Begin():
			c++
		case t.End():
			c--
		}
		if c == 0 {
			return nil
		}
	}
}

// unmarshalling ---

func (d *Decoder) Unmarshal() (v interface{}, err error) {
	t := d.Token()
	switch t.Kind {
	case Error:
		return nil, t.Err
	case Null:
		return nil, nil
	case String:
		s, _ := t.Str()
		return s, nil
	case Number:
		f, _ := t.Float64()
		return f, nil
	case Boolean:
		b, _ := t.Bool()
		return b, nil
	case ObjBegin:
		m := make(map[string]interface{})
		for {
			t = d.Token()
			if t.Error() {
				return nil, t.Err
			}
			if t.Kind == ObjEnd {
				return m, nil
			}
			key, _ := t.Str()
			v, err := d.Unmarshal()
			if err != nil {
				return nil, err
			}
			m[key] = v
		}
	case ArrBegin:
		a := make([]interface{}, 0)
		for {
			v, err := d.Unmarshal()
			if err != nil {
				return nil, err
			}
			if v == ArrEnd {
				return a, nil
			}
			a = append(a, v)
		}
	case ArrEnd:
		return ArrEnd, nil
	default:
		panic(fmt.Sprintln("BUG: got", t))
	}
}

// ---

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

func (d *Decoder) error(c byte, context string) Kind {
	d.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, int64(d.pos)}
	return Error
}

var unexpectedEOF = "unexpected end of JSON input"

func IsUnexpectedEOF(err error) bool {
	e, ok := err.(*SyntaxError)
	return ok && e.msg == unexpectedEOF
}

func (d *Decoder) unexpectedEOF() Kind {
	d.err = &SyntaxError{unexpectedEOF, int64(d.pos)}
	return Error
}

// quoteChar formats c as a quoted character literal
func quoteChar(c byte) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}
