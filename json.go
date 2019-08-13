package json

type Decoder struct {
	buf   []byte
	pos   int
	stack []byte
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

func (t Token) Boolean() bool {
	return t.Data[0] == 't'
}

type Type int

const (
	EOF Type = iota
	ObjBegin
	ObjEnd
	ArrBegin
	ArrEnd
	String
	Number
	Null
	Boolean
)

func (d *Decoder) Token() Token {
	d.whitespace()
	if len(d.stack) > 0 {
		b := d.stack[len(d.stack)-1]
		switch b {
		case 0:
			return Token{EOF, nil}
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
					d.comma()
					return Token{ObjEnd, nil}
				} else if b == '[' {
					d.match(']')
					d.stack = d.stack[:len(d.stack)-1]
					d.comma()
					return Token{ArrEnd, nil}
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
				d.comma()
				return Token{ObjEnd, nil}
			}
		case '[':
			switch d.peek() {
			case ']':
				d.stack = d.stack[:len(d.stack)-1]
				d.comma()
				return Token{ArrEnd, nil}
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
		return Token{ObjBegin, nil}
	case '[':
		d.stack = append(d.stack, '[')
		d.next()
		return Token{ArrBegin, nil}
	case '"':
		t := d.string()
		d.comma()
		return t
	case 'n':
		d.match('n')
		d.match('u')
		d.match('l')
		d.match('l')
		d.comma()
		return Token{Null, nil}
	case 't':
		pos := d.pos
		d.match('t')
		d.match('r')
		d.match('u')
		d.match('e')
		d.comma()
		return Token{Boolean, d.buf[pos:d.pos]}
	case 'f':
		pos := d.pos
		d.match('f')
		d.match('a')
		d.match('l')
		d.match('s')
		d.match('e')
		d.comma()
		return Token{Boolean, d.buf[pos:d.pos]}
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

func (d *Decoder) comma() {
	if l := len(d.stack); l > 0 {
		b := d.stack[l-1]
		if b == '{' || b == '[' {
			d.stack = append(d.stack, ',')
		}
	} else {
		d.stack = append(d.stack, 0)
	}
}
