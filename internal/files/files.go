package files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var skipDirs []string = []string{".git", "node_modules"}

func GetProjectFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing path %q: %v\n", path, err)
			return err
		}
		if d.IsDir() {
			if slices.Contains(skipDirs, d.Name()) {
				return fs.SkipDir
			}
		} else {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func CommonFileExtensionMap(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case "js":
		return "javascript code"
	case "jsx":
		return "javascript/react code"
	case "ts":
		return "typescript code"
	case "tsx":
		return "typescript/react code"
	case "mjs":
		return "javascript code"
	case "json":
		return "json data"
	case "jsonl":
		return "json lines data"
	case "go":
		return "golang code"
	case "sh":
		return "shell script/bash code"
	case "py":
		return "python code"
	case "xml":
		return "xml data"
	case "yaml":
		return "yaml data"
	case "md":
		return "markdown document"
	default:
		return ""
	}
}

// given a file type, resolve it to its standardized display form
func FileTypeResolver(fileType string) string {
	fileType = strings.ToLower(fileType)
	switch fileType {
	// CONFIG TYPES
	case "yaml", "yml":
		return "YAML"
	case "xml":
		return "XML"
	case "json", "jsonl":
		return "JSON data"
	// TEXT TYPES
	case "markdown", "md":
		return "markdown"
	case "text/plain", "plain text", "text", "txt":
		return "plain text"
	// CODE TYPES
	case "html", "html5", "hypertext markup language":
		return "HTML"
	case "css", "style sheet", "styles":
		return "CSS"
	case "python", "py":
		return "python code"
	case "go", "golang":
		return "golang code"
	case "bash", "sh", "shell":
		return "bash/shell script"
	case "js", "mjs", "javascript":
		return "javascript code"
	case "jsx", "tsx", "react":
		return "react code"
	case "ts", "typescript":
		return "typescript code"
	case "cs", "c-sharp", "c sharp", "c#":
		return "c-sharp code"
	default:
		return fileType
	}
}

func ReservedFileMap(filename string) string {
	filename = strings.ToLower(filename)
	switch filename {
	case "go.mod":
		return "golang module manifest"
	case "go.sum":
		return "golang dependency checksum lockfile"
	case "package.json":
		return "javascript/typescript package manifest"
	case "package-lock.json":
		return "javascript/typescript project dependency lockfile"
	case ".gitignore":
		return "gitignore file"
	case "readme.md", "readme.txt", "readme":
		return "readme file containing project information/documentation"
	default:
		return ""
	}
}
