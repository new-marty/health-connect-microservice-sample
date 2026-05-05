// Package toolspec converts a Swagger 2.0 (swag-generated) spec into LLM
// tool-calling schemas: OpenAI function-calling and Anthropic tool-use shapes.
//
// The LLM is expected to invoke the actual REST routes directly; we encode the
// (method, path) hint at the end of each tool description so the model knows
// which endpoint to call.
package toolspec

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Convert parses a Swagger 2.0 JSON spec and returns:
//   - OpenAI function-calling tool array: [{type:"function", function:{name, description, parameters}}, ...]
//   - Anthropic tool-use tool array:      [{name, description, input_schema}, ...]
func Convert(specJSON []byte) (openai, anthropic []byte, err error) {
	var spec swaggerSpec
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, nil, fmt.Errorf("parse swagger spec: %w", err)
	}

	tools := buildTools(&spec)

	openaiArr := make([]openaiTool, len(tools))
	anthropicArr := make([]anthropicTool, len(tools))
	for i, t := range tools {
		openaiArr[i] = openaiTool{
			Type: "function",
			Function: openaiFunction{
				Name:        t.name,
				Description: t.description,
				Parameters:  t.schema,
			},
		}
		anthropicArr[i] = anthropicTool{
			Name:        t.name,
			Description: t.description,
			InputSchema: t.schema,
		}
	}

	openai, err = json.MarshalIndent(openaiArr, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	anthropic, err = json.MarshalIndent(anthropicArr, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return openai, anthropic, nil
}

type tool struct {
	name        string
	description string
	schema      map[string]any
}

func buildTools(spec *swaggerSpec) []tool {
	basePath := strings.TrimRight(spec.BasePath, "/")

	type pathOp struct {
		path   string
		method string
		op     *swaggerOperation
	}
	var ops []pathOp
	for path, item := range spec.Paths {
		for method, op := range item.methodMap() {
			if op == nil {
				continue
			}
			ops = append(ops, pathOp{path: path, method: method, op: op})
		}
	}
	// Stable order: by tool name.
	sort.Slice(ops, func(i, j int) bool {
		return toolName(ops[i].op, ops[i].method, ops[i].path) < toolName(ops[j].op, ops[j].method, ops[j].path)
	})

	tools := make([]tool, 0, len(ops))
	for _, po := range ops {
		fullPath := basePath + po.path
		name := toolName(po.op, po.method, po.path)
		tools = append(tools, tool{
			name:        name,
			description: buildDescription(po.op, po.method, fullPath),
			schema:      buildParameterSchema(po.op, spec.Definitions),
		})
	}
	return tools
}

func toolName(op *swaggerOperation, method, path string) string {
	if op != nil && op.OperationID != "" {
		return op.OperationID
	}
	// Synthesize: "get_api_v1_oura_sleep" from method+path.
	clean := strings.TrimPrefix(path, "/")
	clean = strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_").Replace(clean)
	return strings.ToLower(method) + "_" + clean
}

func buildDescription(op *swaggerOperation, method, fullPath string) string {
	var parts []string
	if op.Summary != "" {
		parts = append(parts, op.Summary)
	}
	if op.Description != "" && op.Description != op.Summary {
		parts = append(parts, op.Description)
	}
	if len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%s %s", strings.ToUpper(method), fullPath))
	}
	parts = append(parts, fmt.Sprintf("Invoke: %s %s", strings.ToUpper(method), fullPath))
	return strings.Join(parts, " — ")
}

func buildParameterSchema(op *swaggerOperation, defs map[string]json.RawMessage) map[string]any {
	properties := map[string]any{}
	required := []string{}

	for _, p := range op.Parameters {
		switch p.In {
		case "body":
			// Inline the body schema's properties at the top level.
			body := resolveSchema(p.Schema, defs)
			if bodyProps, ok := body["properties"].(map[string]any); ok {
				for k, v := range bodyProps {
					properties[k] = v
				}
			}
			if bodyReq, ok := body["required"].([]any); ok {
				for _, r := range bodyReq {
					if s, ok := r.(string); ok {
						required = append(required, s)
					}
				}
			}
		case "query", "path", "formData":
			prop := map[string]any{}
			if p.Type != "" {
				prop["type"] = p.Type
			} else {
				prop["type"] = "string"
			}
			if p.Description != "" {
				prop["description"] = p.Description
			}
			if p.Format != "" {
				prop["format"] = p.Format
			}
			if len(p.Enum) > 0 {
				prop["enum"] = p.Enum
			}
			properties[p.Name] = prop
			if p.Required || p.In == "path" {
				required = append(required, p.Name)
			}
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	if len(properties) == 0 {
		// Anthropic + OpenAI both accept an empty object schema for no-arg tools.
		schema["properties"] = map[string]any{}
	}
	return schema
}

func resolveSchema(s json.RawMessage, defs map[string]json.RawMessage) map[string]any {
	if len(s) == 0 {
		return map[string]any{}
	}
	var raw map[string]any
	if err := json.Unmarshal(s, &raw); err != nil {
		return map[string]any{}
	}
	if ref, ok := raw["$ref"].(string); ok {
		const prefix = "#/definitions/"
		if strings.HasPrefix(ref, prefix) {
			defName := strings.TrimPrefix(ref, prefix)
			if def, ok := defs[defName]; ok {
				var resolved map[string]any
				if err := json.Unmarshal(def, &resolved); err == nil {
					return resolved
				}
			}
		}
		return map[string]any{}
	}
	return raw
}

// --- swagger 2.0 minimal schema ---

type swaggerSpec struct {
	BasePath    string                     `json:"basePath"`
	Paths       map[string]swaggerPathItem `json:"paths"`
	Definitions map[string]json.RawMessage `json:"definitions"`
}

type swaggerPathItem struct {
	Get     *swaggerOperation `json:"get,omitempty"`
	Post    *swaggerOperation `json:"post,omitempty"`
	Put     *swaggerOperation `json:"put,omitempty"`
	Patch   *swaggerOperation `json:"patch,omitempty"`
	Delete  *swaggerOperation `json:"delete,omitempty"`
	Options *swaggerOperation `json:"options,omitempty"`
	Head    *swaggerOperation `json:"head,omitempty"`
}

func (i swaggerPathItem) methodMap() map[string]*swaggerOperation {
	return map[string]*swaggerOperation{
		"get":    i.Get,
		"post":   i.Post,
		"put":    i.Put,
		"patch":  i.Patch,
		"delete": i.Delete,
	}
}

type swaggerOperation struct {
	OperationID string             `json:"operationId"`
	Summary     string             `json:"summary"`
	Description string             `json:"description"`
	Tags        []string           `json:"tags"`
	Parameters  []swaggerParameter `json:"parameters"`
}

type swaggerParameter struct {
	Name        string          `json:"name"`
	In          string          `json:"in"`
	Required    bool            `json:"required"`
	Type        string          `json:"type"`
	Format      string          `json:"format"`
	Description string          `json:"description"`
	Enum        []any           `json:"enum"`
	Schema      json.RawMessage `json:"schema"`
}

// --- output shapes ---

type openaiTool struct {
	Type     string         `json:"type"`
	Function openaiFunction `json:"function"`
}

type openaiFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}
