package metrics

import (
	"fmt"
	"time"
)

// if true, LLM usage stats will be recorded
var LOG_LLM_USAGE bool = true

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

var modelUsageMap map[string]*modelUsage = map[string]*modelUsage{}

func RecordModelUsage(modelName string, startTime time.Time) {
	if _, ok := modelUsageMap[modelName]; !ok {
		modelUsageMap[modelName] = &modelUsage{}
	}

	modelUsageMap[modelName].RecordUsage(startTime)
}

func ShowAllModelUsageMetrics() {
	for modelName, usage := range modelUsageMap {
		fmt.Println(modelName)
		fmt.Printf("%s\n", usage)
	}
}

func ResetModelUsageStats() {
	modelUsageMap = map[string]*modelUsage{}
}
