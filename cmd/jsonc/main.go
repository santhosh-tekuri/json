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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	pkg     *packages.Package
	buf     = new(bytes.Buffer)
	imports = make(map[string]string)
)

func usage() {
	errln(`usage: jsonc [-tags 'tag list'] type ...

-tags 'tag list'
    a space-separated list of build tags to consider for finding files.`)
}

func newPackage(path, name string) *types.Package {
	return types.NewPackage(path, name)
}

func main() {
	output := flag.String("o", "", "output file")
	tags := flag.String("tags", "", "space-separated list of build tags")
	flag.Parse()
	flag.Usage = usage
	types := flag.Args()
	if len(types) == 0 {
		errln("no types specified")
		errln()
		errln()
		usage()
		os.Exit(1)
	}

	cfg := &packages.Config{
		Mode:       packages.LoadSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", *tags)},
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		errln(err)
		os.Exit(1)
	}
	if len(pkgs) != 1 {
		errln(len(pkgs), "packages found")
		os.Exit(1)
	}
	pkg = pkgs[0]

	if *tags != "" {
		printf("// +build %s\n\n", *tags)
	}
	printf("// Code generated by jsonc; DO NOT EDIT.\n\n")
	println("package ", pkg.Name)
	getPackageIdentifier(newPackage("github.com/santhosh-tekuri/json", "json"))
	println(`import "github.com/santhosh-tekuri/json"`)

	for _, typ := range types {
		s := findStruct(typ)
		if s == nil {
			errln("type", typ, "not found")
			errln()
			os.Exit(1)
		}
		println()
		generate(s, typ)
	}

	// fix imports if necessary
	if len(imports) > 1 {
		src := buf.String()
		imp := `import "github.com/santhosh-tekuri/json"`
		i := strings.Index(src, imp)
		buf.Reset()
		buf.WriteString(src[0:i])
		println("import(")
		for path, name := range imports {
			if path == name || strings.HasSuffix(path, "/"+name) {
				printf("%q\n", path)
			} else {
				printf("%s %q\n", name, path)
			}
		}
		println(")")
		buf.WriteString(src[i+len(imp):])
	}

	b, err := format.Source(buf.Bytes())
	if err != nil {
		errln(buf)
		errln("COULD NOT GOFMT")
		os.Exit(1)
	}
	if *output != "" {
		if err := ioutil.WriteFile(*output, b, 0666); err != nil {
			errln(err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("%s", b)
	}
	if todo > 0 {
		errln("CHECK TODOS")
		os.Exit(1)
	}
}

func generate(s *types.Struct, sname string) {
	r := strings.ToLower(sname[:1])
	r = newVar(r)
	printf("func (%s *%s) DecodeJSON(de json.Decoder) error {\n", r, sname)
	printf(`return `)
	unmarshalStruct(s, r, sname)
	println(`}`)
	releaseVar(r)
}

func unmarshalStruct(s *types.Struct, lhs, context string) {
	printf(`json.DecodeObj("%s", de, func(de json.Decoder, prop json.Token) (err error) {
            switch {
    `, context)

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		if !field.Exported() {
			continue
		}
		fname := field.Name()
		prop := fname
		if s.Tag(i) != "" {
			tag := reflect.StructTag(s.Tag(i)).Get("json")
			opts := strings.Split(tag, ",")
			if len(opts) > 0 && opts[0] != "" {
				prop = opts[0]
				if prop == "-" {
					continue
				}
			}
		}
		printf(`case prop.Eq("%s"):`, prop)
		lhs := lhs + "." + fname
		context := context + "." + fname
		unmarshal(true, lhs, "=", context, field.Type())
	}
	println(`
        default:
            err = de.Skip()
        }
        return
    });`)
}

func unmarshal(checkNull bool, lhs, equals, context string, t types.Type) {
	star := false
	if s, ok := t.(*types.Pointer); ok {
		star = true
		t = s.Elem()
		checkNull = true
	}

	if types.Implements(types.NewPointer(t), jsonUnmarshaller) {
		if star {
			if equals == ":=" {
				printf(`var %s *%s;`, lhs, type2String(t))
			} else {
				printf("%s = nil;", lhs)
			}
			println("if !de.Peek().Null() {")
			printf(`%s = &%s{};`, lhs, type2String(t))
			println("}")
		} else {
			if equals == ":=" {
				printf(`%s := %s{};`, lhs, type2String(t))
			}
		}
		b := newVar("b")
		printf("var %s []byte;", b)
		printf(`%s, err %s de.Marshal();`, b, equals)
		printf("if err==nil {")
		printf(`err = %s.UnmarshalJSON(%s);`, lhs, b)
		printf("};")
		releaseVar(b)
		return
	}

	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.String, types.Float64, types.Bool, types.Int, types.Int64:
			if star {
				if equals == ":=" {
					printf("var %s *%s;", lhs, type2String(t))
					printf("var err error;")
				} else {
					printf("%s = nil;", lhs)
				}
			}
			val := "de.Token()"
			if checkNull {
				val = "val"
				printf(`if val:=de.Token(); !val.Null() {`)
			}
			method := strings.ToUpper(t.Name()[:1]) + t.Name()[1:]
			if star {
				printf("var pval %s;", type2String(t))
				printf(`%s, err %s %s.%s("%s");`, "pval", "=", val, method, context)
				printf("%s = &pval;", lhs)
			} else {
				printf(`%s, err %s %s.%s("%s");`, lhs, equals, val, method, context)
			}
			if checkNull {
				printf("};")
			}
		default:
			printf("\n//%s %s %#v\n", lhs, t.String(), t)
			notImplemented()
		}
	case *types.Named:
		if star {
			if equals == ":=" {
				printf(`var %s *%s;`, lhs, type2String(t))
			} else {
				printf("%s = nil;", lhs)
			}
			println("if !de.Peek().Null() {")
			printf(`%s = &%s{};`, lhs, type2String(t))
			println("}")
		} else {
			if equals == ":=" {
				printf(`%s := %s{};`, lhs, type2String(t))
			}
		}
		printf(`err %s %s.DecodeJSON(de);`, equals, lhs)
	case *types.Struct:
		if star {
			if equals == ":=" {
				printf(`var %s *%s;`, lhs, type2String(t))
			} else {
				printf("%s = nil;", lhs)
			}
			println("if !de.Peek().Null() {")
			printf(`%s = &%s{};`, lhs, type2String(t))
			println("}")
		} else {
			if equals == ":=" {
				printf(`%s := %s{};`, lhs, type2String(t))
			}
		}
		printf("err %s", equals)
		unmarshalStruct(t, lhs, context)
	case *types.Slice:
		if equals == ":=" {
			printf(`var %s []%s;`, lhs, type2String(t.Elem()))
			equals = "="
		} else {
			println("if de.Peek().Null() {")
			printf(`%s = nil;`, lhs)
			println("} else {")
			printf(`%s = []%s{};`, lhs, type2String(t.Elem()))
			println("}")
		}
		printf(`err %s json.DecodeArr("%s", de, func(de json.Decoder) error {`, equals, context)
		println()
		item := newVar("item")
		unmarshal(false, item, ":=", context+"[]", t.Elem())
		printf(`%s = append(%s, %s);`, lhs, lhs, item)
		printf("return err;")
		printf("});")
		releaseVar(item)
	case *types.Map:
		ktype := type2String(t.Key())
		if ktype != "string" {
			printf("\n// map with non-string key not implemented\n")
			notImplemented()
			return
		}
		if equals == ":=" {
			printf(`var %s map[string]%s;`, lhs, type2String(t.Elem()))
			equals = "="
		} else {
			println("if de.Peek().Null() {")
			printf(`%s = nil;`, lhs)
			printf("} else if %s == nil {", lhs)
			printf(`%s = map[string]%s{};`, lhs, type2String(t.Elem()))
			println("}")
		}
		printf(`err %s json.DecodeObj("%s", de, func(de json.Decoder, prop json.Token) (err error) {`, equals, context)
		println()
		printf(`k, _ := prop.String("");`)
		v := newVar("v")
		unmarshal(false, v, ":=", context+"{}", t.Elem())
		printf(`%s[k] = %s;`, lhs, v)
		printf("return err;")
		printf("});")
		releaseVar(v)
	case *types.Interface:
		// todo check empty interface
		printf(`%s, err %s de.Decode();`, lhs, equals)
	default:
		printf("\n//%s %s %#v\n", lhs, t.String(), t)
		notImplemented()
	}
}

