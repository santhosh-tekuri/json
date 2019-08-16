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
		for d.pos < len(d.buf) && digit(d.buf[d.pos]) {
			d.pos++
		}
	default:
		if d.mark == d.pos {
			return d.error("looking for beginning of value")
		}
		return d.error("in numeric literal")
	}

	// fraction: dot followed by one or more digits
	if d.pos < len(d.buf) && d.buf[d.pos] == '.' {
		d.pos++
		if d.pos == len(d.buf) || !digit(d.buf[d.pos]) {
			return d.error("after decimal point in numeric literal")
		}
		for d.pos < len(d.buf) && digit(d.buf[d.pos]) {
			d.pos++
		}
	}

	// exponent: e/E optional(+/-) followed by one or more digits
	if d.pos < len(d.buf) {
		if b = d.buf[d.pos]; b == 'e' || b == 'E' {
			d.pos++
			if d.pos < len(d.buf) {
				if b = d.buf[d.pos]; b == '+' || b == '-' {
					d.pos++
				}
			}
			if d.pos == len(d.buf) || !digit(d.buf[d.pos]) {
				return d.error("in exponent of numeric literal")
			}
			for d.pos < len(d.buf) && digit(d.buf[d.pos]) {
				d.pos++
			}
		}
	}

	return Token{Kind: Number, Data: d.buf[d.mark:d.pos]}
}

func digit(p byte) bool {
	return '0' <= p && p <= '9'
}
