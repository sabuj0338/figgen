package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteComponent writes the generated React component code to the appropriate file path.
func WriteComponent(outDir string, name string, isShadcn bool, code string) error {
	var targetDir string
	if isShadcn {
		targetDir = filepath.Join(outDir, "src", "components", "ui")
	} else {
		targetDir = filepath.Join(outDir, "src", "components", "common")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	filePath := filepath.Join(targetDir, fmt.Sprintf("%s.tsx", name))
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// WritePage writes the generated Next.js page code.
func WritePage(outDir string, routeName string, code string) error {
	// For Next.js app router, routeName like "Dashboard" might map to "src/app/dashboard/page.tsx"
	
	// Default route mapping (e.g., Home -> src/app/page.tsx)
	targetDir := filepath.Join(outDir, "src", "app")
	if routeName != "Home" && routeName != "index" && routeName != "/" {
		targetDir = filepath.Join(targetDir, strings.ToLower(routeName))
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	filePath := filepath.Join(targetDir, "page.tsx")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}
