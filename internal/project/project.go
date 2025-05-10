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

func AnalyzeFileBasic(filePath string, fileName string) (BasicFileAnalysisResponse, error) {
	sysPrompt := prompts.P_ANALYZE_FILE_01

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
			Type:        t,
			Description: "",
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
			Type:        fileType,
			Description: desc,
		}, nil
	}

	prompt := fmt.Sprintf("File name: %s\n\nFile content:\n\n%s", fileName, string(fileContent))

	var responseJson BasicFileAnalysisResponse
	err = llm.GenerateCompletionJson(sysPrompt, prompt, BasicFileAnalysisSchema, &responseJson)
	if err != nil {
		log.Println("filePath:", filePath)
		return BasicFileAnalysisResponse{}, errors.Join(errors.New("analyze file: error while generating completion;"), err)
	}

	responseJson.Type = files.FileTypeResolver(responseJson.Type)

	return responseJson, nil
}

func AnalyzeDirectory(root string) error {
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

	startTime := time.Now()

	for i, file := range fileList {
		// show progress
		percent := float64(i) / float64(len(fileList)) * 100
		if i > 0 {
			averageMs := time.Since(startTime) / time.Duration(i+1)
			remainingMs := averageMs * time.Duration(len(fileList)-i)
			remainingDisplay := fmt.Sprintf("%.0fs", remainingMs.Seconds())
			if remainingMs.Minutes() > 120 {
				remainingDisplay = fmt.Sprintf("%.0fh", remainingMs.Hours())
			} else if remainingMs.Seconds() > 120 {
				remainingDisplay = fmt.Sprintf("%.0fm", remainingMs.Minutes())
			}
			fmt.Printf("\rProcessing: %.0f%% (~ %s remaining)", percent, remainingDisplay)
		} else {
			fmt.Printf("\rProcessing: %.0f%%", percent)
		}

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
		fileDataList = append(fileDataList, fileData)
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
		trimmedPath := strings.TrimPrefix(filedata.FullPath, root)
		trimmedPath = filepath.Join(filepath.Base(root), trimmedPath)
		fmt.Println(trimmedPath, fmt.Sprintf("(%s)", filedata.Type))
		projectMap = fmt.Sprintf("%s\n%s (%s) - %s", projectMap, trimmedPath, filedata.Type, filedata.Description)
		fmt.Println()
		fmt.Println(filedata.Description)
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
