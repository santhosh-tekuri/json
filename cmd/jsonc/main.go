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
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

var pkg *packages.Package
var buf = new(bytes.Buffer)

func usage() {
	errln(`usage: jsonc [-tags 'tag list'] type ...

-tags 'tag list'
	a space-separated list of build tags to consider for finding files.`)
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

func generate(s *ast.StructType, sname string) {
	r := strings.ToLower(sname[:1])
	printf("func (%s *%s) DecodeJSON(de json.Decoder) error {\n", r, sname)
	printf(`return `)
	unmarshalStruct(s, r, sname)
	println(`}`)
}

func unmarshalStruct(s *ast.StructType, lhs, context string) {
	printf(`json.DecodeObj("%s", de, func(de json.Decoder, prop json.Token) (err error) {
			switch {
	`, context)

	for _, field := range s.Fields.List {
		fname := field.Names[0].Name
		if !ast.IsExported(fname) {
			continue
		}
		prop := fname
		if field.Tag != nil {
			tag, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				errln("struct tag", fmt.Sprintf("%s.%s: %s", context, fname, err))
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
		lhs := lhs + "." + fname
		context := context + "." + fname
		unmarshal(true, lhs, "=", context, field.Type)
	}
	println(`
		default:
			err = de.Skip()
		}
		return
	});`)
}

func unmarshal(checkNull bool, lhs, equals, context string, t ast.Expr) {
	star := false
	if s, ok := t.(*ast.StarExpr); ok {
		star = true
		t = s.X
		checkNull = true
	}
	switch t := t.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string", "float64", "bool", "int", "int64":
			if star {
				if equals == ":=" {
					printf("var %s *%s;", lhs, expr2String(t))
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
			method := strings.ToUpper(t.Name[:1]) + t.Name[1:]
			if star {
				printf("var pval %s;", expr2String(t))
				printf(`%s, err %s %s.%s("%s");`, "pval", "=", val, method, context)
				printf("%s = &pval;", lhs)
			} else {
				printf(`%s, err %s %s.%s("%s");`, lhs, equals, val, method, context)
			}
			if checkNull {
				printf("};")
			}
		default:
			if equals == ":=" {
				printf(`%s := %s{};`, lhs, t.Name)
			}
			printf(`err %s %s.DecodeJSON(de);`, equals, lhs)
		}
	case *ast.InterfaceType:
		printf(`%s, err %s de.Decode();`, lhs, equals)
	case *ast.ArrayType:
		if equals == ":=" {
			printf(`var %s []%s;`, lhs, expr2String(t.Elt))
			equals = "="
		}
		printf(`err %s json.DecodeArr("%s", de, func(de json.Decoder) error {`, equals, context)
		println()
		item := newVar("item")
		unmarshal(false, item, ":=", context+"[]", t.Elt)
		printf(`%s = append(%s, %s);`, lhs, lhs, item)
		printf("return err;")
		printf("});")
		releaseVar(item)
	case *ast.MapType:
		ktype := expr2String(t.Key)
		if ktype != "string" {
			printf("\n// map with non-string key not implemented\n")
			notImplemented()
			return
		}
		printf(`%s %s make(%s);`, lhs, equals, expr2String(t))
		printf(`err %s json.DecodeObj("%s", de, func(de json.Decoder, prop json.Token) (err error) {`, equals, context)
		println()
		printf(`k, _ := prop.String("");`)
		v := newVar("v")
		unmarshal(false, v, ":=", context+"{}", t.Value)
		printf(`%s[k] = %s;`, lhs, v)
		printf("return err;")
		printf("});")
		releaseVar(v)
	case *ast.SelectorExpr:
		if expr2String(t) != "json.RawMessage" {
			printf("\n//%s %s %#v\n", lhs, expr2String(t), t)
			notImplemented()
			return
		}
		printf(`%s, err %s de.Marshal();`, lhs, equals)
	case *ast.StructType:
		printf("err %s", equals)
		unmarshalStruct(t, lhs, context)
	default:
		printf("\n//%s %s %#v\n", lhs, expr2String(t), t)
		notImplemented()
	}
}

var scope = make(map[string]struct{})

func newVar(name string) string {
	if _, ok := scope[name]; !ok {
		return name
	}
	i := 1
	for {
		n := fmt.Sprintf("%s%d", name, i)
		if _, ok := scope[n]; !ok {
			return n
		}
		i++
	}
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

func expr2String(t ast.Expr) string {
	switch t := t.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ArrayType:
		return "[]" + expr2String(t.Elt)
	case *ast.MapType:
		return "map[" + expr2String(t.Key) + "]" + expr2String(t.Value)
	case *ast.SelectorExpr:
		return expr2String(t.X) + "." + expr2String(t.Sel)
	case *ast.StructType:
		var buf = strings.Builder{}
		buf.WriteString("struct {")
		for _, f := range t.Fields.List {
			buf.WriteString(f.Names[0].Name)
			buf.WriteString(" ")
			buf.WriteString(expr2String(f.Type))
			if f.Tag != nil {
				buf.WriteString(" ")
				buf.WriteString(f.Tag.Value)
			}
			buf.WriteString(";")
		}
		buf.WriteString("}")
		return buf.String()
	case *ast.StarExpr:
		return "*" + expr2String(t.X)
	default:
		panic(fmt.Sprintf("expr2String(%T)", t))
	}
}
