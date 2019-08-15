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
	"io/ioutil"
	"path/filepath"
	"reflect"
	"runtime"
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
		{"object_invalid_comma1", `{,}`},
		{"object_invalid_comma2", `{"key":"value",}`},
		{"object_invalid_comma3", `{"key":"value",,}`},
		{"object_invalid_comma4", `{,"key":"value"}`},
		{"array_invalid_comma1", `[,]`},
		{"array_invalid_comma2", `[1,]`},
		{"array_invalid_comma3", `[1,,]`},
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
		{"number", `123`},
		{"number_negative", `-123`},
		{"number_starring_with_zero", `0123`},
		{"number_valid01", "0"},
		{"number_valid02", "-0"},
		{"number_valid03", "1"},
		{"number_valid04", "-1"},
		{"number_valid05", "0.1"},
		{"number_valid06", "-0.1"},
		{"number_valid07", "1234"},
		{"number_valid08", "-1234"},
		{"number_valid09", "12.34"},
		{"number_valid10", "-12.34"},
		{"number_valid11", "12E0"},
		{"number_valid12", "12E1"},
		{"number_valid13", "12e34"},
		{"number_valid14", "12E-0"},
		{"number_valid15", "12e+1"},
		{"number_valid16", "12e-34"},
		{"number_valid17", "-12E0"},
		{"number_valid18", "-12E1"},
		{"number_valid19", "-12e34"},
		{"number_valid20", "-12E-0"},
		{"number_valid21", "-12e+1"},
		{"number_valid22", "-12e-34"},
		{"number_valid23", "1.2E0"},
		{"number_valid24", "1.2E1"},
		{"number_valid25", "1.2e34"},
		{"number_valid26", "1.2E-0"},
		{"number_valid27", "1.2e+1"},
		{"number_valid28", "1.2e-34"},
		{"number_valid29", "-1.2E0"},
		{"number_valid30", "-1.2E1"},
		{"number_valid31", "-1.2e34"},
		{"number_valid32", "-1.2E-0"},
		{"number_valid33", "-1.2e+1"},
		{"number_valid34", "-1.2e-34"},
		{"number_valid35", "0E0"},
		{"number_valid36", "0E1"},
		{"number_valid37", "0e34"},
		{"number_valid38", "0E-0"},
		{"number_valid39", "0e+1"},
		{"number_valid40", "0e-34"},
		{"number_valid41", "-0E0"},
		{"number_valid42", "-0E1"},
		{"number_valid43", "-0e34"},
		{"number_valid44", "-0E-0"},
		{"number_valid45", "-0e+1"},
		{"number_valid46", "-0e-34"},
		{"number_valid47", "-61657.61667E+61673"},
		{"number_invalid01", `{"n":1.0.1}`},
		{"number_invalid02", `{"n":1..1}`},
		{"number_invalid03", `{"n":-1-2}`},
		{"number_invalid04", `{"n":012a42}`},
		{"number_invalid05", `{"n":01.2}`},
		{"number_invalid06", `{"n":012}`},
		{"number_invalid07", `{"n":12E12.12}`},
		{"number_invalid08", `{"n":1e2e3}`},
		{"number_invalid09", `{"n":1e+-2}`},
		{"number_invalid10", `{"n":1e--23}`},
		{"number_invalid11", `{"n":1e}`},
		{"number_invalid12", `{"n":e1}`},
		{"number_invalid13", `{"n":1e+}`},
		{"number_invalid14", `{"n":1ea}`},
		{"number_invalid15", `{"n":1a}`},
		{"number_invalid16", `{"n":1.a}`},
		{"number_invalid17", `{"n":1.}`},
		{"number_invalid18", `{"n":01}`},
		{"number_invalid19", `{"n":1.e1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			de := json.NewDecoder([]byte(tt.doc))
			gode := gojson.NewDecoder(strings.NewReader(tt.doc))
			gode.UseNumber()
			for {
				tok := de.Token()
				// t.Logf("%s %q %v", tok.Kind, tok.Data, tok.Err)
				if tok.EOD() {
					tok = de.Token()
					// t.Logf("%s %q %v", tok.Kind, tok.Data, tok.Err)
				}
				gotok, goerr := gode.Token()
				switch {
				case tok.Error():
					if goerr == nil {
						t.Fatal(tok.Err)
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
				case tok.EOF():
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
					if s, ok := gotok.(gojson.Number); !ok || string(s) != string(tok.Data) {
						t.Fatalf("number: got %q want %q", string(tok.Data), s)
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

func TestDecoder_Unmarshal(t *testing.T) {
	ff, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range ff {
		t.Run(filepath.Base(f), func(t *testing.T) {
			doc, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			de := json.NewDecoder(doc)
			got, err := de.Unmarshal()
			if err != nil {
				t.Fatal(err)
			}
			var want interface{}
			if err := gojson.Unmarshal(doc, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Error("value does not match")
			}
		})
	}
}

func TestNoAllocs(t *testing.T) {
	ff, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range ff {
		t.Run(filepath.Base(f), func(t *testing.T) {
			doc, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			de := json.NewDecoder(doc)
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			allocs := memStats.Mallocs

			for {
				tok := de.Token()
				if tok.Error() {
					t.Fatal(tok.Err)
				}
				if tok.EOF() {
					break
				}
			}

			runtime.ReadMemStats(&memStats)
			if d := memStats.Mallocs - allocs; d != 0 {
				t.Fatalf("%d allocs detected", d)
			}
		})
	}
}

func BenchmarkDecoder(b *testing.B) {
	ff, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil {
		b.Fatal(err)
	}
	for _, f := range ff {
		doc, err := ioutil.ReadFile(f)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(filepath.Base(f), func(b *testing.B) {
			de := json.NewDecoder(doc)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for {
					t := de.Token()
					if t.Error() {
						b.Fatal(t.Err)
					}
					if t.EOF() {
						break
					}
				}
				de.Reset(doc)
			}
		})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	ff, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil {
		b.Fatal(err)
	}
	var v interface{}
	for _, f := range ff {
		doc, err := ioutil.ReadFile(f)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(filepath.Base(f), func(b *testing.B) {
			b.Run("mine", func(b *testing.B) {
				de := json.NewDecoder(doc)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					v, err = de.Unmarshal()
					if err != nil {
						b.Fatal(err)
					}
					de.Reset(doc)
				}
			})
			b.Run("std", func(b *testing.B) {
				v = nil
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					err = gojson.Unmarshal(doc, &v)
					if err != nil {
						b.Fatal(err)
					}
					v = nil
				}
			})
		})
	}
}
