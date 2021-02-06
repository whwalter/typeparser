package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"go/ast"
	"io/ioutil"
	"os"
)

func main() {
	fset := token.NewFileSet() // positions are relative to fset
	src := `package foo

import (
	"fmt"
	"time"
)

type bar struct {
	A string
}
type foo struct {
	bar
	t int
}
func bbar() {
	fmt.Println(time.Now())
}`

	src = src
	file, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	fmt.Println(string(file))
	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", string(file), parser.DeclarationErrors)

	if err != nil {
		fmt.Println(err)
		return
	}

	fl := Filer{
		Pkgs: map[string]Thinger{},
	}

	tl := []*ast.TypeSpec{}
	for _, s := range f.Imports {
		if s.Path.Value == targetIm {
			t := Thinger{}
			if s.Name != nil {
				t.ImportName = s.Name.Name
			}
			t.Types = f.Decls
			tl = extractTypes(f.Decls)
			fl.Pkgs[f.Name.Name] = t
		}
	}

	wrappables := map[string]ast.Node{}
	for _, t := range tl {
		switch t.Type.(type) {
		case *ast.StructType:
			fmt.Printf("Struct: %s\n", t.Name.Name)
			if objectMetaEmbedFilter(t.Type, "ObjectMeta") {
				wrappables[t.Name.Name] = t.Type
			}
		}
	}

	for k := range wrappables {
		fmt.Printf("Templateable Type %s\n", k)
	}
/*
	for _, types := range fl.Pkgs {
		for _, t := range types.Types {
			//wrappable, err := ast.Inspect(t, runtimeObjectFilter)
//			ast.Inspect(t, runtimeObjectFilter)
		}
	}
	*/
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
type Filer struct {
	Pkgs map[string]Thinger
}

type Thinger struct {
	ImportName string
	Types []ast.Decl
}
const targetIm = "\"k8s.io/apimachinery/pkg/apis/meta/v1\""
const targetImport = "\"fmt\""
