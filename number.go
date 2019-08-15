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

func (d *Decoder) number() Kind {
	d.mark = d.pos

	// optional -
	if d.hasMore() && d.buf[d.pos] == '-' {
		d.pos++
	}

	// digits
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	b := d.buf[d.pos]
	switch {
	case b == '0':
		d.pos++
	case '1' <= b && b <= '9':
		d.pos++
		d.digits()
	default:
		return d.error(b, "in numeric literal")
	}
	if d.hasMore() {
		if d.buf[d.pos] == '.' {
			d.pos++
			if d.oneOrMoreDigits("after decimal point in numeric literal") == Error {
				return Error
			}
		}
		if d.hasMore() {
			p := d.buf[d.pos]
			if p == 'e' || p == 'E' {
				d.pos++
				if d.hasMore() {
					p = d.buf[d.pos]
					if p == '+' || p == '-' {
						d.pos++
					}
				}
				if d.oneOrMoreDigits("in exponent of numeric literal") == Error {
					return Error
				}
			}
		}
	}
	return Number
}

func (d *Decoder) digits() {
	for d.hasMore() {
		p := d.buf[d.pos]
		if '0' <= p && p <= '9' {
			d.pos++
		} else {
			return
		}
	}
}

func (d *Decoder) oneOrMoreDigits(context string) Kind {
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	b := d.buf[d.pos]
	if !('0' <= b && b <= '9') {
		return d.error(b, context)
	}
	d.pos++
	d.digits()
	return none
}
