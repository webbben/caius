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

		testsPath, err := filepath.Abs("cmd/tests")
		if err != nil {
			panic(err)
		}

		path := ""
		if len(args) > 0 {
			path = args[0]
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

		const N = 10

		config.BASIC_FILE_ANALYSIS_MODEL = llm.Models.Llama3
		for i := range N {
			_, err := project.AnalyzeFileBasic(path, "index.js")
			if err != nil {
				panic(err)
			}
			utils.Terminal.Lowkey(fmt.Sprintf("Model: %s, Runs %v/%v", "Llama3", i+1, N))
		}
		config.BASIC_FILE_ANALYSIS_MODEL = llm.Models.CodeLlama
		for i := range N {
			_, err := project.AnalyzeFileBasic(path, "index.js")
			if err != nil {
				panic(err)
			}
			utils.Terminal.Lowkey(fmt.Sprintf("Model: %s, Runs %v/%v", "CodeLlama", i+1, N))
		}
		config.BASIC_FILE_ANALYSIS_MODEL = llm.Models.CodeLlama13b
		for i := range N {
			_, err := project.AnalyzeFileBasic(path, "index.js")
			if err != nil {
				panic(err)
			}
			utils.Terminal.Lowkey(fmt.Sprintf("Model: %s, Runs %v/%v", "CodeLlama13b", i+1, N))
		}
		config.BASIC_FILE_ANALYSIS_MODEL = llm.Models.DeepSeek
		for i := range N {
			_, err := project.AnalyzeFileBasic(path, "index.js")
			if err != nil {
				panic(err)
			}
			utils.Terminal.Lowkey(fmt.Sprintf("Model: %s, Runs %v/%v", "DeepSeek", i+1, N))
		}
		config.BASIC_FILE_ANALYSIS_MODEL = llm.Models.DeepSeek14b
		for i := range N {
			_, err := project.AnalyzeFileBasic(path, "index.js")
			if err != nil {
				panic(err)
			}
			utils.Terminal.Lowkey(fmt.Sprintf("Model: %s, Runs %v/%v", "DeepSeek14b", i+1, N))
		}

		metrics.ModelUsageStats.ShowAllMetrics()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
