package cmd

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/figma"
	"github.com/sabujislam/figgen/internal/github"
	"github.com/sabujislam/figgen/internal/logger"
	"github.com/sabujislam/figgen/internal/state"
	"github.com/spf13/cobra"
)

var (
	listenConfigPath string
	listenProvider   string
	listenModel      string
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow from local Figma plugin
	},
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Start a local WebSocket server to receive data directly from the Figma Plugin",
	Run: func(cmd *cobra.Command, args []string) {
		outDir := globalOutDir
		cfg, err := config.LoadConfig(listenConfigPath)
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

		if listenProvider == "" {
			listenProvider = os.Getenv("DEFAULT_PROVIDER")
			if listenProvider == "" {
				listenProvider = "gemini"
			}
		}
		if listenModel == "" {
			listenModel = os.Getenv("DEFAULT_MODEL")
		}

		aiClient, err := agents.NewProvider(ctx, listenProvider, listenModel)
		if err != nil {
			logger.Fatal("AI Provider initialization failed: %v", err)
		}

		os.MkdirAll(outDir+"/.figgen", os.ModePerm)

		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logger.Warn("WebSocket upgrade error: %v", err)
				return
			}
			defer c.Close()

			logger.Success("Figma Plugin Connected!")

			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					break
				}

				logger.Step("Received design payload from Figma Plugin (%d bytes)", len(message))

				// Save the full context locally for 'figgen run' to use
				contextFile := outDir + "/.figgen/figma_context.json"
				err = os.WriteFile(contextFile, message, 0644)
				if err != nil {
					logger.Warn("Failed to save local figma context: %v", err)
				} else {
					logger.Info("Saved local design context for Coder.")
				}

				logger.Step("Running Architecture Planner...")
				
				prunedJSON, err := figma.PruneFigmaData(string(message))
				if err != nil {
					prunedJSON = string(message)
				}

				plan, err := agents.RunPlanner(ctx, aiClient, cfg, prunedJSON)
				if err != nil {
					logger.Error("AI Planner failed: %v", err)
					// We continue listening instead of fatal
					continue
				}

				fileKey := "local"
				figmaURL := os.Getenv("FIGMA_URL")
				if figmaURL != "" {
					parsedKey, _, err := figma.ParseURL(figmaURL)
					if err == nil && parsedKey != "" {
						fileKey = parsedKey
					}
				}

				err = state.InitState(outDir, fileKey, plan)
				if err != nil {
					logger.Error("Failed to save state: %v", err)
					continue
				}

				logger.Success("Planning complete! Identified %d components and %d pages.", len(plan.Components), len(plan.Pages))
				
				// Prompt user to execute run
				logger.Prompt("Do you want to run code generation now? [r = run, a = run --all, s = skip/keep listening]: ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				
				if response == "r" || response == "run" {
					ExecuteRun(outDir, false, listenConfigPath, listenProvider, listenModel)
				} else if response == "a" || response == "all" {
					ExecuteRun(outDir, true, listenConfigPath, listenProvider, listenModel)
				} else {
					logger.Info("Skipped generation. Waiting for more design payloads on WebSocket...")
				}
			}
		})

		port := "8080"
		logger.Info("Figgen WebSocket Listener running on ws://localhost:%s/ws", port)
		logger.Info("Open your local Figgen Exporter plugin in Figma and click Export!")
		
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logger.Fatal("Server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
	listenCmd.Flags().StringVar(&listenConfigPath, "config", "figgen.config.yaml", "Path to configuration file")
	listenCmd.Flags().StringVar(&listenProvider, "provider", "", "LLM Provider")
	listenCmd.Flags().StringVar(&listenModel, "model", "", "LLM Model Name")
}
