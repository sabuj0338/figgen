package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CloneRepository clones the boilerplate URL to the target directory.
func CloneRepository(url string, destPath string) error {
	// Ensure destPath exists
	if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Check if boilerplate is already cloned (e.g., package.json exists)
	packageJsonPath := filepath.Join(destPath, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		fmt.Printf("Boilerplate already exists in %s (package.json found). Skipping clone.\n", destPath)
		return nil
	}

	fmt.Printf("Cloning boilerplate from %s to %s...\n", url, destPath)

	// Create temp dir in the same parent directory to ensure os.Rename works
	parentDir := filepath.Dir(destPath)
	tempDir, err := os.MkdirTemp(parentDir, "figgen-clone-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("git", "clone", url, tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Move contents from tempDir to destPath
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		
		srcPath := filepath.Join(tempDir, entry.Name())
		dstPath := filepath.Join(destPath, entry.Name())
		
		// If the destination already exists (e.g. .figgen), we skip replacing it or we can remove it.
		// Usually, the boilerplate won't have .figgen, but let's be safe.
		if _, err := os.Stat(dstPath); err == nil {
			if entry.Name() == ".figgen" {
				continue // Keep existing .figgen state
			}
			os.RemoveAll(dstPath) // Overwrite other existing files/folders
		}

		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
		}
	}

	return nil
}
