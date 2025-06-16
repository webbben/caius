package llm

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/webbben/caius/internal/metrics"
	ollamawrapper "github.com/webbben/ollama-wrapper"
)

type models struct {
	Llama3       string
	DeepSeek     string
	DeepSeek14b  string
	CodeLlama    string
	CodeLlama13b string
}

var Models models = models{
	Llama3:       "llama3.2:3b",
	DeepSeek:     "deepseek-r1:7b",
	DeepSeek14b:  "deepseek-r1:14b",
	CodeLlama:    "codellama:7b",
	CodeLlama13b: "codellama:13b",
}

func RecordLLMUsage(startTime time.Time) {
	curModel := ollamawrapper.GetModel()
	switch curModel {
	case "llama3.2:3b":
		metrics.ModelUsageStats.Llama3.RecordUsage(startTime)
	case "deepseek-r1:7b":
		metrics.ModelUsageStats.DeepSeek.RecordUsage(startTime)
	case "deepseek-r1:14b":
		metrics.ModelUsageStats.DeepSeek14b.RecordUsage(startTime)
	case "codellama:7b":
		metrics.ModelUsageStats.CodeLlama.RecordUsage(startTime)
	case "codellama:13b":
		metrics.ModelUsageStats.CodeLlama13b.RecordUsage(startTime)
	default:
		log.Println("RecordLLMUsage: current model not found in switch statement! Do you need to add a new one?")
	}
}

func StartServer() (int32, error) {
	pid, err := ollamawrapper.StartServer()
	return pid, err
}

func SetModel(model string) {
	ollamawrapper.SetModel(model)
}

func GenerateCompletionJson(systemPrompt string, prompt string, formatSchema json.RawMessage, v any) error {
	start := time.Now()
	client, err := ollamawrapper.GetClient()
	if err != nil {
		return errors.Join(errors.New("GenerateCompletionJson: error getting client;"), err)
	}

	response, err := ollamawrapper.GenerateCompletionOptsFormat(client, systemPrompt, prompt, map[string]interface{}{
		"temperature": 0.0,
	}, formatSchema)
	if err != nil {
		return errors.Join(errors.New("GenerateCompletionJson: error generating completion;"), err)
	}

	if response == "" {
		return errors.New("GenerateCompletionJson: error generating completion; no data returned")
	}

	err = json.Unmarshal([]byte(response), &v)
	if err != nil {
		log.Printf("\nresponse:\n%s\n", response)
		return errors.Join(errors.New("GenerateCompletionJson: error unmarshalling JSON in LLM response;"), err)
	}

	RecordLLMUsage(start)
	return nil
}
