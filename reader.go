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
	"bytes"
	"fmt"
	"io"
)

type ReadDecoder struct {
	r io.Reader
	d *ByteDecoder
}

func NewReadDecoder(r io.Reader) *ReadDecoder {
	return &ReadDecoder{r, NewByteDecoder(nil)}
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
		if t.Kind == Num {
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
			} else {
				buf := d.d.buf
				if n == cap(d.d.buf) {
					buf = make([]byte, 1024)
				}
				copy(buf, d.d.buf[pos:])
				d.d.buf = buf[:n]
			}
			d.d.pos = n
			if len(t.Data) > 0 {
				t.Data = d.d.buf[n-len(t.Data):]
			}
			r := 0
			r, err = d.r.Read(d.d.buf[n:cap(d.d.buf)])
			if r > 0 {
				d.d.buf = d.d.buf[:n+r]
				d.d.sep, d.d.empty = sep, empty
				d.d.pos, pos = 0, 0
				continue
			}
		}
		if err == io.EOF {
			return t
		}
		return Token{Kind: Error, Err: err}
	}
}

func (d *ReadDecoder) Marshal() ([]byte, error) {
	t := d.Token()
	switch t.Kind {
	case Error:
		return nil, t.Err
	case Null:
		return []byte("null"), nil
	case Str, Num, Bool:
		buf := make([]byte, len(t.Data))
		copy(buf, t.Data)
		return buf, nil
	default:
		d.d.peek = t
		buf := new(bytes.Buffer)
		if err := d.marshal(buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

func (d *ReadDecoder) marshal(buf *bytes.Buffer) error {
	t := d.Token()
	switch t.Kind {
	case Error:
		return t.Err
	case Null:
		buf.WriteString("null")
	case Str, Num, Bool:
		buf.Write(t.Data)
	case ObjBegin:
		buf.WriteByte('{')
		comma := false
		for {
			t = d.Token()
			switch t.Kind {
			case Error:
				return t.Err
			case ObjEnd:
				buf.WriteByte('}')
				return nil
			}
			if comma {
				buf.WriteByte(',')
			}
			comma = true
			buf.Write(t.Data) // key
			buf.WriteByte(':')
			if err := d.marshal(buf); err != nil {
				return err
			}
		}
	case ArrBegin:
		buf.WriteByte('[')
		comma := false
		for {
			switch d.Peek().Kind {
			case Error:
				d.Token()
				return t.Err
			case ArrEnd:
				d.Token()
				buf.WriteByte(']')
				return nil
			}
			if comma {
				buf.WriteByte(',')
			}
			comma = true
			if err := d.marshal(buf); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *ReadDecoder) Unmarshal(useNumber bool) (v interface{}, err error) {
	t := d.Token()
	switch t.Kind {
	case Error:
		return nil, t.Err
	case Null:
		return nil, nil
	case Str:
		return t.String("")
	case Num:
		if useNumber {
			return t.Number("")
		}
		return t.Float64("")
	case Bool:
		return t.Bool("")
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
			key, _ := t.String("")
			v, err := d.Unmarshal(useNumber)
			if err != nil {
				return nil, err
			}
			m[key] = v
		}
	case ArrBegin:
		a := make([]interface{}, 0)
		for {
			v, err := d.Unmarshal(useNumber)
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

func (d *ReadDecoder) Skip() error {
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
