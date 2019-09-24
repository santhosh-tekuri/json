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

type ValueDecoder interface {
	DecodeJSON(de Decoder) error
}

type Decoder interface {
	Token() Token
	Peek() Token
	Skip() error
	Marshal() ([]byte, error)
	// UseNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64.
	UseNumber()
	Decode() (interface{}, error)
}

type ByteDecoder struct {
	buf   []byte
	pos   int
	stack []byte
	mark  int
	empty Kind // tells action to take when stack is empty
	peek  Token

	// sep contains either 0 or ',' or ':'; default is ','
	// if !stack.empty {
	//    if comma then read ',' if present
	//    if colon read ':'
	// }
	sep byte

	useNumber bool
}

func NewByteDecoder(b []byte) *ByteDecoder {
	return &ByteDecoder{buf: b, stack: make([]byte, 0, 50), empty: none, sep: ','}
}

func (d *ByteDecoder) Reset(b []byte) {
	d.buf, d.pos, d.stack, d.empty, d.sep = b, 0, d.stack[:0], none, ','
}

func (d *ByteDecoder) Peek() Token {
	if d.peek.Kind == none {
		d.peek = d.Token()
	}
	return d.peek
}

func (d *ByteDecoder) Token() Token {
	if d.peek.Kind != none {
		t := d.peek
		d.peek = Token{}
		return t
	}
	d.mark = -1
	// skip whitespace
	for d.pos < len(d.buf) && whitespace(d.buf[d.pos]) {
		d.pos++
	}
	if len(d.stack) == 0 {
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
		if d.pos == len(d.buf) {
			return d.unexpectedEOF()
		}
		s, b := d.stack[len(d.stack)-1], d.buf[d.pos]
		if d.sep == ':' {
			d.sep = ','
			if b != ':' {
				return d.error("after object key")
			}
			d.pos++ // read colon
			// skip whitespace
			for d.pos < len(d.buf) && whitespace(d.buf[d.pos]) {
				d.pos++
			}
		} else {
			comma := d.sep == ','
			d.sep = ','
			if b == s+2 {
				d.pos++
				d.stack = d.stack[:len(d.stack)-1]
				return Token{Kind: Kind(s + 2)}
			}
			if comma {
				if b != ',' {
					if s == '{' {
						return d.error("after object key:value pair")
					}
					return d.error("after array element")
				}
				d.pos++ // read comma
				// skip whitespace
				for d.pos < len(d.buf) && whitespace(d.buf[d.pos]) {
					d.pos++
				}
			}
			if s == '{' {
				if d.pos == len(d.buf) {
					return d.unexpectedEOF()
				}
				if b := d.buf[d.pos]; b != '"' {
					return d.error("looking for beginning of object key string")
				}
				d.sep = ':'
			}
		}
	}

	// read value ---
	if d.pos == len(d.buf) {
		return d.unexpectedEOF()
	}
	switch d.buf[d.pos] {
	case '{', '[':
		b := d.buf[d.pos]
		d.stack = append(d.stack, b)
		d.pos++
		d.sep = 0
		return Token{Kind: Kind(b)}
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
			return Token{Kind: Bool, Data: d.buf[d.mark:d.pos]}
		}
		return d.error("in literal true")
	case 'f':
		d.mark = d.pos
		d.pos++
		if d.match('a') && d.match('l') && d.match('s') && d.match('e') {
			return Token{Kind: Bool, Data: d.buf[d.mark:d.pos]}
		}
		return d.error("in literal false")
	default:
		return d.number()
	}
}

func (d *ByteDecoder) match(m byte) bool {
	if d.pos < len(d.buf) && d.buf[d.pos] == m {
		d.pos++
		return true
	}
	return false
}

func whitespace(p byte) bool {
	return p == ' ' || p == '\t' || p == '\r' || p == '\n'
}

func (d *ByteDecoder) Skip() error {
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

func (d *ByteDecoder) MarshalInternal() ([]byte, error) {
	if d.peek.Kind == none {
		d.Peek()
	}
	pos := d.pos
	switch d.peek.Kind {
	case ObjBegin, ObjEnd, ArrBegin, ArrEnd:
		pos--
	case Null:
		pos -= 4
	default:
		pos -= len(d.peek.Data)
	}
	if err := d.Skip(); err != nil {
		return nil, err
	}
	return d.buf[pos:d.pos], nil
}

func (d *ByteDecoder) Marshal() ([]byte, error) {
	interal, err := d.MarshalInternal()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, len(interal))
	copy(buf, interal)
	return buf, nil
}

// unmarshalling ---

// errors ---

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return fmt.Sprintf("json: %s", e.msg) }

func (d *ByteDecoder) error(context string) Token {
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

func (d *ByteDecoder) unexpectedEOF() Token {
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
