package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"
)

type tpl struct {
	Value      string
	TypeName   string
	MethodName string
	ParamName  string
	FieldName  string
	IsInt      bool
	Slice      []string
}

var (
	funcMap = template.FuncMap{
		"toLower":   strings.ToLower,
		"joinComma": func(slice []string) string { return strings.Join(slice, ", ") },
	}

	serveTplOpen = template.Must(template.New("serveTplOpen").Parse(`
// ServeHTTP for {{ .TypeName }}
func (node *{{ .TypeName }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
`))
	serveTplClose = template.Must(template.New("serveTplClose").Parse(`}
`))
	methodWrapOpen = template.Must(template.New("methodWrapOpen").Parse(`// [Wrapper for {{ .TypeName }}] method: {{ .MethodName }}
func (node *{{ .TypeName }}) wrapper{{ .MethodName }}(w http.ResponseWriter, r *http.Request) {
`))
	methodWrapClose = template.Must(template.New("methodWrapClose").Parse(`}
`))
	tplServeHTTP = template.Must(template.New("tplServeHTTP").Parse(
		`		node.wrapper{{ .MethodName }}(w, r)
`))
	tplAuth = template.Must(template.New("tplAuth").Parse(
		`	// Authorization checker	
	authGood := "100500"
	auth := r.Header.Get("X-Auth")
	if auth != authGood {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resValue{"error": "unauthorized"})
		io.WriteString(w, string(data))
		return
	} 	

`))
	tplMethod = template.Must(template.New("tplMethod").Parse(
		`	// Method checker
	if r.Method != {{ .Value }} {
		w.WriteHeader(http.StatusNotAcceptable)
		data, _ := json.Marshal(resValue{"error": "bad method"})
		io.WriteString(w, string(data))	
		return
	}

`))

	tplUnkMethod = template.Must(template.New("tplUnkMethod").Parse(
		`		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resValue{"error": "unknown method"})
		io.WriteString(w, string(data))
`))

	tplGetParam = template.Must(template.New("tplGetParam").Parse(
		`	param{{.FieldName}} := r.Form.Get("{{ .ParamName }}")
`))
	// FieldName
	tplRequired = template.Must(template.New("tplUnkMethod").Funcs(funcMap).Parse(
		`	// tplRequired
	if param{{ .FieldName }} == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ .FieldName | toLower }} must me not empty"})
		io.WriteString(w, string(data))
		return
	}

`))
	// Slice | FieldName
	tplEnum = template.Must(template.New("tplEnum").Funcs(funcMap).Parse(
		`	// tplEnum
	enumFlag := false	
	{{ range $val := .Slice }}if param{{ $.FieldName }} == "{{ $val }}" {
		enumFlag = true
	}
	{{ end }}if !enumFlag {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ .FieldName | toLower }} must be one of [{{ .Slice | joinComma }}]"})
		io.WriteString(w, string(data))
		return
	}

`))
	// FieldName | Value
	tplDefault = template.Must(template.New("tplDefault").Parse(
		`	// tplDefault	
	if param{{ .FieldName }} == "" {
		param{{.FieldName}} = "{{ .Value }}"
	}

`))
	// IsInt | FieldName | Value
	tplMin = template.Must(template.New("tmpMin").Funcs(funcMap).Parse(
		`	// tplMin
	{{ if .IsInt }}param{{ $.FieldName }}IntMin, err := strconv.Atoi(param{{ $.FieldName }})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ $.FieldName | toLower }} must be int"})
		io.WriteString(w, string(data))
		return
	}
	if param{{ .FieldName }}IntMin < {{ .Value }} {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ $.FieldName | toLower }} must be >= {{ $.Value }}"})
		io.WriteString(w, string(data))
		return 
	}
	{{ else }}if len([]rune(param{{ .FieldName }})) < {{ .Value }} {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ $.FieldName | toLower }} len must be >= {{ $.Value }}"})
		io.WriteString(w, string(data))
		return 
	}
	{{end}}
`))
	// IsInt | FieldName | Value
	tplMax = template.Must(template.New("tmpMax").Funcs(funcMap).Parse(
		`	// tplMax
	{{ if .IsInt }}param{{ $.FieldName }}IntMax, err := strconv.Atoi(param{{ $.FieldName }})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ $.FieldName | toLower }} must be int"})
		io.WriteString(w, string(data))
		return
	}
	if param{{ .FieldName }}IntMax > {{ .Value }} {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ $.FieldName | toLower }} must be <= {{ $.Value }}"})
		io.WriteString(w, string(data))
		return 
	}
	{{ else }}if len([]rune(param{{ .FieldName }})) > {{ .Value }} {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resValue{"error": "{{ .FieldName | toLower}} len must be <= {{ $.Value }}"})
		io.WriteString(w, string(data))
		return 
	}
	{{ end }}
`))

	tplResponseMethod = template.Must(template.New("tplResponseMethod").Parse(
		`	ctx := r.Context()
	response, err := node.{{ .MethodName }}(ctx, params)
	if err != nil {
		switch err.(type) {
		case ApiError:
			data, _ := json.Marshal(resValue{"error": err.(ApiError).Err.Error()})
			w.WriteHeader(err.(ApiError).HTTPStatus)
			io.WriteString(w, string(data))
		default:
			data, _ := json.Marshal(resValue{"error": err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, string(data))
		}
		return
	}
	data, _ := json.Marshal(resValue{"error": "", "response": response})
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(data))	
`))
)

