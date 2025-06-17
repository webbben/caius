/*
Copyright © 2025 Ben Webb ben.webb340@gmail.com
*/
package main

import (
	"fmt"
	"os"

	"github.com/webbben/caius/cmd"
	"github.com/webbben/caius/internal/config"
	"github.com/webbben/caius/internal/llm"
	"github.com/webbben/caius/internal/metrics"
	ollamawrapper "github.com/webbben/ollama-wrapper"
)

func main() {
	_, err := llm.StartServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start ollama server: %q", err)
		os.Exit(1)
	}

	// ensure all models are pulled
	ollamawrapper.EnsureModelIsPulled(llm.Models.DeepSeek, true, func(prp ollamawrapper.PullRequestProgress) {
		fmt.Printf("\rPulling model: %v/%v (%s)", prp.Completed, prp.Total, prp.Status)
	})
	llm.SetModel(llm.Models.DeepSeek)

	cmd.Execute()

	if config.SHOW_FUNCTION_METRICS {
		fmt.Println()
	}
	if config.SHOW_LLM_METRICS {
		fmt.Println()
		metrics.ShowAllModelUsageMetrics()
	}
}
