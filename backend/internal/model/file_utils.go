package model

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ============ MIME 类型常量 ============

const (
	MimeTypePDF       = "application/pdf"
	MimeTypeZIP       = "application/zip"
	MimeTypeJSON      = "application/json"
	MimeTypeOctet     = "application/octet-stream"
	MimeTypeTextPlain = "text/plain"
	MimeTypeTextHTML  = "text/html"
	MimeTypeTextCSS   = "text/css"
	MimeTypeTextJS    = "text/javascript"
	MimeTypeTextMD    = "text/markdown"
)

// ============ 预览类型 ============

type PreviewType string

const (
	PreviewTypePDF      PreviewType = "pdf"
	PreviewTypeImage    PreviewType = "image"
	PreviewTypeText     PreviewType = "text"
	PreviewTypeCode     PreviewType = "code"
	PreviewTypeMarkdown PreviewType = "markdown"
	PreviewTypeNone     PreviewType = "none" // 不支持预览，需下载
)

// ============ 文件类型分类 ============

type FileCategory string

const (
	FileCategoryDocument FileCategory = "document" // 文档
	FileCategoryImage    FileCategory = "image"    // 图片
	FileCategoryVideo    FileCategory = "video"    // 视频
	FileCategoryAudio    FileCategory = "audio"    // 音频
	FileCategoryArchive  FileCategory = "archive"  // 压缩包
	FileCategoryCode     FileCategory = "code"     // 代码
	FileCategoryOther    FileCategory = "other"    // 其他
)

// ============ 预览支持的 MIME 类型 ============

// previewableMimeTypes 支持预览的 MIME 类型映射
var previewableMimeTypes = map[string]PreviewType{
	// PDF
	MimeTypePDF: PreviewTypePDF,
	// 图片
	"image/jpeg":    PreviewTypeImage,
	"image/png":     PreviewTypeImage,
	"image/gif":     PreviewTypeImage,
	"image/webp":    PreviewTypeImage,
	"image/svg+xml": PreviewTypeImage,
	"image/bmp":     PreviewTypeImage,
	// 文本
	MimeTypeTextPlain: PreviewTypeText,
	MimeTypeTextHTML:  PreviewTypeCode,
	MimeTypeTextCSS:   PreviewTypeCode,
	MimeTypeTextJS:    PreviewTypeCode,
	MimeTypeTextMD:    PreviewTypeMarkdown,
	MimeTypeJSON:      PreviewTypeCode,
	// 代码文件
	"text/x-python":      PreviewTypeCode,
	"text/x-java":        PreviewTypeCode,
	"text/x-c":           PreviewTypeCode,
	"text/x-go":          PreviewTypeCode,
	"application/x-sh":   PreviewTypeCode,
	"application/xml":    PreviewTypeCode,
	"application/x-yaml": PreviewTypeCode,
}

// codeExtensions 代码文件扩展名
var codeExtensions = map[string]bool{
	".go":     true,
	".py":     true,
	".js":     true,
	".ts":     true,
	".jsx":    true,
	".tsx":    true,
	".java":   true,
	".c":      true,
	".cpp":    true,
	".h":      true,
	".hpp":    true,
	".cs":     true,
	".rb":     true,
	".php":    true,
	".rs":     true,
	".swift":  true,
	".kt":     true,
	".scala":  true,
	".sh":     true,
	".bash":   true,
	".zsh":    true,
	".sql":    true,
	".html":   true,
	".css":    true,
	".scss":   true,
	".less":   true,
	".json":   true,
	".xml":    true,
	".yaml":   true,
	".yml":    true,
	".toml":   true,
	".ini":    true,
	".conf":   true,
	".vue":    true,
	".svelte": true,
}

// ============ 文件元数据工具函数 ============

// GetPreviewType 根据 MIME 类型和扩展名获取预览类型
func GetPreviewType(mimeType, extension string) PreviewType {
	// 优先根据 MIME 类型判断
	if pt, ok := previewableMimeTypes[mimeType]; ok {
		return pt
	}

	// 根据扩展名判断
	ext := strings.ToLower(extension)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	// 检查是否是代码文件
	if codeExtensions[ext] {
		return PreviewTypeCode
	}

	// Markdown 文件
	if ext == ".md" || ext == ".markdown" {
		return PreviewTypeMarkdown
	}

	// 图片类型（根据扩展名）
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".ico"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return PreviewTypeImage
		}
	}

	return PreviewTypeNone
}

// GetFileCategory 根据 MIME 类型获取文件分类
func GetFileCategory(mimeType string) FileCategory {
	// 图片
	if strings.HasPrefix(mimeType, "image/") {
		return FileCategoryImage
	}
	// 视频
	if strings.HasPrefix(mimeType, "video/") {
		return FileCategoryVideo
	}
	// 音频
	if strings.HasPrefix(mimeType, "audio/") {
		return FileCategoryAudio
	}
	// 压缩包
	archiveMimes := []string{
		"application/zip",
		"application/x-rar-compressed",
		"application/x-7z-compressed",
		"application/x-tar",
		"application/gzip",
		"application/x-bzip2",
	}
	for _, m := range archiveMimes {
		if mimeType == m {
			return FileCategoryArchive
		}
	}
	// 文档
	docMimes := []string{
		MimeTypePDF,
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}
	for _, m := range docMimes {
		if mimeType == m {
			return FileCategoryDocument
		}
	}
	// 代码/文本
	if strings.HasPrefix(mimeType, "text/") {
		return FileCategoryCode
	}

	return FileCategoryOther
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// GetExtension 从文件名获取扩展名（包含点号，小写）
func GetExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// GetBaseName 获取不带扩展名的文件名
func GetBaseName(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// SanitizeFileName 清理文件名中的非法字符
func SanitizeFileName(filename string) string {
	// 替换非法字符
	illegal := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	result := filename
	for _, char := range illegal {
		result = strings.ReplaceAll(result, char, "_")
	}
	// 移除首尾空格
	result = strings.TrimSpace(result)
	// 如果为空，使用默认名称
	if result == "" {
		result = "unnamed"
	}
	return result
}

// IsImageMimeType 检查是否是图片 MIME 类型
func IsImageMimeType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsPDFMimeType 检查是否是 PDF MIME 类型
func IsPDFMimeType(mimeType string) bool {
	return mimeType == MimeTypePDF
}

// IsTextMimeType 检查是否是文本 MIME 类型
func IsTextMimeType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "text/")
}

// ============ 文件路径工具 ============

// BuildStoragePath 构建存储路径
// basePath: 基础存储路径（如 /data/openshare/repository）
// folderPath: 文件夹相对路径（如 /课程资料/数学）
// fileName: 文件名
func BuildStoragePath(basePath, folderPath, fileName string) string {
	if folderPath == "" || folderPath == "/" {
		return filepath.Join(basePath, fileName)
	}
	return filepath.Join(basePath, folderPath, fileName)
}

// BuildTrashPath 构建回收站路径（带时间戳避免冲突）
func BuildTrashPath(trashBasePath, originalPath string, timestamp int64) string {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	newName := fmt.Sprintf("%s_%d%s", name, timestamp, ext)
	return filepath.Join(trashBasePath, dir, newName)
}

// ============ 文件哈希 ============

// HashAlgorithm 哈希算法类型
type HashAlgorithm string

const (
	HashAlgorithmMD5    HashAlgorithm = "md5"
	HashAlgorithmSHA256 HashAlgorithm = "sha256"
)

// DefaultHashAlgorithm 默认哈希算法
const DefaultHashAlgorithm = HashAlgorithmSHA256
