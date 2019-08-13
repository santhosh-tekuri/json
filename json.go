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
	"errors"
	"fmt"
	"strconv"
)

type Decoder struct {
	buf   []byte
	pos   int
	stack []byte
	mark  int
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{buf: b, stack: make([]byte, 0, 50)}
}

func (d *Decoder) Reset(b []byte) {
	d.buf, d.pos, d.stack = b, 0, d.stack[:0]
}

func (d *Decoder) token(t Type) Token {
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
		return Token{t, d.buf[d.mark:d.pos]}
	default:
		return Token{t, nil}
	}
}

func (d *Decoder) Token() Token {
	d.whitespace()
	if len(d.stack) > 0 {
		s := d.stack[len(d.stack)-1]
		switch s {
		case 0:
			return d.token(EOF)
		case ':':
			d.stack = d.stack[:len(d.stack)-1]
			d.match(':', "after object key")
			d.whitespace()
		case ',':
			d.stack = d.stack[:len(d.stack)-1]
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
					d.match('}', "after object key:value pair")
					d.stack = d.stack[:len(d.stack)-1]
					return d.token(ObjEnd)
				} else if s == '[' {
					d.match(']', "after array element")
					d.stack = d.stack[:len(d.stack)-1]
					return d.token(ArrEnd)
				}
			}
		case '{':
			switch d.peek() {
			case '}':
				d.stack = d.stack[:len(d.stack)-1]
				return d.token(ObjEnd)
			default:
				t := d.string()
				d.stack = append(d.stack, ':')
				return t
			}
		case '[':
			switch d.peek() {
			case ']':
				d.stack = d.stack[:len(d.stack)-1]
				return d.token(ArrEnd)
			}
		}
	}
	return d.value()
}

func (d *Decoder) value() Token {
	switch d.peek() {
	case '{':
		d.stack = append(d.stack, '{')
		d.next()
		return d.token(ObjBegin)
	case '[':
		d.stack = append(d.stack, '[')
		d.next()
		return d.token(ArrBegin)
	case '"':
		t := d.string()
		return t
	case 'n':
		d.match('n', "in literal null")
		d.match('u', "in literal null")
		d.match('l', "in literal null")
		d.match('l', "in literal null")
		return d.token(Null)
	case 't':
		d.mark = d.pos
		d.match('t', "in literal true")
		d.match('r', "in literal true")
		d.match('u', "in literal true")
		d.match('e', "in literal true")
		return d.token(Boolean)
	case 'f':
		d.mark = d.pos
		d.match('f', "in literal false")
		d.match('a', "in literal false")
		d.match('l', "in literal false")
		d.match('s', "in literal false")
		d.match('e', "in literal false")
		return d.token(Boolean)
	default:
		p := d.peek()
		if p == '-' || ('0' <= p && p <= '9') {
			return d.number()
		}
		panic(d.error(d.peek(), "looking for beginning of value"))
	}
}

func (d *Decoder) hasMore() bool {
	return d.pos < len(d.buf)
}

func (d *Decoder) peek() byte {
	if !d.hasMore() {
		panic(&SyntaxError{"unexpected end of JSON input", int64(d.pos)})
	}
	return d.buf[d.pos]
}

func (d *Decoder) next() byte {
	if !d.hasMore() {
		panic(&SyntaxError{"unexpected end of JSON input", int64(d.pos)})
	}
	d.pos++
	return d.buf[d.pos-1]
}

func (d *Decoder) match(m byte, context string) {
	if b := d.next(); b != m {
		panic(d.error(b, context))
	}
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

func (d *Decoder) Skip() {
	c := 0
	for {
		t := d.Token()
		switch {
		case t.Begin():
			c++
		case t.End():
			c--
		}
		if c == 0 {
			break
		}
	}
}

// unmarshalling ---

func (d *Decoder) Unmarshal() (v interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(*SyntaxError); ok {
				err = e
			} else {
				err = errors.New(r.(string)) // todo re-panic
			}
		}
	}()
	return d.unmarshal(), nil
}

func (d *Decoder) unmarshal() interface{} {
	t := d.Token()
	switch t.Type {
	case Null:
		return nil
	case String:
		return t.Str()
	case Number:
		return t.Float64()
	case Boolean:
		return t.Bool()
	case ObjBegin:
		m := make(map[string]interface{})
		for {
			t = d.Token()
			if t.Type == ObjEnd {
				return m
			}
			m[t.Str()] = d.unmarshal()
		}
	case ArrBegin:
		a := []interface{}(nil)
		for {
			v := d.unmarshal()
			if v == ArrEnd {
				return a
			}
			a = append(a, v)
		}
	case ArrEnd:
		return ArrEnd
	default:
		panic(fmt.Sprintln("got", t))
	}
}

// ---

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

func (d *Decoder) error(c byte, context string) error {
	return &SyntaxError{"invalid character " + quoteChar(c) + " " + context, int64(d.pos)}
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
