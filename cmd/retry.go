package cmd

import (
	"github.com/sabujislam/figgen/internal/logger"
	"github.com/sabujislam/figgen/internal/state"
	"github.com/spf13/cobra"
)

var retryIncludeStuck bool

var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Reset failed tasks back to pending so they can be re-run",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := state.LoadState(globalOutDir)
		if err != nil {
			logger.Fatal("Failed to load state: %v. Did you run 'figgen plan' first?", err)
		}

		reset := 0
		for i := range st.Tasks {
			status := st.Tasks[i].Status
			if status == "failed" || (retryIncludeStuck && status == "in_progress") {
				st.Tasks[i].Status = "pending"
				reset++
			}
		}

		if reset == 0 {
			logger.Info("Nothing to retry. No failed tasks found.")
			return
		}

		if err := state.SaveState(globalOutDir, st); err != nil {
			logger.Fatal("Failed to save state: %v", err)
		}

		logger.Success("Reset %d task(s) to pending. Run 'figgen run --all' to continue.", reset)
	},
}

func init() {
	rootCmd.AddCommand(retryCmd)
	retryCmd.Flags().BoolVar(&retryIncludeStuck, "stuck", false, "Also reset tasks stuck in the 'in_progress' state (e.g. after a crash)")
}
