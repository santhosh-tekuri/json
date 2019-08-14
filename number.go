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
	if d.hasMore() && d.peek() == '-' {
		d.next()
	}

	// digits
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	b := d.next()
	switch {
	case b == '0':
	case '1' <= b && b <= '9':
		d.digits()
	default:
		return d.error(b, "in numeric literal")
	}
	if d.hasMore() {
		if d.peek() == '.' {
			d.next()
			if d.oneOrMoreDigits() == Error {
				return Error
			}
		}
		if d.hasMore() {
			p := d.peek()
			if p == 'e' || p == 'E' {
				d.next()
				if d.hasMore() {
					p = d.peek()
					if p == '+' || p == '-' {
						d.next()
					}
				}
				if d.oneOrMoreDigits() == Error {
					return Error
				}
			}
		}
	}
	return Number
}

func (d *Decoder) digits() {
	for d.hasMore() {
		p := d.peek()
		if '0' <= p && p <= '9' {
			d.next()
		} else {
			return
		}
	}
}

func (d *Decoder) oneOrMoreDigits() Kind {
	if !d.hasMore() {
		return d.unexpectedEOF()
	}
	b := d.next()
	if !('0' <= b && b <= '9') {
		return d.error(b, "in numeric literal")
	}
	d.digits()
	return noError
}
