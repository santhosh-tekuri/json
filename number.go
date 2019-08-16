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

func (d *Decoder) number() Token {
	if d.pos == len(d.buf) {
		return d.unexpectedEOF()
	}
	d.mark = d.pos

	// optional -
	if d.buf[d.pos] == '-' {
		d.pos++
		if d.pos == len(d.buf) {
			return d.unexpectedEOF()
		}
	}

	// digits
	b := d.buf[d.pos]
	switch {
	case b == '0':
		d.pos++
	case '1' <= b && b <= '9':
		d.pos++
		d.digits()
	default:
		if d.mark == d.pos {
			return d.error("looking for beginning of value")
		}
		return d.error("in numeric literal")
	}
	if d.pos < len(d.buf) {
		if d.buf[d.pos] == '.' {
			d.pos++
			if t := d.oneOrMoreDigits("after decimal point in numeric literal"); t.Kind == Error {
				return t
			}
		}
		if d.pos < len(d.buf) {
			p := d.buf[d.pos]
			if p == 'e' || p == 'E' {
				d.pos++
				if d.pos < len(d.buf) {
					p = d.buf[d.pos]
					if p == '+' || p == '-' {
						d.pos++
					}
				}
				if t := d.oneOrMoreDigits("in exponent of numeric literal"); t.Kind == Error {
					return t
				}
			}
		}
	}
	return Token{Kind: Number, Data: d.buf[d.mark:d.pos]}
}

func (d *Decoder) digits() {
	for d.pos < len(d.buf) {
		p := d.buf[d.pos]
		if '0' <= p && p <= '9' {
			d.pos++
		} else {
			return
		}
	}
}

func (d *Decoder) oneOrMoreDigits(context string) Token {
	if d.pos == len(d.buf) {
		return d.unexpectedEOF()
	}
	b := d.buf[d.pos]
	if !('0' <= b && b <= '9') {
		return d.error(context)
	}
	d.pos++
	d.digits()
	return Token{Kind: none}
}
