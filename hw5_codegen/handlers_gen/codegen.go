package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"text/template"
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

	auth = `	if r.Header.Get("X-Auth") != "100500" {
		http.Error(w, ` + "`{\"error\": \"unauthorized\"}`" + `, http.StatusForbidden)
		return
	}`

	endFunc = `
	if err != nil {
		var e ApiError
		errText := fmt.Sprintf(` + "`{\"error\": \"%s\"}`" + `, err)
		if errors.As(err, &e) {
			http.Error(w, errText, e.HTTPStatus)
			return
		}
		http.Error(w, errText, http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"error":    "",
		"response": &u,
	}

	answer, err := json.Marshal(result)
	if err != nil {
		http.Error(w, ` + "`{\"error\": \"error marshaling answer\"}`" + `, http.StatusInternalServerError)
	}
	w.Write(answer)
}
`

	validateRequiredTpl = template.Must(template.New("validateRequiredTpl").Parse(`
	if params.{{.FieldName}} == "" {
		http.Error(w, ` + "`{\"error\": \"{{.ParamName}} must me not empty\"}`" + `, http.StatusBadRequest)
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
		http.Error(w, ` + "`{\"error\": \"{{.ParamName}} len must be >= {{.ParamValue}}\"}`" + `, http.StatusBadRequest)
		return
	}
	`))

	validateMinIntTpl = template.Must(template.New("validateMinIntTpl").Parse(`
	if params.{{.FieldName}} < {{.ParamValue}} {
		http.Error(w, ` + "`{\"error\": \"{{.ParamName}} must be >= {{.ParamValue}}\"}`" + `, http.StatusBadRequest)
		return
	}
	`))

	validateMaxIntTpl = template.Must(template.New("validateMaxIntTpl").Parse(`
	if params.{{.FieldName}} > {{.ParamValue}} {
		http.Error(w, ` + "`{\"error\": \"{{.ParamName}} must be <= {{.ParamValue}}\"}`" + `, http.StatusBadRequest)
		return
	}
	`))

	validateEnumStrTpl = template.Must(template.New("validateEnumStrTpl").Parse(`
	if !({{$fname := .FieldName}}{{range $i, $v := .ParamEnum}}{{if eq $i 0}}params.{{$fname}} == "{{$v}}"{{else}} || params.{{$fname}} == "{{$v}}"{{end}}{{end}}) {
		http.Error(w, ` + "`{\"error\": \"{{.ParamName}} must be one of [{{range $i, $v := .ParamEnum}}{{if eq $i 0}}{{$v}}{{else}}, {{$v}}{{end}}{{end}}]\"}`" + `, http.StatusBadRequest)
		return
	}
	`))

	validateEnumIntTpl = template.Must(template.New("validateEnumIntTpl").Parse(`
	if !({{$fname := .FieldName}}{{range $i, $v := .ParamEnum}}{{if eq $i 0}}params.{{$fname}} == {{$v}}{{else}} || params.{{$fname}} == {{$v}}{{end}}{{end}}) {
		http.Error(w, ` + "`{\"error\": \"{{.ParamName}} must be one of [{{range $i, $v := .ParamEnum}}{{if eq $i 0}}{{$v}}{{else}}, {{$v}}{{end}}{{end}}]\"}`" + `, http.StatusBadRequest)
		return
	}
	`))

	getIntTpl = template.Must(template.New("getIntTpl").Parse(`
	{{.ParamName}}, err := strconv.Atoi(r.FormValue("{{.ParamName}}"))
	if err != nil {
		http.Error(w, ` + "`{\"error\": \"{{.ParamName}} must be int\"}`" + `, http.StatusBadRequest)
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
		fmt.Println(err)
		return
	}

	out, _ := os.Create(os.Args[2])

	prepFunc := make(map[string][]GenerateFunc)
	prepStruct := make(map[string][]GenerateFields)

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
										tagArr[ind], tagArr[0] = tagArr[0], tagArr[ind]
										continue
									}
									err := validateDefaultIntTpl.Execute(&res, tmpl)
									if err != nil {
										fmt.Println("template execute error: ", err.Error())
										continue
									}
									tagArr[ind] = res.String()
									tagArr[ind], tagArr[0] = tagArr[0], tagArr[ind]
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
		fmt.Fprintf(out, "func (srv *%v) ServeHTTP(w http.ResponseWriter, r *http.Request) {\n", k)
		fmt.Fprintln(out, `	switch r.URL.Path {`)
		for _, param := range v {
			fmt.Fprintf(out, "\tcase \"%v\":\n", param.URL)
			fmt.Fprintf(out, "\t\tsrv.handler%v(w, r)\n", param.Name)
		}
		fmt.Fprintln(out, `	default:
		http.Error(w, `+"`{\"error\": \"unknown method\"}`"+`, http.StatusNotFound)
	}
}
`)
		fmt.Fprint(out)
	}

	// render functions
	for k, v := range prepFunc {
		for _, param := range v {
			fmt.Fprintf(out, "func (srv *%v) handler%v(w http.ResponseWriter, r *http.Request) {\n", k, param.Name)
			fmt.Fprintln(out, `	w.Header().Set("Content-Type", "application/json")
			`)

			if param.Method != "" {
				fmt.Fprintf(out, "\tif r.Method != \"%v\" {\n", param.Method)
				fmt.Fprintln(out, `	http.Error(w, `+"`{\"error\": \"bad method\"}`"+`, http.StatusNotAcceptable)
		return
	}`)
			}

			if param.Auth {
				fmt.Fprintln(out, auth)
			}

			structPar := prepStruct[param.InParam]
			// optimize: don't collect, range map again
			var validations []string
			// optimize strings.Builder
			params := fmt.Sprintf("\tparams := %v{\n", param.InParam)
			var res bytes.Buffer
			for _, strP := range structPar {
				res.Reset()
				if strP.Type == "int" {
					err := getIntTpl.Execute(&res, tpl{ParamName: strP.ParamName})
					if err != nil {
						fmt.Println("template execute error: ", err.Error())
						continue
					}
					fmt.Fprintln(out, res.String())
					params += fmt.Sprintf("\t\t%v: %v,\n", strP.Name, strP.ParamName)
				} else {
					params += fmt.Sprintf("\t\t%v: r.FormValue(\"%v\"),\n", strP.Name, strP.ParamName)
				}
				validations = append(validations, strP.Validations...)
			}
			params += "\t}"
			fmt.Fprintln(out, params)

			for _, s := range validations {
				fmt.Fprint(out, s)
			}

			fmt.Fprintf(out, "\n\tu, err := srv.%v(context.Background(), params)", param.Name)
			fmt.Fprintln(out, endFunc)
		}
	}
}
