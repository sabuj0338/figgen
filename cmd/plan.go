package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/figma"
	"github.com/sabujislam/figgen/internal/github"
	"github.com/sabujislam/figgen/internal/state"
	"github.com/spf13/cobra"
)

var (
	planFigmaURL   string
	planPage       string
	planConfigPath string
	planProvider   string
	planModel      string
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Extract Figma and generate an execution plan (.figgen/tasks.md)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting planning process for Figma URL: %s\n", planFigmaURL)
		
		cfg, err := config.LoadConfig(planConfigPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}

		outDir := "./out" // Default output directory
		if cfg.BoilerplateURL != "" {
			err = github.CloneRepository(cfg.BoilerplateURL, outDir)
			if err != nil {
				log.Fatalf("Failed to clone boilerplate: %v", err)
			}
		}

		figmaClient, err := figma.NewClient()
		if err != nil {
			log.Fatalf("Figma setup failed: %v", err)
		}

		fileKey, nodeID, err := figma.ParseURL(planFigmaURL)
		if err != nil {
			log.Fatalf("Invalid Figma URL: %v", err)
		}

		fmt.Printf("Fetching Figma file %s (node: %s)...\n", fileKey, nodeID)
		fileData, err := figmaClient.FetchFile(fileKey, nodeID)
		if err != nil {
			log.Fatalf("Failed to fetch Figma data: %v", err)
		}

		ctx := context.Background()
		
		if planProvider == "" {
			planProvider = os.Getenv("DEFAULT_PROVIDER")
			if planProvider == "" {
				planProvider = "gemini"
			}
		}
		if planModel == "" {
			planModel = os.Getenv("DEFAULT_MODEL")
		}

		aiClient, err := agents.NewProvider(ctx, planProvider, planModel)
		if err != nil {
			log.Fatalf("AI Provider initialization failed: %v", err)
		}

		// Strip styles to drastically reduce token count to fit within free tier limits
		var stripStyles func(nodes []figma.Node)
		stripStyles = func(nodes []figma.Node) {
			for i := range nodes {
				nodes[i].Style = nil
				if len(nodes[i].Children) > 0 {
					stripStyles(nodes[i].Children)
				}
			}
		}
		stripStyles(fileData.Document.Children)

		fmt.Println("Running Architecture Planner...")
		plan, err := agents.RunPlanner(ctx, aiClient, cfg, fileData)
		if err != nil {
			log.Fatalf("AI Planner failed: %v", err)
		}

		err = state.InitState(outDir, plan)
		if err != nil {
			log.Fatalf("Failed to save state: %v", err)
		}

		fmt.Printf("🎉 Planning complete! Identified %d components and %d pages.\n", len(plan.Components), len(plan.Pages))
		fmt.Println("Check ./out/.figgen/tasks.md and then execute 'figgen run' to begin generation.")
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVar(&planFigmaURL, "figma", "", "Figma URL to generate from (required)")
	planCmd.Flags().StringVar(&planPage, "page", "", "Specific page name in Figma")
	planCmd.Flags().StringVar(&planConfigPath, "config", "figgen.config.yaml", "Path to configuration file")
	planCmd.Flags().StringVar(&planProvider, "provider", "", "LLM Provider (gemini, openai, anthropic, ollama). Overrides .env")
	planCmd.Flags().StringVar(&planModel, "model", "", "LLM Model Name. Overrides .env")
	planCmd.MarkFlagRequired("figma")
}
