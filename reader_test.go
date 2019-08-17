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
