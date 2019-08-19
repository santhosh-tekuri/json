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
	"strconv"
)

// Kind tells the kind of json token.
type Kind byte

const (
	none Kind = iota
	EOD
	EOF
	Str
	Num
	Bool
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

// Token holds a value of json token.
type Token struct {
	Kind Kind
	Data []byte
	Err  error
}

// Error tells whether token is an error.
func (t Token) Error() bool {
	return t.Kind == Error
}

// UnexpectedEOF tells whether token is an unexpectedEOF error.
func (t Token) UnexpectedEOF() bool {
	return t.Kind == Error && IsUnexpectedEOF(t.Err)
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
		return &UnmarshalError{context, "object"}
	}
}

func (t Token) Arr(context string) error {
	switch t.Kind {
	case Error:
		return t.Err
	case ArrBegin:
		return nil
	default:
		return &UnmarshalError{context, "array"}
	}
}

func (t Token) String(context string) (string, error) {
	switch t.Kind {
	case Error:
		return "", t.Err
	case Null:
		return "", nil
	case Str:
		s, _ := unquoteBytes(t.Data)
		return string(s), nil
	default:
		return "", &UnmarshalError{context, "string"}
	}
}

func (t Token) Number(context string) (Number, error) {
	switch t.Kind {
	case Error:
		return "", t.Err
	case Null:
		return Number("0"), nil
	case Num:
		return Number(t.Data), nil
	default:
		return "", &UnmarshalError{context, "number"}
	}
}

// Float64 returns the number as a float64.
func (t Token) Float64(context string) (float64, error) {
	switch t.Kind {
	case Error:
		return 0, t.Err
	case Null:
		return 0, nil
	case Num:
		return strconv.ParseFloat(string(t.Data), 64)
	default:
		return 0, &UnmarshalError{context, "number"}
	}
}

// Int64 returns the number as an int64.
func (t Token) Int64(context string) (int64, error) {
	switch t.Kind {
	case Error:
		return 0, t.Err
	case Null:
		return 0, nil
	case Num:
		i, err := strconv.ParseInt(string(t.Data), 10, 64)
		if err != nil {
			return 0, &UnmarshalError{context, "integer, got number " + string(t.Data)}
		}
		return i, nil
	default:
		return 0, &UnmarshalError{context, "integer"}
	}
}

// Int returns the number as an int.
func (t Token) Int(context string) (int, error) {
	i, err := t.Int64(context)
	return int(i), err
}

// Bool returns the boolean value. For Null token it returns false.
// If token is not boolean/null, returns *UnmarshalError with context provided.
func (t Token) Bool(context string) (bool, error) {
	switch t.Kind {
	case Error:
		return false, t.Err
	case Null:
		return false, nil
	case Bool:
		return t.Data[0] == 't', nil
	default:
		return false, &UnmarshalError{context, "boolean"}
	}
}

// number ---

// A Number represents a JSON number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

// Int returns the number as an int.
func (n Number) Int() (int, error) {
	i, err := n.Int64()
	return int(i), err
}
