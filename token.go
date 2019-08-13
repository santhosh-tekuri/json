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

type Type int

const (
	EOF Type = iota + 1
	ObjBegin
	ObjEnd
	ArrBegin
	ArrEnd
	String
	Number
	Null
	Boolean
)

type Token struct {
	Type Type
	Data []byte
}

// query token type ---

func (t Token) Begin() bool {
	return t.Type == ObjBegin || t.Type == ArrBegin
}

func (t Token) End() bool {
	return t.Type == ObjEnd || t.Type == ArrEnd
}

func (t Token) Null() bool {
	return t.Type == Null
}

// assert token type ---

func (t Token) Obj() {
	if t.Type != ObjBegin {
		panic("expected object")
	}
}

func (t Token) Arr() {
	if t.Type != ArrBegin {
		panic("expected array")
	}
}

// assert type and extract data ---

func (t Token) Str() string {
	if t.Type != String {
		panic("expected string")
	}
	s, _ := unquoteBytes(t.Data)
	return string(s)
}

// Float64 returns the number as a float64.
func (t Token) Float64() float64 {
	if t.Type != Number {
		panic("expected number")
	}
	f, _ := strconv.ParseFloat(string(t.Data), 64)
	return f
}

// Int64 returns the number as an int64.
func (t Token) Int64() int64 {
	if t.Type != Number {
		panic("expected number")
	}
	i, _ := strconv.ParseInt(string(t.Data), 10, 64)
	return i
}

// Int returns the number as an int.
func (t Token) Int() int {
	if t.Type != Number {
		panic("expected number")
	}
	return int(t.Int64())
}

// Bool returns the boolean value.
func (t Token) Bool() bool {
	if t.Type != Boolean {
		panic("boolean expected")
	}
	return t.Data[0] == 't'
}
