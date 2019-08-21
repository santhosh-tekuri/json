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

package json_test

import (
	"bytes"
	gojson "encoding/json"
	"reflect"
	"testing"

	"github.com/santhosh-tekuri/json"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
	}{
		{"null", nil},
		{"nilArray", []interface{}(nil)},
		{"emptyArray", []interface{}{}},
		{"array", []interface{}{nil, 0, true, "hello", .23}},
		{"nilMap", map[string]interface{}(nil)},
		{"emptyMap", map[string]interface{}{}},
		{"singleKeyMap", map[string]interface{}{"key": "value"}},
		{"string", "hello world"},
		{"string1", "hello \n\r\t\\\"\bworld"},
		{"true", true},
		{"false", false},
		{"int", 100},
		{"int8", int8(100)},
		{"int16", int16(100)},
		{"int32", int32(100)},
		{"int64", int64(100)},
		{"uint", uint(100)},
		{"uint8", uint8(100)},
		{"uint16", uint16(100)},
		{"uint32", uint32(100)},
		{"uint64", uint64(100)},
		{"float32", float32(1.234)},
		{"float64", 1.234},
		{"raw", gojson.RawMessage([]byte(`{"key":"value"}`))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.val)
			if err != nil {
				t.Fatal(err)
			}
			want, err := gojson.Marshal(tt.val)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Log(" got", string(got))
				t.Log("want", string(want))
				t.Fatal("got!=want")
			}
		})
	}

	t.Run("map", func(t *testing.T) {
		want := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		b, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.NewByteDecoder(b).Unmarshal()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Log(" got", got)
			t.Log("want", want)
			t.Fatal("got!=want")
		}
	})
}
