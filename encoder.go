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
	"encoding/base64"
	"io"
	"math"
	"reflect"
	"strconv"
	"sync"
	"unicode/utf8"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func Marshal(v interface{}) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	w := NewWriter(buf)
	w.Value(v)
	if w.Err != nil {
		buf.Reset()
		bufferPool.Put(buf)
		return nil, w.Err
	}
	b := append([]byte(nil), buf.Bytes()...)
	buf.Reset()
	bufferPool.Put(buf)
	return b, nil
}

type Encoder struct {
	w *Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{NewWriter(w)}
}

func (e *Encoder) Encode(v interface{}) error {
	if e.w.Err == nil {
		e.w.Value(v)
	}
	return e.w.Err
}

func (e *Encoder) NewLine() error {
	e.w.write(newLine)
	return e.w.Err
}

// writer ---

var (
	null      = []byte(`null`)
	boolTrue  = []byte(`true`)
	boolFalse = []byte(`false`)
	comma     = []byte(`,`)
	colon     = []byte(`:`)
	newLine   = []byte("\n")
	arrOpen   = []byte(`[`)
	arrClose  = []byte(`]`)
	objOpen   = []byte(`{`)
	objClose  = []byte(`}`)
	quote     = []byte(`"`)
	u00       = []byte(`\u00`)
	u202      = []byte(`\u202`)
	ufffd     = []byte(`\ufffd`)
	escSlash  = []byte(`\\`)
	escQuote  = []byte(`\"`)
	escLF     = []byte(`\n`)
	escCR     = []byte(`\r`)
	escFF     = []byte(`\f`)
	escTAB    = []byte(`\t`)
)

type Writer struct {
	scratch  [64]byte
	w        io.Writer
	writeStr func(string) (int, error)
	Err      error
}

func NewWriter(w io.Writer) *Writer {
	jw := &Writer{w: w}
	if sw, ok := w.(io.StringWriter); ok {
		jw.writeStr = sw.WriteString
	} else {
		jw.writeStr = func(s string) (int, error) {
			return w.Write([]byte(s))
		}
	}
	return jw
}

func (w *Writer) Value(v interface{}) {
	if w.Err != nil {
		return
	}
	switch v := v.(type) {
	case ValueEncoder:
		v.EncodeJSON(w)
	case Marshaler:
		b, err := v.MarshalJSON()
		if err == nil {
			w.write(b)
		}
		w.Err = err
	case nil:
		w.Null()
	case bool:
		w.Bool(v)
	case int:
		w.Int(int64(v))
	case int8:
		w.Int(int64(v))
	case int16:
		w.Int(int64(v))
	case int32:
		w.Int(int64(v))
	case int64:
		w.Int(v)
	case uint:
		w.Uint(uint64(v))
	case uint8:
		w.Uint(uint64(v))
	case uint16:
		w.Uint(uint64(v))
	case uint32:
		w.Uint(uint64(v))
	case uint64:
		w.Uint(v)
	case float32:
		w.Float32(v)
	case float64:
		w.Float64(v)
	case string:
		w.String(v)
	case []interface{}:
		w.Array(v)
	case map[string]interface{}:
		w.Object(v)
	case []byte:
		w.Bytes(v)
	default:
		w.Err = UnsupportedTypeError(reflect.TypeOf(v).String())
	}
}

func (w *Writer) Null() {
	w.write(null)
}

func (w *Writer) Bool(v bool) {
	if v {
		w.write(boolTrue)
	} else {
		w.write(boolFalse)
	}
}

func (w *Writer) Int(v int64) {
	b := strconv.AppendInt(w.scratch[:0], v, 10)
	w.write(b)
}

func (w *Writer) Uint(v uint64) {
	b := strconv.AppendUint(w.scratch[:0], v, 10)
	w.write(b)
}

func (w *Writer) Float32(f float32) {
	w.float(float64(f), 32)
}

