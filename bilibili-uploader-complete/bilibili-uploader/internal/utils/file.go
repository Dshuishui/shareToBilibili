package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// GetFileExtension 获取文件扩展名
func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// GetFileNameWithoutExt 获取不带扩展名的文件名
func GetFileNameWithoutExt(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// IsVideoFile 检查文件是否为视频文件
func IsVideoFile(filename string) bool {
	ext := GetFileExtension(filename)
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", ".m4v"}
	
	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// IsImageFile 检查文件是否为图片文件
func IsImageFile(filename string) bool {
	ext := GetFileExtension(filename)
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	
	for _, imageExt := range imageExts {
		if ext == imageExt {
			return true
		}
	}
	return false
}

// GetMimeType 根据文件扩展名获取MIME类型
func GetMimeType(filename string) string {
	ext := GetFileExtension(filename)
	
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".mkv":  "video/x-matroska",
		".webm": "video/webm",
		".m4v":  "video/x-m4v",
	}
	
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream"
}

// SaveUploadedFile 保存上传的文件
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// ReadFileChunk 读取文件的指定分片
func ReadFileChunk(filename string, offset, size int64) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 移动到指定偏移量
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, err
	}

	// 读取指定大小的数据
	chunk := make([]byte, size)
	n, err := file.Read(chunk)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return chunk[:n], nil
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CalculateChunks 计算文件需要分成多少片
func CalculateChunks(fileSize, chunkSize int64) int {
	chunks := int(fileSize / chunkSize)
	if fileSize%chunkSize != 0 {
		chunks++
	}
	return chunks
}

// FormatFileSize 格式化文件大小显示
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

