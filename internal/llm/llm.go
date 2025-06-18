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
	// Ave/Min/Max call time: 1701 ms / 1452 ms / 3902 ms (#2)
	//
	// Seems to be the fastest, but I have an impression that with longer text, it will not perform as well.
	Llama3 string
	// Ave/Min/Max call time: 2818 ms / 2298 ms / 7370 ms (#3)
	//
	// Currently this seems to be the best overall; somewhat fast, but still quite accurate
	DeepSeek string
	// Ave/Min/Max call time: 6538 ms / 5341 ms / 16165 ms (#6)
	//
	// This performs considerably slower than the smaller DeepSeek, so I think it may be off the table for now.
	DeepSeek14b string
	// Ave/Min/Max call time: 529 ms / 360 ms / 2023 ms (#1 - fastest!)
	//
	// Size: 776 MB
	DeepSeekCoder string
	// Ave/Min/Max call time: 3141 ms / 2366 ms / 9989 ms (#4)
	//
	// Size: 3.8 GB
	DeepSeekCoder6b string
	// Ave/Min/Max call time: 5237 ms / 4532 ms / 11425 ms (#5)
	//
	// This model will probably be most useful for writing actual code, as it apparently has a lot of pre-trained functionality for this.
	// Read more about it here: https://ollama.com/blog/how-to-prompt-code-llama
	//
	// For now, it just seems to perform slower than the alternatives for tasks like describing a code block
	CodeLlama string
	// Ave/Min/Max call time: 7805 ms / 6274 ms / 19519 ms (#7 - slowest)
	//
	// This is the slowest model, so I don't think I'll be making use of it.
	// Maybe on beefier systems it would be useful for writing code though?
	CodeLlama13b string
}

var Models models = models{
	Llama3:          "llama3.2:3b",
	DeepSeek:        "deepseek-r1:7b",
	DeepSeek14b:     "deepseek-r1:14b",
	DeepSeekCoder:   "deepseek-coder:1.3b",
	DeepSeekCoder6b: "deepseek-coder:6.7b",
	CodeLlama:       "codellama:7b",
	CodeLlama13b:    "codellama:13b",
}

func RecordLLMUsage(startTime time.Time) {
	if !metrics.LOG_LLM_USAGE {
		return
	}
	curModel := ollamawrapper.GetModel()
	if curModel == "" {
		log.Println("failed to log model usage; no model name found")
		return
	}

	metrics.RecordModelUsage(curModel, startTime)
}

func StartServer() (int32, error) {
	pid, err := ollamawrapper.StartServer()
	return pid, err
}

func SetModel(model string) {
	ollamawrapper.SetModel(model)
}

func WakeUp() error {
	client, err := ollamawrapper.GetClient()
	if err != nil {
		return err
	}

	_, err = ollamawrapper.GenerateCompletion(client, "say hi", "hi!")
	return err
}

var EmptyResponseError = errors.New("GenerateCompletionJson: no data returned by LLM")

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
		return EmptyResponseError
	}

	err = json.Unmarshal([]byte(response), &v)
	if err != nil {
		log.Printf("\nresponse:\n%s\n", response)
		return errors.Join(errors.New("GenerateCompletionJson: error unmarshalling JSON in LLM response;"), err)
	}

	RecordLLMUsage(start)
	return nil
}
