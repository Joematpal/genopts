package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type OAPIFile struct {
	Paths map[string]Path `json:"paths,omitempty" yaml:"paths"`
}

type Path struct {
	Get    *Endpoint `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *Endpoint `json:"post,omitempty" yaml:"post,omitempty"`
	Patch  *Endpoint `json:"patch,omitempty" yaml:"patch,omitempty"`
	Put    *Endpoint `json:"put,omitempty" yaml:"put,omitempty"`
	Delete *Endpoint `json:"delete,omitempty" yaml:"delete,omitempty"`
}

type Endpoint struct {
	Tags        []string            `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string              `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses   map[string]Response `json:"responses,omitempty" yaml:"responses,omitempty"`
	Produces    []string            `json:"produces,omitempty" yaml:"produces,omitempty"`
	OperationId string              `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	RequestBody Request             `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
}

func (e Endpoint) GetMethodName(method, pathName string) string {
	// List, Get, Create, Update, Patch, Delete
	// Remove query params. they need to be arguments
	if e.OperationId != "" {
		return cases.Title(language.English, cases.NoLower).String(e.OperationId)
	}

	switch strings.ToUpper(method) {
	case http.MethodGet:
		pn := strings.ReplaceAll(pathName, "/", "_")
		for _, parameters := range e.Parameters {
			if parameters.In != "path" {
				pn = strings.ReplaceAll(pn, fmt.Sprintf("_%s", parameters.Name), "")
			}
		}
		pn = strcase.ToCamel(pn)
		if resp, ok := e.Responses["200"]; ok {
			if resp.Schema.Type == SwaggerTypes_array {
				return fmt.Sprintf("List%s", pn)
			}
		}
		// Find out if it is a Get or a List?
		// is only list on an item that is an slice of
		return fmt.Sprintf("Get%s", pn)
	case http.MethodPost:
		pn := strings.ReplaceAll(pathName, "/", "_")
		for _, parameters := range e.Parameters {
			if parameters.In != "path" {
				pn = strings.ReplaceAll(pn, fmt.Sprintf("_%s", parameters.Name), "")
			}
		}
		pn = strcase.ToCamel(pn)

		// Find out if it is a Get or a List?
		// is only list on an item that is an slice of
		return fmt.Sprintf("Add%s", pn)
	}
	return ""
}

func (e Endpoint) GetFuncParameters(method, pathName string) ParameterTypes {
	out := ParameterTypes{
		Parameters: []ParameterType{},
	}

	if !reflect.ValueOf(e.RequestBody).IsZero() {
		isObject := false
		typ := e.RequestBody.DeriveType()
		name := e.RequestBody.GetName()
		if typ == "object" {
			if name == "" {
				name = "req"
			}
			isObject = true
			typ = e.GetMethodName(method, pathName) + "JSONBody"
		}

		out.Parameters = append(out.Parameters, ParameterType{
			Name:     name,
			Type:     typ,
			In:       SwaggerParameterTypes_body,
			IsObject: isObject,
		})
	}

	for _, param := range e.Parameters {
		out.Parameters = append(out.Parameters, ParameterType{
			Name: param.Name,
			Type: param.DeriveType(),
			In:   param.In,
		})
	}
	return out
}

// is it a struct?
// is it just query params?
// is it just path params?
// is it path and query params
type ParameterTypes struct {
	// SwaggerParameterTypes // this needs to be a bitwise operator
	Parameters []ParameterType
}

func (fp ParameterTypes) String() string {
	return ""
}

type FactoryParameter interface {
	GetName() string
	GetType() string
}

type ParameterType struct {
	Name string
	// type needs to be the converted type of the swagger type and format.
	Type string                // this should be the full computed type?
	In   SwaggerParameterTypes // the location where the parameter is used
	// there is also a use case where a schema does not have a ref. and only has inline properties like an object.
	// inline object
	IsObject bool
}

func (pt ParameterType) String() string {
	return ""
}

func (pt ParameterType) GetName() string {
	return pt.Name
}

func (pt ParameterType) GetType() string {
	return pt.Type
}

// func (pt ParameterType) GetSubType() FactoryParameter {
// 	return pt.SubType
// }

type SwaggerTypes = string

const (
	SwaggerTypes_string  = "string"  // (this includes dates and files)
	SwaggerTypes_number  = "number"  //
	SwaggerTypes_integer = "integer" //
	SwaggerTypes_boolean = "boolean" //
	SwaggerTypes_array   = "array"   //
	SwaggerTypes_object  = "object"  //
)

type SwaggerFormats = string

const (
	SwaggerFormats_float  = "float"
	SwaggerFormats_double = "double"
	SwaggerFormats_int32  = "int32"
	SwaggerFormats_int64  = "int64"
	SwaggerFormats_binary = "binary"
	// common string formats
	SwaggerFormats_email = "email"
	SwaggerFormats_uuid  = "uuid"
)

type ContextType = string

const (
	ContentType_applicationJson = "application/json"
)

// there needs to be a place for the body
type Response = ReqOrRes
type Request = ReqOrRes

type ReqOrRes struct {
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]Content `json:"content,omitempty" yaml:"content,omitempty"`
	Type        SwaggerTypes       `json:"type,omitempty"`
	Format      SwaggerFormats     `json:"format,omitempty"`
	Schema      Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Ref         string             `json:"ref,omitempty"`
}

func (p ReqOrRes) GetContent() map[string]Content {
	return p.Content
}

func (p ReqOrRes) GetSchema() TypeGetter {
	return p.Schema
}

func (p ReqOrRes) GetName() string {
	return DeriveName(p)
}

type RefGetter interface {
	GetRef() string
}

type ContentGetter interface {
	GetContent() map[string]Content
}

type SchemaGetter interface {
	GetSchema() TypeGetter
}

type NameDeriver interface {
	RefGetter
	SchemaGetter
	GetContent() map[string]Content
}

func DeriveName[T NameDeriver](t T) string {
	if s := strcase.ToLowerCamel(t.GetRef()); s != "" {
		return s
	}

	if s := t.GetSchema(); s != nil {
		v := DeriveName(s)
		if v != "" {
			return v
		}
	}

	if v, ok := t.GetContent()[ContentType_applicationJson]; ok {
		return DeriveName(v)
	}
	return ""
}

func (p ReqOrRes) GetType() string {
	return p.Type
}

func (p ReqOrRes) GetFormat() string {
	return p.Format
}

func (p ReqOrRes) GetItems() TypeGetter {
	return nil
}

func (p ReqOrRes) GetRef() string {
	if p.Ref != "" {
		return filepath.Base(p.Ref)
	}
	return ""
}

func (p ReqOrRes) DeriveType() string {
	return DeriveType(p)
}

type Content struct {
	ContentType string         `json:"content_type,omitempty" yaml:"content_type,omitempty"`
	Type        SwaggerTypes   `json:"type,omitempty"`
	Format      SwaggerFormats `json:"format,omitempty"`
	Schema      Schema         `json:"schema,omitempty" yaml:"schema,omitempty"`
	Ref         string         `json:"ref,omitempty"`
}

func (c Content) GetContent() map[string]Content {
	return map[string]Content{}
}

func (c Content) GetName() string {
	return DeriveName(c)
}
func (c Content) GetSchema() TypeGetter {
	return c.Schema
}

func (c Content) GetType() string {
	return c.Type
}

func (c Content) GetFormat() string {
	return c.Format
}

func (c Content) GetItems() TypeGetter {
	return nil
}

func (c Content) GetRef() string {
	if c.Ref != "" {
		return filepath.Base(c.Ref)
	}
	return ""
}

func (c Content) DeriveType() string {
	return DeriveType(c)
}

type Parameter struct {
	Name             string                `json:"name,omitempty" yaml:"name,omitempty"`
	In               SwaggerParameterTypes `json:"in,omitempty" yaml:"in,omitempty"`
	Description      string                `json:"description,omitempty"`
	Required         bool                  `json:"required,omitempty" yaml:"required,omitempty"`
	Style            SwaggerStyles         `json:"style,omitempty" yaml:"style,omitempty"`
	Explode          bool                  `json:"explode,omitempty" yaml:"explode,omitempty"`
	Schema           Schema                `json:"schema,omitempty" yaml:"schema,omitempty"`
	Type             SwaggerTypes          `json:"type,omitempty" yaml:"type,omitempty"`
	Format           string                `json:"format,omitempty" yaml:"format,omitempty"`
	Items            Items                 `json:"items,omitempty" yaml:"items,omitempty"`
	CollectionFormat string                `json:"collection_format,omitempty" yaml:"collection_format,omitempty"`
	Ref              string                `json:"ref,omitempty"`
}

func (p Parameter) GetSchema() TypeGetter {
	return p.Schema
}

func (p Parameter) GetType() string {
	return p.Type
}

func (p Parameter) GetFormat() string {
	return p.Format
}

func (p Parameter) GetItems() TypeGetter {
	return p.Items
}

func (p Parameter) GetContent() map[string]Content {
	return map[string]Content{}
}

func (p Parameter) GetRef() string {
	if p.Ref != "" {
		return filepath.Base(p.Ref)
	}
	return ""
}

func (p Parameter) DeriveType() string {
	return DeriveType(p)
}

type Items struct {
	Type    SwaggerTypes   `json:"type,omitempty" yaml:"type,omitempty"`
	Format  SwaggerFormats `json:"format,omitempty" yaml:"format,omitempty"`
	Enum    []string       `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default string         `json:"default,omitempty" yaml:"default,omitempty"`
	Ref     string         `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

func (i Items) GetFormat() string {
	return i.Format
}
func (i Items) GetItems() TypeGetter {
	return nil
}

func (i Items) GetSchema() TypeGetter {
	return nil
}

func (i Items) GetType() string {
	return i.Type
}

func (i Items) GetContent() map[string]Content {
	return map[string]Content{}
}

func (i Items) GetRef() string {
	if i.Ref != "" {
		return filepath.Base(i.Ref)
	}
	return ""
}

type Schema struct {
	Type       SwaggerTypes        `json:"type,omitempty" yaml:"type,omitempty"`
	Format     SwaggerFormats      `json:"format,omitempty"`
	Nullable   bool                `json:"nullable,omitempty"`
	Ref        string              `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Items      Items               `json:"items,omitempty" yaml:"items,omitempty"`
	Enum       []string            `json:"enum,omitempty"`
	Properties map[string]Property `json:"properties,omitempty"`
}

