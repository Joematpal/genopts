package main

//go:generate sh -c "oapi-codegen --package main --generate types swagger.yaml > ./openapi.gen.go"

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var filename = flag.String("file", "", "Source file to process")

type TemplateData struct {
	Package  string
	OAPIFile *OAPIFile
}

//go:embed *.templ
var templates embed.FS

func main() {

	flag.Parse()
	if *filename == "" {
		log.Fatal("Missing -file flag")
	}

	oapiFile, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	oapi := &OAPIFile{}

	switch filepath.Ext(*filename) {
	case ".json":
		if err := json.NewDecoder(oapiFile).Decode(oapi); err != nil {
			log.Fatal(err)
		}
	case ".yml", ".yaml":
		if err := yaml.NewDecoder(oapiFile).Decode(oapi); err != nil {
			log.Fatal(err)
		}

	}

	buf, err := generateClientCode(TemplateData{
		Package:  "client",
		OAPIFile: oapi,
	})
	if err != nil {
		log.Fatal(err)
	}

	io.Copy(os.Stdout, buf)
}

func generateClientCode(data TemplateData) (*bytes.Buffer, error) {
	tmpl := template.New("templates").Funcs(template.FuncMap{
		"ToCamel": strcase.ToLowerCamel,
		"Title": func(s string) string {
			return cases.Title(language.English, cases.NoLower).String(s)
		},
		"GetMethodName": func(method string, path string, ep *Endpoint) string {
			return ep.GetMethodName(method, path)
		},
		"dict":                  dict,
		"getEndpointParameters": getEndpointParameters,
		"getEndpointResponse":   getEndpointResponse,
		"getEndpointOnPath":     getEndpointOnPath,
	})

	tmpl, err := tmpl.ParseFS(templates, "*.templ")
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	buf := bytes.NewBuffer([]byte{})

	err = tmpl.ExecuteTemplate(buf, "client", data)
	if err != nil {
		return nil, fmt.Errorf("exec: %w", err)
	}
	return buf, nil
}

func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid dict call: must pass even number of key-value pairs")
	}
	m := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings, got %T", values[i])
		}
		m[key] = values[i+1]
	}
	return m, nil
}

func getEndpointParameters(method, pathName string, ep *Endpoint) ParameterTypes {
	return ep.GetFuncParameters(method, pathName)
}

func getEndpointResponse(ep *Endpoint) string {
	return ep.Responses["200"].Schema.GetType()
}

func getEndpointOnPath(p Path) map[string]*Endpoint {
	out := map[string]*Endpoint{}
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		val := v.FieldByName(field.Name).Interface().(*Endpoint)
		if val == nil {
			continue
		}
		out[strings.ToUpper(field.Name)] = val
	}
	return out
}
