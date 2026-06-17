package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteComponent writes the generated React component code to the appropriate file path.
func WriteComponent(outDir string, name string, isShadcn bool, code string) (string, error) {
	var targetDir string
	if isShadcn {
		targetDir = filepath.Join(outDir, "src", "components", "ui")
	} else {
		targetDir = filepath.Join(outDir, "src", "components", "common")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	filePath := filepath.Join(targetDir, fmt.Sprintf("%s.tsx", name))
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return filePath, nil
}

// WritePage writes the generated Next.js page code.
func WritePage(outDir string, routeName string, code string) (string, error) {
	// For Next.js app router, routeName like "Dashboard" might map to "src/app/dashboard/page.tsx"
	
	// Default route mapping (e.g., Home -> src/app/page.tsx)
	targetDir := filepath.Join(outDir, "src", "app")
	if routeName != "Home" && routeName != "index" && routeName != "/" {
		targetDir = filepath.Join(targetDir, strings.ToLower(routeName))
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	filePath := filepath.Join(targetDir, "page.tsx")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return filePath, nil
}

// InjectTranslations merges AI generated translations into en.json
func InjectTranslations(outDir string, namespace string, newTranslations map[string]interface{}) error {
	messagesPath := filepath.Join(outDir, "src", "messages", "en.json")
	
	// Create messages directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(messagesPath), 0755); err != nil {
		return err
	}

	// Read existing JSON
	data, err := os.ReadFile(messagesPath)
	if err != nil {
		if os.IsNotExist(err) {
			data = []byte("{}")
		} else {
			return err
		}
	}

	var messages map[string]interface{}
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("invalid json in en.json: %w", err)
	}

	// Helper function to unflatten dot-notation keys and merge into target
	var unflattenAndMerge func(target map[string]interface{}, source map[string]interface{})
	unflattenAndMerge = func(target map[string]interface{}, source map[string]interface{}) {
		for k, v := range source {
			keys := strings.Split(k, ".")
			current := target
			for i := 0; i < len(keys)-1; i++ {
				key := keys[i]
				if current[key] == nil {
					current[key] = make(map[string]interface{})
				}
				if nextMap, ok := current[key].(map[string]interface{}); ok {
					current = nextMap
				} else {
					// Overwrite if it was a string but now needs to be an object
					newMap := make(map[string]interface{})
					current[key] = newMap
					current = newMap
				}
			}
			
			lastKey := keys[len(keys)-1]
			if vMap, ok := v.(map[string]interface{}); ok {
				if current[lastKey] == nil {
					current[lastKey] = make(map[string]interface{})
				}
				if targetMap, ok := current[lastKey].(map[string]interface{}); ok {
					unflattenAndMerge(targetMap, vMap)
				} else {
					current[lastKey] = vMap
				}
			} else {
				current[lastKey] = v
			}
		}
	}

	// Clean up newTranslations by unflattening any dot notations first
	cleanTranslations := make(map[string]interface{})
	unflattenAndMerge(cleanTranslations, newTranslations)

	ns := strings.ToLower(namespace)

	// If the AI already nested everything under the namespace, unwrap it to prevent double nesting.
	if nsVal, ok := cleanTranslations[ns]; ok && len(cleanTranslations) == 1 {
		if nsMap, isMap := nsVal.(map[string]interface{}); isMap {
			cleanTranslations = nsMap
		}
	}

	// Ensure the namespace object exists in the global messages
	if messages[ns] == nil {
		messages[ns] = make(map[string]interface{})
	}
	
	targetMap, ok := messages[ns].(map[string]interface{})
	if !ok {
		targetMap = make(map[string]interface{})
		messages[ns] = targetMap
	}

	// Deep merge the cleaned translations into the target namespace
	unflattenAndMerge(targetMap, cleanTranslations)

	// Write back
	outBytes, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(messagesPath, outBytes, 0644)
}
