package main

import (
	"testing"
)

func TestEndpoint_GetMethodName(t *testing.T) {
	type args struct {
		method   string
		pathName string
	}
	tests := []struct {
		name string
		e    Endpoint
		args args
		want string
	}{
		{
			name: "should pass",
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
			args: args{
				method:   "get",
				pathName: "/pet/findByStatus",
			},
			want: "FindPetsByStatus",
		},
		{
			name: "should pass; list response",
			e: Endpoint{
				Responses: map[string]Response{
					"200": {
						Schema: Schema{
							Type: "array",
						},
					},
				},
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
			args: args{
				method:   "get",
				pathName: "/pet/findByStatus",
			},
			want: "ListPetFindByStatus",
		},
		{
			name: "should pass",
			e: Endpoint{
				Responses: map[string]Response{},
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
			args: args{
				method:   "get",
				pathName: "/pet/findByStatus",
			},
			want: "GetPetFindByStatus",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.GetMethodName(tt.args.method, tt.args.pathName); got != tt.want {
				t.Errorf("Endpoint.GetMethodName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_generatePostMethod(t *testing.T) {
// 	type args struct {
// 		path       string
// 		httpMethod string
// 		ep         *Endpoint
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		{
// 			name: "should pass",
// 			args: args{
// 				path:       "",
// 				httpMethod: "http.MethodGet",
// 				ep: &Endpoint{
// 					OperationId: "findPetsByStatus",
// 					Parameters: []Parameter{
// 						{
// 							Name:        "status",
// 							In:          "query",
// 							Description: "Status values that need to be considered for filter",
// 							Required:    true,
// 							Type:        "array",
// 							Item: Item{
// 								Type: "string",
// 								Enum: []string{
// 									"available",
// 									"pending",
// 									"sold",
// 								},
// 								Default: "available",
// 							},
// 							CollectionFormat: "multi",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := generateMethod(tt.args.path, tt.args.httpMethod, tt.args.ep); got != tt.want {
// 				t.Errorf("generatePostMethod() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
