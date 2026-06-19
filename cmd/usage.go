package cmd

import (
	"fmt"

	"github.com/sabujislam/figgen/internal/logger"
	"github.com/sabujislam/figgen/internal/telemetry"
	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show recorded LLM token usage for this project",
	Run: func(cmd *cobra.Command, args []string) {
		summary, _, err := telemetry.Load(globalOutDir)
		if err != nil {
			logger.Fatal("No usage data found: %v. Run 'figgen plan' or 'figgen run' first.", err)
		}

		total := summary.InputTokens + summary.OutputTokens
		logger.Info("LLM usage across %d call(s):", summary.Calls)
		fmt.Printf("  Input tokens:  %d\n", summary.InputTokens)
		fmt.Printf("  Output tokens: %d\n", summary.OutputTokens)
		fmt.Printf("  Cached (read): %d\n", summary.CachedTokens)
		fmt.Printf("  Total tokens:  %d\n", total)

		if len(summary.ByStage) > 0 {
			fmt.Println("\n  By stage (input+output):")
			for stage, tokens := range summary.ByStage {
				fmt.Printf("    %-6s %d\n", stage, tokens)
			}
		}

		if summary.InputTokens > 0 {
			pct := float64(summary.CachedTokens) / float64(summary.InputTokens) * 100
			fmt.Printf("\n  Cache hit rate (of input): %.1f%%\n", pct)
		}
	},
}

func init() {
	rootCmd.AddCommand(usageCmd)
}
