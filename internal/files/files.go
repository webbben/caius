package files

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
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

func UnableToProcessTypes(filename string) string {
	filename = strings.ToLower(filename)

	if !strings.Contains(filename, ".") {
		return ""
	}
	ext := strings.Split(filename, ".")[1]

	switch ext {
	// image types
	case "jpg", "jpeg", "jfif", "pjpeg", "pjp":
		return "image/jpg"
	case "png", "apng":
		return "image/png"
	case "heic":
		return "image/heic"
	case "pdf":
		return "image/pdf"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	case "webp":
		return "image/webp"
	case "ico", "cur":
		return "image/x-icon"
	case "bmp":
		return "image/bmp"
	case "tif", "tiff":
		return "image/tiff"
	// video types
	case "avi":
		return "video/x-msvideo"
	case "mp4":
		return "video/mp4"
	case "mpeg":
		return "video/mpeg"
	case "webm":
		return "video/webm"
	// audio types
	case "aac":
		return "audio/aac"
	case "mid", "midi":
		return "audio/midi"
	case "mp3":
		return "audio/mp3"
	case "wav":
		return "audio/wav"
	case "weba":
		return "audio/webm"
	// compressed data
	case "gz":
		return "gzip compressed file"
	case "zip":
		return "zip compressed file"
	case "7z":
		return "7z compressed file"
	case "rar":
		return "rar archive file"
	case "tar":
		return "tar archive file"
	case "jar":
		return "java archive file"
	// binaries, executables
	case "dll":
		return "windows dynamic library file"
	case "exe":
		return "windows executable file"
	default:
		return ""
	}
}

func IsProbablyBinaryData(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	// limit data for efficiency
	if len(data) > 8000 {
		data = data[:8000]
	}

	valid, invalid := 0, 0
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			invalid++
		} else {
			valid++
		}
		data = data[size:]
	}

	total := valid + invalid
	if total == 0 {
		// buffer was empty (not enough data to decode even a single rune)
		return false
	}

	// if 5% or more of the content is invalid utf8, it's probably binary
	return float64(invalid)/float64(total) > 0.05
}

func IsFileExecutable(fileInfo os.FileInfo) bool {
	mode := fileInfo.Mode()
	return mode&0111 != 0
}
