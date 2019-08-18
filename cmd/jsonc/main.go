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
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

var pkg *packages.Package
var buf = new(bytes.Buffer)

func usage() {
	fmt.Fprintln(os.Stderr, `usage: jsonc [-tags 'tag list'] type ...

-tags 'tag list'
	a space-separated list of build tags to consider for finding files.`)
}

func main() {
	tags := flag.String("tags", "", "space-separated list of build tags")
	flag.Parse()
	flag.Usage = usage
	types := flag.Args()
	if len(types) == 0 {
		fmt.Fprintf(os.Stderr, "no types specified\n\n")
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(pkgs) != 1 {
		fmt.Fprintf(os.Stderr, "%d packages found", len(pkgs))
		os.Exit(1)
	}
	pkg = pkgs[0]

	if *tags != "" {
		printf("// +build %s\n\n", *tags)
	}
	printf("// Code generated by jsonc; DO NOT EDIT.\n\n")
	println("package ", pkg.Name)
	println(`import "github.com/santhosh-tekuri/json"`)

	for _, typ := range types {
		s := findStruct(typ)
		if s == nil {
			fmt.Fprintf(os.Stderr, "type %s not found\n", typ)
			os.Exit(1)
		}
		generate(s, typ)
	}
	b, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, buf)
		fmt.Fprintln(os.Stderr, "COULD NOT GOFMT")
		os.Exit(1)
	}
	fmt.Printf("%s", b)
	if todo > 0 {
		fmt.Fprintln(os.Stderr, "CHECK TODOS")
	}
}

func generate(s *ast.StructType, sname string) {
	r := strings.ToLower(sname[:1])
	printf(`
	func (%s *%s) Unmarshal(de json.Decoder) error {
		return json.UnmarshalObj("%s", de, func(de json.Decoder, prop json.Token) (err error) {
			switch {
	`, r, sname, sname)

	for _, field := range s.Fields.List {
		fname := field.Names[0].Name
		if !ast.IsExported(fname) {
			continue
		}
		prop := fname
		if field.Tag != nil {
			tag, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "struct tag %s.%s: %s", sname, fname, err)
				os.Exit(1)
			}
			tag = reflect.StructTag(tag).Get("json")
			opts := strings.Split(tag, ",")
			if len(opts) > 0 && opts[0] != "" {
				prop = opts[0]
				if prop == "-" {
					continue
				}
			}
		}
		printf(`case prop.Eq("%s"):`, prop)
		rfield := r + "." + fname
		context := sname + "." + fname
		switch t := field.Type.(type) {
		case *ast.Ident:
			switch t.Name {
			case "string", "float64", "bool", "int", "int64":
				method := strings.ToUpper(t.Name[:1]) + t.Name[1:]
				printf(`%s, err = de.Token().%s("%s");`, rfield, method, context)
			default:
				printf(`err = %s.Unmarshal(de);`, rfield)
			}
		case *ast.ArrayType:
			printf(`err = json.UnmarshalArr("%s", de, func(de json.Decoder) error {`, context)
			switch t := t.Elt.(type) {
			case *ast.Ident:
				switch t.Name {
				case "string", "float64", "bool", "int", "int64":
					method := strings.ToUpper(t.Name[:1]) + t.Name[1:]
					printf(`item, err := de.Token().%s("%s[]");`, method, context)
					printf(`%s = append(%s, item);`, rfield, rfield)
					println("return err;")
				default:
					printf(`item := %s{};`, t.Name)
					printf(`err := item.Unmarshal(de);`)
					printf(`%s = append(%s, item);`, rfield, rfield)
					println("return err;")
				}
			default:
				printf("\n//%s %T\n", fname, t)
				notImplemented()
			}
			println("})")
		default:
			printf("\n//%s %T\n", fname, t)
			notImplemented()
		}

	}
	println(`
		default:
			err = de.Skip()
		}
		return
	})
	}`)
}

// helpers ---

func println(a ...interface{}) {
	fmt.Fprintln(buf, a...)
}

func printf(format string, a ...interface{}) {
	fmt.Fprintf(buf, format, a...)
}

var todo int

func notImplemented() {
	todo++
	printf(`panic("TODO: NOT IMPLEMENTED YET");`)
}

func findStruct(name string) *ast.StructType {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.GenDecl); ok {
				if decl.Tok == token.TYPE {
					ts := decl.Specs[0].(*ast.TypeSpec)
					if s, ok := ts.Type.(*ast.StructType); ok && ts.Name.Name == name {
						return s
					}
				}
			}
		}
	}
	return nil
}
