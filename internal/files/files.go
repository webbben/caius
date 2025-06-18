package files

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

var skipDirs []string = []string{".git", "node_modules"}

type GetProjectFilesOptions struct {
	SkipDotfiles bool // if true, "dotfiles" (files or directories starting with a period) will be skipped
}

func GetProjectFiles(root string, op GetProjectFilesOptions) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing path %q: %v\n", path, err)
			return err
		}
		if d.IsDir() {
			if op.SkipDotfiles && d.Name()[0] == '.' {
				return fs.SkipDir
			}
			if slices.Contains(skipDirs, d.Name()) {
				return fs.SkipDir
			}
		} else {
			if op.SkipDotfiles && filepath.Base(path)[0] == '.' {
				return nil
			}
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func IgnoreFiles(filename string) bool {
	filename = strings.ToLower(filename)

	if filename == "license" {
		return true
	}

	if !strings.Contains(filename, ".") {
		return false
	}
	ext := strings.Split(filename, ".")[1]

	switch ext {
	case "ds_store":
		return true
	default:
		return false
	}
}

// given a file type (name, file extension, etc), resolve it to its standardized display form.
// returns the resolved type, and a boolean indicating if a match was found or not (if not, its just the original input string).
func FileTypeResolver(fileType string) (string, bool) {
	fileType = strings.ToLower(fileType)
	switch fileType {
	// CONFIG TYPES
	case "yaml", "yml":
		return "YAML", true
	case "xml":
		return "XML", true
	case "json", "jsonl":
		return "JSON data", true
	// TEXT TYPES
	case "markdown", "md":
		return "markdown", true
	case "text/plain", "plain text", "text", "txt":
		return "plain text", true
	case "readme.md", "readme.txt", "readme":
		return "readme file", true
	// CODE TYPES
	case "html", "html5", "hypertext markup language":
		return "HTML", true
	case "css", "style sheet", "styles":
		return "CSS", true
	case "python", "py":
		return "python code", true
	case "go", "golang":
		return "golang code", true
	case "bash", "sh", "shell":
		return "bash/shell script", true
	case "js", "mjs", "javascript":
		return "javascript code", true
	case "jsx", "tsx", "react":
		return "react code", true
	case "ts", "typescript":
		return "typescript code", true
	case "cs", "c-sharp", "c sharp", "c#":
		return "c-sharp code", true
	default:
		return fileType, false
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

func CheckShebang(fileData []byte) string {
	// detect shebang on the first line
	lines := bytes.SplitN(fileData, []byte("\n"), 2)
	if len(lines) == 0 {
		return ""
	}
	firstLine := string(lines[0])
	// no shebang detected
	if !strings.HasPrefix(firstLine, "#!/") {
		return ""
	}

	pathParts := strings.Split(firstLine, "/")
	lastPart := pathParts[len(pathParts)-1]
	lastPart, _ = strings.CutPrefix(lastPart, "env ")

	// the last part of the path should indicate which programming language is used
	switch lastPart {
	case "bash", "sh":
		return "bash"
	case "python", "python3":
		return "python"
	case "node":
		return "javascript"
	default:
		return lastPart
	}
}
