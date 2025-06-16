package metrics

import (
	"fmt"
	"time"
)

type FileContext struct {
	CurrentFilepath string
	FileBytes       int64
}

func (fc FileContext) String() string {
	s := ""
	if fc.CurrentFilepath != "" {
		s += fc.CurrentFilepath + "\n"
	}
	if fc.FileBytes != 0 {
		s += fmt.Sprintf("%v bytes\n", fc.FileBytes)
	}
	return s
}

type executionStatVals struct {
	numExecutions        int
	totalExecutionTimeMs int64
	minExecutionTimeMs   int64
	maxExecutionTimeMs   int64
	maxExecFileContext   FileContext // note information to investigate a high execution time
	init                 bool
}

func (e *executionStatVals) AddNewRecord(timeMs int64, ctx FileContext) {
	if !e.init {
		e.minExecutionTimeMs = timeMs
		e.init = true
	}
	e.numExecutions++
	e.totalExecutionTimeMs += timeMs

	if timeMs < e.minExecutionTimeMs {
		e.minExecutionTimeMs = timeMs
	}
	if timeMs > e.maxExecutionTimeMs {
		e.maxExecutionTimeMs = timeMs
		e.maxExecFileContext = ctx
	}
}

func (e executionStatVals) AveExecTime() int {
	return int(float64(e.totalExecutionTimeMs) / float64(e.numExecutions))
}

func (e executionStatVals) String() string {
	if e.numExecutions == 0 {
		return "No recorded executions."
	}
	aveExecTime := e.AveExecTime()

	s := fmt.Sprintf("Ave/Min/Max: %v ms / %v ms / %v ms (%v total executions)", aveExecTime, e.minExecutionTimeMs, e.maxExecutionTimeMs, e.numExecutions)
	s += fmt.Sprintf("\nMax exec info:\n%s", e.maxExecFileContext)

	return s
}

type functionExecutionStats struct {
	DetectFileTypeLLM   executionStatVals
	AnalyzeFileBasicLLM executionStatVals
}

// for counting stats and metrics globally for execution of different functions
var ExecStats functionExecutionStats = functionExecutionStats{}

func (f functionExecutionStats) ShowAllMetrics() {
	if f.AnalyzeFileBasicLLM.numExecutions > 0 {
		fmt.Println("AnalyzeFileBasicLLM")
		fmt.Println(f.AnalyzeFileBasicLLM.String())
	}
	if f.DetectFileTypeLLM.numExecutions > 0 {
		fmt.Println("DetectFileTypeLLM")
		fmt.Println(f.DetectFileTypeLLM.String())
	}
}

type modelUsage struct {
	callCount         int
	totalCallDuration time.Duration
	minCallDuration   time.Duration
	maxCallDuration   time.Duration
	init              bool
}

func (m modelUsage) String() string {
	s := fmt.Sprintf("call count: %v", m.callCount)
	ave := m.totalCallDuration.Milliseconds() / int64(m.callCount)
	s += fmt.Sprintf("\nAve/Min/Max call time: %v ms / %v ms / %v ms", ave, m.minCallDuration.Milliseconds(), m.maxCallDuration.Milliseconds())
	return s
}

func (m *modelUsage) RecordUsage(startTime time.Time) {
	m.callCount++
	duration := time.Since(startTime)
	m.totalCallDuration += duration

	if !m.init {
		m.minCallDuration = duration
		m.init = true
	}

	if duration > m.maxCallDuration {
		m.maxCallDuration = duration
	}
	if duration < m.minCallDuration {
		m.minCallDuration = duration
	}
}

type modelUsageStats struct {
	Llama3       modelUsage
	DeepSeek     modelUsage
	DeepSeek14b  modelUsage
	CodeLlama    modelUsage
	CodeLlama13b modelUsage
}

func (m modelUsageStats) ShowAllMetrics() {
	s := ""
	if m.Llama3.callCount > 0 {
		s += "Llama3\n"
		s += fmt.Sprintf("%s\n", m.Llama3)
	}
	if m.DeepSeek.callCount > 0 {
		s += "DeepSeek\n"
		s += fmt.Sprintf("%s\n", m.DeepSeek)
	}
	if m.DeepSeek14b.callCount > 0 {
		s += "DeepSeek14b\n"
		s += fmt.Sprintf("%s\n", m.DeepSeek14b)
	}
	if m.CodeLlama.callCount > 0 {
		s += "CodeLlama\n"
		s += fmt.Sprintf("%s\n", m.CodeLlama)
	}
	if m.CodeLlama13b.callCount > 0 {
		s += "CodeLlama13b\n"
		s += fmt.Sprintf("%s\n", m.CodeLlama13b)
	}
	if s == "" {
		s = "No LLM models called."
	}
	fmt.Println(s)
}

var ModelUsageStats modelUsageStats = modelUsageStats{}
