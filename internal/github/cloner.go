package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CloneRepository clones the boilerplate URL to the target directory.
func CloneRepository(url string, destPath string) error {
	// Check if directory already exists
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		fmt.Printf("Destination directory %s already exists. Skipping clone.\n", destPath)
		return nil
	}

	fmt.Printf("Cloning boilerplate from %s to %s...\n", url, destPath)

	cmd := exec.Command("git", "clone", url, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Remove the .git folder so the user can initialize their own repo
	gitPath := filepath.Join(destPath, ".git")
	if err := os.RemoveAll(gitPath); err != nil {
		fmt.Printf("Warning: failed to remove .git directory: %v\n", err)
	}

	return nil
}
