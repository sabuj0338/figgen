package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "figgen",
	Short: "Figgen is an AI-powered CLI to convert Figma to production-ready Next.js code",
	Long: `Figgen automates frontend development by analyzing your Figma design,
mapping it to an existing UI library like shadcn/ui, and generating
production-ready Next.js code adhering to strict architecture rules.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
