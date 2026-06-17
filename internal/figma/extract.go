package figma

import (
	"gopkg.in/yaml.v3"
)

// ExtractChildIDs parses a depth=1 Figma YAML string and returns a list of child node IDs.
// If the node has no children (e.g. it's a leaf component), it returns an empty slice.
func ExtractChildIDs(rawYAML string) ([]string, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(rawYAML), &data); err != nil {
		return nil, err
	}

	var childIDs []string

	nodes, ok := data["nodes"].([]interface{})
	if !ok || len(nodes) == 0 {
		return childIDs, nil
	}

	firstNode, ok := nodes[0].(map[string]interface{})
	if !ok {
		return childIDs, nil
	}

	children, ok := firstNode["children"].([]interface{})
	if !ok {
		return childIDs, nil
	}

	for _, child := range children {
		if cMap, ok := child.(map[string]interface{}); ok {
			if id, ok := cMap["id"].(string); ok {
				childIDs = append(childIDs, id)
			}
		}
	}

	return childIDs, nil
}

// FindNodeByID recursively searches the JSON tree (map[string]interface{}) for a node with the matching ID.
func FindNodeByID(root interface{}, targetID string) map[string]interface{} {
	switch v := root.(type) {
	case map[string]interface{}:
		if id, ok := v["id"].(string); ok && id == targetID {
			return v
		}
		if children, ok := v["children"].([]interface{}); ok {
			for _, child := range children {
				if found := FindNodeByID(child, targetID); found != nil {
					return found
				}
			}
		}
	case []interface{}:
		for _, item := range v {
			if found := FindNodeByID(item, targetID); found != nil {
				return found
			}
		}
	}
	return nil
}
