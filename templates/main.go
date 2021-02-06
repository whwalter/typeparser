package main

import (
	"io/ioutil"
	"text/template"
	"strings"
	"os"
)
type T struct {
	Name string
	Pkg string
}
func main() {
	fm := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	file, err := ioutil.ReadFile("./object_wrapper.tmpl")
	if err != nil {
		panic(err)
	}

	temp, err := template.New("thing").Funcs(fm).Parse(string(file))
	if err != nil {
		panic(err)
	}

	t := T{
		Name: "TestFunc",
		Pkg: "thingv1",
	}
	err = temp.Execute(os.Stdout, []T{t})
	if err != nil {
		panic(err)
	}
}
