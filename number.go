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
	d.mark = d.pos

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
	return d.token(Number)
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