type methodOptions struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type genMethod struct {
	Name      string        // method name
	Node      *ast.FuncDecl // method node
	ValidName string        // second param (for validation) name
	Options   methodOptions // getted JSON options from comment
}

type tag struct {
	Name  string
	Value string
}

type field struct {
	FieldName string
	ParamName string
	IsInt     bool
	Tags      []tag
}

func typeName(typ ast.Expr) string {
	if p, ok := typ.(*ast.StarExpr); ok {
		typ = p.X
	}
	id, _ := typ.(*ast.Ident)
	return id.Name
}

/*

Нам доступны следующие метки валидатора-заполнятора `apivalidator`:
* `required` - поле не должно быть пустым (не должно иметь значение по-умолчанию)
* `paramname` - если указано - то брать из параметра с этим именем, иначе `lowercase` от имени
* `default` - если указано и приходит пустое значение (значение по-умолчанию) - устанавливать то что написано указано в `default`
* `enum` - "одно из"
* `min` - >= X для типа `int`, для строк `len(str)` >=
* `max` - <= X для типа `int`

*/

var (
	validPriority = map[string]int{
		"required":  1,
		"paramname": 2,
		"default":   3,
		"enum":      4,
		"min":       5,
		"max":       5,
	}
)

func validGen(out *os.File, name string, fields []field) {
	fmt.Printf("\t\tgenerating validation of params for %s\n\n", name)
	fmt.Fprintf(out, "\t// validation %s\n", name)
	fmt.Fprintf(out, "\tr.ParseForm()\n")
	for _, field := range fields {
		paramname := strings.ToLower(field.FieldName)
		if field.ParamName != "" {
			paramname = field.ParamName
		}
		tplGetParam.Execute(out, tpl{FieldName: field.FieldName, ParamName: paramname})
		sort.Slice(field.Tags, func(i, j int) bool {
			return validPriority[field.Tags[i].Name] < validPriority[field.Tags[j].Name]
		})
		for _, curTag := range field.Tags {
			switch curTag.Name {
			case "required":
				tplRequired.Execute(out, tpl{FieldName: field.FieldName})
			case "default":
				tplDefault.Execute(out, tpl{FieldName: field.FieldName, Value: curTag.Value})
			case "enum":
				tagValueSlice := strings.Split(strings.Trim(curTag.Value, "()"), "|")
				tplEnum.Execute(out, tpl{FieldName: field.FieldName, Slice: tagValueSlice})
			case "min":
				tplMin.Execute(out, tpl{IsInt: field.IsInt, FieldName: field.FieldName, Value: curTag.Value})
			case "max":
				tplMax.Execute(out, tpl{IsInt: field.IsInt, FieldName: field.FieldName, Value: curTag.Value})
			default:
			}
		}
	}
}

func responseGen(out *os.File, methodName string, name string, fields []field) {
	for _, field := range fields {
		if field.IsInt {
			fmt.Fprintf(out, "\tparam%sInt, _ := strconv.Atoi(param%s)\n", field.FieldName, field.FieldName)
		}
	}
	fmt.Fprintf(out, "\tparams := %s{\n", name)
	for _, field := range fields {
		fmt.Fprintf(out, "\t\t%s: param%s", field.FieldName, field.FieldName)
		if field.IsInt {
			fmt.Fprintf(out, "Int")
		}
		fmt.Fprintf(out, ",\n")
	}
	fmt.Fprintf(out, "\t}\n")
	tplResponseMethod.Execute(out, tpl{MethodName: methodName})
}

