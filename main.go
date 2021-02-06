package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"go/ast"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

func main() {

	fset := token.NewFileSet() // positions are relative to fset

	file, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", string(file), parser.DeclarationErrors)

	if err != nil {
		fmt.Println(err)
		return
	}


	tl := []*ast.TypeSpec{}
	for _, s := range f.Imports {
		if s.Path.Value == targetIm {
			tl = extractTypes(f.Decls)
		}
	}

	wrappables := map[string]ast.Node{}
	wrap := []w{}
	for _, t := range tl {
		switch t.Type.(type) {
		case *ast.StructType:
			fmt.Printf("Struct: %s\n", t.Name.Name)
			if objectMetaEmbedFilter(t.Type, "ObjectMeta") {
				wrappables[t.Name.Name] = t.Type
				wrap = append(wrap, w{ Name: t.Name.Name, Pkg: f.Name.Name})
			}
		}
	}

	fm := template.FuncMap{
		"ToLower": strings.ToLower,
	}
	tFile, err := ioutil.ReadFile("./templates/object_wrapper.tmpl")
	if err != nil {
		panic(err)
	}

	temp, err := template.New("thing").Funcs(fm).Parse(string(tFile))
	if err != nil {
		panic(err)
	}

	oFile, err := os.OpenFile("./output.go", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer oFile.Close()

	err = temp.Execute(oFile, wrap)
	for k := range wrappables {
		fmt.Printf("Templateable Type %s\n", k)

	}
}

type w struct {
	Name string
	Pkg string
}
func extractTypes(decls []ast.Decl) []*ast.TypeSpec {
	tl := []*ast.TypeSpec{}
	for _, decl := range decls {
		switch decl.(type) {
		case *ast.GenDecl:
			if decl.(*ast.GenDecl).Tok == token.TYPE {
				for _, t := range decl.(*ast.GenDecl).Specs {
					tl = append(tl, t.(*ast.TypeSpec))
				}
			}

		}
	}
	return tl
}


func runtimeObjectFilter(n ast.Node) bool {
	_, ok := n.(*ast.GenDecl)
	if ok { fmt.Println("castable")}
	switch n.(type) {
	case *ast.StructType:
//		return objectMetaEmbedFilter(n)
	default:
		fmt.Println(n)
	}
	return false
}

func objectMetaEmbedFilter(n ast.Node, embedName string) bool {
	for _, field := range n.(*ast.StructType).Fields.List {
		var name string
		switch field.Type.(type) {
		case *ast.SelectorExpr:
			name = field.Type.(*ast.SelectorExpr).Sel.Name
		case *ast.Ident:
			name = field.Type.(*ast.Ident).Name
		}
		if name == embedName{
			return true
		}
	}
	return false
}

const targetIm = "\"k8s.io/apimachinery/pkg/apis/meta/v1\""
const targetImport = "\"fmt\""
