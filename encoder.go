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
	"unicode/utf8"
)

func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type Encoder struct {
	scratch  [64]byte
	w        io.Writer
	writeStr func(string) (int, error)
	err      error
}

func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{w: w}
	if sw, ok := w.(io.StringWriter); ok {
		e.writeStr = sw.WriteString
	} else {
		e.writeStr = func(s string) (int, error) {
			return w.Write([]byte(s))
		}
	}
	return e
}

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
)

func (e *Encoder) Encode(v interface{}) (err error) {
	switch v := v.(type) {
	case Marshaler:
		b, err := v.MarshalJSON()
		if err != nil {
			return err
		}
		e.write(b)
	case nil:
		e.write(null)
	case bool:
		if v {
			e.write(boolTrue)
		} else {
			e.write(boolFalse)
		}
	case int:
		b := strconv.AppendInt(e.scratch[:0], int64(v), 10)
		e.write(b)
	case int8:
		b := strconv.AppendInt(e.scratch[:0], int64(v), 10)
		e.write(b)
	case int16:
		b := strconv.AppendInt(e.scratch[:0], int64(v), 10)
		e.write(b)
	case int32:
		b := strconv.AppendInt(e.scratch[:0], int64(v), 10)
		e.write(b)
	case int64:
		b := strconv.AppendInt(e.scratch[:0], v, 10)
		e.write(b)
	case uint:
		b := strconv.AppendUint(e.scratch[:0], uint64(v), 10)
		e.write(b)
	case uint8:
		b := strconv.AppendUint(e.scratch[:0], uint64(v), 10)
		e.write(b)
	case uint16:
		b := strconv.AppendUint(e.scratch[:0], uint64(v), 10)
		e.write(b)
	case uint32:
		b := strconv.AppendUint(e.scratch[:0], uint64(v), 10)
		e.write(b)
	case uint64:
		b := strconv.AppendUint(e.scratch[:0], v, 10)
		e.write(b)
	case float32:
		err = e.encodeFloat(float64(v), 32)
	case float64:
		err = e.encodeFloat(v, 64)
	case string:
		e.encodeString(v)
	case []interface{}:
		if v == nil {
			e.write(null)
			break
		}
		e.write(arrOpen)
		for i, elem := range v {
			if i > 0 {
				e.write(comma)
			}
			if err = e.Encode(elem); err != nil {
				return err
			}
		}
		e.write(arrClose)
	case map[string]interface{}:
		if v == nil {
			e.write(null)
			break
		}
		e.write(objOpen)
		i := 0
		for key, val := range v {
			if i > 0 {
				e.write(comma)
			}
			e.encodeString(key)
			e.write(colon)
			if err = e.Encode(val); err != nil {
				return err
			}
			i++
		}
		e.write(objClose)
	case []byte:
		e.write(quote)
		_, err = base64.NewEncoder(base64.StdEncoding, e.w).Write(v)
		e.write(quote)
	default:
		return UnsupportedTypeError(reflect.TypeOf(v).String())
	}
	if err != nil {
		return err
	}
	return e.err
}

func (e *Encoder) NewLine() error {
	e.write(newLine)
	return e.err
}

// ---

func (e *Encoder) encodeFloat(f float64, bits int) error {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return UnsupportedValueError(strconv.FormatFloat(f, 'g', -1, bits))
	}

	abs := math.Abs(f)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}

	b := e.scratch[:0]
	b = strconv.AppendFloat(b, f, fmt, -1, bits)
	if fmt == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}
	e.write(b)
	return nil
}

// ---

var (
	quote    = []byte(`"`)
	u00      = []byte(`\u00`)
	u202     = []byte(`\u202`)
	ufffd    = []byte(`\ufffd`)
	escSlash = []byte(`\\`)
	escQuote = []byte(`\"`)
	escLF    = []byte(`\n`)
	escCR    = []byte(`\r`)
	escFF    = []byte(`\f`)
	escTAB   = []byte(`\t`)
)

func (e *Encoder) encodeString(s string) {
	e.write(quote)
	start := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b < 0x20 {
			if start < i {
				e.writeString(s[start:i])
			}
			switch b {
			case '\n':
				e.write(escLF)
			case '\r':
				e.write(escCR)
			case '\f':
				e.write(escFF)
			case '\t':
				e.write(escTAB)
			default:
				e.write(u00)
				e.write(hex(b >> 4))
				e.write(hex(b & 0xF))
			}
			i++
			start = i
			continue
		}
		if b < utf8.RuneSelf {
			if b == '\\' || b == '"' {
				if start < i {
					e.writeString(s[start:i])
				}
				switch b {
				case '\\':
					e.write(escSlash)
				case '"':
					e.write(escQuote)
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
				e.writeString(s[start:i])
			}
			e.write(ufffd)
			i += size
			start = i
			continue
		}
		if r == '\u2028' || r == '\u2029' {
			if start < i {
				e.writeString(s[start:i])
			}
			e.write(u202)
			e.write(hex(uint8(r & 0xF)))
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.writeString(s[start:])
	}
	e.write(quote)
}

var hexBytes = []byte("0123456789abcdef")

func hex(i uint8) []byte {
	return hexBytes[i : i+1]
}

func (e *Encoder) write(b []byte) {
	if e.err == nil {
		_, e.err = e.w.Write(b)
	}
}

func (e *Encoder) writeString(s string) {
	if e.err == nil {
		_, e.err = e.writeStr(s)
	}
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
