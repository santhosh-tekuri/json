package json_test

import (
	gojson "encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/json"
)

type singleByteReader struct {
	r io.Reader
}

func (r singleByteReader) Read(b []byte) (int, error) {
	return r.r.Read(b[:1])
}

func TestReader_Decode(t *testing.T) {
	tests := []struct {
		name string
		doc  string
	}{
		{"numbers", "0123  456"},
		{"strings", `"one""two"  "three"  "four"`},
		{"mixed", `{}   123  "one" truefalse"two" []`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := singleByteReader{strings.NewReader(tt.doc)}
			de := json.NewReadDecoder(r)
			gode := gojson.NewDecoder(strings.NewReader(tt.doc))
			gode.UseNumber()
			for {
				tok := de.Token()
				t.Logf("%s `%s` %v", tok.Kind, tok.Data, tok.Err)
				if tok.EOD() {
					tok = de.Token()
					t.Logf("%s `%s` %v", tok.Kind, tok.Data, tok.Err)
				}
				gotok, goerr := gode.Token()
				switch tok.Kind {
				case json.Error:
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
				case json.EOF:
					if goerr != io.EOF {
						t.Fatal()
					}
					return
				case json.ObjBegin:
					if gotok != gojson.Delim('{') {
						t.Fatal(gotok)
					}
				case json.ObjEnd:
					if gotok != gojson.Delim('}') {
						t.Fatal()
					}
				case json.ArrBegin:
					if gotok != gojson.Delim('[') {
						t.Fatal()
					}
				case json.ArrEnd:
					if gotok != gojson.Delim(']') {
						t.Fatal()
					}
				case json.String:
					if s, ok := gotok.(string); !ok || !tok.Eq(s) {
						tok.Eq(s)
						t.Fatal()
					}
				case json.Number:
					if s, ok := gotok.(gojson.Number); !ok || string(s) != string(tok.Data) {
						t.Fatalf("number: got %q want %q", string(tok.Data), s)
					}
				case json.Boolean:
					b, _ := tok.Bool("")
					if b1, ok := gotok.(bool); !ok || b != b1 {
						t.Fatal()
					}
				case json.Null:
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

func TestReader_Unmarshal(t *testing.T) {
	ff, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range ff {
		if !strings.HasSuffix(f, "small.json") {
			continue
		}
		t.Run(filepath.Base(f), func(t *testing.T) {
			r, err := os.Open(f)
			if err != nil {
				t.Fatal(err)
			}
			de := json.NewReadDecoder(singleByteReader{r})
			got, err := de.Unmarshal()
			if err != nil {
				t.Fatal(err)
			}
			doc, err := ioutil.ReadFile(f)
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
