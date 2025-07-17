package main

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEndpoint_GetFuncParameters(t *testing.T) {
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		e    Endpoint
		want ParameterTypes
		args args
	}{
		{
			name: "",
			e: Endpoint{
				OperationId: "findPetsByStatus",
				Parameters: []Parameter{
					{
						Name:        "status",
						In:          "query",
						Description: "Status values that need to be considered for filter",
						Required:    true,
						Type:        "array",
						Items: Items{
							Type: "string",
							Enum: []string{
								"available",
								"pending",
								"sold",
							},
							Default: "available",
						},
						CollectionFormat: "multi",
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "status",
						In:   "query",
						Type: "[]string",
					},
				},
			},
		},
		{
			e: Endpoint{
				Parameters: []Parameter{
					{
						Name: "id",
						In:   "path",
						Schema: Schema{
							Type:   "integer",
							Format: "int64",
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "id",
						In:   "path",
						Type: "int64",
					},
				},
			},
		},
		{
			name: "should pass: but should never happen in a parameter",
			e: Endpoint{
				Parameters: []Parameter{
					{
						Name: "id",
						In:   "path",
						Schema: Schema{
							Type:   SwaggerTypes_string,
							Format: SwaggerFormats_binary,
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "id",
						In:   "path",
						Type: "[]byte",
					},
				},
			},
		},
		{
			e: Endpoint{
				Parameters: []Parameter{
					{
						Name: "id",
						In:   "path",
						Schema: Schema{
							Type:   SwaggerTypes_string,
							Format: SwaggerFormats_email,
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "id",
						In:   "path",
						Type: "string",
					},
				},
			},
		},
		{
			e: Endpoint{
				Parameters: []Parameter{
					{
						Name: "id",
						In:   "path",
						Schema: Schema{
							Type:   SwaggerTypes_number,
							Format: SwaggerFormats_double,
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "id",
						In:   "path",
						Type: "float64",
					},
				},
			},
		},
		{
			e: Endpoint{
				Parameters: []Parameter{
					{
						Name: "id",
						In:   "path",
						Schema: Schema{
							Type:   SwaggerTypes_number,
							Format: SwaggerFormats_float,
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "id",
						In:   "path",
						Type: "float32",
					},
				},
			},
		},
		{
			e: Endpoint{
				RequestBody: Request{
					Content: map[string]Content{
						ContentType_applicationJson: {
							Schema: Schema{
								Type: SwaggerTypes_string,
							},
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "",
						In:   "body",
						Type: "string",
					},
				},
			},
		},
		{
			e: Endpoint{
				RequestBody: Request{
					Content: map[string]Content{
						ContentType_applicationJson: {
							Schema: Schema{
								Ref: "#/components/schemas/NewPet",
							},
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "newPet",
						In:   "body",
						Type: "NewPet",
					},
				},
			},
		},
		{
			e: Endpoint{
				RequestBody: Request{
					Content: map[string]Content{
						ContentType_applicationJson: {
							Ref: "#/components/schemas/NewPet",
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name: "newPet",
						In:   "body",
						Type: "NewPet",
					},
				},
			},
		},
		{
			args: args{
				method: http.MethodPost,
				path:   "/pet",
			},
			e: Endpoint{
				RequestBody: ReqOrRes{
					Content: map[string]Content{
						ContentType_applicationJson: {
							Schema: Schema{
								Type: "object",
								Properties: map[string]Property{
									"name": {
										Type:   SwaggerTypes_string,
										Format: SwaggerFormats_binary,
									},
									"age": {
										Type:   SwaggerTypes_integer,
										Format: SwaggerFormats_int32,
									},
								},
							},
						},
					},
				},
			},
			want: ParameterTypes{
				Parameters: []ParameterType{
					{
						Name:     "req",
						In:       "body",
						Type:     "AddPetJSONBody",
						IsObject: true,
					},
				},
			},
		},
	}
	for _, tt := range tests[len(tests)-1:] {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.e.GetFuncParameters(tt.args.method, tt.args.path)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Endpoint.GetFuncParameters() = %v", diff)
			}
		})
	}
}
