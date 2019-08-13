package json

import "strconv"

func (d *Decoder) number() Token {
	pos := d.pos

	// optional -
	if d.peek() == '-' {
		d.next()
	}

	// digits
	b := d.next()
	switch {
	case b == '0':
	case '1' <= b && b <= '9':
		d.digits()
	default:
		panic("invalid number")
	}
	if d.hasMore() {
		if d.peek() == '.' {
			d.next()
			d.oneOrMoreDigits()
		}
		if d.hasMore() {
			b = d.peek()
			if b == 'e' || b == 'E' {
				d.next()
				b = d.peek()
				if b == '+' || b == '-' {
					d.next()
				}
				d.oneOrMoreDigits()
			}
		}
	}
	return Token{Number, d.buf[pos:d.pos]}
}

func (d *Decoder) digits() {
	for d.hasMore() {
		b := d.peek()
		if '0' <= b && b <= '9' {
			d.next()
		} else {
			return
		}
	}
}

func (d *Decoder) oneOrMoreDigits() {
	b := d.next()
	if !('0' <= b && b <= '9') {
		panic("invalid number")
	}
	d.digits()
}

// Float64 returns the number as a float64.
func (t Token) Float64() float64 {
	f, _ := strconv.ParseFloat(string(t.Data), 64)
	return f
}

// Int64 returns the number as an int64.
func (t Token) Int64() int64 {
	i, _ := strconv.ParseInt(string(t.Data), 10, 64)
	return i
}

// Int returns the number as an int.
func (t Token) Int() int {
	return int(t.Int64())
}
