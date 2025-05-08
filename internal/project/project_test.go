package project

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/webbben/caius/internal/llm"
)

type analyzeFileBasicTestCase struct {
	CaseName        string
	FileName        string
	MockFilename    string
	Type            []string
	Keywords        []string
	ExcludeKeywords []string
}

var testAnalyzeFileBasicTestCases []analyzeFileBasicTestCase = []analyzeFileBasicTestCase{
	{
		CaseName:        "[JS 1] login page",
		FileName:        "javascript01.txt",
		MockFilename:    "login.js",
		Type:            []string{"javascript", "react"},
		Keywords:        []string{"login", "log", "welcome", "react", "component"},
		ExcludeKeywords: []string{"logout", "vue", "angular", "typescript", "html"},
	},
	{
		CaseName:        "[JS 2] calculator",
		FileName:        "javascript02.txt",
		MockFilename:    "calculator.js",
		Type:            []string{"javascript"},
		Keywords:        []string{"calculator", "math", "arithmetic"},
		ExcludeKeywords: []string{"component", "react", "typescript", "python"},
	},
	{
		CaseName:        "[JS 3] todo list",
		FileName:        "javascript03.txt",
		MockFilename:    "todoApp.js",
		Type:            []string{"javascript"},
		Keywords:        []string{"todo", "task", "list", "manage"},
		ExcludeKeywords: []string{"react", "typescript", "python"},
	},
	{
		CaseName:        "[JS 4] theme toggler",
		FileName:        "javascript04.txt",
		MockFilename:    "themeToggle.js",
		Type:            []string{"javascript"},
		Keywords:        []string{"theme", "dark", "toggle", "event"},
		ExcludeKeywords: []string{"react", "typescript", "python", "html"},
	},
}

func TestAnalyzeFileBasicLlama3(t *testing.T) {
	t.Run("Llama3 AnalyzeFileBasic", func(t *testing.T) {
		llm.SetModel(llm.Models.Llama3)
		TestAnalyzeFileBasic(t)
	})
}

func TestAnalyzeFileBasicDeepSeek(t *testing.T) {
	t.Run("DeepSeek AnalyzeFileBasic", func(t *testing.T) {
		llm.SetModel(llm.Models.DeepSeek)
		TestAnalyzeFileBasic(t)
	})
}

// go test -run ^TestAnalyzeFileBasic$ github.com/webbben/caius/internal/project
func TestAnalyzeFileBasic(t *testing.T) {
	pass := 0
	i := 0
	for _, testCase := range testAnalyzeFileBasicTestCases {
		i++
		log.Println(testCase.CaseName)

		response, err := AnalyzeFileBasic(fmt.Sprintf("tests/%s", testCase.FileName), testCase.MockFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to analyze file: %q", err)
		}
		fileType := strings.ToLower(response.Type)
		if !slices.Contains(testCase.Type, fileType) {
			log.Println("fail: detected file type is incorrect.")
			log.Println("got:", fileType, " | expected:", testCase.Type)
			continue
		}

		desc := strings.ToLower(response.Description)
		for _, keyword := range testCase.Keywords {
			if strings.Contains(desc, keyword) {
				// finding one satisfies this check
				break
			}
		}
		for _, exclude := range testCase.ExcludeKeywords {
			if strings.Contains(desc, exclude) {
				// fail
				log.Println("fail: detected an excluded keyword:", exclude)
				log.Println("description given:", desc)
				continue
			}
		}

		pass++
	}

	log.Printf("pass: %v/%v\n", pass, i)
}
