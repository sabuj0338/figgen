package figma

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"gopkg.in/yaml.v3"
)

// PruneFigmaData aggressively strips out heavy layout and style data
// from the raw Figma YAML/JSON returned by the MCP server, to save LLM tokens.
func PruneFigmaData(raw string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		if err := yaml.Unmarshal([]byte(raw), &data); err != nil {
			return "", err
		}
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
		keysToRemove := []string{
			"style", "fills", "strokes", "effects",
			"absoluteBoundingBox", "absoluteRenderBounds",
			"geometry", "fillGeometry", "strokeGeometry",
			"blendMode", "exportSettings", "constraints",
			"transitionNodeID", "transitionDuration", "transitionEasing",
			"preserveRatio", "layoutAlign", "layoutGrow",
			"css",
		}

		for _, k := range keysToRemove {
			delete(v, k)
		}

		for key, val := range v {
			if arr, ok := val.([]interface{}); ok && len(arr) == 0 {
				delete(v, key)
			} else if str, ok := val.(string); ok && (str == "" || (key == "layoutMode" && str == "NONE")) {
				delete(v, key)
			} else if val == nil {
				delete(v, key)
			} else {
				v[key] = walkAndPrune(val)
			}
		}
		return v

	case []interface{}:
		var newList []interface{}
		for _, item := range v {
			pruned := walkAndPrune(item)
			if pruned != nil {
				newList = append(newList, pruned)
			}
		}
		return newList

	case string:
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

	minified, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(minified), nil
}

func walkAndPruneForCoder(node interface{}) interface{} {
	switch v := node.(type) {
	case map[string]interface{}:
		// Check if it's a color map
		r, rok := getFloat(v, "r")
		g, gok := getFloat(v, "g")
		b, bok := getFloat(v, "b")
		if rok && gok && bok {
			return rgbaToHex(r, g, b)
		}

		keysToRemove := []string{
			"absoluteBoundingBox", "absoluteRenderBounds",
			"geometry", "fillGeometry", "strokeGeometry",
			"exportSettings", "blendMode",
		}

		for _, k := range keysToRemove {
			delete(v, k)
		}

		if styleMap, ok := v["style"].(map[string]interface{}); ok {
			tailwind := mapStyleToTailwind(styleMap)
			if tailwind != "" {
				v["tailwind_text"] = tailwind
			}
			delete(v, "style")
		}

		for key, val := range v {
			if arr, ok := val.([]interface{}); ok && len(arr) == 0 {
				delete(v, key)
			} else if str, ok := val.(string); ok && (str == "" || (key == "layoutMode" && str == "NONE")) {
				delete(v, key)
			} else if val == nil {
				delete(v, key)
			} else {
				v[key] = walkAndPruneForCoder(val)
			}
		}
		return v

	case []interface{}:
		var newList []interface{}
		for _, item := range v {
			pruned := walkAndPruneForCoder(item)
			if pruned != nil {
				newList = append(newList, pruned)
			}
		}
		return newList

	case string:
		if len(v) > 500 {
			return v[:500] + "...(truncated)"
		}
		return strings.TrimSpace(v)

	default:
		return v
	}
}

func getFloat(m map[string]interface{}, key string) (float64, bool) {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f, true
		}
		if i, ok := val.(int); ok {
			return float64(i), true
		}
	}
	return 0, false
}

func rgbaToHex(r, g, b float64) string {
	ir := int(math.Round(r * 255))
	ig := int(math.Round(g * 255))
	ib := int(math.Round(b * 255))
	return fmt.Sprintf("#%02X%02X%02X", ir, ig, ib)
}

func mapStyleToTailwind(style map[string]interface{}) string {
	var classes []string

	fs, _ := getFloat(style, "fontSize")
	if fs > 0 {
		classes = append(classes, fmt.Sprintf("text-[%dpx]", int(fs)))
	}

	fw, _ := getFloat(style, "fontWeight")
	if fw > 0 {
		if fw == 400 {
			classes = append(classes, "font-normal")
		} else if fw == 500 {
			classes = append(classes, "font-medium")
		} else if fw == 600 {
			classes = append(classes, "font-semibold")
		} else if fw == 700 {
			classes = append(classes, "font-bold")
		} else {
			classes = append(classes, fmt.Sprintf("font-[%d]", int(fw)))
		}
	}

	lh, _ := getFloat(style, "lineHeightPx")
	if lh > 0 {
		classes = append(classes, fmt.Sprintf("leading-[%dpx]", int(lh)))
	}

	ls, _ := getFloat(style, "letterSpacing")
	if ls != 0 {
		classes = append(classes, fmt.Sprintf("tracking-[%.2fpx]", ls))
	}

	return strings.Join(classes, " ")
}
