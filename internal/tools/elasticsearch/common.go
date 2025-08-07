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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/googleapis/genai-toolbox/internal/tools"
)

// ReplaceQueryParams replaces the placeholders in the query string with the actual values from the parameters.
func ReplaceQueryParams(query string, params tools.Parameters, paramValues tools.ParamValues) string {
	paramsMap := paramValues.AsMapWithDollarPrefix()
	typeMap := make(map[string]string, len(params))
	for _, p := range params {
		placeholder := "$" + p.GetName()
		typeMap[placeholder] = p.GetType()
	}

	newQuery := query
	// For each parameter, replace its placeholder in the query
	for placeholder, value := range paramsMap {
		if value == nil {
			continue
		}
		if typeMap[placeholder] == "array" {
			// If the parameter is an array, join its values with a comma
			arr, ok := value.([]string)
			if !ok {
				continue // skip if not a []string
			}
			newQuery = strings.ReplaceAll(newQuery, placeholder, fmt.Sprintf("%s", strings.Join(arr, ",")))
		} else {
			newQuery = strings.ReplaceAll(newQuery, placeholder, fmt.Sprintf("%v", value))
		}
	}
	return newQuery
}

// ReplaceQueryDSLParams replaces the placeholders in the DSL query string with the actual values from the parameters.
func ReplaceQueryDSLParams(query string, params tools.Parameters, paramValues tools.ParamValues) (string, error) {
	paramsMap := paramValues.AsMapWithDollarPrefix()
	newQuery := query
	// For each parameter, replace its placeholder in the query
	for placeholder, value := range paramsMap {
		data, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("failed to marshal value for placeholder %s: %w", placeholder, err)
		}
		newQuery = strings.ReplaceAll(newQuery, placeholder, fmt.Sprintf("%v", string(data)))
	}
	return newQuery, nil
}

// RetrieveIndices extracts the indices from the provided parameters.
func RetrieveIndices(params tools.ParamValues) ([]string, error) {
	paramsMap := params.AsMap()
	index, ok := paramsMap["index"]
	if ok {
		if str, ok := index.(string); ok {
			return []string{str}, nil
		}
		return nil, fmt.Errorf("invalid type for index: expected string, got %T", index)
	}

	anyIndices, ok := paramsMap["indices"].([]any)
	if !ok {
		return nil, fmt.Errorf("missing required parameter: indices, got %T", paramsMap["indices"])
	}
	var indices []string
	for _, index := range anyIndices {
		if str, ok := index.(string); ok {
			indices = append(indices, str)
		} else {
			return nil, fmt.Errorf("invalid type for indices: expected []string, got %T", index)
		}
	}
	return indices, nil
}
