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
	"errors"
	"strconv"
)

type Kind byte

const (
	none Kind = iota
	EOD
	EOF
	String
	Number
	Boolean
	Null
	Error
	ObjBegin, ObjEnd Kind = '{', '}'
	ArrBegin, ArrEnd Kind = '[', ']'
)

var kindNames = []string{
	`<none>`,
	`<eod>`, `<eof>`,
	`<string>`, `<number>`, `<boolean>`, `<null>`,
	`<error>`,
}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return string(k)
}

type Token struct {
	Kind Kind
	Data []byte
	Err  error
}

func (t Token) Error() bool {
	return t.Kind == Error
}

func (t Token) EOD() bool {
	return t.Kind == EOD
}

func (t Token) EOF() bool {
	return t.Kind == EOF
}

func (t Token) Begin() bool {
	return t.Kind == ObjBegin || t.Kind == ArrBegin
}

func (t Token) End() bool {
	return t.Kind == ObjEnd || t.Kind == ArrEnd
}

func (t Token) Null() bool {
	return t.Kind == Null
}

func (t Token) Obj(context string) error {
	switch t.Kind {
	case Error:
		return t.Err
	case ObjBegin:
		return nil
	default:
		return errors.New(context + " expects object")
	}
}

func (t Token) Arr(context string) error {
	switch t.Kind {
	case Error:
		return t.Err
	case ArrBegin:
		return nil
	default:
		return errors.New(context + " expects array")
	}
}

func (t Token) Str(context string) (string, error) {
	switch t.Kind {
	case Error:
		return "", t.Err
	case String:
		s, _ := unquoteBytes(t.Data)
		return string(s), nil
	default:
		return "", errors.New(context + " expects string")
	}
}

func (t Token) Number(context string) (string, error) {
	switch t.Kind {
	case Error:
		return "", t.Err
	case Number:
		return string(t.Data), nil
	default:
		return "", errors.New(context + " expects number")
	}
}

// Float64 returns the number as a float64.
func (t Token) Float64(context string) (float64, error) {
	switch t.Kind {
	case Error:
		return 0, t.Err
	case Number:
		return strconv.ParseFloat(string(t.Data), 64)
	default:
		return 0, errors.New(context + " expects number")
	}
}

// Int64 returns the number as an int64.
func (t Token) Int64(context string) (int64, error) {
	switch t.Kind {
	case Error:
		return 0, t.Err
	case Number:
		i, err := strconv.ParseInt(string(t.Data), 10, 64)
		if err != nil {
			return 0, errors.New(context + " expects integer")
		}
		return i, nil
	default:
		return 0, errors.New(context + " expects integer")
	}
}

// Int returns the number as an int.
func (t Token) Int(context string) (int, error) {
	i, err := t.Int64(context)
	return int(i), err
}

// Bool returns the boolean value.
func (t Token) Bool(context string) (bool, error) {
	switch t.Kind {
	case Error:
		return false, t.Err
	case Boolean:
		return t.Data[0] == 't', nil
	default:
		return false, errors.New(context + " expects boolean")
	}
}
