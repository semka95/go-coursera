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

// GenerateParams stores params for generator
type GenerateParams struct {
	URL      string `json:"url"`
	Auth     bool   `json:"auth"`
	Method   string `json:"method"`
	Name     string
	InParam  string
	RetParam string
}

type GenerateFields struct {
	Name        string
	Type        string
	Validations []string
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	prepFunc := make(map[string][]GenerateParams)
	prepStruct := make(map[string][]GenerateFields)

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, imp)
	fmt.Fprintln(out)

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

			p := &GenerateParams{}
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
			prepFunc[reciever] = []GenerateParams{*p}
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
							tmpl := tpl{
								FieldName: field.Names[0].Name,
							}
							fType := field.Type.(*ast.Ident).Name
							var res bytes.Buffer
							for ind, val := range tagArr {
								if strings.HasPrefix(val, "required") {
									tmpl.ParamName = "required"
									validateRequiredTpl.Execute(&res, tmpl) // check err !!!
									tagArr[ind] = res.String()
									continue
								}
								if strings.HasPrefix(val, "default") {
									tmpl.ParamValue = val[8:]
									if fType == "string" {
										validateDefaultStrTpl.Execute(&res, tmpl) // check err !!!
										tagArr[ind] = res.String()
										continue
									}
									validateDefaultIntTpl.Execute(&res, tmpl) // check err !!!
									tagArr[ind] = res.String()
									continue
								}
								if strings.HasPrefix(val, "min") {
									tmpl.ParamValue = val[4:]
									if fType == "string" {
										validateMinStrTpl.Execute(&res, tmpl) // check err !!!
										tagArr[ind] = res.String()
										continue
									}
									validateMinIntTpl.Execute(&res, tmpl) // check err !!!
									tagArr[ind] = res.String()
									continue
								}
								if strings.HasPrefix(val, "max") {
									tmpl.ParamValue = val[4:]
									validateMaxIntTpl.Execute(&res, tmpl) // check err !!!
									tagArr[ind] = res.String()
									continue
								}
							}
							arr := prepStruct[structName]
							prepStruct[structName] = append(arr, GenerateFields{Name: field.Names[0].Name, Type: fType, Validations: tagArr})
						}
					}

					// fieldName := field.Names[0].Name
					// fileType := field.Type.(*ast.Ident).Name

					// fmt.Printf("\tgenerating code for field %s.%s\n", currType.Name.Name, fieldName)

					// switch fileType {
					// case "int":
					// 	intTpl.Execute(out, tpl{fieldName})
					// case "string":
					// 	strTpl.Execute(out, tpl{fieldName})
					// default:
					// 	log.Fatalln("unsupported", fileType)
					// }
				}

			}
		}

	}
}
