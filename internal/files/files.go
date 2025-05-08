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
	case "readme.md":
		return "readme file containing project information/documentation"
	case "readme.txt":
		return "readme file containing project information/documentation"
	default:
		return ""
	}
}
