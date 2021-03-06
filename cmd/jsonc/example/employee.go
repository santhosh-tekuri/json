package example

import "encoding/json"

//go:generate jsonc -o employee_json.go employee address

type employee struct {
	Name       string
	Sirname    string `json:"sirName"`
	firstName  string
	LastName   string `json:"-"`
	ShortName  *string
	ShortNames []*string
	Permanent  bool
	Height     float64
	Weight     int
	NickNames  []string
	Address    address
	Addresses  []address
	//Addr       *address
	Notes1     interface{}
	Notes2     []interface{}
	Notes3     map[string]interface{}
	Contacts   map[string][]string
	Raw        json.RawMessage
	Department struct {
		Name    string `json:"name"`
		Manager string
	}
}

type address struct {
	Street string
}
