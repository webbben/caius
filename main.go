/*
Copyright © 2025 Ben Webb ben.webb340@gmail.com
*/
package main

import (
	"fmt"
	"os"

	"github.com/webbben/caius/cmd"
	"github.com/webbben/caius/internal/llm"
)

func main() {
	llmCmd, err := llm.StartServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start llm server: %q", err)
		os.Exit(1)
	}
	defer llm.StopServer(llmCmd)

	llm.SetModel(llm.Models.DeepSeek)
	cmd.Execute()
}
