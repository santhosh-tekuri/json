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

func (d *ByteDecoder) Unmarshal(useNumber bool) (v interface{}, err error) {
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

type PropUnmarshaller func(de Decoder, prop Token) error

func UnmarshalObj(context string, d Decoder, f PropUnmarshaller) error {
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
			if d.Peek().Null() {
				d.Token()
				continue
			}
			if err = f(d, t); err != nil {
				return err
			}
		}
	}
}

type ItemUnmarshaller func(de Decoder) error

func UnmarshalArr(context string, d Decoder, f ItemUnmarshaller) error {
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
