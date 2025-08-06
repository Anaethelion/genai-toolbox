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

package elasticsearchlistindices

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"io"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/sources"
	es "github.com/googleapis/genai-toolbox/internal/sources/elasticsearch"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

const kind string = "elasticsearch-list-indices"

func init() {
	if !tools.Register(kind, newConfig) {
		panic(fmt.Sprintf("tool kind %q already registered", kind))
	}
}

type compatibleSource interface {
	// Add methods if needed for more tool types
}

var compatibleSources = [...]string{es.SourceKind}

type Config struct {
	Name         string           `yaml:"name" validate:"required"`
	Kind         string           `yaml:"kind" validate:"required"`
	Source       string           `yaml:"source" validate:"required"`
	Description  string           `yaml:"description" validate:"required"`
	Timeout      int              `yaml:"timeout"`
	AuthRequired []string         `yaml:"authRequired"`
	Parameters   tools.Parameters `yaml:"parameters"`
}

var _ tools.ToolConfig = Config{}

func (c Config) ToolConfigKind() string {
	return kind
}

func newConfig(ctx context.Context, name string, decoder *yaml.Decoder) (tools.ToolConfig, error) {
	actual := Config{Name: name}
	if err := decoder.DecodeContext(ctx, &actual); err != nil {
		return nil, err
	}
	return actual, nil
}

type Tool struct {
	Name         string           `yaml:"name"`
	Kind         string           `yaml:"kind"`
	AuthRequired []string         `yaml:"authRequired"`
	Timeout      int              `yaml:"timeout"`
	Parameters   tools.Parameters `yaml:"parameters"`

	manifest    tools.Manifest
	mcpManifest tools.McpManifest
	Src         *es.Source
}

var _ tools.Tool = Tool{}

func (c Config) Initialize(srcs map[string]sources.Source) (tools.Tool, error) {
	src, ok := srcs[c.Source]
	if !ok {
		return nil, fmt.Errorf("source %q not found", c.Source)
	}
	esSrc, ok := src.(*es.Source)
	if !ok {
		return nil, fmt.Errorf("source %q is not elasticsearch", c.Source)
	}

	mcpManifest := tools.McpManifest{
		Name:        c.Name,
		Description: c.Description,
		InputSchema: c.Parameters.McpManifest(),
	}

	return Tool{
		Name:         c.Name,
		Kind:         kind,
		Parameters:   c.Parameters,
		Timeout:      c.Timeout,
		AuthRequired: c.AuthRequired,
		Src:          esSrc,
		manifest:     tools.Manifest{Description: c.Description, Parameters: c.Parameters.Manifest(), AuthRequired: c.AuthRequired},
		mcpManifest:  mcpManifest,
	}, nil
}

func (t Tool) Invoke(ctx context.Context, params tools.ParamValues) (any, error) {
	var cancel context.CancelFunc
	if t.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
		defer cancel()
	} else {
		ctx, cancel = context.WithTimeout(ctx, time.Minute)
		defer cancel()
	}

	indices, err := t.RetrieveIndices(params)
	if err != nil {
		return nil, err
	}

	res, err := esapi.CatIndicesRequest{
		Index:      indices,
		H:          []string{"index", "status", "docs.count"},
		Format:     "json",
		Instrument: t.Src.Client.InstrumentationEnabled(),
	}.Do(ctx, t.Src.Client)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("[%s] %s", res.Status(), string(bodyBytes))
	}

	return []any{string(bodyBytes)}, nil
}

func (t Tool) ParseParams(data map[string]any, claims map[string]map[string]any) (tools.ParamValues, error) {
	return tools.ParseParams(t.Parameters, data, claims)
}

func (t Tool) Manifest() tools.Manifest {
	return t.manifest
}

func (t Tool) McpManifest() tools.McpManifest {
	return t.mcpManifest
}

func (t Tool) Authorized(verifiedAuthServices []string) bool {
	return tools.IsAuthorized(t.AuthRequired, verifiedAuthServices)
}

// RetrieveIndices extracts the indices from the provided parameters.
func (t Tool) RetrieveIndices(params tools.ParamValues) ([]string, error) {
	paramsMap := params.AsMap()
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
