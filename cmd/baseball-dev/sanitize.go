package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func sanitizeOpenAPI(inputPath, outputPath string) error {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return err
	}

	sanitizeSchema(doc)
	sanitizePaths(doc)

	out, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, out, 0o644); err != nil {
		return err
	}

	fmt.Printf("sanitized OpenAPI written to %s\n", outputPath)
	return nil
}

func sanitizeSchema(node any) {
	switch v := node.(type) {
	case map[string]any:
		if ap, ok := v["additionalProperties"]; ok {
			if _, isBool := ap.(bool); isBool {
				v["additionalProperties"] = map[string]any{"type": "string"}
			}
		}
		for _, child := range v {
			sanitizeSchema(child)
		}
	case []any:
		for _, child := range v {
			sanitizeSchema(child)
		}
	}
}

func sanitizePaths(doc map[string]any) {
	pathsRaw, ok := doc["paths"]
	if !ok {
		return
	}
	paths, ok := pathsRaw.(map[string]any)
	if !ok {
		return
	}

	methods := map[string]struct{}{
		"get": {}, "put": {}, "post": {}, "delete": {},
		"options": {}, "head": {}, "patch": {}, "trace": {},
	}

	for _, pathItemRaw := range paths {
		pathItem, ok := pathItemRaw.(map[string]any)
		if !ok {
			continue
		}

		sanitizeParameterList(pathItem["parameters"])

		for method, operationRaw := range pathItem {
			if _, ok := methods[strings.ToLower(method)]; !ok {
				continue
			}
			operation, ok := operationRaw.(map[string]any)
			if !ok {
				continue
			}
			sanitizeParameterList(operation["parameters"])
		}
	}
}

func sanitizeParameterList(paramsRaw any) {
	params, ok := paramsRaw.([]any)
	if !ok {
		return
	}
	for _, paramRaw := range params {
		param, ok := paramRaw.(map[string]any)
		if !ok {
			continue
		}
		schemaRaw, ok := param["schema"]
		if !ok {
			continue
		}
		schema, ok := schemaRaw.(map[string]any)
		if !ok {
			continue
		}
		enumRaw, ok := schema["enum"]
		if !ok {
			continue
		}
		enumValues, ok := enumRaw.([]any)
		if !ok {
			continue
		}
		nonString := false
		for _, v := range enumValues {
			if _, isString := v.(string); !isString {
				nonString = true
				break
			}
		}
		if nonString {
			delete(schema, "enum")
		}
	}
}
