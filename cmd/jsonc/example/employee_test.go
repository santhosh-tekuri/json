package example

import (
	gojson "encoding/json"
	"reflect"
	"testing"

	"github.com/santhosh-tekuri/json"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		val  employee
	}{
		{"string_prop", `{"Name":"Alice"}`, employee{}},
		{"bool_prop", `{"Permanent":true}`, employee{}},
		{"float64_prop", `{"Height":5.4}`, employee{}},
		{"int_prop", `{"Weight":5}`, employee{}},
		{"int_prop_float64", `{"Weight":5.4}`, employee{}},
		{"ignore_unexported", `{"first_name":"Alice"}`, employee{}},
		{"override_name", `{"sirName":"Alice"}`, employee{}},
		{"override_ignore", `{"LastName":"Alice"}`, employee{}},
		{"ignore_nullprop", `{"Name": null}`, employee{Name: "Alice"}},
		{"stringarr_prop", `{"NickNames": ["one", "two"]}`, employee{}},
		{"stringarr_nullprop", `{"NickNames": null}`, employee{}},
		{"stringarr_nullitem", `{"NickNames": ["one", null, "three"]}`, employee{}},
		{"obj_prop", `{"Address": {"Street": "HSR"}}`, employee{}},
		{"obj_nullprop", `{"Address": null}`, employee{Address: address{Street: "HSR"}}},
		{"objarr_prop", `{"Addresses": [{"Street": "HSR"}, {"Street": "BEML"}]}`, employee{}},
		{"objarr_nullitem", `{"Addresses": [{"Street": "HSR"}, null, {"Street": "BEML"}]}`, employee{}},
		{"interface_prop", `{"Notes1": [{"Street": "HSR"}, null, {"Street": "BEML"}]}`, employee{}},
		{"interfacearr_prop", `{"Notes2": [{"Street": "HSR"}, null, {"Street": "BEML"}]}`, employee{}},
		{"mapstrinterface_prop", `{"Notes3": {"Street": "HSR", "City": "BEML"}}`, employee{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val
			de := json.NewByteDecoder([]byte(tt.doc))
			gerr := got.Unmarshal(de)
			if gerr != nil {
				t.Log("gerr:", gerr)
			}

			want := tt.val
			werr := gojson.Unmarshal([]byte(tt.doc), &want)
			if werr != nil {
				t.Log("werr:", werr)
			}
			if (gerr != nil) != (werr != nil) {
				t.Fatal("errors did not match")
			}
			if gerr == nil && !reflect.DeepEqual(got, want) {
				t.Log("got:", got)
				t.Log("want:", want)
				t.Fatal("values did not match")
			}
		})
	}
}
