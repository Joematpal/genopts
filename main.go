package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var filename = flag.String("file", "", "Source file to process")

func main() {
	flag.Parse()
	if *filename == "" {
		log.Fatal("Missing -file flag")
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	filePath := filepath.Join(cwd, *filename)

	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	pkgName := node.Name.Name
	constructors := map[string]bool{}

	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "New") {
			constructors[fn.Name.Name] = true
		}
	}

	var structs []StructData

	fieldsCheck := map[string]map[string]struct{}{}
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			var fields []Field
			for _, field := range st.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag := strings.Trim(field.Tag.Value, "`")
				if strings.Contains(tag, `with:"-"`) {
					for _, name := range field.Names {
						_, ok := fieldsCheck[name.Name]
						if ok {
							fieldsCheck[name.Name][ts.Name.Name] = struct{}{}
						} else {
							fieldsCheck[name.Name] = map[string]struct{}{
								ts.Name.Name: {},
							}
						}

						fields = append(fields, Field{
							Name: name.Name,
							Type: exprString(field.Type),
						})
					}
				}
			}

			if len(fields) == 0 {
				continue
			}

			// Check if there are any field duplications across the other struts.
			// If so we need to prepend a struct name to the with func: ${StructName}_With${FieldName}

			hasFieldDuplicationAcrossStructsInPackage := false
			for _, field := range fields {
				if structs, ok := fieldsCheck[field.Name]; ok && len(structs) > 1 {
					if _, ok := structs[ts.Name.Name]; ok {
						hasFieldDuplicationAcrossStructsInPackage = true
					}
				}
			}

			structName := ts.Name.Name
			optionName := structName + "Option"
			funcName := toCamelCase(structName) + "OptionFunc"
			ctorName := "New" + structName
			structs = append(structs, StructData{
				Name:        structName,
				OptionName:  optionName,
				FuncName:    funcName,
				OptionType:  ctorName,
				Fields:      fields,
				HasCtorFunc: constructors[ctorName],
				HasFieldDup: hasFieldDuplicationAcrossStructsInPackage,
			})
		}
	}

	if len(structs) == 0 {
		return
	}

	outputFile := strings.TrimSuffix(filePath, ".go") + ".gen.go"
	out, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	tmpl := template.Must(template.New("code").Funcs(template.FuncMap{
		"toStartCase": toStartCase,
	}).Parse(tmplSrc))

	var buf bytes.Buffer
	data := struct {
		Package string
		Structs []StructData
	}{
		Package: pkgName,
		Structs: structs,
	}

	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatal(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("failed to format: %v", err)
	}
	_, err = out.Write(formatted)
	if err != nil {
		log.Fatal(err)
	}
}

type Field struct {
	Name string
	Type string
}

type StructData struct {
	Name        string
	OptionName  string
	FuncName    string
	OptionType  string
	Fields      []Field
	HasCtorFunc bool
	HasFieldDup bool
}

func exprString(e ast.Expr) string {
	var buf bytes.Buffer
	_ = format.Node(&buf, token.NewFileSet(), e)
	return buf.String()
}

const tmplSrc = `// Code generated by generateopts; DO NOT EDIT.

package {{.Package}}

{{range .Structs}}
type {{.OptionName}} interface {
	apply(*{{.Name}}) error
}

type {{.FuncName}} func(*{{.Name}}) error

func (f {{.FuncName}}) apply(s *{{.Name}}) error {
	return f(s)
}

{{- $optName := .OptionName -}}
{{- $funcName := .FuncName -}}
{{- $structName := .Name -}}
{{- $hasFieldDup := .HasFieldDup -}}

{{range .Fields}}

func {{if $hasFieldDup}}{{$structName}}_{{end -}}With{{toStartCase .Name}}(v {{.Type}}) {{$optName}} {
	return {{$funcName}}(func(s *{{$structName}}) error {
		s.{{.Name}} = v
		return nil
	})
}
{{end}}

{{if not .HasCtorFunc}}
func {{.OptionType}}(opts ...{{.OptionName}}) (*{{.Name}}, error) {
	obj := &{{.Name}}{}
	for _, opt := range opts {
		if err := opt.apply(obj); err != nil {
			return nil, err
		}
	}
	return obj, nil
}
{{end}}

{{end}}`

// TODO: make a unit test compairing the start and camel case funcs
func toStartCase(s string, opts ...language.Tag) string {
	return cases.Title(language.English, cases.NoLower).String(s)
}

func toCamelCase(str string) string {
	if str == "" {
		return ""
	}

	// Remove separators and normalize chunks
	words := strings.FieldsFunc(str, func(r rune) bool {
		return r == '_' || r == '-' || unicode.IsSpace(r)
	})

	if len(words) == 0 {
		return ""
	}

	// First word: lowercase first letter, preserve rest
	first := []rune(words[0])
	first[0] = unicode.ToLower(first[0])
	result := string(first)

	// Append subsequent words exactly as-is
	for _, word := range words[1:] {
		result += word
	}

	return result
}
