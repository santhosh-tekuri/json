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

import "strconv"

type Kind byte

const (
	noError Kind = iota
	Error
	EOD
	EOF
	ObjBegin
	ObjEnd
	ArrBegin
	ArrEnd
	String
	Number
	Boolean
	Null
)

var kindNames = []string{
	`<continue>`,
	`<error>`,
	`<eod>`,
	`<eof>`,
	`'{'`, `'}'`, `'['`, `']'`,
	`<string>`, `<number>`, `<boolean>`, `<null>`,
}

func (k Kind) String() string {
	return kindNames[k]
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

func (t Token) Obj() bool {
	return t.Kind == ObjBegin
}

func (t Token) Arr() bool {
	return t.Kind == ArrBegin
}

func (t Token) Str() (string, bool) {
	if t.Kind != String {
		return "", false
	}
	s, _ := unquoteBytes(t.Data)
	return string(s), true
}

// Float64 returns the number as a float64.
func (t Token) Float64() (float64, bool) {
	if t.Kind != Number {
		return 0, false
	}
	f, _ := strconv.ParseFloat(string(t.Data), 64)
	return f, true
}

// Int64 returns the number as an int64.
func (t Token) Int64() (int64, bool) {
	if t.Kind != Number {
		return 0, false
	}
	i, _ := strconv.ParseInt(string(t.Data), 10, 64)
	return i, true
}

// Int returns the number as an int.
func (t Token) Int() (int, bool) {
	i, ok := t.Int64()
	return int(i), ok
}

// Bool returns the boolean value.
func (t Token) Bool() (bool, bool) {
	if t.Kind != Boolean {
		return false, false
	}
	return t.Data[0] == 't', true
}
