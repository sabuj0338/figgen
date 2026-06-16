package cmd

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/filesystem"
	"github.com/sabujislam/figgen/internal/executor"
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
		outDir := "./out"
		cfg, err := config.LoadConfig(runConfigPath)
		if err != nil {
			logger.Fatal("Failed to load configuration: %v", err)
		}

		ctx := context.Background()
		
		if runProvider == "" {
			runProvider = os.Getenv("DEFAULT_PROVIDER")
			if runProvider == "" {
				runProvider = "gemini"
			}
		}
		if runModel == "" {
			runModel = os.Getenv("DEFAULT_MODEL")
		}

		aiClient, err := agents.NewProvider(ctx, runProvider, runModel)
		if err != nil {
			logger.Fatal("AI Provider initialization failed: %v", err)
		}

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
			
			// Mark as in-progress
			st.Tasks[taskIndex].Status = "in_progress"
			state.SaveState(outDir, st)

			// Execute
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
					codeResp, err = agents.RunCoderForComponent(ctx, aiClient, cfg, *targetTask.ComponentPlan)
					if err == nil {
						targetFilePath, err = filesystem.WriteComponent(outDir, targetTask.Name, targetTask.IsShadcn, codeResp.Code)
					}
				}
			} else {
				codeResp, err = agents.RunCoderForPage(ctx, aiClient, cfg, *targetTask.PagePlan)
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
				
				if !runAll {
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

			if !runAll {
				logger.Info("Run 'figgen run' again to execute the next task, or 'figgen run --all' to execute all.")
				break
			}

			// Respect Gemini free tier limits (15 Requests Per Minute)
			if runProvider == "gemini" {
				logger.Info("Waiting 4 seconds to respect Gemini API free-tier rate limits...")
				time.Sleep(4 * time.Second)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&runAll, "all", false, "Run all pending tasks continuously")
	runCmd.Flags().StringVar(&runConfigPath, "config", "figgen.config.yaml", "Path to configuration file")
	runCmd.Flags().StringVar(&runProvider, "provider", "", "LLM Provider (gemini, openai, anthropic, ollama). Overrides .env")
	runCmd.Flags().StringVar(&runModel, "model", "", "LLM Model Name. Overrides .env")
}
