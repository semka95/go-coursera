package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"reflect"
	"strings"
)

type tpl struct {
	FieldName  string
	ParamName  string
	ParamValue string
	ParamEnum  []string
}

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

	validateDefaultStrTpl = template.Must(template.New("validateDefaultStrTpl").Parse(`
	if params.{{.FieldName}} == "" {
		params.{{.FieldName}} = "{{.ParamValue}}"
	}
	`))

	validateDefaultIntTpl = template.Must(template.New("validateDefaultIntTpl").Parse(`
	if params.{{.FieldName}} == 0 {
		params.{{.FieldName}} = {{.ParamValue}}
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

	validateEnumStrTpl = template.Must(template.New("validateEnumStrTpl").Parse(`
	if !({{$fname := .FieldName}}{{range $i, $v := .ParamEnum}}{{if eq $i 0}}params.{{$fname}} == "{{$v}}"{{else}} || params.{{$fname}} == "{{$v}}"{{end}}{{end}}) {
		http.Error(w, ` + `{"error": "{{.ParamName}} must be one of [{{range $i, $v := .ParamEnum}}{{if eq $i 0}}{{$v}}{{else}}, {{$v}}{{end}}{{end}}]"}` + `, http.StatusBadRequest)
		return
	}
	`))

	validateEnumIntTpl = template.Must(template.New("validateEnumIntTpl").Parse(`
	if !({{$fname := .FieldName}}{{range $i, $v := .ParamEnum}}{{if eq $i 0}}params.{{$fname}} == {{$v}}{{else}} || params.{{$fname}} == {{$v}}{{end}}{{end}}) {
		http.Error(w, ` + `{"error": "{{.ParamName}} must be one of [{{range $i, $v := .ParamEnum}}{{if eq $i 0}}{{$v}}{{else}}, {{$v}}{{end}}{{end}}]"}` + `, http.StatusBadRequest)
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

// GenerateFunc stores params for generator
type GenerateFunc struct {
	URL      string `json:"url"`
	Auth     bool   `json:"auth"`
	Method   string `json:"method"`
	Name     string
	InParam  string
	RetParam string
}

// GenerateFields stores params of struct fields
type GenerateFields struct {
	Name        string
	ParamName   string
	Type        string
	Validations []string
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	prepFunc := make(map[string][]GenerateFunc)
	prepStruct := make(map[string][]GenerateFields)

	out, _ := os.Create(os.Args[2])

	for _, f := range node.Decls {
		if g, ok := f.(*ast.FuncDecl); ok {
			funcName := g.Name.Name
			if g.Doc == nil {
				fmt.Printf("SKIP func %#v doesn't have comments\n", funcName)
				continue
			}

			ind := 0
			needCodegen := false
			for i, comment := range g.Doc.List {
				if needCodegen = strings.HasPrefix(comment.Text, "// apigen:api"); needCodegen {
					ind = i
					break
				}
			}
			if !needCodegen {
				fmt.Printf("SKIP func %#v doesn't have apigen mark\n", funcName)
				continue
			}

			p := &GenerateFunc{}
			err := json.Unmarshal([]byte(g.Doc.List[ind].Text[13:]), p)
			if err != nil {
				fmt.Printf("SKIP func %#v wrong params\n", funcName)
				continue
			}

			reciever := g.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
			p.Name = funcName
			p.InParam = g.Type.Params.List[1].Type.(*ast.Ident).Name
			p.RetParam = g.Type.Results.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
			fmt.Printf("FOUND func %#v\n", funcName)
			if v, ok := prepFunc[reciever]; ok {
				prepFunc[reciever] = append(v, *p)
				continue
			}
			prepFunc[reciever] = []GenerateFunc{*p}
		}

		if g, ok := f.(*ast.GenDecl); ok {
			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
					continue
				}

				structName := currType.Name.Name
				fmt.Printf("process struct %s\n", structName)

				for _, field := range currStruct.Fields.List {

					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
						if v, ok := tag.Lookup("apivalidator"); ok {
							tagArr := strings.Split(v, ",")
							fType := field.Type.(*ast.Ident).Name
							fName := field.Names[0].Name
							tmpl := tpl{
								FieldName: fName,
								ParamName: strings.ToLower(fName),
							}

							var res bytes.Buffer

							for ind, val := range tagArr {
								res.Reset()
								if strings.HasPrefix(val, "paramname") {
									tmpl.ParamName = val[10:]
									tagArr[ind] = ""
									continue
								}
								if strings.HasPrefix(val, "required") {
									err := validateRequiredTpl.Execute(&res, tmpl)
									if err != nil {
										fmt.Println("template execute error: ", err.Error())
										continue
									}
									tagArr[ind] = res.String()
									continue
								}
								if strings.HasPrefix(val, "default") {
									tmpl.ParamValue = val[8:]
									if fType == "string" {
										err := validateDefaultStrTpl.Execute(&res, tmpl)
										if err != nil {
											fmt.Println("template execute error: ", err.Error())
											continue
										}
										tagArr[ind] = res.String()
										continue
									}
									err := validateDefaultIntTpl.Execute(&res, tmpl)
									if err != nil {
										fmt.Println("template execute error: ", err.Error())
										continue
									}
									tagArr[ind] = res.String()
									continue
								}
								if strings.HasPrefix(val, "min") {
									tmpl.ParamValue = val[4:]
									if fType == "string" {
										err := validateMinStrTpl.Execute(&res, tmpl)
										if err != nil {
											fmt.Println("template execute error: ", err.Error())
											continue
										}
										tagArr[ind] = string(res.Bytes())
										continue
									}
									err := validateMinIntTpl.Execute(&res, tmpl)
									if err != nil {
										fmt.Println("template execute error: ", err.Error())
										continue
									}
									tagArr[ind] = string(res.Bytes())
									continue
								}
								if strings.HasPrefix(val, "max") {
									tmpl.ParamValue = val[4:]
									err := validateMaxIntTpl.Execute(&res, tmpl)
									if err != nil {
										fmt.Println("template execute error: ", err.Error())
										continue
									}
									tagArr[ind] = string(res.Bytes())
									continue
								}
								if strings.HasPrefix(val, "enum") {
									tmpl.ParamEnum = strings.Split(val[5:], "|")
									if fType == "string" {
										err := validateEnumStrTpl.Execute(&res, tmpl)
										if err != nil {
											fmt.Println("template execute error: ", err.Error())
											continue
										}
										tagArr[ind] = string(res.Bytes())
										fmt.Print(tagArr[ind])
										continue
									}
									err := validateEnumIntTpl.Execute(&res, tmpl)
									if err != nil {
										fmt.Println("template execute error: ", err.Error())
										continue
									}
									tagArr[ind] = string(res.Bytes())
									fmt.Print(tagArr[ind])
									continue
								}
							}
							arr := prepStruct[structName]
							prepStruct[structName] = append(arr, GenerateFields{Name: fName, ParamName: tmpl.ParamName, Type: fType, Validations: tagArr})
						}
					}
				}

			}
		}

	}

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, imp)
	fmt.Fprintln(out)

	// render ServeHTTP
	for k, v := range prepFunc {

	}
}
