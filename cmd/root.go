package cmd

import (
	"fmt"
	"os"

	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/figma"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "figgen",
	Short: "Figgen is an AI-powered CLI to convert Figma to production-ready Next.js code",
	Long: `Figgen automates frontend development by analyzing your Figma design,
mapping it to an existing UI library like shadcn/ui, and generating
production-ready Next.js code adhering to strict architecture rules.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		agents.Debug = globalDebug
		figma.MaxChildrenPerNode = globalMaxChildren
	},
}

var (
	globalOutDir      string
	globalDebug       bool
	globalMaxChildren int
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&globalOutDir, "out", "o", "./out", "Output directory for the generated Next.js project")
	rootCmd.PersistentFlags().BoolVar(&globalDebug, "debug", false, "Write raw model responses to disk for troubleshooting")
	rootCmd.PersistentFlags().IntVar(&globalMaxChildren, "max-children", 0, "Max child nodes kept per Figma node when pruning (0 = unlimited)")
}
