package metrics

import (
	"fmt"
	"log"
	"time"
)

type FileContext struct {
	Filepath  string
	FileBytes int64
}

func (fc FileContext) String() string {
	s := ""
	if fc.Filepath != "" {
		s += fc.Filepath + "\n"
	}
	if fc.FileBytes != 0 {
		s += fmt.Sprintf("%v bytes\n", fc.FileBytes)
	}
	return s
}

type speedRecord struct {
	count          int
	totalDuration  time.Duration
	maxDuration    time.Duration
	minDuration    time.Duration
	maxDurationCtx FileContext
}

func (s *speedRecord) AddRecord(startTime time.Time, ctx FileContext) {
	s.count++
	d := time.Since(startTime)
	s.totalDuration += d

	if d > s.maxDuration {
		s.maxDuration = d
		s.maxDurationCtx = ctx
	}
	if s.minDuration == 0 || d < s.minDuration {
		s.minDuration = d
	}
}

func (s speedRecord) GetAverageDuration() time.Duration {
	if s.count == 0 {
		return 0
	}
	return s.totalDuration / time.Duration(s.count)
}

func (s speedRecord) CalculateTimeEstimate(count int) time.Duration {
	return (s.GetAverageDuration() * time.Duration(count)).Round(time.Second)
}

var speedRecordMap map[string]*speedRecord = map[string]*speedRecord{}

func AddSpeedRecord(functionName string, startTime time.Time, ctx FileContext) {
	if _, ok := speedRecordMap[functionName]; !ok {
		speedRecordMap[functionName] = &speedRecord{}
	}

	speedRecordMap[functionName].AddRecord(startTime, ctx)
}

// Just for viewing data - use AddSpeedRecord for saving new data
func SpeedRecord(functionName string) speedRecord {
	record, exists := speedRecordMap[functionName]
	if exists {
		return *record
	}

	// create if it doesn't exist yet, just to prevent errors and stuff
	speedRecordMap[functionName] = &speedRecord{}
	log.Println("created new speed record in GetSpeedRecord; this is probably wrong...")
	return speedRecord{}
}
