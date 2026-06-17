package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/figma"
	"github.com/sabujislam/figgen/internal/github"
	"github.com/sabujislam/figgen/internal/logger"
	"github.com/sabujislam/figgen/internal/mcp"
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
		if planFigmaURL == "" {
			planFigmaURL = os.Getenv("FIGMA_URL")
		}
		if planFigmaURL == "" {
			logger.Fatal("FIGMA_URL environment variable or --figma flag is required. Please add FIGMA_URL to your .env file or pass --figma.")
		}

		logger.Step("Starting planning process for Figma URL: %s", planFigmaURL)
		
		cfg, err := config.LoadConfig(planConfigPath)
		if err != nil {
			logger.Fatal("Failed to load configuration: %v", err)
		}

		outDir := globalOutDir
		if cfg.BoilerplateURL != "" {
			err = github.CloneRepository(cfg.BoilerplateURL, outDir)
			if err != nil {
				logger.Fatal("Failed to clone boilerplate: %v", err)
			}
		}

		figmaClient, err := figma.NewClient()
		if err != nil {
			logger.Fatal("Figma setup failed: %v", err)
		}

		fileKey, nodeID, err := figma.ParseURL(planFigmaURL)
		if err != nil {
			logger.Fatal("Invalid Figma URL: %v", err)
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
			logger.Fatal("AI Provider initialization failed: %v", err)
		}

		logger.Step("Connecting to Figma MCP Server for semantic context...")
		mcpClient, err := mcp.NewClient("npx", []string{"-y", "figma-developer-mcp", "--stdio", "--figma-api-key", figmaClient.Token}, nil)
		if err != nil {
			logger.Fatal("Failed to start MCP client: %v", err)
		}
		defer mcpClient.Close()
		
		err = mcpClient.Initialize(ctx)
		if err != nil {
			logger.Fatal("Failed to initialize MCP client: %v", err)
		}

		logger.Step("Fetching semantic design context via MCP...")
		mcpArgs := map[string]interface{}{
			"fileKey": fileKey,
			"depth":   1, // Fetch top-level structural layout first
		}
		if nodeID != "" {
			mcpArgs["nodeId"] = nodeID
		}

		mcpResultBytes, err := mcpClient.Call(ctx, "tools/call", map[string]interface{}{
			"name":      "get_figma_data",
			"arguments": mcpArgs,
		})
		if err != nil {
			logger.Fatal("MCP get_figma_data failed: %v", err)
		}

		var mcpResult struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		}
		if err := json.Unmarshal(mcpResultBytes, &mcpResult); err != nil {
			logger.Fatal("Failed to parse MCP result: %v", err)
		}

		var rawYAML string
		if len(mcpResult.Content) > 0 {
			rawYAML = mcpResult.Content[0].Text
		}

		childIDs, err := figma.ExtractChildIDs(rawYAML)
		if err != nil {
			logger.Info("Failed to extract children IDs from YAML: %v", err)
		}

		// If no children were found, fall back to just processing the selected node ID itself
		targetIDs := childIDs
		if len(targetIDs) == 0 {
			targetIDs = []string{nodeID}
		}

		logger.Step("Identified %d structural segments to plan. Running Architecture Planner...", len(targetIDs))

		masterPlan := &agents.PlannerResponse{}

		for i, id := range targetIDs {
			logger.Info("Planning chunk %d/%d (Node %s)...", i+1, len(targetIDs), id)
			chunkArgs := map[string]interface{}{
				"fileKey": fileKey,
				"depth":   2,
				"nodeId":  id,
			}
			chunkBytes, err := mcpClient.Call(ctx, "tools/call", map[string]interface{}{
				"name":      "get_figma_data",
				"arguments": chunkArgs,
			})
			if err != nil {
				logger.Info("Warning: Failed to fetch chunk %s: %v", id, err)
				continue
			}

			var chunkResult struct {
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			}
			json.Unmarshal(chunkBytes, &chunkResult)

			if len(chunkResult.Content) > 0 {
				chunkRaw := chunkResult.Content[0].Text
				prunedJSON, err := figma.PruneFigmaData(chunkRaw)
				if err != nil {
					prunedJSON = chunkRaw
				}

				plan, err := agents.RunPlanner(ctx, aiClient, cfg, prunedJSON)
				if err != nil {
					logger.Info("Warning: AI Planner failed for chunk %s: %v", id, err)
					continue
				}

				// Merge into master plan
				if plan != nil {
					masterPlan.Components = append(masterPlan.Components, plan.Components...)
					masterPlan.Pages = append(masterPlan.Pages, plan.Pages...)
				}
			}
		}

		err = state.InitState(outDir, fileKey, masterPlan)
		if err != nil {
			logger.Fatal("Failed to save state: %v", err)
		}

		logger.Success("Planning complete! Identified %d components and %d pages.", len(masterPlan.Components), len(masterPlan.Pages))
		logger.Info("Check ./out/.figgen/tasks.md and then execute 'figgen run' to begin generation.")
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().StringVar(&planFigmaURL, "figma", "", "Figma URL to generate from (can also use FIGMA_URL env var)")
	planCmd.Flags().StringVar(&planPage, "page", "", "Specific page name in Figma")
	planCmd.Flags().StringVar(&planConfigPath, "config", "figgen.config.yaml", "Path to configuration file")
	planCmd.Flags().StringVar(&planProvider, "provider", "", "LLM Provider (gemini, openai, anthropic, ollama). Overrides .env")
	planCmd.Flags().StringVar(&planModel, "model", "", "LLM Model Name. Overrides .env")
}
