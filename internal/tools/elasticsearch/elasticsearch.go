// Copyright 2024 Google LLC
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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"io"

	yaml "github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/sources"
	es "github.com/googleapis/genai-toolbox/internal/sources/elasticsearch"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

const kind string = "elasticsearch-search"

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
	Name        string           `yaml:"name" validate:"required"`
	Kind        string           `yaml:"kind" validate:"required"`
	Source      string           `yaml:"source" validate:"required"`
	Description string           `yaml:"description" validate:"required"`
	Index       string           `yaml:"index" validate:"required"`
	Query       map[string]any   `yaml:"query" validate:"required"`
	Parameters  tools.Parameters `yaml:"parameters"`
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
	Name         string
	Kind         string
	Description  string
	Index        string
	Query        map[string]any
	Parameters   tools.Parameters
	AuthRequired []string
	manifest     tools.Manifest
	mcpManifest  tools.McpManifest
	Src          *es.Source
}

var _ tools.Tool = &Tool{}

func (c Config) Initialize(srcs map[string]sources.Source) (tools.Tool, error) {
	src, ok := srcs[c.Source]
	if !ok {
		return nil, fmt.Errorf("source %q not found", c.Source)
	}
	esSrc, ok := src.(*es.Source)
	if !ok {
		return nil, fmt.Errorf("source %q is not elasticsearch", c.Source)
	}
	return &Tool{
		Name:        c.Name,
		Kind:        kind,
		Description: c.Description,
		Index:       c.Index,
		Query:       c.Query,
		Parameters:  c.Parameters,
		Src:         esSrc,
	}, nil
}

func (t *Tool) Invoke(ctx context.Context, params tools.ParamValues) ([]any, error) {
	// Marshal query to JSON
	body, err := json.Marshal(t.Query)
	if err != nil {
		return nil, err
	}

	res, err := esapi.SearchRequest{
		Index: []string{t.Index},
		Body:  bytes.NewReader(body),
	}.Do(ctx, t.Src.Client)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return []any{string(bodyBytes)}, nil
}

func (t *Tool) ParseParams(data map[string]any, claims map[string]map[string]any) (tools.ParamValues, error) {
	// For now, just return the input data as params
	return nil, nil
}

func (t *Tool) Manifest() tools.Manifest {
	return t.manifest
}

func (t *Tool) McpManifest() tools.McpManifest {
	return t.mcpManifest
}

func (t *Tool) Authorized(verifiedAuthServices []string) bool {
	// For now, always authorized (customize as needed)
	return true
}