var scope = make(map[string]struct{})

func newVar(name string) string {
	if _, ok := scope[name]; ok {
		i := 1
		for {
			n := fmt.Sprintf("%s%d", name, i)
			if _, ok := scope[n]; !ok {
				name = n
				break
			}
			i++
		}
	}
	scope[name] = struct{}{}
	return name
}

func releaseVar(name string) {
	delete(scope, name)
}

// helpers ---

func errln(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func println(a ...interface{}) {
	_, _ = fmt.Fprintln(buf, a...)
}

func printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(buf, format, a...)
}

var todo int

func notImplemented() {
	todo++
	printf(`panic("TODO: NOT IMPLEMENTED YET");`)
}

func findStruct(name string) *types.Struct {
	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		return nil
	}
	if s, ok := obj.Type().Underlying().(*types.Struct); ok {
		return s
	}
	return nil
}

func getPackageIdentifier(p *types.Package) string {
	if pkg.PkgPath == p.Path() {
		return ""
	}
	if id, ok := imports[p.Path()]; ok {
		return id
	}
	name := newVar(p.Name())
	imports[p.Path()] = name
	return name
}

func type2String(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		return t.Name()
	case *types.Slice:
		return "[]" + type2String(t.Elem())
	case *types.Pointer:
		return "*" + type2String(t.Elem())
	case *types.Named:
		return types.TypeString(t, getPackageIdentifier)
	case *types.Interface:
		return t.String()
	case *types.Struct:
		var buf = strings.Builder{}
		buf.WriteString("struct {\n")
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)
			buf.WriteString(f.Name())
			buf.WriteString(" ")
			buf.WriteString(type2String(f.Type()))
			if t.Tag(i) != "" {
				buf.WriteString(" ")
				buf.WriteString(t.Tag(i))
			}
			buf.WriteString(";")
		}
		buf.WriteString("}")
		return buf.String()
	default:
		panic(fmt.Sprintf("type2String(%T)", t))
	}
}

var jsonUnmarshaller = getIface()

func getIface() *types.Interface {
	const file = `package json
type Unmarshaler interface {
    UnmarshalJSON([]byte) error
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "iface.go", file, 0)
	if err != nil {
		panic(err)
	}

	config := &types.Config{
		Error: func(e error) {
			fmt.Println(e)
		},
		Importer: importer.Default(),
	}

	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	pkg, e := config.Check("genval", fset, []*ast.File{f}, &info)
	if e != nil {
		fmt.Println(e)
	}

	return pkg.Scope().Lookup("Unmarshaler").Type().Underlying().(*types.Interface)
}
