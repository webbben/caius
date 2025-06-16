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
	"github.com/webbben/caius/internal/utils"
	"github.com/webbben/caius/prompts"
)

type FileData struct {
	Filename          string
	Type              string
	Description       string
	FullPath          string
	SkipLLMProcessing bool
	SizeBytes         int64
}

type BasicFileAnalysisResponse struct {
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

// GetProcessableBytes counts the bytes of file data for files that are text based and processed by LLMs.
// Mainly used for calculating processing time estimates.
func GetProcessableBytes(fileList []string) (int64, error) {
	var b int64 = 0
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
			return 0, err
		}
		b += info.Size()
	}

	return b, nil
}

func DetectFileType(filename string, fileContent []byte) (string, error) {
	// check file name, in case it has a specific type (e.g. readme files)
	filetype := files.FileTypeResolver(filename)
	if filetype != "" {
		return filetype, nil
	}

	// check file extension
	if strings.Contains(filename, ".") {
		parts := strings.Split(filename, ".")
		// handle for filenames with multiple periods (get last extension)
		ext := parts[len(parts)-1]
		filetype = files.FileTypeResolver(ext)
		if filetype != "" {
			return filetype, nil
		}
	}

	// check if file has a shebang that indicates a programming language script
	shebangType := files.CheckShebang(fileContent)
	if shebangType != "" {
		return files.FileTypeResolver(shebangType), nil
	}

	// failed to determine filetype by name, extension, content, etc. Last resort: use LLM
	// we should avoid LLM calls whenever possible, since it is relatively expensive in terms of processing time
	fileTypeResp, err := DetectFileTypeLLM(fileContent)
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

func DetectFileTypeLLM(fileData []byte) (DetectFileTypeLLMResponse, error) {
	var responseJson DetectFileTypeLLMResponse
	llm.SetModel(llm.Models.CodeLlama)

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
	responseJson.Type = files.FileTypeResolver(responseJson.Type)
	return responseJson, nil
}

func AnalyzeFileBasic(filePath string, fileName string) (BasicFileAnalysisResponse, error) {
	sysPrompt := prompts.P_ANALYZE_FILE_01

	if files.IgnoreFiles(fileName) {
		return BasicFileAnalysisResponse{
			SkipLLMProcessing: true,
		}, nil
	}

	// check for reserved file types (pre-defined type/description)
	reservedFileType := files.ReservedFileMap(fileName)
	if reservedFileType != "" {
		return BasicFileAnalysisResponse{
			Type:        reservedFileType,
			Description: reservedFileType,
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
	filetype, err := DetectFileType(fileName, fileContent)
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
	llm.SetModel(llm.Models.CodeLlama)
	err = llm.GenerateCompletionJson(sysPrompt, prompt, BasicFileAnalysisSchema, &responseJson)
	if err != nil {
		log.Println("filePath:", filePath)
		return BasicFileAnalysisResponse{}, errors.Join(errors.New("analyze file: error while generating completion;"), err)
	}

	// use the predetermined filetype if existing
	// filetype from BasicFileAnalysis seems to be inaccurate sometimes
	// TODO: remove file type detection from BasicFileAnalysis?
	if filetype != "" {
		responseJson.Type = filetype
	} else {
		responseJson.Type = files.FileTypeResolver(responseJson.Type)
	}
	responseJson.SizeBytes = fileInfo.Size()

	return responseJson, nil
}

func AnalyzeDirectory(root string) error {
	fileList, err := files.GetProjectFiles(root)
	if err != nil {
		return err
	}

	// get number of bytes that will be calculated, for time estimate purposes
	processBytesTotal, err := GetProcessableBytes(fileList)
	if err != nil {
		return utils.WrapError("error calculating processable bytes", err)
	}
	var totalBytesProcessed int64 = 0
	llmProcessedFileCount := 0

	fileDataList := make([]FileData, 0)

	startTime := time.Now()

	for i, file := range fileList {
		// show progress
		// percentage of files processed
		percent := float64(i) / float64(len(fileList)) * 100

		// we calculate time estimate by average ms per byte of data processed by LLM
		// LLM processing of data is the only time consuming part of this, and it takes longer for larger files
		utils.Terminal.ClearScreen()
		if i > 0 && totalBytesProcessed > 0 {
			// average ms per byte
			averageMs := time.Since(startTime) / time.Duration(totalBytesProcessed)
			remainingMs := averageMs * time.Duration(processBytesTotal-totalBytesProcessed)
			remainingDisplay := fmt.Sprintf("%.0fs", remainingMs.Seconds())
			if remainingMs.Minutes() > 120 {
				remainingDisplay = fmt.Sprintf("%.0fh", remainingMs.Hours())
			} else if remainingMs.Seconds() > 120 {
				remainingDisplay = fmt.Sprintf("%.0fm", remainingMs.Minutes())
			}
			fmt.Printf("Processing: %v/%v (%.0f%% ~ %s remaining, ave speed %v ms/byte)", i, len(fileList), percent, remainingDisplay, averageMs)
		} else {
			fmt.Printf("Processing: %v/%v (%.0f%%)", i, len(fileList), percent)
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
			return err
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
		utils.Terminal.Lowkey(fmt.Sprintf("%s (%s)", trimmedPath, filedata.Type))
		projectMap = fmt.Sprintf("%s\n%s (%s) - %s", projectMap, trimmedPath, filedata.Type, filedata.Description)
		fmt.Println()
		utils.Terminal.Lowkey(filedata.Description)
		fmt.Println()
	}

	projectMap = strings.TrimSpace(projectMap)

	// get AI description of entire directory, based on combined file analyses
	projectDesc, err := DescribeProject(projectMap)
	if err != nil {
		return errors.Join(errors.New("error generating project description"), err)
	}

	fmt.Println(projectDesc)

	return nil
}

func DescribeProject(projectMapString string) (string, error) {
	var resp DescribeProjectResponse
	err := llm.GenerateCompletionJson(prompts.P_ANALYZE_FILE_MAP_01, projectMapString, DescribeProjectSchema, &resp)
	if err != nil {
		return "", err
	}
	return resp.Description, nil
}