func main() {
	in := token.NewFileSet()

	node, err := parser.ParseFile(in, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	// IMPORT LIST
	importList := []string{"strconv", "encoding/json", "io", "net/http"}
	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "import (")
	for _, item := range importList {
		fmt.Fprintln(out, `	"`+item+`"`)
	}
	fmt.Fprintln(out, ")")

	mapStrMethod := make(map[string][]genMethod)
	mapStrByName := make(map[string](*ast.StructType))
	mapGenValid := make(map[string]bool)
	fmt.Printf("Reading file...\n\n")
	for _, decl := range node.Decls {
		if now, ok := decl.(*ast.FuncDecl); ok {
			if now.Recv != nil {
				if strings.HasPrefix(now.Doc.Text(), "apigen:api ") {
					strJson := now.Doc.Text()[len("apigen:api "):]
					data := methodOptions{}
					err := json.Unmarshal([]byte(strJson), &data)
					fmt.Printf("\tcommented JSON: %s", strJson)
					if err != nil {
						fmt.Printf("\tcant UNPACK from JSON to DATA: %s\n\n", now.Name.Name)
					} else {
						fmt.Printf("\tgetted JSON from: %s\n", now.Name.Name)
						fmt.Printf("\t%#v\n\n", data)
					}
					strValidName := ""
					inputStruct := now.Type.Params.List[1]
					if validStruct, ok := inputStruct.Type.(*ast.Ident); ok {
						strValidName = validStruct.Name
					}
					mapStrMethod[typeName(now.Recv.List[0].Type)] = append(mapStrMethod[typeName(now.Recv.List[0].Type)], genMethod{
						Name:      now.Name.Name,
						Node:      now,
						ValidName: strValidName,
						Options:   data,
					})
					mapGenValid[strValidName] = true
				}
			}
		}
		if genNode, ok := decl.(*ast.GenDecl); ok {
			if typeNode, ok := genNode.Specs[0].(*ast.TypeSpec); ok {
				if strNode, ok := typeNode.Type.(*ast.StructType); ok {
					mapStrByName[typeNode.Name.Name] = strNode
				}
			}
		}
	}
	fmt.Printf("File readed!\n\n")

	fmt.Printf("Reading structs for validation\n")
	mapStructFields := make(map[string][]field)
	for structName := range mapGenValid {
		// можно убрать mapStrMethod и хранить вместо bool в мапе саму структуру
		node := mapStrByName[structName]
		fmt.Printf("\t | generating validation for %s\n", structName)
		structFields := []field{}
		for _, curField := range node.Fields.List {
			f := field{}
			tagSlice := strings.Split(strings.TrimPrefix(strings.Trim(curField.Tag.Value, `"`+"`"), `apivalidator:"`), ",")
			fieldTags := []tag{}
			for _, curTag := range tagSlice {
				if strings.HasPrefix(curTag, "paramname") {
					f.ParamName, _ = strings.CutPrefix(curTag, "paramname=")
				} else {
					t := tag{}
					switch {
					case strings.HasPrefix(curTag, "required"):
						t.Name = "required"
						t.Value, _ = strings.CutPrefix(curTag, "required")
					case strings.HasPrefix(curTag, "enum="):
						t.Name = "enum"
						t.Value, _ = strings.CutPrefix(curTag, "enum=")
					case strings.HasPrefix(curTag, "default="):
						t.Name = "default"
						t.Value, _ = strings.CutPrefix(curTag, "default=")
					case strings.HasPrefix(curTag, "min="):
						t.Name = "min"
						t.Value, _ = strings.CutPrefix(curTag, "min=")
					case strings.HasPrefix(curTag, "max="):
						t.Name = "max"
						t.Value, _ = strings.CutPrefix(curTag, "max=")
					}
					if t.Name != "" {
						fieldTags = append(fieldTags, t)
					}
				}
			}
			f.FieldName = curField.Names[0].Name
			f.Tags = fieldTags
			f.IsInt = curField.Type.(*ast.Ident).Name == "int"
			structFields = append(structFields, f)
		}
		mapStructFields[structName] = structFields
	}
	fmt.Printf("Structs reading done!\n\n")

	fmt.Println("Generating started")
	fmt.Fprintf(out, "\n// Result from wrappers\n")
	fmt.Fprintf(out, "type resValue map[string]interface{}\n")
	for structName, methodSlice := range mapStrMethod {
		fmt.Fprintf(out, "\n// ...\n// generated for type: %s\n// ...\n", structName)
		for _, method := range methodSlice {
			fmt.Printf("\tgenerate method %s: \n", method.Name)
			fmt.Fprintf(out, "\n// %#v\n", method.Options)
			methodWrapOpen.Execute(out, tpl{
				TypeName:   structName,
				MethodName: method.Name,
			})
			// Генерация враппера (проверки и т.п.)
			if method.Options.Auth {
				tplAuth.Execute(out, tpl{})
			}
			if method.Options.Method == http.MethodPost {
				tplMethod.Execute(out, tpl{Value: "http.MethodPost"})
			} else if method.Options.Method == http.MethodGet {
				tplMethod.Execute(out, tpl{Value: "http.MethodGet"})
			}
			validGen(out, method.ValidName, mapStructFields[method.ValidName])
			responseGen(out, method.Name, method.ValidName, mapStructFields[method.ValidName])
			methodWrapClose.Execute(out, tpl{})
		}
		// to template
		serveTplOpen.Execute(out, tpl{TypeName: structName})
		fmt.Fprintf(out, "\tswitch r.URL.Path {\n")
		for _, method := range methodSlice {
			fmt.Fprintln(out, `	case "`+method.Options.URL+`":`)
			tplServeHTTP.Execute(out, tpl{MethodName: method.Name})
		}
		fmt.Fprintf(out, "\tdefault:\n")
		tplUnkMethod.Execute(out, tpl{})
		fmt.Fprintf(out, "\t}\n")
		serveTplClose.Execute(out, tpl{})
		// end to template
	}

	fmt.Printf("All done!\n")
	fmt.Printf("by @kayot123")
}