func (s Schema) GetSchema() TypeGetter {
	return nil
}

func (s Schema) GetType() string {
	return s.Type
}

func (s Schema) GetRef() string {
	if s.Ref != "" {
		return filepath.Base(s.Ref)
	}
	return ""
}

func (s Schema) GetContent() map[string]Content {
	return map[string]Content{}
}

func (s Schema) GetFormat() string {
	return s.Format
}
func (s Schema) GetItems() TypeGetter {
	return s.Items
}

type Components struct {
	Schemas map[string]Schema
}

type Property struct {
	Type     SwaggerTypes   `json:"type,omitempty"`
	Format   SwaggerFormats `json:"format,omitempty"`
	Nullable bool           `json:"nullable,omitempty"`
	Required []string       `json:"required,omitempty"`
}

type TypeGetter interface {
	GetType() SwaggerFormats
	GetFormat() SwaggerFormats
	SchemaGetter
	GetItems() TypeGetter
	ContentGetter
	RefGetter
}

func DeriveType[T TypeGetter](t T) string {
	var sb strings.Builder

	if content := t.GetContent(); len(content) != 0 {
		json, ok := content[ContentType_applicationJson]
		if ok {
			return DeriveType(json)
		}
	}

	if ref := t.GetRef(); ref != "" {
		return ref
	}

	// check if schema is not set
	if v := t.GetSchema(); v != nil && !reflect.ValueOf(v).IsZero() {
		return DeriveType(t.GetSchema())
	}

	typ := t.GetType()
	switch typ {
	case SwaggerTypes_array:
		sb.WriteString("[]")
		if v := t.GetItems(); v != nil {
			if s := DeriveType(v); s != "" {
				sb.WriteString(s)
			}
		}
	case SwaggerTypes_integer:
		sb.WriteString(t.GetFormat())
	case SwaggerTypes_string:
		if f := t.GetFormat(); f == SwaggerFormats_binary {
			sb.WriteString("[]byte")
		} else {
			sb.WriteString(typ)
		}
	case SwaggerTypes_number:
		f := t.GetFormat()

		switch f {
		// what about in the case of sending double over the wire? shouldn't they be strings?
		case SwaggerFormats_double:
			sb.WriteString("float64")
		case SwaggerFormats_float:
			sb.WriteString("float32")
		}

	default:
		sb.WriteString(typ)
	}
	return sb.String()
}

type SwaggerStyles = string

const (
	SwaggerStyles_form           = "form"
	SwaggerStyles_spaceDelimited = "spaceDelimited"
	SwaggerStyles_pipeDelimited  = "pipeDelimited"
	SwaggerStyles_deepObject     = "deepObject"
	SwaggerStyles_simple         = "simple"
	SwaggerStyles_label          = "label"
	SwaggerStyles_matrix         = "matrix"
)

type SwaggerParameterTypes = string

const (
	SwaggerParameterTypes_path   = "path"
	SwaggerParameterTypes_query  = "query"
	SwaggerParameterTypes_header = "header"
	SwaggerParameterTypes_cookie = "cookie"
	// Deprecated
	SwaggerParameterTypes_body = "body" // no there anymore.
)
