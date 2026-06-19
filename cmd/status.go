package cmd

import (
	"fmt"
	"sort"

	"github.com/sabujislam/figgen/internal/logger"
	"github.com/sabujislam/figgen/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current generation progress from the state file",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := state.LoadState(globalOutDir)
		if err != nil {
			logger.Fatal("Failed to load state: %v. Did you run 'figgen plan' first?", err)
		}

		if len(st.Tasks) == 0 {
			logger.Info("No tasks found in the state file.")
			return
		}

		counts := map[string]int{}
		for _, t := range st.Tasks {
			counts[t.Status]++
		}

		logger.Info("Progress: %d/%d completed | %d pending | %d in_progress | %d failed",
			counts["completed"], len(st.Tasks), counts["pending"], counts["in_progress"], counts["failed"])

		// Group tasks by category for a readable breakdown.
		categories := make(map[string][]state.Task)
		for _, t := range st.Tasks {
			cat := t.Category
			if cat == "" {
				cat = "Uncategorized"
			}
			categories[cat] = append(categories[cat], t)
		}

		var sortedCats []string
		for cat := range categories {
			sortedCats = append(sortedCats, cat)
		}
		sort.Strings(sortedCats)

		for _, cat := range sortedCats {
			fmt.Printf("\n%s\n", cat)
			for _, t := range categories[cat] {
				fmt.Printf("  %s %-9s %s\n", statusGlyph(t.Status), t.Type, t.Name)
			}
		}

		if counts["failed"] > 0 {
			fmt.Println()
			logger.Warn("%d task(s) failed. Run 'figgen retry' to reset them, then 'figgen run --all'.", counts["failed"])
		}
	},
}

func statusGlyph(status string) string {
	switch status {
	case "completed":
		return "[x]"
	case "in_progress":
		return "[/]"
	case "failed":
		return "[!]"
	default:
		return "[ ]"
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
