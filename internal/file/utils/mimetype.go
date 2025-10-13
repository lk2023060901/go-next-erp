package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

// DetectMIMEType 检测文件的 MIME 类型
func DetectMIMEType(reader io.Reader) (string, error) {
	// Read first 512 bytes for detection
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for MIME detection: %w", err)
	}

	// Detect MIME type
	mtype := mimetype.Detect(buffer[:n])
	return mtype.String(), nil
}

// DetectMIMETypeFromFile 从文件路径检测 MIME 类型
func DetectMIMETypeFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	mtype, err := mimetype.DetectReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to detect MIME type: %w", err)
	}

	return mtype.String(), nil
}

// DetectMIMETypeFromBytes 从字节数组检测 MIME 类型
func DetectMIMETypeFromBytes(data []byte) string {
	mtype := mimetype.Detect(data)
	return mtype.String()
}

// GetMIMEExtension 根据 MIME 类型获取推荐的文件扩展名
func GetMIMEExtension(mimeType string) string {
	mtype := mimetype.Lookup(mimeType)
	if mtype == nil {
		return ""
	}
	return mtype.Extension()
}

// IsSafeMIMEType 检查 MIME 类型是否安全（不包含可执行文件）
func IsSafeMIMEType(mimeType string) bool {
	dangerousMIME := []string{
		"application/x-msdownload",      // .exe
		"application/x-msdos-program",   // .com
		"application/x-sh",              // .sh
		"application/x-executable",      // executable
		"application/x-sharedlib",       // .so
		"application/vnd.microsoft.portable-executable", // .exe
		"text/x-python",                 // .py (可选，取决于策略)
		"text/x-shellscript",            // shell script
	}

	for _, dangerous := range dangerousMIME {
		if mimeType == dangerous {
			return false
		}
	}

	return true
}
