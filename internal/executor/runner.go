package executor

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/sabujislam/figgen/internal/logger"
)

// InstallDependencies runs the package manager install command
func InstallDependencies(outDir string, pkgManager string, deps []string) error {
	if len(deps) == 0 {
		return nil
	}

	logger.Step("Installing dependencies: %s", strings.Join(deps, ", "))

	cmdArgs := []string{"install"}
	if pkgManager == "npm" {
		cmdArgs = []string{"install", "--save"} // Standard fallback
	} else if pkgManager == "pnpm" {
		cmdArgs = []string{"add"}
	} else if pkgManager == "yarn" {
		cmdArgs = []string{"add"}
	} else if pkgManager == "bun" {
		cmdArgs = []string{"add"}
	}

	cmdArgs = append(cmdArgs, deps...)

	cmd := exec.Command(pkgManager, cmdArgs...)
	cmd.Dir = outDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install dependencies (%s): %v\nOutput: %s", pkgManager, err, string(output))
	}
	
	logger.Success("Installed dependencies successfully!")
	return nil
}

// InstallShadcn runs the shadcn add command
func InstallShadcn(outDir string, pkgManager string, comps []string) error {
	if len(comps) == 0 {
		return nil
	}

	logger.Step("Installing shadcn/ui components: %s", strings.Join(comps, ", "))

	// We use dlx/npx to ensure we hit the latest shadcn-ui without global install
	var runner string
	if pkgManager == "npm" || pkgManager == "yarn" {
		runner = "npx"
	} else if pkgManager == "pnpm" {
		runner = "pnpm" // pnpm dlx
	} else if pkgManager == "bun" {
		runner = "bunx"
	} else {
		runner = "npx"
	}

	cmdArgs := []string{}
	if runner == "pnpm" {
		cmdArgs = append(cmdArgs, "dlx")
	}
	cmdArgs = append(cmdArgs, "shadcn-ui@latest", "add")
	cmdArgs = append(cmdArgs, comps...)
	
	// Bypass interactive prompts
	cmdArgs = append(cmdArgs, "--yes", "--overwrite")

	cmd := exec.Command(runner, cmdArgs...)
	cmd.Dir = outDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install shadcn components: %v\nOutput: %s", err, string(output))
	}

	logger.Success("Installed shadcn/ui components successfully!")
	return nil
}

// LintFile formats the generated file
func LintFile(outDir string, pkgManager string, filePath string) error {
	logger.Step("Formatting code with prettier: %s", filePath)

	var runner string
	if pkgManager == "bun" {
		runner = "bunx"
	} else {
		runner = "npx" // fallback to npx for prettier
	}

	cmd := exec.Command(runner, "prettier", "--write", filePath)
	cmd.Dir = outDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("prettier formatting failed: %v\nOutput: %s", err, string(output))
	}

	logger.Success("Code formatted successfully!")
	return nil
}
