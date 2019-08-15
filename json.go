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
	empty Kind // tells action to take when stack is empty
	comma bool // if comma && !stack.empty then read ','
	colon bool // if true read ':'
	peek  Token
}

func NewDecoder(b []byte) *Decoder {
	return &Decoder{buf: b, stack: make([]byte, 0, 50), empty: none, comma: true, colon: false}
}

func (d *Decoder) Reset(b []byte) {
	d.buf, d.pos, d.stack, d.empty, d.comma, d.colon = b, 0, d.stack[:0], none, true, false
}

func (d *Decoder) Peek() Token {
	if d.peek.Kind == none {
		d.peek = d.Token()
	}
	return d.peek
}

func (d *Decoder) Token() Token {
	if d.peek.Kind != none {
		t := d.peek
		d.peek = Token{}
		return t
	}
	d.mark = -1
	d.whitespace()
	if d.colon {
		d.colon = false
		if !d.match(':') {
			return d.error("after object key")
		}
		d.whitespace()
	} else if len(d.stack) == 0 {
		switch d.empty {
		case none:
			d.empty = EOD
		case EOD:
			d.empty = EOF
			return Token{Kind: EOD}
		case EOF:
			if d.pos == len(d.buf) {
				return Token{Kind: EOF}
			}
			d.empty = EOD
		}
	} else {
		s := d.stack[len(d.stack)-1]
		if d.comma {
			if d.pos == len(d.buf) {
				return d.unexpectedEOF()
			}
			if d.buf[d.pos] != ',' {
				if s == '{' {
					if !d.match('}') {
						return d.error("after object key:value pair")
					}
					d.stack = d.stack[:len(d.stack)-1]
					return Token{Kind: ObjEnd}
				} else if s == '[' {
					if !d.match(']') {
						return d.error("after array element")
					}
					d.stack = d.stack[:len(d.stack)-1]
					return Token{Kind: ArrEnd}
				}
			}
			// has comma
			d.pos++
			d.whitespace()
			if s == '{' {
				t := d.string()
				d.colon = true
				return t
			}
		} else {
			// it is next token after '{' or '['
			d.comma = true
			switch s {
			case '{':
				if d.pos == len(d.buf) {
					return d.unexpectedEOF()
				}
				switch d.buf[d.pos] {
				case '}':
					d.pos++
					d.stack = d.stack[:len(d.stack)-1]
					return Token{Kind: ObjEnd}
				default:
					t := d.string()
					d.colon = true
					return t
				}
			case '[':
				if d.pos == len(d.buf) {
					return d.unexpectedEOF()
				}
				if d.buf[d.pos] == ']' {
					d.pos++
					d.stack = d.stack[:len(d.stack)-1]
					return Token{Kind: ArrEnd}
				}
			}
		}
	}

	// read value ---
	if d.pos == len(d.buf) {
		return d.unexpectedEOF()
	}
	switch d.buf[d.pos] {
	case '{':
		d.stack = append(d.stack, '{')
		d.pos++
		d.comma = false
		return Token{Kind: ObjBegin}
	case '[':
		d.stack = append(d.stack, '[')
		d.pos++
		d.comma = false
		return Token{Kind: ArrBegin}
	case '"':
		return d.string()
	case 'n':
		d.pos++
		if d.match('u') && d.match('l') && d.match('l') {
			return Token{Kind: Null}
		}
		return d.error("in literal null")
	case 't':
		d.mark = d.pos
		d.pos++
		if d.match('r') && d.match('u') && d.match('e') {
			return Token{Kind: Boolean, Data: d.buf[d.mark:d.pos]}
		}
		return d.error("in literal true")
	case 'f':
		d.mark = d.pos
		d.pos++
		if d.match('a') && d.match('l') && d.match('s') && d.match('e') {
			return Token{Kind: Boolean, Data: d.buf[d.mark:d.pos]}
		}
		return d.error("in literal false")
	default:
		p := d.buf[d.pos]
		if p == '-' || ('0' <= p && p <= '9') {
			return d.number()
		}
		return d.error("looking for beginning of value")
	}
}

func (d *Decoder) match(m byte) bool {
	if d.pos < len(d.buf) && d.buf[d.pos] == m {
		d.pos++
		return true
	}
	return false
}

func (d *Decoder) whitespace() {
	for d.pos < len(d.buf) {
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
		s, _ := t.Str("")
		return s, nil
	case Number:
		f, _ := t.Float64("")
		return f, nil
	case Boolean:
		b, _ := t.Bool("")
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
			key, _ := t.Str("")
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

type PropUnmarshaller func(de *Decoder, prop Token) error

func (d *Decoder) UnmarshalObj(context string, f PropUnmarshaller) error {
	if err := d.Token().Obj(context); err != nil {
		return err
	}
	var err error
	for {
		t := d.Token()
		switch {
		case t.Error():
			return t.Err
		case t.End():
			return nil
		default:
			if err = f(d, t); err != nil {
				return err
			}
		}
	}
}

type ArrUnmarshaller func(de *Decoder) error

func (d *Decoder) UnmarshalArr(context string, f ArrUnmarshaller) error {
	if err := d.Token().Arr(context); err != nil {
		return err
	}
	for !d.Peek().End() {
		if err := f(d); err != nil {
			return err
		}
	}
	return d.Token().Err
}

// errors ---

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

func (d *Decoder) error(context string) Token {
	if d.pos == len(d.buf) {
		return d.unexpectedEOF()
	}
	return Token{
		Kind: Error,
		Err:  &SyntaxError{"invalid character " + quoteChar(d.buf[d.pos]) + " " + context, int64(d.pos)},
	}
}

var unexpectedEOF = "unexpected end of JSON input"

func IsUnexpectedEOF(err error) bool {
	e, ok := err.(*SyntaxError)
	return ok && e.msg == unexpectedEOF
}

func (d *Decoder) unexpectedEOF() Token {
	return Token{
		Kind: Error,
		Err:  &SyntaxError{unexpectedEOF, int64(d.pos)},
	}
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
