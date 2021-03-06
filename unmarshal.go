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

import "fmt"

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64.
func (d *ByteDecoder) UseNumber() {
	d.useNumber = true
}

func (d *ByteDecoder) Decode() (v interface{}, err error) {
	t := d.Token()
	switch t.Kind {
	case Error:
		return nil, t.Err
	case Null:
		return nil, nil
	case Str:
		return t.String("")
	case Num:
		if d.useNumber {
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
			key, err := t.String("")
			if err != nil {
				return nil, err
			}
			v, err := d.Decode()
			if err != nil {
				return nil, err
			}
			m[key] = v
		}
	case ArrBegin:
		a := make([]interface{}, 0)
		for {
			v, err := d.Decode()
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

type DecodeProp func(de Decoder, prop Token) error

func DecodeObj(context string, d Decoder, f DecodeProp) error {
	t := d.Token()
	if t.Null() {
		return nil
	}
	if err := t.Obj(context); err != nil {
		return err
	}
	var err error
	for {
		t := d.Token()
		switch {
		case t.Error():
			return t.Err
		case t.End():
			return nil
		default:
			if err = f(d, t); err != nil {
				return err
			}
		}
	}
}

type DecodeItem func(de Decoder) error

func DecodeArr(context string, d Decoder, f DecodeItem) error {
	t := d.Token()
	if t.Null() {
		return nil
	}
	if err := t.Arr(context); err != nil {
		return err
	}
	for !d.Peek().End() {
		if err := f(d); err != nil {
			return err
		}
	}
	return d.Token().Err
}

// --

type UnmarshalError struct {
	Context string
	Type    string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("json: %s expects %s", e.Context, e.Type)
}
