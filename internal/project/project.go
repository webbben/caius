package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/webbben/caius/internal/config"
	"github.com/webbben/caius/internal/files"
	"github.com/webbben/caius/internal/llm"
	"github.com/webbben/caius/internal/metrics"
	"github.com/webbben/caius/internal/utils"
	"github.com/webbben/caius/prompts"
)

type FileData struct {
	Filename          string
	Type              string // The type of data in this file
	Description       string // Description of what this file contains
	FullPath          string
	SkipLLMProcessing bool // Indicates if LLM should not bother analyzing file content
	SizeBytes         int64
}

type BasicFileAnalysisResponse struct {
	SKIP              bool   // if true, this file data will be discarded and not used anywhere
	Type              string `json:"file_type"`
	Description       string `json:"description"`
	SkipLLMProcessing bool   `json:"skip_llm_processing"`
	SizeBytes         int64  `json:"size_bytes"`
}

type DetectFileTypeLLMResponse struct {
	Category string `json:"category"`
	Type     string `json:"type"`
}

var DetectFileTypeLLMSchema json.RawMessage = json.RawMessage(`{
	"type": "object",
	"properties": {
		"category": {
			"type": "string"
		},
		"type": {
			"type": "string"
		}
	},
	"required": ["category", "type"]
}`)

var BasicFileAnalysisSchema json.RawMessage = json.RawMessage(`{
	"type": "object",
	"properties": {
		"file_type": {
			"type": "string"
		},
		"description": {
			"type": "string"
		}
	},
	"required": ["file_type", "description"]
}`)

type DescribeProjectResponse struct {
	Description string `json:"description"`
}

var DescribeProjectSchema json.RawMessage = json.RawMessage(`{
	"type": "object",
	"properties": {
		"description": {
			"type": "string"
		}
	},
	"required": ["description"]
}`)

// GetProcessableFileInfo counts the number of files (and their size) for files that are text based and processed by LLMs.
// Mainly used for calculating processing time estimates.
func GetProcessableFileInfo(fileList []string) (int, int64, error) {
	var b int64 = 0
	fileCount := 0
	for _, file := range fileList {
		filename := filepath.Base(file)
		if files.IgnoreFiles(filename) {
			continue
		}
		if files.ReservedFileMap(filename) != "" {
			continue
		}
		if files.UnableToProcessTypes(filename) != "" {
			continue
		}

		info, err := os.Stat(file)
		if err != nil {
			return 0, 0, err
		}
		b += info.Size()
		fileCount++
	}

	return fileCount, b, nil
}

func DetectFileType(filename string, fileContent []byte, ctx *metrics.FileContext) (string, error) {
	// check file name, in case it has a specific type (e.g. readme files)
	filetype, matchFound := files.FileTypeResolver(filename)
	if matchFound {
		return filetype, nil
	}

	// check file extension
	if strings.Contains(filename, ".") {
		parts := strings.Split(filename, ".")
		// handle for filenames with multiple periods (get last extension)
		ext := parts[len(parts)-1]
		filetype, matchFound = files.FileTypeResolver(ext)
		if matchFound {
			return filetype, nil
		}
	}

	// check if file has a shebang that indicates a programming language script
	shebangType := files.CheckShebang(fileContent)
	if shebangType != "" {
		filetype, _ = files.FileTypeResolver(shebangType)
		return filetype, nil
	}

	// failed to determine filetype by name, extension, content, etc. Last resort: use LLM
	// we should avoid LLM calls whenever possible, since it is relatively expensive in terms of processing time
	fileTypeResp, err := DetectFileTypeLLM(fileContent, ctx)
	if err != nil {
		return "", err
	}
	if fileTypeResp.Type != "" {
		return fileTypeResp.Type, nil
	}
	if fileTypeResp.Category != "" {
		return fileTypeResp.Category, nil
	}

	log.Println("DetectFileTypeLLM: no type or category returned")
	return "", nil
}

