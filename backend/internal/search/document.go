package search

import (
	"path"
	"strings"
	"time"

	"openshare/backend/internal/model"
	"openshare/backend/internal/resources"
)

const (
	SearchDocumentPrimaryKey = "id"

	SearchDocumentTypeFile   = "file"
	SearchDocumentTypeFolder = "folder"

	SearchFileKindPDF      = "pdf"
	SearchFileKindOffice   = "office"
	SearchFileKindMarkdown = "markdown"
	SearchFileKindText     = "text"
	SearchFileKindImage    = "image"
	SearchFileKindArchive  = "archive"
	SearchFileKindCode     = "code"
	SearchFileKindWeb      = "web"
	SearchFileKindNotebook = "notebook"
	SearchFileKindCAD      = "cad"
	SearchFileKindModel3D  = "model_3d"
	SearchFileKindMedia    = "media"
	SearchFileKindBinary   = "binary"
	SearchFileKindOther    = "other"

	SearchContentStatusNone    = "none"
	SearchContentStatusPending = "pending"
	SearchContentStatusReady   = "ready"
	SearchContentStatusFailed  = "failed"
)

var (
	SearchDocumentSearchableAttributes = []string{
		"name",
		"name_tokens",
		"course",
		"path",
		"path_segments",
		"path_tokens",
		"material_type",
		"description",
		"readme",
		"content_text",
	}

	SearchDocumentFilterableAttributes = []string{
		"resource_id",
		"type",
		"root_folder_id",
		"folder_id",
		"parent_folder_id",
		"extension",
		"file_kind",
		"category",
		"course",
		"material_type",
		"content_status",
	}

	SearchDocumentSortableAttributes = []string{
		"updated_at",
		"download_count",
		"size",
	}

	SearchDocumentRankingRules = []string{
		"words",
		"typo",
		"proximity",
		"attribute",
		"sort",
		"exactness",
	}
)

var searchIgnoredPathSegments = map[string]struct{}{
	".git":         {},
	".vs":          {},
	"__pycache__":  {},
	"bin":          {},
	"debug":        {},
	"ipch":         {},
	"node_modules": {},
	"obj":          {},
	"release":      {},
	"temp":         {},
	"tmp":          {},
	"x64":          {},
}

var searchIgnoredFileExtensions = map[string]struct{}{
	"cache":          {},
	"dll":            {},
	"exe":            {},
	"idb":            {},
	"ilk":            {},
	"lastbuildstate": {},
	"lib":            {},
	"log":            {},
	"o":              {},
	"obj":            {},
	"pch":            {},
	"pdb":            {},
	"suo":            {},
	"tlog":           {},
	"user":           {},
}

