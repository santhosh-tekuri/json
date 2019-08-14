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
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{buf: b, stack: make([]byte, 0, 50)}
}

func (d *Decoder) Reset(b []byte) {
	d.buf, d.pos, d.stack = b, 0, d.stack[:0]
}

func (d *Decoder) Token() Token {
	d.err, d.mark = nil, -1
	t := d.token()
	if len(d.stack) > 0 {
		s := d.stack[len(d.stack)-1]
		if s == '{' || s == '[' {
			switch t {
			case ObjEnd, ArrEnd, String, Number, Null, Boolean: // value end
				d.stack = append(d.stack, ',')
			}
		}
	} else {
		d.stack = append(d.stack, 0)
	}
	switch t {
	case String, Number, Boolean:
		return Token{t, d.buf[d.mark:d.pos], d.err}
	default:
		return Token{t, nil, d.err}
	}
}

func (d *Decoder) token() Type {
	d.whitespace()
	if len(d.stack) > 0 {
		s := d.stack[len(d.stack)-1]
		switch s {
		case 0:
			return EOF
		case ':':
			d.stack = d.stack[:len(d.stack)-1]
			if d.match(':', "after object key") == Error {
				return Error
			}
			d.whitespace()
		case ',':
			d.stack = d.stack[:len(d.stack)-1]
			if !d.hasMore() {
				return d.unexpectedEOF()
			}
			if d.peek() == ',' {
				d.next()
				d.whitespace()
				s := d.stack[len(d.stack)-1]
				if s == '{' {
					t := d.string()
					d.stack = append(d.stack, ':')
					return t
				}
			} else {
				s := d.stack[len(d.stack)-1]
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
		case '{':
			if !d.hasMore() {
				return d.unexpectedEOF()
			}
			switch d.peek() {
			case '}':
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
			if d.peek() == ']' {
				d.stack = d.stack[:len(d.stack)-1]
				return ArrEnd
			}
		}
	}
	return d.value()
}

func (d *Decoder) value() Type {
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	switch d.peek() {
	case '{':
		d.stack = append(d.stack, '{')
		d.next()
		return ObjBegin
	case '[':
		d.stack = append(d.stack, '[')
		d.next()
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
		p := d.peek()
		if p == '-' || ('0' <= p && p <= '9') {
			return d.number()
		}
		return d.error(d.peek(), "looking for beginning of value")
	}
}

func (d *Decoder) hasMore() bool {
	return d.pos < len(d.buf)
}

func (d *Decoder) peek() byte {
	return d.buf[d.pos]
}

func (d *Decoder) next() byte {
	d.pos++
	return d.buf[d.pos-1]
}

func (d *Decoder) match(m byte, context string) Type {
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	if b := d.next(); b != m {
		return d.error(b, context)
	}
	return noError
}

func (d *Decoder) whitespace() {
	for d.hasMore() {
		if p := d.peek(); p == ' ' || p == '\t' || p == '\r' || p == '\n' {
			d.next()
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
	switch t.Type {
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
			if t.Type == ObjEnd {
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
		a := []interface{}(nil)
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

func (d *Decoder) error(c byte, context string) Type {
	d.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, int64(d.pos)}
	return Error
}

func (d *Decoder) unexpectedEOF() Type {
	d.err = &SyntaxError{"unexpected end of JSON input", int64(d.pos)}
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