func DetectFileTypeLLM(fileData []byte, ctx *metrics.FileContext) (DetectFileTypeLLMResponse, error) {
	start := time.Now()
	var responseJson DetectFileTypeLLMResponse
	llm.SetModel(config.DETECT_FILE_TYPE_MODEL)

	sysPrompt := prompts.P_ANALYZE_FILE_TYPE_01
	sampleData := fileData
	if len(sampleData) > config.MAX_BYTES_BASIC_ANALYSIS {
		sampleData = sampleData[:config.MAX_BYTES_BASIC_ANALYSIS]
	}
	prompt := string(sampleData)

	err := llm.GenerateCompletionJson(sysPrompt, prompt, DetectFileTypeLLMSchema, &responseJson)
	if err != nil {
		return DetectFileTypeLLMResponse{}, utils.WrapError("detectFileTypeLLM: error while generating completion", err)
	}

	responseJson.Type, _ = files.FileTypeResolver(responseJson.Type)

	metrics.AddSpeedRecord("DetectFileTypeLLM", start, *ctx)
	return responseJson, nil
}

func AnalyzeFileBasic(filePath string, fileName string) (BasicFileAnalysisResponse, error) {
	start := time.Now()
	ctx := &metrics.FileContext{}
	ctx.Filepath = filePath

	if files.IgnoreFiles(fileName) {
		return BasicFileAnalysisResponse{
			SkipLLMProcessing: true,
		}, nil
	}

	// check for reserved file types (pre-defined type/description)
	reservedFileType := files.ReservedFileMap(fileName)
	if reservedFileType != "" {
		return BasicFileAnalysisResponse{
			Type:              reservedFileType,
			Description:       reservedFileType,
			SkipLLMProcessing: true, // to skip insertion in basic project map
		}, nil
	}

	// check for unprocessable file types
	t := files.UnableToProcessTypes(fileName)
	if t != "" {
		return BasicFileAnalysisResponse{
			Type:              t,
			Description:       "",
			SkipLLMProcessing: true,
		}, nil
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return BasicFileAnalysisResponse{}, errors.Join(errors.New("analyze file: failed to get file info;"), err)
	}
	if fileInfo.Size() == 0 {
		// empty file - ignore
		return BasicFileAnalysisResponse{SKIP: true}, nil
	}
	ctx.FileBytes = fileInfo.Size()

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return BasicFileAnalysisResponse{}, errors.Join(errors.New("analyze file: error reading file "+filePath), err)
	}

	// do a quick scan of the data to confirm it's not compiled binary or unreadable
	if files.IsProbablyBinaryData(fileContent) {
		fileType := "binary"
		desc := "a file containing binary or non-utf8 data."
		if files.IsFileExecutable(fileInfo) {
			desc = "an executable file containing binary data"
		}
		return BasicFileAnalysisResponse{
			Type:              fileType,
			Description:       desc,
			SkipLLMProcessing: true,
		}, nil
	}

	// detect file type
	filetype, err := DetectFileType(fileName, fileContent, ctx)
	if err != nil {
		return BasicFileAnalysisResponse{}, utils.WrapError("error while detecting filetype in AnalyzeFileBasic:", err)
	}

	// do LLM analysis of file
	header := fmt.Sprintf("File name: %s", fileName)
	if filetype != "" {
		header = fmt.Sprintf("%s\nFile type: %s", header, filetype)
	}
	prompt := fmt.Sprintf("%s\n\n(file content below)\n\n%s", header, string(fileContent))

	var responseJson BasicFileAnalysisResponse
	llm.SetModel(config.BASIC_FILE_ANALYSIS_MODEL)
	sysPrompt := prompts.P_ANALYZE_FILE_01
	err = llm.GenerateCompletionJson(sysPrompt, prompt, BasicFileAnalysisSchema, &responseJson)
	if err != nil {
		log.Println("filePath:", filePath)
		if err == llm.EmptyResponseError {
			log.Println(err)
			log.Println("skipping file")
			return BasicFileAnalysisResponse{SKIP: true}, nil
		} else {
			return BasicFileAnalysisResponse{}, errors.Join(errors.New("analyze file: error while generating completion;"), err)
		}
	}

	// use the predetermined filetype if existing
	// filetype from BasicFileAnalysis seems to be inaccurate sometimes
	// TODO: remove file type detection from BasicFileAnalysis?
	if filetype != "" {
		responseJson.Type = filetype
	} else {
		responseJson.Type, _ = files.FileTypeResolver(responseJson.Type)
	}
	responseJson.SizeBytes = fileInfo.Size()

	// clean up descriptions to be more concise
	// Note: not removing capitalization since it could give meaning to some parts of the description.
	desc := responseJson.Description

	// "this file contains ..."
	desc, _ = strings.CutPrefix(desc, "This file contains")
	// "this is ..."
	desc, _ = strings.CutPrefix(desc, "This is")
	// "this file is ..."
	desc, _ = strings.CutPrefix(desc, "This file is")

	responseJson.Description = strings.TrimSpace(desc)

	metrics.AddSpeedRecord("AnalyzeFileBasic", start, *ctx)
	return responseJson, nil
}

