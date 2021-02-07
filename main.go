package main

import (
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"go/ast"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	log "github.com/sirupsen/logrus"
)

func main() {

	// get the list of input packages
	pkgInfo := parseInput(os.Args[2:]...)


	templateMap := map[string][]TypeSpec{}
	//Parse the source files for each package
	for _, pkg := range pkgInfo {
		wrapables, err := parsePkg(pkg)
		if err != nil {
			panic(err)
		}

		templateMap[pkg.ImportPath] = wrapables
	}
	for pkg, ts := range templateMap {

		fmt.Printf("Package: %s\nTemplated types: %s\n", pkg, ts)
		templatePkg(pkg, ts)
	}

}

type TypeSpec struct {
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

func parseInput(inputs ...string) map[string]*build.Package {
	ctx := build.Default
	importPaths := map[string]*build.Package{}
	for _, input := range inputs {
		pkg, err := ctx.Import(input, ".", build.ImportComment)
		if err != nil {
			log.Errorf("Failure for %s\nerr\ntry go get %s first.\n", input, err.Error(), input)
		}
		importPaths[input] = pkg
	}
	return importPaths
}

func parsePkg(input *build.Package) ([]TypeSpec, error) {

	wrapables := []TypeSpec{}
	for _, source := range input.GoFiles {
		wraps, err := parseFile(fmt.Sprintf("%s%s%s", input.Dir, string(os.PathSeparator), source))
		if err != nil {
			return nil, err
		}
		wrapables = append(wrapables, wraps...)
	}
	return wrapables, nil
}

func parseFile(input string) ([]TypeSpec, error) {

	fset := token.NewFileSet() // positions are relative to fset
	file, err := ioutil.ReadFile(input)
	if err != nil {
		return nil, err
	}
	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", string(file), parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}


	tl := []*ast.TypeSpec{}
	for _, s := range f.Imports {
		if s.Path.Value == targetIm {
			tl = extractTypes(f.Decls)
		}
	}

	wrap := []TypeSpec{}
	for _, t := range tl {
		switch t.Type.(type) {
		case *ast.StructType:
			if objectMetaEmbedFilter(t.Type, "ObjectMeta") && objectMetaEmbedFilter(t.Type, "TypeMeta"){
				wrap = append(wrap, TypeSpec{ Name: t.Name.Name, Pkg: f.Name.Name})
			}
		}
	}
	return wrap, nil
}

func templatePkg(pkg string, wrapables []TypeSpec) {
	path := fmt.Sprintf(".%s%s%s",string(os.PathSeparator), pkg, string(os.PathSeparator))
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}

	oFile, err := os.OpenFile(fmt.Sprintf("%szz_generated_types.go", path), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer oFile.Close()

	testOutFile, err := os.OpenFile(fmt.Sprintf("%szz_generated_types_test.go", path), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer testOutFile.Close()

	jsonOutFile, err := os.OpenFile(fmt.Sprintf("%szz_generated_json.go", path), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer jsonOutFile.Close()

	jsonTestOutFile, err := os.OpenFile(fmt.Sprintf("%szz_generated_json_test.go", path), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer jsonTestOutFile.Close()

	fm := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	// Load all my templaes
	headerFile, err := ioutil.ReadFile("./templates/header.tmpl")
	if err != nil {
		panic(err)
	}

	//TODO: figurre out a way to embed the templates, I suppose const strings works
	tFile, err := ioutil.ReadFile("./templates/object_wrapper.tmpl")
	if err != nil {
		panic(err)
	}

	//TODO: figurre out a way to embed the templates, I suppose const strings works
	testHeadFile, err := ioutil.ReadFile("./templates/tests_header.tmpl")
	if err != nil {
		panic(err)
	}

	//TODO: figurre out a way to embed the templates, I suppose const strings works
	testFile, err := ioutil.ReadFile("./templates/test_types.tmpl")
	if err != nil {
		panic(err)
	}

	jsonHeaderFile, err := ioutil.ReadFile("./templates/json_header.tmpl")
	if err != nil {
		panic(err)
	}

	jsonFile, err := ioutil.ReadFile("./templates/json.tmpl")
	if err != nil {
		panic(err)
	}

	jsonTestHeadFile, err := ioutil.ReadFile("./templates/json_test_header.tmpl")
	if err != nil {
		panic(err)
	}

	jsonTestFile, err := ioutil.ReadFile("./templates/json_tests.tmpl")
	if err != nil {
		panic(err)
	}

	temp, err := template.New("thing").Funcs(fm).Parse(string(headerFile))
	if err != nil {
		panic(err)
	}

	testTemp, err := template.New("thing").Funcs(fm).Parse(string(testHeadFile))
	if err != nil {
		panic(err)
	}

	jsonTemp, err := template.New("thing").Funcs(fm).Parse(string(jsonHeaderFile))
	if err != nil {
		panic(err)
	}

	jsonTestTemp, err := template.New("thing").Funcs(fm).Parse(string(jsonTestHeadFile))
	if err != nil {
		panic(err)
	}

	pSpec := TypeSpec{
		Name: "v1",
		Pkg: pkg,
	}

	err = temp.Execute(oFile, pSpec)
	if err != nil {
		panic(err)
	}

	err = testTemp.Execute(testOutFile, pSpec)
	if err != nil {
		panic(err)
	}

	err = jsonTemp.Execute(jsonOutFile, pSpec)
	if err != nil {
		panic(err)
	}

	err = jsonTestTemp.Execute(jsonTestOutFile, pSpec)
	if err != nil {
		panic(err)
	}

	temp, err = template.New("thing").Funcs(fm).Parse(string(tFile))
	if err != nil {
		panic(err)
	}

	err = temp.Execute(oFile, wrapables)
	if err != nil {
		panic(err)
	}

	testTemp, err = template.New("thing").Funcs(fm).Parse(string(testFile))
	if err != nil {
		panic(err)
	}

	err = testTemp.Execute(testOutFile, wrapables)
	if err != nil {
		panic(err)
	}

	jsonTemp, err = template.New("thing").Funcs(fm).Parse(string(jsonFile))
	if err != nil {
		panic(err)
	}
	err = jsonTemp.Execute(jsonOutFile, wrapables)
	if err != nil {
		panic(err)
	}

	jsonTestTemp, err = template.New("thing").Funcs(fm).Parse(string(jsonTestFile))
	if err != nil {
		panic(err)
	}
	err = jsonTestTemp.Execute(jsonTestOutFile, wrapables)
	if err != nil {
		panic(err)
	}
}
