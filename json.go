package json

import (
	"errors"
	"fmt"
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

type Token struct {
	Type Type
	Data []byte
}

func (t Token) Bool() bool {
	return t.Data[0] == 't'
}

type Type int

const (
	EOF Type = iota + 1
	ObjBegin
	ObjEnd
	ArrBegin
	ArrEnd
	String
	Number
	Null
	Boolean
)

func (d *Decoder) token(t Type) Token {
	if len(d.stack) > 0 {
		b := d.stack[len(d.stack)-1]
		if b == '{' || b == '[' {
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
		b := d.stack[len(d.stack)-1]
		switch b {
		case 0:
			return d.token(EOF)
		case ':':
			d.stack = d.stack[:len(d.stack)-1]
			d.match(':')
			d.whitespace()
		case ',':
			d.stack = d.stack[:len(d.stack)-1]
			if d.peek() == ',' {
				d.match(',')
				d.whitespace()
				b := d.stack[len(d.stack)-1]
				if b == '{' {
					t := d.string()
					d.stack = append(d.stack, ':')
					return t
				}
			} else {
				b := d.stack[len(d.stack)-1]
				if b == '{' {
					d.match('}')
					d.stack = d.stack[:len(d.stack)-1]
					return d.token(ObjEnd)
				} else if b == '[' {
					d.match(']')
					d.stack = d.stack[:len(d.stack)-1]
					return d.token(ArrEnd)
				}
			}
		case '{':
			switch d.peek() {
			case '"':
				t := d.string()
				d.stack = append(d.stack, ':')
				return t
			case '}':
				d.stack = d.stack[:len(d.stack)-1]
				return d.token(ObjEnd)
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
		d.match('n')
		d.match('u')
		d.match('l')
		d.match('l')
		return d.token(Null)
	case 't':
		d.mark = d.pos
		d.match('t')
		d.match('r')
		d.match('u')
		d.match('e')
		return d.token(Boolean)
	case 'f':
		d.mark = d.pos
		d.match('f')
		d.match('a')
		d.match('l')
		d.match('s')
		d.match('e')
		return d.token(Boolean)
	default:
		panic("value expected")
	}
}

func (d *Decoder) hasMore() bool {
	return d.pos < len(d.buf)
}

func (d *Decoder) peek() byte {
	if !d.hasMore() {
		panic("unexpeted EOF")
	}
	return d.buf[d.pos]
}

func (d *Decoder) next() byte {
	if !d.hasMore() {
		panic("unexpeted EOF")
	}
	d.pos++
	return d.buf[d.pos-1]
}

func (d *Decoder) match(b byte) {
	if d.next() != b {
		panic("expecting '" + string(b) + "'")
	}
}

func (d *Decoder) whitespace() {
	for d.hasMore() {
		if b := d.peek(); b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			d.next()
		} else {
			break
		}
	}
}

func (d *Decoder) Unmarshal() (v interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(r.(string))
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
