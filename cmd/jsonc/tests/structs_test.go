package tests

import (
	gojson "encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/santhosh-tekuri/json"
)

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		val  json.ValueDecoder
	}{
		{"stringVal_1", `{"Field":"hello"}`, &stringVal{}},
		{"stringVal_2", `{"Field": null}`, &stringVal{Field: "hello"}},
		{"stringVal_3", `{}`, &stringVal{Field: "hello"}},
		{"structTag_1", `{"Name":"hello"}`, &structTag{}},
		{"structTag_2", `{"Name": null}`, &structTag{Field: "hello"}},
		{"structTag_3", `{}`, &structTag{Field: "hello"}},
		{"structTag_4", `{"Field":"hello"}`, &structTag{}},
		{"excludeTag_1", `{"Field":"hello"}`, &excludeTag{}},
		{"excludeTag_2", `{"Field": null}`, &excludeTag{Field: "hello"}},
		{"excludeTag_3", `{}`, &excludeTag{Field: "hello"}},
		{"unexported", `{"field": "hello"}`, &unexported{}},
		{"arrString_1", `{"Field":["hello", "world"]}`, &arrString{}},
		{"arrString_2", `{"Field":["hello", "world"]}`, &arrString{Field: []string{"one"}}},
		{"arrString_3", `{"Field": null}`, &arrString{Field: []string{"hello"}}},
		{"arrString_4", `{"Field": null}`, &arrString{}},
		{"arrString_5", `{"Field": []}`, &arrString{Field: []string{"hello"}}},
		{"arrString_6", `{"Field": []}`, &arrString{}},
		{"arrString_7", `{}`, &arrString{Field: []string{"hello"}}},
		{"arrString_8", `{}`, &arrString{}},
		{"ptrString_1", `{"Field":"hello"}`, &ptrString{}},
		{"ptrString_2", `{"Field": null}`, &ptrString{Field: addrString("hello")}},
		{"ptrString_3", `{}`, &ptrString{Field: addrString("hello")}},
		{"arrPtrString_1", `{"Field":["hello", null, "world"]}`, &arrPtrString{}},
		{"arrPtrString_2", `{"Field": null}`, &arrPtrString{Field: []*string{addrString("hello")}}},
		{"arrPtrString_3", `{}`, &arrPtrString{Field: []*string{addrString("hello")}}},
		{"arrPtrString_4", `{"Field": ["hello", null, "world"]}`, &arrPtrString{Field: []*string{addrString("one")}}},
		{"interfaceVal_1", `{"Field":["hello"]}`, &interfaceVal{}},
		{"interfaceVal_2", `{"Field":null}`, &interfaceVal{Field: "hello"}},
		{"interfaceVal_3", `{}`, &interfaceVal{Field: "hello"}},
		{"arrInterface_1", `{"Field": [{"Street": "HSR"}, null, {"Street": "BEML"}]}`, &arrInterface{}},
		{"arrInterface_2", `{"Field": null}`, &arrInterface{Field: []interface{}{"hello", nil, "world"}}},
		{"arrInterface_3", `{}`, &arrInterface{Field: []interface{}{"hello", nil, "world"}}},
		{"arrInterface_4", `{"Field": [{"Street": "HSR"}, null, {"Street": "BEML"}]}`, &arrInterface{Field: []interface{}{"hello", nil, "world"}}},
		{"structVal_1", `{"Field":{"Field":"hello"}}`, &structVal{}},
		{"structVal_2", `{"Field":null}`, &structVal{Field: stringVal{Field: "hello"}}},
		{"structVal_3", `{}`, &structVal{Field: stringVal{Field: "hello"}}},
		{"arrStruct_1", `{"Field":[{"Field":"hello"}, null, {"Field": "world"}]}`, &arrStruct{}},
		{"arrStruct_2", `{"Field":null}`, &arrStruct{Field: []stringVal{{Field: "one"}}}},
		{"arrStruct_3", `{}`, &arrStruct{Field: []stringVal{{Field: "one"}}}},
		{"arrStruct_4", `{"Field":[{"Field":"hello"}, null, {"Field":"world"}]}`, &arrStruct{Field: []stringVal{{Field: "one"}}}},
		{"ptrStruct_1", `{"Field":{"Field":"hello"}}`, &ptrStruct{}},
		{"ptrStruct_2", `{"Field":null}`, &ptrStruct{Field: &stringVal{Field: "one"}}},
		{"ptrStruct_3", `{}`, &ptrStruct{Field: &stringVal{Field: "one"}}},
		{"arrPtrStruct_1", `{"Field":[{"Field":"hello"}, null,{"Field":"world"}]}`, &arrPtrStruct{}},
		{"arrPtrStruct_2", `{"Field": null}`, &arrPtrStruct{Field: []*stringVal{&stringVal{Field: "hello"}}}},
		{"arrPtrStruct_3", `{}`, &arrPtrStruct{Field: []*stringVal{&stringVal{Field: "hello"}}}},
		{"arrPtrStruct_4", `{"Field":[{"Field":"hello"}, null,{"Field":"world"}]}`, &arrPtrStruct{Field: []*stringVal{&stringVal{Field: "hello"}}}},
		{"anonStructVal_1", `{"Field":{"Field":"hello"}}`, &anonStruct{}},
		{"anonStructVal_2", `{"Field":null}`, &anonStruct{Field: struct{ Field string }{Field: "hello"}}},
		{"anonStructVal_3", `{}`, &anonStruct{Field: struct{ Field string }{Field: "hello"}}},
		{"arrAnonStruct_1", `{"Field":[{"Field":"hello"}, null, {"Field": "world"}]}`, &arrAnonStruct{}},
		{"arrAnonStruct_2", `{"Field":null}`, &arrAnonStruct{Field: []struct{ Field string }{{Field: "one"}}}},
		{"arrAnonStruct_3", `{}`, &arrAnonStruct{Field: []struct{ Field string }{{Field: "one"}}}},
		{"arrAnonStruct_4", `{"Field":[{"Field":"hello"}, null, {"Field":"world"}]}`, &arrAnonStruct{Field: []struct{ Field string }{{Field: "one"}}}},
		{"ptrAnonStruct_1", `{"Field":{"Field":"hello"}}`, &ptrAnonStruct{}},
		{"ptrAnonStruct_2", `{"Field":null}`, &ptrAnonStruct{Field: &struct{ Field string }{Field: "one"}}},
		{"ptrAnonStruct_3", `{}`, &ptrAnonStruct{Field: &struct{ Field string }{Field: "one"}}},
		{"arrPtrAnonStruct_1", `{"Field":[{"Field":"hello"}, null,{"Field":"world"}]}`, &arrPtrAnonStruct{}},
		{"arrPtrAnonStruct_2", `{"Field": null}`, &arrPtrAnonStruct{Field: []*struct{ Field string }{{Field: "hello"}}}},
		{"arrPtrAnonStruct_3", `{}`, &arrPtrAnonStruct{Field: []*struct{ Field string }{{Field: "hello"}}}},
		{"arrPtrAnonStruct_4", `{"Field":[{"Field":"hello"}, null,{"Field":"world"}]}`, &arrPtrAnonStruct{Field: []*struct{ Field string }{{Field: "hello"}}}},
		{"mapString_1", `{"Field":{"k1": "v1"}}`, &mapString{}},
		{"mapString_2", `{"Field":{"k1": "v1"}}`, &mapString{Field: map[string]string{"k2": "v2"}}},
		{"mapString_3", `{"Field":null}`, &mapString{Field: map[string]string{"k2": "v2"}}},
		{"mapString_4", `{"Field":null}`, &mapString{}},
		{"timeVal_1", `{"Field":"2018-09-22T12:42:31Z"}`, &timeVal{}},
		{"timeVal_2", `{"Field":null}`, &timeVal{Field: time.Now()}},
		{"timeVal_3", `{}`, &timeVal{Field: time.Now()}},
		{"ptrTime_1", `{"Field":"2018-09-22T12:42:31Z"}`, &ptrTime{}},
		{"ptrTime_2", `{"Field": null}`, &ptrTime{Field: addrTime(time.Now())}},
		{"ptrTime_3", `{}`, &ptrTime{Field: addrTime(time.Now())}},
	}
	for _, tt := range tests {
		f := func(t *testing.T, de json.Decoder) {
			got := deepcopy.Copy(tt.val).(json.ValueDecoder)
			gerr := got.DecodeJSON(de)
			if gerr != nil {
				t.Log("gerr:", gerr)
			}
			want := deepcopy.Copy(tt.val)
			werr := gojson.Unmarshal([]byte(tt.doc), &want)
			if werr != nil {
				t.Log("werr:", werr)
			}
			if (gerr != nil) != (werr != nil) {
				t.Fatal("errors did not match")
			}
			if gerr == nil && !reflect.DeepEqual(got, want) {
				t.Logf("got: %#v\n", got)
				t.Logf("want: %#v\n", want)
				t.Fatal("values did not match")
			}
		}
		t.Run("bytes_"+tt.name, func(t *testing.T) {
			f(t, json.NewByteDecoder([]byte(tt.doc)))
		})
		t.Run("reader_"+tt.name, func(t *testing.T) {
			f(t, json.NewReadDecoder(strings.NewReader(tt.doc)))
		})
	}
}

func addrString(s string) *string {
	return &s
}

func addrTime(t time.Time) *time.Time {
	return &t
}
