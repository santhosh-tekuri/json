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
	"io"
)

type ReadDecoder struct {
	r io.Reader
	d *Decoder
}

func NewReadDecoder(r io.Reader) *ReadDecoder {
	return &ReadDecoder{r, NewDecoder(nil)}
}

func (d *ReadDecoder) Reset(r io.Reader) {
	d.r = r
	d.d.Reset(d.d.buf[:0])
}

func (d *ReadDecoder) Peek() Token {
	if d.d.peek.Kind == none {
		d.d.peek = d.Token()
	}
	return d.d.peek
}

func (d *ReadDecoder) Token() Token {
	if d.d.peek.Kind != none {
		t := d.d.peek
		d.d.peek = Token{}
		return t
	}
	var err error
	pos, sep, empty := d.d.pos, d.d.sep, d.d.empty
	for {
		var t Token
		if d.d.buf == nil {
			t = d.d.unexpectedEOF()
			d.d.buf = make([]byte, 0, 4*1024)
		} else {
			t = d.d.Token()
		}
		if t.Kind == Number {
			if d.d.pos < len(d.d.buf) {
				return t
			}
		} else if !t.EOF() && !t.UnexpectedEOF() {
			return t
		}
		if err == nil {
			n := len(d.d.buf) - pos
			if n == 0 {
				d.d.buf = d.d.buf[:0]
				d.d.pos = 0
			} else {
				buf := d.d.buf
				if n == cap(d.d.buf) {
					buf = make([]byte, 1024)
				}
				copy(buf, d.d.buf[pos:])
				d.d.buf = buf[:n]
				d.d.pos = 0
			}
			r := 0
			r, err = d.r.Read(d.d.buf[n:cap(d.d.buf)])
			if r > 0 {
				d.d.buf = d.d.buf[:n+r]
				d.d.sep, d.d.empty = sep, empty
				pos = 0
				continue
			}
		}
		if err == io.EOF {
			return t
		}
		return Token{Kind: Error, Err: err}
	}
}

func (d *ReadDecoder) Unmarshal() (v interface{}, err error) {
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
