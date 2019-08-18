package example

//go:generate jsonc -o employee_json.go employee address

type employee struct {
	Name      string
	Sirname   string `json:"sirName"`
	firstName string
	LastName  string `json:"-"`
	Permanent bool
	Height    float64
	Weight    int
	NickNames []string
	Address   address
	Addresses []address
}

type address struct {
	Street string
}
