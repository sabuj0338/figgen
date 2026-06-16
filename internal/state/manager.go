package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sabujislam/figgen/internal/agents"
)

type Task struct {
	ID        string `json:"id"`
	Type      string `json:"type"`   // "component" or "page"
	Name      string `json:"name"`
	Status    string `json:"status"` // "pending", "in_progress", "completed", "failed"
	IsShadcn  bool   `json:"is_shadcn,omitempty"`
	
	// Payload keeps the original plan
	ComponentPlan *agents.ComponentPlan `json:"component_plan,omitempty"`
	PagePlan      *agents.PagePlan      `json:"page_plan,omitempty"`
}

type State struct {
	Tasks []Task `json:"tasks"`
}

// InitState takes the PlannerResponse and creates the initial state files
func InitState(outDir string, plan *agents.PlannerResponse) error {
	stateDir := filepath.Join(outDir, ".figgen")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	var state State
	for i, comp := range plan.Components {
		c := comp // copy
		state.Tasks = append(state.Tasks, Task{
			ID:            fmt.Sprintf("comp_%d", i),
			Type:          "component",
			Name:          comp.Name,
			Status:        "pending",
			IsShadcn:      comp.IsShadcn,
			ComponentPlan: &c,
		})
	}

	for i, page := range plan.Pages {
		p := page // copy
		state.Tasks = append(state.Tasks, Task{
			ID:       fmt.Sprintf("page_%d", i),
			Type:     "page",
			Name:     page.Name,
			Status:   "pending",
			PagePlan: &p,
		})
	}

	return SaveState(outDir, &state)
}

// LoadState reads the current task state
func LoadState(outDir string) (*State, error) {
	path := filepath.Join(outDir, ".figgen", "tasks.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	return &state, nil
}

// SaveState saves both the JSON and the human-readable Markdown tracker
func SaveState(outDir string, state *State) error {
	stateDir := filepath.Join(outDir, ".figgen")
	
	// Write JSON
	jsonPath := filepath.Join(stateDir, "tasks.json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return err
	}

	// Write Markdown tracker
	mdPath := filepath.Join(stateDir, "tasks.md")
	mdContent := "# Figma Generation Tasks\n\n"
	
	mdContent += "## Components\n"
	for _, t := range state.Tasks {
		if t.Type == "component" {
			check := " "
			if t.Status == "completed" {
				check = "x"
			} else if t.Status == "in_progress" {
				check = "/"
			} else if t.Status == "failed" {
				check = "!"
			}
			mdContent += fmt.Sprintf("- [%s] Component: %s\n", check, t.Name)
		}
	}

	mdContent += "\n## Pages\n"
	for _, t := range state.Tasks {
		if t.Type == "page" {
			check := " "
			if t.Status == "completed" {
				check = "x"
			} else if t.Status == "in_progress" {
				check = "/"
			} else if t.Status == "failed" {
				check = "!"
			}
			mdContent += fmt.Sprintf("- [%s] Page: %s\n", check, t.Name)
		}
	}

	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		return err
	}

	return nil
}
