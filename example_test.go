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
	"fmt"

	"github.com/santhosh-tekuri/json"
)

type employee struct {
	name      string
	age       int
	permanent bool
	details   details
	addrs     []address
}

type details struct {
	height float64
	weight float64
}

type address struct {
	street string
	city   string
	state  string
}

func (a *address) Unmarshal(de json.Decoder) error {
	return json.UnmarshalObj("address", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("street"):
			a.street, err = de.Token().Str("address.street")
		case prop.Eq("city"):
			a.city, err = de.Token().Str("address.city")
		case prop.Eq("state"):
			a.state, err = de.Token().Str("address.state")
		default:
			err = de.Skip()
		}
		return
	})
}

func (d *details) Unmarshal(de json.Decoder) error {
	return json.UnmarshalObj("details", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("height"):
			d.height, err = de.Token().Float64("details.height")
		case prop.Eq("weight"):
			d.weight, err = de.Token().Float64("details.weight")
		default:
			err = de.Skip()
		}
		return
	})
}

func (e *employee) Unmarshal(de json.Decoder) error {
	return json.UnmarshalObj("employee", de, func(de json.Decoder, prop json.Token) (err error) {
		switch {
		case prop.Eq("name"):
			e.name, err = de.Token().Str("employee.name")
		case prop.Eq("age"):
			e.age, err = de.Token().Int("employee.age")
		case prop.Eq("permanent"):
			e.permanent, err = de.Token().Bool("employee.permanent")
		case prop.Eq("details"):
			return e.details.Unmarshal(de)
		case prop.Eq("addresses"):
			err = json.UnmarshalArr("employee.addresses", de, func(de json.Decoder) error {
				addr := address{}
				if err := addr.Unmarshal(de); err != nil {
					return err
				}
				e.addrs = append(e.addrs, addr)
				return nil
			})
		default:
			err = de.Skip()
		}
		return
	})
}

func ExampleUnmarshal() {
	doc := `{
		"name": "Santhosh Kumar Tekuri",
		"age": 30,
		"junk1": "junk",
		"junk2": 0,
		"junk3": true,
		"junk4": null,
		"junk5": {"k1": "v1", "k2": 0},
		"junk6": ["junk", 1, true, null, ["junk"], {"k":"v"}],
		"permanent": true,
		"addresses": [
			{
				"street": "HSR Layout",
				"city": "Bangalore",
				"state": "Karnataka"
			}
		],
		"details": {
			"height": 100,
			"weight": 200
		}
	}`
	de := json.NewByteDecoder([]byte(doc))
	emp := employee{}
	if err := emp.Unmarshal(de); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", emp)
	// Output:
	// json_test.employee{name:"Santhosh Kumar Tekuri", age:30, permanent:true, details:json_test.details{height:100, weight:200}, addrs:[]json_test.address{json_test.address{street:"HSR Layout", city:"Bangalore", state:"Karnataka"}}}
}
