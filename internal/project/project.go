package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/webbben/caius/internal/files"
	"github.com/webbben/caius/internal/llm"
	"github.com/webbben/caius/prompts"
)

type FileData struct {
	Filename    string
	Type        string
	Description string
	FullPath    string
}

type BasicFileAnalysisResponse struct {
	Type        string `json:"file_type"`
	Description string `json:"description"`
}

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

func AnalyzeFileBasic(filePath string, fileName string) (BasicFileAnalysisResponse, error) {
	sysPrompt := prompts.P_ANALYZE_FILE_01

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return BasicFileAnalysisResponse{}, err
	}
	prompt := fmt.Sprintf("File name: %s\n\nFile content:\n\n%s", fileName, string(fileContent))

	var responseJson BasicFileAnalysisResponse
	err = llm.GenerateCompletionJson(sysPrompt, prompt, BasicFileAnalysisSchema, &responseJson)
	if err != nil {
		return BasicFileAnalysisResponse{}, err
	}

	return responseJson, nil
}

func Analyze(root string) error {
	fileList, err := files.GetProjectFiles(root)
	if err != nil {
		return err
	}

	// for each file found, do some basic analysis:
	// - detect what type of code or data the file represents
	//   - if it has a common file extension (e.g. js, ts, go, sh, json, etc) we will handle that programmatically
	//   - if it doesn't have a recognized extension, then the LLM will analyze the file contents to decide
	// - AI will analyze the content of the file to give a brief description of its purpose
	fileDataList := make([]FileData, 0)

	for _, file := range fileList {
		filename := filepath.Base(file)
		fileData := FileData{
			Filename: filename,
			FullPath: file,
		}

		// check if file is a "reserved type" - e.g. one that we will assign a predefined description, and skip analysis for.
		reservedFileType := files.ReservedFileMap(filename)
		if reservedFileType != "" {
			fileData.Type = reservedFileType
			fileData.Description = reservedFileType
			fileDataList = append(fileDataList, fileData)
			continue
		}

		// LLM analysis of file
		fileAnalysisResponse, err := AnalyzeFileBasic(file, filename)
		if err != nil {
			// handle error
		}
		fileData.Type = fileAnalysisResponse.Type
		fileData.Description = fileAnalysisResponse.Description
		fileDataList = append(fileDataList, fileData)
	}

	// once we have finished analysis of files, create a document that stores all of this information in an easy-to-digest format for LLMs.
	// idea:
	// - a diagram of the directory and file structure
	// - short, high-level overview of what the project (or contents within) represent altogether
	// - for each file, the analytical information:
	//   - file type
	//   - brief description

	return nil
}
