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
	gojson "encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/json"
)

func TestDecoder(t *testing.T) {
	tests := []struct {
		name string
		doc  string
	}{
		{"empty", ``},
		{"empty_obj", `{}`},
		{"empty_arr", `[]`},
		{"null", `null`},
		{"string_empty", `""`},
		{"string", `"this is message"`},
		{"string_with_escapes", `"message\"\\\/\b\f\n\r\t"`},
		{"string_with_invalid_escape", `"\x"`},
		{"string_with_hex", `"1\u00357"`},
		{"string_with_tab", `"12	34"`},
		{"string_multiline", `"line1
		line2"`},
		{"string_partial", `"this is message`},
		{"string_partial_escape1", `"this is message\"`},
		{"string_partial_escape2", `"this is message\`},
		{"string_partial_hex1", `"this is message\u00`},
		{"string_partial_hex2", `"this is message\u00"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			de := json.NewDecoder([]byte(tt.doc))
			gode := gojson.NewDecoder(strings.NewReader(tt.doc))
			gode.UseNumber()
			for {
				gotok, goerr := gode.Token()
				tok := de.Token()
				//t.Logf("%s %q %v", tok.Kind, tok.Data, tok.Err)
				switch {
				case tok.Error():
					if goerr == nil {
						t.Fatal()
					}
					t.Logf("\n  err %v\ngoerr %v", tok.Err, goerr)
					if goerr == io.EOF || goerr == io.ErrUnexpectedEOF {
						if !json.IsUnexpectedEOF(tok.Err) {
							t.Fatal()
						}
						return
					}
					got := tok.Err.(*json.SyntaxError).Offset
					want := goerr.(*gojson.SyntaxError).Offset
					diff := got - want
					if diff < -1 || diff > 1 {
						t.Fatalf("error offset: got %d want %d", got, want)
					}
					return
				case tok.EOD():
					if goerr != io.EOF {
						t.Fatal()
					}
					return
				case tok.Obj():
					if gotok != gojson.Delim('{') {
						t.Fatal(gotok)
					}
				case tok.Kind == json.ObjEnd:
					if gotok != gojson.Delim('}') {
						t.Fatal()
					}
				case tok.Arr():
					if gotok != gojson.Delim('[') {
						t.Fatal()
					}
				case tok.Kind == json.ArrEnd:
					if gotok != gojson.Delim(']') {
						t.Fatal()
					}
				case tok.Kind == json.String:
					if s, ok := gotok.(string); !ok || !tok.Eq(s) {
						tok.Eq(s)
						t.Fatal()
					}
				case tok.Number():
					if s, ok := gotok.(string); !ok || s != string(tok.Data) {
						t.Fatal()
					}
				case tok.Kind == json.Boolean:
					b, ok := tok.Bool()
					if !ok {
						t.Fatal()
					}
					if b1, ok := gotok.(bool); !ok || b != b1 {
						t.Fatal()
					}
				case tok.Null():
					if gotok != nil {
						t.Fatal()
					}
				default:
					t.Fatalf("unexpected token: %s %q %v", tok.Kind, tok.Data, tok.Err)
				}
			}
		})
	}
}