type SearchDocument struct {
	ID             string   `json:"id"`
	Type           string   `json:"type"`
	ResourceID     string   `json:"resource_id"`
	RootFolderID   string   `json:"root_folder_id,omitempty"`
	FolderID       string   `json:"folder_id,omitempty"`
	ParentFolderID string   `json:"parent_folder_id,omitempty"`
	Name           string   `json:"name"`
	Extension      string   `json:"extension,omitempty"`
	FileKind       string   `json:"file_kind,omitempty"`
	MimeType       string   `json:"mime_type,omitempty"`
	Description    string   `json:"description"`
	Readme         string   `json:"readme"`
	Path           string   `json:"path"`
	PathSegments   []string `json:"path_segments"`
	PathTokens     []string `json:"path_tokens,omitempty"`
	NameTokens     []string `json:"name_tokens,omitempty"`
	Category       string   `json:"category,omitempty"`
	Course         string   `json:"course,omitempty"`
	MaterialType   string   `json:"material_type,omitempty"`
	Size           int64    `json:"size"`
	DownloadCount  int64    `json:"download_count"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
	ContentText    string   `json:"content_text"`
	ContentStatus  string   `json:"content_status"`
}

type FileSearchDocumentInput struct {
	File          model.File
	RootFolderID  string
	FolderPath    string
	Readme        string
	ContentText   string
	ContentStatus string
}

type FolderSearchDocumentInput struct {
	Folder        model.Folder
	RootFolderID  string
	FolderPath    string
	Readme        string
	ContentText   string
	ContentStatus string
}

type SearchPathInfo struct {
	Category     string
	Course       string
	MaterialType string
}

type SearchDocumentSemantics interface {
	ShouldSkipFile(FileSearchDocumentInput) bool
	ShouldSkipFolder(FolderSearchDocumentInput) bool
	InferPathInfo([]string) SearchPathInfo
	PathTokens(SearchPathInfo, []string) []string
	NameTokens(string) []string
}

type SearchDocumentBuilder struct {
	semantics SearchDocumentSemantics
}

func NewSearchDocumentBuilder(semantics SearchDocumentSemantics) *SearchDocumentBuilder {
	if semantics == nil {
		semantics = NewGenericSearchSemantics()
	}
	return &SearchDocumentBuilder{semantics: semantics}
}

func BuildFileSearchDocument(input FileSearchDocumentInput) *SearchDocument {
	return NewSearchDocumentBuilder(nil).BuildFile(input)
}

func BuildFolderSearchDocument(input FolderSearchDocumentInput) *SearchDocument {
	return NewSearchDocumentBuilder(nil).BuildFolder(input)
}

func (b *SearchDocumentBuilder) BuildFile(input FileSearchDocumentInput) *SearchDocument {
	file := input.File
	if b.semantics.ShouldSkipFile(input) {
		return nil
	}

	folderPath := normalizeSearchDocumentPath(input.FolderPath)
	segments := splitSearchDocumentPath(folderPath)
	pathInfo := b.semantics.InferPathInfo(segments)
	extension := normalizeSearchDocumentExtension(file.Extension)
	documentPath := joinSearchDocumentPath(folderPath, file.Name)

	return &SearchDocument{
		ID:            SearchDocumentID(SearchDocumentTypeFile, file.ID),
		Type:          SearchDocumentTypeFile,
		ResourceID:    strings.TrimSpace(file.ID),
		RootFolderID:  strings.TrimSpace(input.RootFolderID),
		FolderID:      modelValue(file.FolderID),
		Name:          strings.TrimSpace(file.Name),
		Extension:     extension,
		FileKind:      inferSearchFileKind(extension, file.MimeType),
		MimeType:      strings.TrimSpace(file.MimeType),
		Description:   strings.TrimSpace(file.Description),
		Readme:        strings.TrimSpace(input.Readme),
		Path:          documentPath,
		PathSegments:  segments,
		PathTokens:    b.semantics.PathTokens(pathInfo, segments),
		NameTokens:    b.semantics.NameTokens(file.Name),
		Category:      pathInfo.Category,
		Course:        pathInfo.Course,
		MaterialType:  pathInfo.MaterialType,
		Size:          file.Size,
		DownloadCount: file.DownloadCount,
		CreatedAt:     unixSeconds(file.CreatedAt),
		UpdatedAt:     unixSeconds(preferUpdatedTime(file.UpdatedAt, file.CreatedAt)),
		ContentText:   strings.TrimSpace(input.ContentText),
		ContentStatus: normalizeSearchContentStatus(input.ContentStatus),
	}
}

func (b *SearchDocumentBuilder) BuildFolder(input FolderSearchDocumentInput) *SearchDocument {
	folder := input.Folder
	if b.semantics.ShouldSkipFolder(input) {
		return nil
	}

	folderPath := normalizeSearchDocumentPath(input.FolderPath)
	if folderPath == "" {
		folderPath = normalizeSearchDocumentPath(folder.Name)
	}
	segments := splitSearchDocumentPath(folderPath)
	pathInfo := b.semantics.InferPathInfo(segments)

	return &SearchDocument{
		ID:             SearchDocumentID(SearchDocumentTypeFolder, folder.ID),
		Type:           SearchDocumentTypeFolder,
		ResourceID:     strings.TrimSpace(folder.ID),
		RootFolderID:   strings.TrimSpace(input.RootFolderID),
		ParentFolderID: modelValue(folder.ParentID),
		Name:           strings.TrimSpace(folder.Name),
		Description:    strings.TrimSpace(folder.Description),
		Readme:         strings.TrimSpace(input.Readme),
		Path:           folderPath,
		PathSegments:   segments,
		PathTokens:     b.semantics.PathTokens(pathInfo, segments),
		NameTokens:     b.semantics.NameTokens(folder.Name),
		Category:       pathInfo.Category,
		Course:         pathInfo.Course,
		MaterialType:   pathInfo.MaterialType,
		Size:           folder.TotalSize,
		DownloadCount:  folder.DownloadCount,
		CreatedAt:      unixSeconds(folder.CreatedAt),
		UpdatedAt:      unixSeconds(preferUpdatedTime(folder.UpdatedAt, folder.CreatedAt)),
		ContentText:    strings.TrimSpace(input.ContentText),
		ContentStatus:  normalizeSearchContentStatus(input.ContentStatus),
	}
}

func SearchDocumentID(resourceType string, resourceID string) string {
	resourceType = strings.TrimSpace(resourceType)
	resourceID = strings.TrimSpace(resourceID)
	if resourceType == "" {
		resourceType = "resource"
	}
	return resourceType + "_" + resourceID
}

func normalizeSearchDocumentPath(value string) string {
	return resources.NormalizeRelativePathForStorage(value)
}

func splitSearchDocumentPath(value string) []string {
	value = normalizeSearchDocumentPath(value)
	if value == "" {
		return []string{}
	}
	return strings.Split(value, "/")
}

func hasIgnoredSearchPathSegment(value string) bool {
	for _, segment := range splitSearchDocumentPath(value) {
		if isIgnoredSearchName(segment) {
			return true
		}
	}
	return false
}

func isIgnoredSearchName(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}
	_, ignored := searchIgnoredPathSegments[normalized]
	if ignored {
		return true
	}
	if strings.HasSuffix(normalized, ".tlog") {
		return true
	}
	return false
}

func isIgnoredSearchFileExtension(extension string) bool {
	extension = normalizeSearchDocumentExtension(extension)
	_, ignored := searchIgnoredFileExtensions[extension]
	return ignored
}

func joinSearchDocumentPath(parentPath string, name string) string {
	parentPath = normalizeSearchDocumentPath(parentPath)
	name = normalizeSearchDocumentPath(path.Base(strings.TrimSpace(name)))
	if parentPath == "" {
		return name
	}
	if name == "" {
		return parentPath
	}
	return parentPath + "/" + name
}

func normalizeSearchDocumentExtension(value string) string {
	return strings.TrimLeft(strings.ToLower(strings.TrimSpace(value)), ".")
}

func inferSearchFileKind(extension string, mimeType string) string {
	extension = normalizeSearchDocumentExtension(extension)
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))

	switch extension {
	case "pdf":
		return SearchFileKindPDF
	case "doc", "docx", "xls", "xlsx", "ppt", "pptx":
		return SearchFileKindOffice
	case "md", "markdown":
		return SearchFileKindMarkdown
	case "txt", "log", "csv", "tex":
		return SearchFileKindText
	case "jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico":
		return SearchFileKindImage
	case "zip", "rar", "7z", "tar", "gz", "bz2", "xz":
		return SearchFileKindArchive
	case "c", "cc", "cpp", "cxx", "h", "hpp", "go", "java", "js", "ts", "vue", "py", "m", "cs", "asm", "v", "sh", "sql", "json", "xml", "yaml", "yml", "toml", "config", "cfg", "ini", "rc", "mak", "mk":
		return SearchFileKindCode
	case "html", "htm", "css", "asp":
		return SearchFileKindWeb
	case "ipynb", "nb":
		return SearchFileKindNotebook
	case "dwg", "vsdx":
		return SearchFileKindCAD
	case "wrl", "stl", "fbx", "gltf", "glb":
		return SearchFileKindModel3D
	case "mp3", "wav", "mp4", "mov", "avi", "flv", "swf":
		return SearchFileKindMedia
	case "exe", "dll", "lib", "pdb", "ilk", "pch", "idb", "o", "obj":
		return SearchFileKindBinary
	}

	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return SearchFileKindImage
	case strings.HasPrefix(mimeType, "text/"):
		return SearchFileKindText
	case strings.HasPrefix(mimeType, "audio/"), strings.HasPrefix(mimeType, "video/"):
		return SearchFileKindMedia
	default:
		return SearchFileKindOther
	}
}

func normalizeSearchContentStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case SearchContentStatusPending:
		return SearchContentStatusPending
	case SearchContentStatusReady:
		return SearchContentStatusReady
	case SearchContentStatusFailed:
		return SearchContentStatusFailed
	default:
		return SearchContentStatusNone
	}
}

func preferUpdatedTime(updatedAt time.Time, createdAt time.Time) time.Time {
	if !updatedAt.IsZero() {
		return updatedAt
	}
	return createdAt
}

func unixSeconds(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.Unix()
}

func modelValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