func (w *Writer) Float64(f float64) {
	w.float(f, 64)
}

func (w *Writer) float(f float64, bits int) {
	if w.Err != nil {
		return
	}
	if math.IsInf(f, 0) || math.IsNaN(f) {
		w.Err = UnsupportedValueError(strconv.FormatFloat(f, 'g', -1, bits))
		return
	}

	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}

	b := w.scratch[:0]
	b = strconv.AppendFloat(b, f, fmt, -1, bits)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	w.write(b)
}

func (w *Writer) String(s string) {
	if w.Err != nil {
		return
	}
	w.write(quote)
	start := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b < 0x20 {
			if start < i {
				w.writeString(s[start:i])
			}
			switch b {
			case '\n':
				w.write(escLF)
			case '\r':
				w.write(escCR)
			case '\f':
				w.write(escFF)
			case '\t':
				w.write(escTAB)
			default:
				w.write(u00)
				w.write(hex(b >> 4))
				w.write(hex(b & 0xF))
			}
			i++
			start = i
			continue
		}
		if b < utf8.RuneSelf {
			if b == '\\' || b == '"' {
				if start < i {
					w.writeString(s[start:i])
				}
				switch b {
				case '\\':
					w.write(escSlash)
				case '"':
					w.write(escQuote)
				}
				i++
				start = i
				continue
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			if start < i {
				w.writeString(s[start:i])
			}
			w.write(ufffd)
			i += size
			start = i
			continue
		}
		if r == '\u2028' || r == '\u2029' {
			if start < i {
				w.writeString(s[start:i])
			}
			w.write(u202)
			w.write(hex(uint8(r & 0xF)))
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		w.writeString(s[start:])
	}
	w.write(quote)
}

func (w *Writer) Bytes(v []byte) {
	if w.Err == nil {
		w.write(quote)
		_, w.Err = base64.NewEncoder(base64.StdEncoding, w.w).Write(v)
		w.write(quote)
	}
}

func (w *Writer) Object(v map[string]interface{}) {
	if v == nil {
		w.write(null)
		return
	}
	w.write(objOpen)
	i := 0
	for key, val := range v {
		if i > 0 {
			w.write(comma)
		}
		w.String(key)
		w.write(colon)
		w.Value(val)
		if w.Err != nil {
			return
		}
		i++
	}
	w.write(objClose)
}

func (w *Writer) Array(v []interface{}) {
	if v == nil {
		w.write(null)
		return
	}
	w.write(arrOpen)
	for i, elem := range v {
		if i > 0 {
			w.write(comma)
		}
		w.Value(elem)
		if w.Err != nil {
			return
		}
	}
	w.write(arrClose)
}

func (w *Writer) Comma() {
	w.write(comma)
}

func (w *Writer) StartObject() {
	w.write(objOpen)
}

func (w *Writer) Prop(s string) {
	w.String(s)
	w.write(colon)
}

func (w *Writer) EndObject() {
	w.write(objClose)
}

func (w *Writer) StartArray() {
	w.write(arrOpen)
}

func (w *Writer) EndArray() {
	w.write(arrClose)
}

var hexBytes = []byte("0123456789abcdef")

func hex(i uint8) []byte {
	return hexBytes[i : i+1]
}

func (w *Writer) Raw(s string) {
	w.writeString(s)
}

func (w *Writer) write(b []byte) {
	if w.Err == nil {
		_, w.Err = w.w.Write(b)
	}
}

func (w *Writer) writeString(s string) {
	if w.Err == nil {
		_, w.Err = w.writeStr(s)
	}
}

type ValueEncoder interface {
	EncodeJSON(*Writer)
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError string

func (e UnsupportedTypeError) Error() string {
	return "json: unsupported type: " + string(e)
}

type UnsupportedValueError string

func (e UnsupportedValueError) Error() string {
	return "json: unsupported value: " + string(e)
}
