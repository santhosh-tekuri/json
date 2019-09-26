package tests

//go:generate jsonc -o structs_json.go stringVal structTag excludeTag unexported arrString ptrString arrPtrString interfaceVal arrInterface structVal arrStruct ptrStruct arrPtrStruct anonStruct arrAnonStruct ptrAnonStruct arrPtrAnonStruct

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

type interfaceVal struct {
	Field interface{}
}

type arrInterface struct {
	Field []interface{}
}

type structVal struct {
	Field stringVal
}

type arrStruct struct {
	Field []stringVal
}

type ptrStruct struct {
	Field *stringVal
}

type arrPtrStruct struct {
	Field []*stringVal
}

type anonStruct struct {
	Field struct {
		Field string
	}
}

type arrAnonStruct struct {
	Field []struct {
		Field string
	}
}

type ptrAnonStruct struct {
	Field *struct {
		Field string
	}
}

type arrPtrAnonStruct struct {
	Field []*struct {
		Field string
	}
}
