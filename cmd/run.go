package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/filesystem"
	"github.com/sabujislam/figgen/internal/executor"
	"github.com/sabujislam/figgen/internal/figma"
	"github.com/sabujislam/figgen/internal/github"
	"github.com/sabujislam/figgen/internal/logger"
	"github.com/sabujislam/figgen/internal/state"
	"github.com/spf13/cobra"
)

var (
	runAll        bool
	runConfigPath string
	runProvider   string
	runModel      string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute tasks from the state file",
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteRun(globalOutDir, runAll, runConfigPath, runProvider, runModel)
	},
}

func ExecuteRun(outDir string, isAll bool, configPath string, provider string, model string) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	if cfg.BoilerplateURL != "" {
		err = github.CloneRepository(cfg.BoilerplateURL, outDir)
		if err != nil {
			logger.Fatal("Failed to clone boilerplate: %v", err)
		}
	}

	ctx := context.Background()

	if provider == "" {
		provider = os.Getenv("DEFAULT_PROVIDER")
		if provider == "" {
			provider = "gemini"
		}
	}
	if model == "" {
		model = os.Getenv("DEFAULT_MODEL")
	}

	aiClient, err := agents.NewProvider(ctx, provider, model)
	if err != nil {
		logger.Fatal("AI Provider initialization failed: %v", err)
	}

	figmaClient, _ := figma.NewClient()

	for {
		st, err := state.LoadState(outDir)
		if err != nil {
			logger.Fatal("Failed to load state: %v. Did you run 'figgen plan' first?", err)
		}

		var targetTask *state.Task
		var taskIndex int
		for i, t := range st.Tasks {
			if t.Status == "pending" {
				targetTask = &st.Tasks[i]
				taskIndex = i
				break
			}
		}

		if targetTask == nil {
			logger.Success("All tasks completed!")
			break
		}

		logger.Step("Executing task: %s [%s]", targetTask.Name, targetTask.Type)

		// 1. Mark as in-progress
		st.Tasks[taskIndex].Status = "in_progress"
		state.SaveState(outDir, st)

		// Extract images and fetch compressed context if FigmaNodeID is present
		var availableImages []string
		var mcpContext string

		// 1. Build MCP Context for this node locally from figma_context.json
		if targetTask.FigmaNodeID != "" {
			contextFile := outDir + "/.figgen/figma_context.json"
			contextData, err := os.ReadFile(contextFile)
			if err == nil {
				var root map[string]interface{}
				if json.Unmarshal(contextData, &root) == nil {
					foundNode := figma.FindNodeByID(root, targetTask.FigmaNodeID)
					if foundNode != nil {
						nodeBytes, _ := json.Marshal(foundNode)
						prunedCtx, err := figma.PruneForCoder(string(nodeBytes))
						if err == nil {
							mcpContext = prunedCtx
						} else {
							mcpContext = string(nodeBytes)
						}

						// 2. Extremely simple local recursive image finder on map[string]interface{}
						var imageNodes []string
						var svgNodes []string
						var findImages func(node map[string]interface{})
						findImages = func(node map[string]interface{}) {
							isImage := false
							if fills, ok := node["fills"].([]interface{}); ok {
								for _, fillRaw := range fills {
									if fill, ok := fillRaw.(map[string]interface{}); ok {
										if fillType, ok := fill["type"].(string); ok && fillType == "IMAGE" {
											isImage = true
											break
										}
									}
								}
							}

							nodeID, _ := node["id"].(string)

							// Prevent the main target node itself from being downloaded as an image
							if isImage && nodeID != "" && nodeID != targetTask.FigmaNodeID {
								imageNodes = append(imageNodes, nodeID)
							}

							if children, ok := node["children"].([]interface{}); ok {
								for _, childRaw := range children {
									if child, ok := childRaw.(map[string]interface{}); ok {
										findImages(child)
									}
								}
							}
						}

						findImages(foundNode)

						// If the task name suggests it is an icon or logo, export it as an SVG.
						nameLower := strings.ToLower(targetTask.Name)
						if strings.Contains(nameLower, "icon") || strings.Contains(nameLower, "logo") {
							if targetTask.FigmaNodeID != "" {
								svgNodes = append(svgNodes, targetTask.FigmaNodeID)
							}
						} else if strings.Contains(nameLower, "illustration") || strings.Contains(nameLower, "image") || strings.Contains(nameLower, "graphic") {
							// If it's an illustration, image, or graphic, prefer PNG
							if targetTask.FigmaNodeID != "" {
								imageNodes = append(imageNodes, targetTask.FigmaNodeID)
							}
						}

						// Now download PNG images
						if len(imageNodes) > 0 && figmaClient != nil && st.FigmaFileKey != "" && st.FigmaFileKey != "local" {
							logger.Step("Found %d image nodes, fetching PNGs from Figma...", len(imageNodes))
							imgURLs, err := figmaClient.FetchImageURLs(st.FigmaFileKey, imageNodes, "png")
							if err == nil {
								imgDir := outDir + "/public/images"
								err = figmaClient.DownloadImages(imgURLs, imgDir, "png")
								if err == nil {
									for id := range imgURLs {
										safeID := strings.ReplaceAll(id, ":", "_")
										availableImages = append(availableImages, safeID+".png")
									}
								}
							} else {
								logger.Warn("Failed to fetch image URLs: %v", err)
							}
						}

						// Now download SVG images (icons, illustrations, logos)
						if len(svgNodes) > 0 && figmaClient != nil && st.FigmaFileKey != "" && st.FigmaFileKey != "local" {
							logger.Step("Found %d vector nodes, fetching SVGs from Figma...", len(svgNodes))
							svgURLs, err := figmaClient.FetchImageURLs(st.FigmaFileKey, svgNodes, "svg")
							if err == nil {
								imgDir := outDir + "/public/images"
								err = figmaClient.DownloadImages(svgURLs, imgDir, "svg")
								if err == nil {
									for id := range svgURLs {
										safeID := strings.ReplaceAll(id, ":", "_")
										availableImages = append(availableImages, safeID+".svg")
									}
								}
							} else {
								logger.Warn("Failed to fetch SVG URLs: %v", err)
							}
						}
					}
				}
			}
		}

		// 3. Generate Code
		var codeResp *agents.CoderResponse
		var targetFilePath string

		if targetTask.Type == "component" {
			if targetTask.IsShadcn {
				// Skip AI generation for shadcn components, just install it
				codeResp = &agents.CoderResponse{
					ShadcnComponents: []string{strings.ToLower(targetTask.Name)},
					Translations:     make(map[string]interface{}),
				}
				targetFilePath = "" // No specific file to format since it's shadcn
			} else {
				codeResp, err = agents.RunCoderForComponent(ctx, aiClient, cfg, *targetTask.ComponentPlan, availableImages, mcpContext)
				if err == nil {
					targetFilePath, err = filesystem.WriteComponent(outDir, targetTask.Name, targetTask.IsShadcn, codeResp.Code)
				}
			}
		} else {
			codeResp, err = agents.RunCoderForPage(ctx, aiClient, cfg, *targetTask.PagePlan, availableImages, mcpContext)
			if err == nil {
				targetFilePath, err = filesystem.WritePage(outDir, targetTask.Name, codeResp.Code)
			}
		}

		// Post-Generation Processing (Executor)
		if err == nil && codeResp != nil {
			_ = executor.InstallDependencies(outDir, cfg.PackageManager, codeResp.Dependencies)
			_ = executor.InstallShadcn(outDir, cfg.PackageManager, codeResp.ShadcnComponents)
			if targetFilePath != "" {
				_ = executor.LintFile(outDir, cfg.PackageManager, targetFilePath)
			}

			// Inject Translations into en.json
			if len(codeResp.Translations) > 0 {
				errTrans := filesystem.InjectTranslations(outDir, targetTask.Name, codeResp.Translations)
				if errTrans != nil {
					logger.Warn("Failed to inject translations for %s: %v", targetTask.Name, errTrans)
				}
			}
		}

		if err != nil {
			logger.Error("Task %s failed: %v", targetTask.Name, err)
			st.Tasks[taskIndex].Status = "failed"
			state.SaveState(outDir, st)

			if !isAll {
				break
			}

			logger.Prompt("Task failed. Do you want to continue executing the remaining tasks? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				logger.Warn("Stopping execution.")
				break
			}
		} else {
			logger.Success("Task %s completed!", targetTask.Name)
			st.Tasks[taskIndex].Status = "completed"
			state.SaveState(outDir, st)
		}

		if !isAll {
			logger.Info("Run 'figgen run' again to execute the next task, or 'figgen run --all' to execute all.")
			break
		}

		// Respect Gemini free tier limits (15 Requests Per Minute)
		if provider == "gemini" {
			logger.Info("Waiting 4 seconds to respect Gemini API free-tier rate limits...")
			time.Sleep(4 * time.Second)
		}
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&runAll, "all", false, "Run all pending tasks continuously")
	runCmd.Flags().StringVar(&runConfigPath, "config", "figgen.config.yaml", "Path to configuration file")
	runCmd.Flags().StringVar(&runProvider, "provider", "", "LLM Provider (gemini, openai, anthropic, ollama). Overrides .env")
	runCmd.Flags().StringVar(&runModel, "model", "", "LLM Model Name. Overrides .env")
}
