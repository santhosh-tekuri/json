package tests

//go:generate jsonc -o structs_json.go stringVal structTag excludeTag unexported arrString ptrString arrPtrString

type stringVal struct {
	Field string
}

type structTag struct {
	Field string `json:"Name"`
}

type excludeTag struct {
	Field string `json:"-"`
}

type unexported struct {
	field string
}

type arrString struct {
	Field []string
}

type ptrString struct {
	Field *string
}

type arrPtrString struct {
	Field []*string
}
