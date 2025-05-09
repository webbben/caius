package project

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/webbben/caius/internal/llm"
)

func resetLog() {
	os.Remove(".log")
}

func writeLog(s string) {
	log.Println(s)
	timestamp := time.Now().Format(time.DateTime)
	line := fmt.Sprintf("%s  %s\n", timestamp, s)

	file, err := os.OpenFile(".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString(line)
	if err != nil {
		log.Fatal(err)
	}
}

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
		Type:            []string{"javascript code", "react"},
		Keywords:        []string{"login", "log", "welcome", "react", "component"},
		ExcludeKeywords: []string{"logout", "vue", "angular", "typescript", "html"},
	},
	{
		CaseName:        "[JS 2] calculator",
		FileName:        "javascript02.txt",
		MockFilename:    "calculator.js",
		Type:            []string{"javascript code"},
		Keywords:        []string{"calculator", "math", "arithmetic"},
		ExcludeKeywords: []string{"component", "react", "typescript", "python"},
	},
	{
		CaseName:        "[JS 3] todo list",
		FileName:        "javascript03.txt",
		MockFilename:    "todoApp.js",
		Type:            []string{"javascript code"},
		Keywords:        []string{"todo", "task", "list", "manage"},
		ExcludeKeywords: []string{"react", "typescript", "python"},
	},
	{
		CaseName:        "[JS 4] theme toggler",
		FileName:        "javascript04.txt",
		MockFilename:    "themeToggle.js",
		Type:            []string{"javascript code"},
		Keywords:        []string{"theme", "dark", "toggle", "event"},
		ExcludeKeywords: []string{"react", "typescript", "python", "html"},
	},
}

type describeProjectTestCase struct {
	CaseName        string
	FileName        string
	Keywords        []string
	ExcludeKeywords []string
}

var describeProjectTestCases []describeProjectTestCase = []describeProjectTestCase{
	{
		CaseName:        "[JS 1] react todo list app",
		FileName:        "01.txt",
		Keywords:        []string{"react", "todo", "task", "javascript"},
		ExcludeKeywords: []string{"typescript", "python", "angular", "vue"},
	},
	{
		CaseName:        "[JS 2] react weather dashboard app",
		FileName:        "02.txt",
		Keywords:        []string{"weather", "dashboard", "javascript", "react"},
		ExcludeKeywords: []string{"typescript", "python", "angular", "vue"},
	},
}

func loadFileText(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
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
	resetLog()
	pass := 0
	i := 0
	for _, testCase := range testAnalyzeFileBasicTestCases {
		i++
		writeLog(testCase.CaseName)

		response, err := AnalyzeFileBasic(fmt.Sprintf("tests/analyzeFile/%s", testCase.FileName), testCase.MockFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to analyze file: %q", err)
		}
		writeLog("LLM response:")
		writeLog(response.Type)
		writeLog(response.Description)

		fileType := strings.ToLower(response.Type)
		if !slices.Contains(testCase.Type, fileType) {
			writeLog("fail: detected file type is incorrect.")
			writeLog("got: " + fileType + " | expected: " + strings.Join(testCase.Type, " or "))
			t.Fail()
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
				writeLog("fail: detected an excluded keyword: " + exclude)
				writeLog("description given: " + desc)
				t.Fail()
				continue
			}
		}

		pass++
	}

	writeLog(fmt.Sprintf("pass: %v/%v\n", pass, i))
}

func TestDescribeProject(t *testing.T) {
	resetLog()
	pass := 0
	i := 0
	for _, testCase := range describeProjectTestCases {
		i++
		writeLog(testCase.CaseName)

		projectMap := loadFileText(fmt.Sprintf("tests/describeProject/%s", testCase.FileName))

		response, err := DescribeProject(projectMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to describe project: %q", err)
		}

		writeLog("LLM response:")
		writeLog(response)

		desc := strings.ToLower(response)
		for _, keyword := range testCase.Keywords {
			if strings.Contains(desc, keyword) {
				// finding one satisfies this check
				break
			}
		}
		for _, exclude := range testCase.ExcludeKeywords {
			if strings.Contains(desc, exclude) {
				// fail
				writeLog("fail: detected an excluded keyword: " + exclude)
				writeLog("description given: " + desc)
				t.Fail()
				continue
			}
		}

		pass++
	}

	writeLog(fmt.Sprintf("pass: %v/%v\n", pass, i))
}
