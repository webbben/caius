package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	ollamawrapper "github.com/webbben/ollama-wrapper"
)

type models struct {
	Llama3    string
	DeepSeek  string
	CodeLlama string
}

var Models models = models{
	Llama3:    "llama3.2:3b",
	DeepSeek:  "deepseek-r1:7b",
	CodeLlama: "codellama:7b",
}

func StartServer() (int32, error) {
	pid, err := ollamawrapper.StartServer()
	return pid, err
}

func SetModel(model string) {
	ollamawrapper.SetModel(model)
}

func GenerateCompletion(systemPrompt string, prompt string) (string, error) {
	client, err := ollamawrapper.GetClient()
	if err != nil {
		return "", err
	}

	response, err := ollamawrapper.GenerateCompletionWithOpts(client, systemPrompt, prompt, map[string]interface{}{
		"temperature": 0.0,
	})
	if err != nil {
		return "", err
	}

	cleanedString, err := extractJsonString(response)
	if err != nil {
		return "", err
	}

	return cleanedString, nil
}

func GenerateCompletionJson(systemPrompt string, prompt string, formatSchema json.RawMessage, v any) error {
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
	return nil
}

// TODO: delete, now that we have built-in Ollama JSON formatting?
func extractJsonString(s string) (string, error) {
	if !strings.Contains(s, "{") || !strings.Contains(s, "}") {
		return "", errors.New("malformed json given; missing opening or closing bracket")
	}
	if s[0] != '{' {
		// cut the first portion before the beginning bracket
		s = strings.Join(strings.Split(s, "{")[1:], "")
	}
	if s[len(s)-1] != '}' {
		parts := strings.Split(s, "}")
		s = strings.Join(parts[:len(parts)-1], "")
	}
	s = strings.TrimSpace(s)
	return fmt.Sprintf("{%s}", s), nil
}
