package figma

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// PruneFigmaData aggressively strips out heavy layout and style data
// from the raw Figma YAML/JSON returned by the MCP server, to save LLM tokens.
func PruneFigmaData(raw string) (string, error) {
	var data interface{}
	if err := yaml.Unmarshal([]byte(raw), &data); err != nil {
		return "", err
	}

	data = walkAndPrune(data)

	// Marshal without indent to minify payload
	minified, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(minified), nil
}

func walkAndPrune(node interface{}) interface{} {
	switch v := node.(type) {
	case map[string]interface{}:
		// Remove keys that consume large amounts of tokens without adding architectural value
		keysToRemove := []string{
			"style", "fills", "strokes", "effects",
			"absoluteBoundingBox", "absoluteRenderBounds",
			"geometry", "fillGeometry", "strokeGeometry",
			"blendMode", "exportSettings", "constraints",
			"transitionNodeID", "transitionDuration", "transitionEasing",
			"preserveRatio", "layoutAlign", "layoutGrow",
			"css", // sometimes injected by MCP plugins
		}

		for _, k := range keysToRemove {
			delete(v, k)
		}

		// Keep essential keys like 'id', 'name', 'type', 'children', 'characters' (for text)
		
		// Recursively prune children if they exist
		for key, val := range v {
			v[key] = walkAndPrune(val)
		}
		return v

	case []interface{}:
		for i, item := range v {
			v[i] = walkAndPrune(item)
		}
		return v

	case string:
		// Clean up massive strings just in case
		if len(v) > 500 {
			return v[:500] + "...(truncated)"
		}
		return strings.TrimSpace(v)

	default:
		return v
	}
}

// PruneForCoder is less aggressive than PruneFigmaData.
// It KEEPS styles, fills, and layout constraints which are necessary
// for the AI Coder to write accurate CSS. It still DELETES massive
// vector paths and bounding box coordinates to save tokens.
func PruneForCoder(raw string) (string, error) {
	var data interface{}
	if err := yaml.Unmarshal([]byte(raw), &data); err != nil {
		return "", err
	}

	data = walkAndPruneForCoder(data)

	// Marshal without indent to minify payload
	minified, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(minified), nil
}

func walkAndPruneForCoder(node interface{}) interface{} {
	switch v := node.(type) {
	case map[string]interface{}:
		// Only remove math/vector data that wastes tokens. KEEP styles/fills.
		keysToRemove := []string{
			"absoluteBoundingBox", "absoluteRenderBounds",
			"geometry", "fillGeometry", "strokeGeometry",
			"exportSettings", "blendMode",
		}

		for _, k := range keysToRemove {
			delete(v, k)
		}

		for key, val := range v {
			v[key] = walkAndPruneForCoder(val)
		}
		return v

	case []interface{}:
		for i, item := range v {
			v[i] = walkAndPruneForCoder(item)
		}
		return v

	case string:
		if len(v) > 500 {
			return v[:500] + "...(truncated)"
		}
		return strings.TrimSpace(v)

	default:
		return v
	}
}

