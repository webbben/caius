/*
Copyright © 2025 Ben Webb ben.webb340@gmail.com
*/
package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/webbben/caius/internal/config"
	"github.com/webbben/caius/internal/llm"
	"github.com/webbben/caius/internal/metrics"
	"github.com/webbben/caius/internal/project"
	"github.com/webbben/caius/internal/utils"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "run a test to see how LLMs are performing",
	Long:  `run a test to see how LLMs are performing`,
	Run: func(cmd *cobra.Command, args []string) {
		config.SHOW_FUNCTION_METRICS = false
		config.SHOW_LLM_METRICS = false

		path := ""
		if len(args) > 0 {
			path = args[0]
		}

		testsPath, err := filepath.Abs("cmd/tests")
		if err != nil {
			panic(err)
		}

		switch path {
		case "short":
			path = filepath.Join(testsPath, "short.txt")
		case "medium":
			path = filepath.Join(testsPath, "medium.txt")
		case "large", "long":
			path = filepath.Join(testsPath, "large.txt")
		default:
			path = filepath.Join(testsPath, "short.txt")
			log.Println("default: short")
		}

		speedTest(path)
	},
}

func speedTest(path string) {
	const N = 10

	var r project.BasicFileAnalysisResponse
	var err error

	models := []string{
		llm.Models.Llama3, llm.Models.CodeLlama, llm.Models.CodeLlama13b, llm.Models.DeepSeek, llm.Models.DeepSeek14b, llm.Models.DeepSeekCoder, llm.Models.DeepSeekCoder6b,
	}

	for _, model := range models {
		config.BASIC_FILE_ANALYSIS_MODEL = model

		// "wake up" the model to ensure it's not running slowly on first call
		utils.Terminal.Lowkey("waking up " + model + "...")
		llm.SetModel(model)
		llm.WakeUp()
		utils.Terminal.Lowkey("... done")

		for i := range N {
			r, err = project.AnalyzeFileBasic(path, "index.js")
			if err != nil {
				panic(err)
			}
			utils.Terminal.Lowkey(fmt.Sprintf("Model: %s, Run %v/%v", model, i+1, N))
		}
		utils.Terminal.Lowkey("output: " + r.Description)
	}

	metrics.ShowAllModelUsageMetrics()

}

func init() {
	rootCmd.AddCommand(testCmd)
}
