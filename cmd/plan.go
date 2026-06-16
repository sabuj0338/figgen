package cmd

import (
	"context"
	"os"

	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/figma"
	"github.com/sabujislam/figgen/internal/github"
	"github.com/sabujislam/figgen/internal/logger"
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

		outDir := "./out" // Default output directory
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

		logger.Step("Fetching Figma file %s (node: %s)...", fileKey, nodeID)
		fileData, err := figmaClient.FetchFile(fileKey, nodeID)
		if err != nil {
			logger.Fatal("Failed to fetch Figma data: %v", err)
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

		// Extract images before compression
		logger.Step("Extracting images from Figma...")
		var imageNodes []string
		var findImages func(nodes []figma.Node)
		findImages = func(nodes []figma.Node) {
			for _, n := range nodes {
				isImage := false
				for _, fill := range n.Fills {
					if fill.Type == "IMAGE" {
						isImage = true
						break
					}
				}
				if isImage && n.ID != "" {
					imageNodes = append(imageNodes, n.ID)
				}
				if len(n.Children) > 0 {
					findImages(n.Children)
				}
			}
		}
		findImages(fileData.Document.Children)

		if len(imageNodes) > 0 {
			logger.Info("Found %d image nodes, downloading...", len(imageNodes))
			// Fetch URLs
			imgURLs, err := figmaClient.FetchImageURLs(fileKey, imageNodes)
			if err != nil {
				logger.Warn("Failed to fetch image URLs: %v", err)
			} else {
				// Download them
				err = figmaClient.DownloadImages(imgURLs, outDir+"/public/images")
				if err != nil {
					logger.Warn("Failed to download images: %v", err)
				} else {
					logger.Success("Images downloaded successfully to %s/public/images", outDir)
				}
			}
		}

		// Aggressively compress Figma JSON to save tokens
		var compressFigmaTree func(nodes []figma.Node) []figma.Node
		compressFigmaTree = func(nodes []figma.Node) []figma.Node {
			var compressed []figma.Node
			for _, n := range nodes {
				// Skip purely decorative/vector nodes that don't represent structural UI
				if n.Type == "VECTOR" || n.Type == "STAR" || n.Type == "LINE" || n.Type == "ELLIPSE" || n.Type == "REGULAR_POLYGON" {
					continue
				}
				
				// Strip ID to save tokens
				n.ID = ""
				
				// Keep only crucial typography hints from Style, discard the rest
				if n.Style != nil {
					filteredStyle := make(map[string]interface{})
					if fontSize, ok := n.Style["fontSize"]; ok {
						filteredStyle["fontSize"] = fontSize
					}
					if fontWeight, ok := n.Style["fontWeight"]; ok {
						filteredStyle["fontWeight"] = fontWeight
					}
					if textAlign, ok := n.Style["textAlignHorizontal"]; ok {
						filteredStyle["textAlignHorizontal"] = textAlign
					}
					
					if len(filteredStyle) > 0 {
						n.Style = filteredStyle
					} else {
						n.Style = nil
					}
				}
				if len(n.Children) > 0 {
					n.Children = compressFigmaTree(n.Children)
				}
				compressed = append(compressed, n)
			}
			return compressed
		}
		
		fileData.Document.Children = compressFigmaTree(fileData.Document.Children)

		logger.Step("Running Architecture Planner...")
		plan, err := agents.RunPlanner(ctx, aiClient, cfg, fileData)
		if err != nil {
			logger.Fatal("AI Planner failed: %v", err)
		}

		err = state.InitState(outDir, plan)
		if err != nil {
			logger.Fatal("Failed to save state: %v", err)
		}

		logger.Success("Planning complete! Identified %d components and %d pages.", len(plan.Components), len(plan.Pages))
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
