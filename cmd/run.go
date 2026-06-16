package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sabujislam/figgen/internal/agents"
	"github.com/sabujislam/figgen/internal/config"
	"github.com/sabujislam/figgen/internal/filesystem"
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
			log.Fatalf("Failed to load configuration: %v", err)
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
			log.Fatalf("AI Provider initialization failed: %v", err)
		}

		for {
			st, err := state.LoadState(outDir)
			if err != nil {
				log.Fatalf("Failed to load state: %v. Did you run 'figgen plan' first?", err)
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
				fmt.Println("🎉 All tasks completed!")
				break
			}

			fmt.Printf("Executing task: %s [%s]\n", targetTask.Name, targetTask.Type)
			
			// Mark as in-progress
			st.Tasks[taskIndex].Status = "in_progress"
			state.SaveState(outDir, st)

			// Execute
			var code string
			if targetTask.Type == "component" {
				code, err = agents.RunCoderForComponent(ctx, aiClient, cfg, *targetTask.ComponentPlan)
				if err == nil {
					err = filesystem.WriteComponent(outDir, targetTask.Name, targetTask.IsShadcn, code)
				}
			} else {
				code, err = agents.RunCoderForPage(ctx, aiClient, cfg, *targetTask.PagePlan)
				if err == nil {
					err = filesystem.WritePage(outDir, targetTask.Name, code)
				}
			}

			if err != nil {
				fmt.Printf("❌ Task %s failed: %v\n", targetTask.Name, err)
				st.Tasks[taskIndex].Status = "failed"
				state.SaveState(outDir, st)
				
				if !runAll {
					break
				}
			} else {
				fmt.Printf("✅ Task %s completed!\n", targetTask.Name)
				st.Tasks[taskIndex].Status = "completed"
				state.SaveState(outDir, st)
			}

			if !runAll {
				fmt.Println("Run 'figgen run' again to execute the next task, or 'figgen run --all' to execute all.")
				break
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
