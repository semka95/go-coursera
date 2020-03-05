package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
)

var (
	imp = `import (
		"context"
		"encoding/json"
		"errors"
		"fmt"
		"net/http"
		"strconv"
	)`

	auth = `if r.Header.Get("X-Auth") != "100500" {
		http.Error(w, ` + `{"error": "unauthorized"}` + `, http.StatusForbidden)
		return
	}`

	validateRequiredTpl = template.Must(template.New("validateRequiredTpl").Parse(`
	if params.{{.FieldName}} == "" {
		http.Error(w, ` + `{"error": "{{.ParamName}} must me not empty"}` + `, http.StatusBadRequest)
		return
	}
	`))

	validateDefaultTpl = template.Must(template.New("validateDefaultTpl").Parse(`
	if params.{{.FieldName}} == "" {
		params.{{.FieldName}} = "{{.ParamName}}"
	}
	`))

	validateMinStrTpl = template.Must(template.New("validateMinStrTpl").Parse(`
	if len(params.{{.FieldName}}) < {{.ParamValue}} {
		http.Error(w, ` + `{"error": "{{.ParamName}} len must be >= {{.ParamValue}}"}` + `, http.StatusBadRequest)
		return
	}
	`))

	validateMinIntTpl = template.Must(template.New("validateMinIntTpl").Parse(`
	if params.{{.FieldName}} < {{.ParamValue}} {
		http.Error(w, ` + `{"error": "{{.ParamName}} must be >= {{.ParamValue}}"}` + `, http.StatusBadRequest)
		return
	}
	`))

	validateMaxIntTpl = template.Must(template.New("validateMaxIntTpl").Parse(`
	if params.{{.FieldName}} > {{.ParamValue}} {
		http.Error(w, ` + `{"error": "{{.ParamName}} must be <= {{.ParamValue}}"}` + `, http.StatusBadRequest)
		return
	}
	`))

	validateEnumTpl = template.Must(template.New("validateEnumTpl").Parse(`
	if !({{range $i, $v := .ParamValue}} {{if $i eq 0}} params.{{.FieldName}} == "{{$v}}" {{else}} || params.{{.FieldName}} == "{{$v}}"{{end}}{{end}}) {
		http.Error(w, ` + `{"error": "{{.ParamName}} must be one of [{{range $i, $v := .ParamValue}} {{if $i eq 0}}{{$v}}{{else}}, {{$v}}{{end}}{{end}}]"}` + `, http.StatusBadRequest)
		return
	}
	`))

	getIntTpl = template.Must(template.New("getIntTpl").Parse(`
	{{.ParamName}}, err := strconv.Atoi(r.FormValue("{{.ParamName}}"))
	if err != nil {
		http.Error(w, ` + `{"error": "{{.ParamName}} must be int"}` + `, http.StatusBadRequest)
		return
	}
	`))

	fillStrTpl = template.Must(template.New("fillStrTpl").Parse(`
	{{.FieldName}}:  r.FormValue("{{.ParamName}}"),
	`))

	fillIntTpl = template.Must(template.New("fillIntTpl").Parse(`
	{{.FieldName}}:  {{.ParamName}},
	`))

	methodTpl = template.Must(template.New("methodTpl").Parse(`
	if r.Method != {{.Method}} {
		http.Error(w, ` + `{"error": "bad method"}` + `, http.StatusNotAcceptable)
		return
	}
	`))
)

// код писать тут
func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, imp)
	fmt.Fprintln(out)

	for _, f := range node.Decls {
		if g, ok := f.(*ast.FuncDecl); ok {

		}

	}
}