func AnalyzeDirectory(root string) (string, error) {
	start := time.Now()
	fileList, err := files.GetProjectFiles(root, files.GetProjectFilesOptions{SkipDotfiles: true})
	if err != nil {
		return "", err
	}

	var totalBytesProcessed int64 = 0
	llmProcessedFileCount := 0

	fileDataList := make([]FileData, 0)

	// get number of LLM processable files, for calculating time estimate
	llmProcessableFileCount, _, err := GetProcessableFileInfo(fileList)
	if err != nil {
		return "", utils.WrapError("error while calculating processable file info;", err)
	}

	for i, file := range fileList {
		// show progress
		// percentage of files processed
		percent := float64(i) / float64(len(fileList)) * 100

		// show time estimate
		utils.Terminal.ClearScreen()
		if i > 0 {
			remainingCount := llmProcessableFileCount - i
			remainingTime := metrics.SpeedRecord("AnalyzeFileBasic").CalculateTimeEstimate(remainingCount)
			estimateString := ""
			if remainingTime > 0 {
				estimateString = fmt.Sprintf("(%.0f%% ~ %s)", percent, remainingTime)
			}
			fmt.Printf("Processing: %v/%v %s", i+1, len(fileList), utils.Terminal.LowkeyS(estimateString))
		} else {
			fmt.Printf("Processing: %v/%v", i, len(fileList))
		}
		utils.Terminal.Lowkey("\n" + file)

		filename := filepath.Base(file)
		fileData := FileData{
			Filename: filename,
			FullPath: file,
		}

		// LLM analysis of file
		fileAnalysisResponse, err := AnalyzeFileBasic(file, filename)
		if err != nil {
			return "", err
		}
		if fileAnalysisResponse.SKIP {
			continue
		}

		fileData.Type = fileAnalysisResponse.Type
		fileData.Description = fileAnalysisResponse.Description
		fileData.SkipLLMProcessing = fileAnalysisResponse.SkipLLMProcessing
		fileData.SizeBytes = fileAnalysisResponse.SizeBytes
		fileDataList = append(fileDataList, fileData)

		if !fileData.SkipLLMProcessing {
			llmProcessedFileCount++
			totalBytesProcessed += fileData.SizeBytes
		}
	}

	utils.Terminal.ClearScreen()

	// once we have finished analysis of files, create a document that stores all of this information in an easy-to-digest format for LLMs.
	// idea:
	// - a diagram of the directory and file structure
	// - short, high-level overview of what the project (or contents within) represent altogether
	// - for each file, the analytical information:
	//   - file type
	//   - brief description

	projectMap := ""

	for _, filedata := range fileDataList {
		if filedata.SkipLLMProcessing {
			continue
		}
		trimmedPath := strings.TrimPrefix(filedata.FullPath, root)
		trimmedPath = filepath.Join(filepath.Base(root), trimmedPath)
		projectMap = fmt.Sprintf("%s\n%s (%s) - %s", projectMap, trimmedPath, filedata.Type, filedata.Description)
	}

	projectMap = strings.TrimSpace(projectMap)

	// get AI description of entire directory, based on combined file analyses
	projectDesc, err := DescribeProject(projectMap)
	if err != nil {
		return "", errors.Join(errors.New("error generating project description"), err)
	}

	fmt.Println(projectDesc)

	metrics.AddSpeedRecord("AnalyzeDirectory", start, metrics.FileContext{})

	return "", nil
}

func DescribeProject(projectMapString string) (string, error) {
	utils.Terminal.Lowkey("project map:")
	utils.Terminal.Lowkey(projectMapString)
	var resp DescribeProjectResponse
	llm.SetModel(llm.Models.DeepSeek)
	err := llm.GenerateCompletionJson(prompts.P_ANALYZE_FILE_MAP_01, projectMapString, DescribeProjectSchema, &resp)
	if err != nil {
		return "", err
	}
	return resp.Description, nil
}
