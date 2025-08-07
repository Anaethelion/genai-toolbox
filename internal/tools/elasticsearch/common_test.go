// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"reflect"
	"testing"

	"github.com/googleapis/genai-toolbox/internal/tools"
)

func TestReplaceQueryParams(t *testing.T) {
	type args struct {
		query       string
		params      tools.Parameters
		paramValues tools.ParamValues
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic replacement",
			args: args{
				query: "FROM $index | KEEP $field | SORT $field DESC",
				params: tools.Parameters{
					tools.NewStringParameter("index", "index description"),
					tools.NewStringParameter("field", "field description"),
				},
				paramValues: tools.ParamValues{
					{Name: "field", Value: "some-field"},
					{Name: "index", Value: "some-index"},
				},
			},
			want: "FROM some-index | KEEP some-field | SORT some-field DESC",
		},
		{
			name: "string array replacement",
			args: args{
				query: "FROM $indices | KEEP $field | SORT $field DESC",
				params: tools.Parameters{
					tools.NewArrayParameter(
						"indices", "array", tools.NewStringParameter("index", "index description"),
					),
					tools.NewStringParameter("field", "field description"),
				},
				paramValues: tools.ParamValues{
					{Name: "indices", Value: []string{"index1", "index2"}},
					{Name: "field", Value: "some-field"},
				},
			},
			want: "FROM index1,index2 | KEEP some-field | SORT some-field DESC",
		},
		{
			name: "integer replacement",
			args: args{
				query: "FROM $indices | KEEP $field | SORT $field DESC | LIMIT $limit",
				params: tools.Parameters{
					tools.NewArrayParameter(
						"indices", "array", tools.NewStringParameter("index", "index description"),
					),
					tools.NewStringParameter("field", "field description"),
					tools.NewIntParameter("limit", "limit description"),
				},
				paramValues: tools.ParamValues{
					{Name: "indices", Value: []string{"index1", "index2"}},
					{Name: "field", Value: "some-field"},
					{Name: "limit", Value: 10},
				},
			},
			want: "FROM index1,index2 | KEEP some-field | SORT some-field DESC | LIMIT 10",
		},
		{
			name: "float replacement",
			args: args{
				query: "LIMIT $limit",
				params: tools.Parameters{
					tools.NewFloatParameter("limit", "float description"),
				},
				paramValues: tools.ParamValues{
					{Name: "limit", Value: 10.1},
				},
			},
			want: "LIMIT 10.1",
		},
		{
			name: "boolean replacement",
			args: args{
				query: "my-field == $bool",
				params: tools.Parameters{
					tools.NewBooleanParameter("bool", "bool description"),
				},
				paramValues: tools.ParamValues{
					{Name: "bool", Value: true},
				},
			},
			want: "my-field == true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceQueryParams(tt.args.query, tt.args.params, tt.args.paramValues); got != tt.want {
				t.Errorf("ReplaceQueryParams() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetrieveIndices(t *testing.T) {
	type args struct {
		params tools.ParamValues
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "single index",
			args: args{
				params: tools.ParamValues{
					{"index", "my-index"},
				},
			},
			want:    []string{"my-index"},
			wantErr: false,
		},
		{
			name: "multiple indices",
			args: args{
				params: tools.ParamValues{
					{"indices", []any{"index1", "index2"}},
				},
			},
			want:    []string{"index1", "index2"},
			wantErr: false,
		},
		{
			name: "missing indices",
			args: args{
				params: tools.ParamValues{
					{"missing", "my-index"},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid index type",
			args: args{
				params: tools.ParamValues{
					{"indices", "not-an-array"},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid single index type",
			args: args{
				params: tools.ParamValues{
					{"index", 12345},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RetrieveIndices(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveIndices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetrieveIndices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceQueryDSLParams(t *testing.T) {
	type args struct {
		query       string
		params      tools.Parameters
		paramValues tools.ParamValues
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic replacement",
			args: args{
				query: `{"query":{"match":{"field":$field}}}`,
				params: tools.Parameters{
					tools.NewStringParameter("field", "field name"),
				},
				paramValues: tools.ParamValues{
					{Name: "field", Value: "some-field"},
				},
			},
			want:    `{"query":{"match":{"field":"some-field"}}}`,
			wantErr: false,
		},
		{
			name: "array replacement",
			args: args{
				query: `{"query":{"terms":{"field":$fields}}}`,
				params: tools.Parameters{
					tools.NewStringParameter("fields", "fields names"),
				},
				paramValues: tools.ParamValues{
					{Name: "fields", Value: []string{"field1", "field2"}},
				},
			},
			want:    `{"query":{"terms":{"field":["field1","field2"]}}}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceQueryDSLParams(tt.args.query, tt.args.params, tt.args.paramValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceQueryDSLParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ReplaceQueryDSLParams() got = %v, want %v", got, tt.want)
			}
		})
	}
}
