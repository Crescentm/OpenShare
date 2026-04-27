package search

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"openshare/backend/internal/model"
)

func TestBuildFileSearchDocumentInfersPathMetadata(t *testing.T) {
	createdAt := time.Unix(1700000000, 0)
	updatedAt := time.Unix(1700100000, 0)

	doc := newTestProfileSearchDocumentBuilder(t).BuildFile(FileSearchDocumentInput{
		File: model.File{
			ID:            "file-1",
			Name:          "2022年数据结构期末试卷.pdf",
			Description:   "历年试题",
			Extension:     ".PDF",
			MimeType:      "application/pdf",
			Size:          1024,
			DownloadCount: 7,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		},
		RootFolderID:  "root-1",
		FolderPath:    " 专业课 / 数据结构 / 试卷 ",
		ContentText:   "图的遍历",
		ContentStatus: "ready",
	})

	if doc == nil {
		t.Fatal("BuildFileSearchDocument() = nil, want document")
	}
	if doc.ID != "file_file-1" {
		t.Fatalf("ID = %q, want file_file-1", doc.ID)
	}
	if doc.Type != SearchDocumentTypeFile {
		t.Fatalf("Type = %q, want %q", doc.Type, SearchDocumentTypeFile)
	}
	if doc.RootFolderID != "root-1" {
		t.Fatalf("RootFolderID = %q, want root-1", doc.RootFolderID)
	}
	if doc.Path != "专业课/数据结构/试卷/2022年数据结构期末试卷.pdf" {
		t.Fatalf("Path = %q", doc.Path)
	}
	wantSegments := []string{"专业课", "数据结构", "试卷"}
	if !reflect.DeepEqual(doc.PathSegments, wantSegments) {
		t.Fatalf("PathSegments = %#v, want %#v", doc.PathSegments, wantSegments)
	}
	if doc.Category != "专业课" {
		t.Fatalf("Category = %q, want 专业课", doc.Category)
	}
	if doc.Course != "数据结构" {
		t.Fatalf("Course = %q, want 数据结构", doc.Course)
	}
	if doc.MaterialType != "试卷" {
		t.Fatalf("MaterialType = %q, want 试卷", doc.MaterialType)
	}
	assertContains(t, doc.PathTokens, "数据结构")
	assertContains(t, doc.NameTokens, "数据结构")
	assertContains(t, doc.NameTokens, "期末试卷")
	if doc.Extension != "pdf" {
		t.Fatalf("Extension = %q, want pdf", doc.Extension)
	}
	if doc.FileKind != SearchFileKindPDF {
		t.Fatalf("FileKind = %q, want %q", doc.FileKind, SearchFileKindPDF)
	}
	if doc.ContentStatus != SearchContentStatusReady {
		t.Fatalf("ContentStatus = %q, want %q", doc.ContentStatus, SearchContentStatusReady)
	}
	if doc.CreatedAt != createdAt.Unix() || doc.UpdatedAt != updatedAt.Unix() {
		t.Fatalf("CreatedAt/UpdatedAt = %d/%d", doc.CreatedAt, doc.UpdatedAt)
	}
}

func TestBuildFolderSearchDocumentInfersMaterialTypeFromFolderPath(t *testing.T) {
	createdAt := time.Unix(1700000000, 0)

	doc := newTestProfileSearchDocumentBuilder(t).BuildFolder(FolderSearchDocumentInput{
		Folder: model.Folder{
			ID:            "folder-1",
			Name:          "复习资料",
			Description:   "考试复习",
			TotalSize:     4096,
			DownloadCount: 12,
			CreatedAt:     createdAt,
		},
		RootFolderID: "root-1",
		FolderPath:   "公共必修/高等数学/复习资料",
		Readme:       "README content",
	})

	if doc == nil {
		t.Fatal("BuildFolderSearchDocument() = nil, want document")
	}
	if doc.ID != "folder_folder-1" {
		t.Fatalf("ID = %q, want folder_folder-1", doc.ID)
	}
	if doc.Type != SearchDocumentTypeFolder {
		t.Fatalf("Type = %q, want %q", doc.Type, SearchDocumentTypeFolder)
	}
	if doc.Path != "公共必修/高等数学/复习资料" {
		t.Fatalf("Path = %q", doc.Path)
	}
	if doc.Category != "公共必修" {
		t.Fatalf("Category = %q, want 公共必修", doc.Category)
	}
	if doc.Course != "高等数学" {
		t.Fatalf("Course = %q, want 高等数学", doc.Course)
	}
	if doc.MaterialType != "复习资料" {
		t.Fatalf("MaterialType = %q, want 复习资料", doc.MaterialType)
	}
	if doc.Size != 4096 || doc.DownloadCount != 12 {
		t.Fatalf("Size/DownloadCount = %d/%d", doc.Size, doc.DownloadCount)
	}
	if doc.UpdatedAt != createdAt.Unix() {
		t.Fatalf("UpdatedAt = %d, want created timestamp fallback", doc.UpdatedAt)
	}
	if doc.ContentStatus != SearchContentStatusNone {
		t.Fatalf("ContentStatus = %q, want %q", doc.ContentStatus, SearchContentStatusNone)
	}
}

func TestBuildFolderSearchDocumentFallsBackToFolderNamePath(t *testing.T) {
	doc := newTestProfileSearchDocumentBuilder(t).BuildFolder(FolderSearchDocumentInput{
		Folder: model.Folder{
			ID:   "folder-1",
			Name: "专业课",
		},
	})

	if doc == nil {
		t.Fatal("BuildFolderSearchDocument() = nil, want document")
	}
	if doc.Path != "专业课" {
		t.Fatalf("Path = %q, want 专业课", doc.Path)
	}
	if doc.Category != "专业课" {
		t.Fatalf("Category = %q, want 专业课", doc.Category)
	}
}

func TestGenericSearchDocumentBuilderDoesNotInferProfileSemantics(t *testing.T) {
	doc := BuildFileSearchDocument(FileSearchDocumentInput{
		File: model.File{
			ID:        "file-1",
			Name:      "2022年数据结构期末试卷.pdf",
			Extension: ".pdf",
		},
		FolderPath: "专业课/数据结构/试卷",
	})

	if doc == nil {
		t.Fatal("BuildFileSearchDocument() = nil, want document")
	}
	if doc.Category != "" || doc.Course != "" || doc.MaterialType != "" {
		t.Fatalf("generic builder inferred profile fields: category=%q course=%q material_type=%q", doc.Category, doc.Course, doc.MaterialType)
	}
	if len(doc.PathTokens) == 0 || len(doc.NameTokens) == 0 {
		t.Fatalf("generic builder should still produce tokens, got path=%#v name=%#v", doc.PathTokens, doc.NameTokens)
	}
}

func TestProfileSearchSemanticsSupportsMaterialTypeAliases(t *testing.T) {
	semantics := NewProfileSearchSemantics(mustLoadTestSemanticProfile(t))
	cases := map[string]string{
		"历年真题":       "试卷",
		"電分試卷":       "试卷",
		"PPT":        "课件",
		"課件":         "课件",
		"homework-1": "作业",
		"实验报告":       "实验",
		"参考书籍":       "教材",
		"课程讲义":       "讲义",
		"题库":         "习题",
		"项目源代码":      "代码",
		"答案与解析":      "答案",
	}

	for input, want := range cases {
		info := semantics.InferPathInfo([]string{"专业课", "测试课程", input})
		if info.MaterialType != want {
			t.Fatalf("InferPathInfo(%q).MaterialType = %q, want %q", input, info.MaterialType, want)
		}
	}
}

func TestSearchDocumentAttributesStayAlignedWithModel(t *testing.T) {
	assertContains(t, SearchDocumentSearchableAttributes, "name_tokens")
	assertContains(t, SearchDocumentSearchableAttributes, "path_tokens")
	assertContains(t, SearchDocumentSearchableAttributes, "content_text")
	assertContains(t, SearchDocumentFilterableAttributes, "root_folder_id")
	assertContains(t, SearchDocumentFilterableAttributes, "file_kind")
	assertContains(t, SearchDocumentSortableAttributes, "download_count")
	if SearchDocumentPrimaryKey != "id" {
		t.Fatalf("SearchDocumentPrimaryKey = %q, want id", SearchDocumentPrimaryKey)
	}
}

func TestInferSearchFileKindCoversCourseResourceFileMix(t *testing.T) {
	cases := map[string]string{
		"pdf":   SearchFileKindPDF,
		"docx":  SearchFileKindOffice,
		"ppt":   SearchFileKindOffice,
		"jpg":   SearchFileKindImage,
		"md":    SearchFileKindMarkdown,
		"cpp":   SearchFileKindCode,
		"htm":   SearchFileKindWeb,
		"ipynb": SearchFileKindNotebook,
		"dwg":   SearchFileKindCAD,
		"wrl":   SearchFileKindModel3D,
		"swf":   SearchFileKindMedia,
		"dll":   SearchFileKindBinary,
		"obj":   SearchFileKindBinary,
	}

	for extension, want := range cases {
		if got := inferSearchFileKind(extension, ""); got != want {
			t.Fatalf("inferSearchFileKind(%q) = %q, want %q", extension, got, want)
		}
	}
}

func TestBuildFileSearchDocumentSkipsBuildArtifacts(t *testing.T) {
	cases := []FileSearchDocumentInput{
		{
			File:       model.File{Name: "hello_world.obj", Extension: ".obj"},
			FolderPath: "专业课/多核架构及编程/代码/Multi_core_homework/Debug",
		},
		{
			File:       model.File{Name: "OpenCVTest.tlog", Extension: ".tlog"},
			FolderPath: "专业课/多核架构及编程/代码/OpenCVTest",
		},
		{
			File:       model.File{Name: "project.pdb", Extension: ".pdb"},
			FolderPath: "专业课/计算机网络/课程设计/bin",
		},
		{
			File:       model.File{Name: "app.exe", Extension: ".exe"},
			FolderPath: "专业课/高频电子线路/课程设计/Release",
		},
		{
			File:       model.File{Name: "database.suo", Extension: ".suo"},
			FolderPath: "专业课/数据库原理/.vs",
		},
	}

	for _, input := range cases {
		if doc := newTestProfileSearchDocumentBuilder(t).BuildFile(input); doc != nil {
			t.Fatalf("BuildFileSearchDocument(%q, %q) returned document, want nil", input.FolderPath, input.File.Name)
		}
	}
}

func TestBuildFileSearchDocumentKeepsCourseMaterialsAndSource(t *testing.T) {
	cases := []FileSearchDocumentInput{
		{
			File:       model.File{Name: "2022年数据结构期末试卷.pdf", Extension: ".pdf"},
			FolderPath: "专业课/数据结构/试卷",
		},
		{
			File:       model.File{Name: "main.cpp", Extension: ".cpp"},
			FolderPath: "专业课/多核架构及编程/代码/Multi_core_homework",
		},
		{
			File:       model.File{Name: "实验报告.md", Extension: ".md"},
			FolderPath: "专业课/网络安全/notes",
		},
	}

	for _, input := range cases {
		if doc := newTestProfileSearchDocumentBuilder(t).BuildFile(input); doc == nil {
			t.Fatalf("BuildFileSearchDocument(%q, %q) = nil, want document", input.FolderPath, input.File.Name)
		}
	}
}

func TestBuildFolderSearchDocumentSkipsBuildArtifactSubtrees(t *testing.T) {
	cases := []FolderSearchDocumentInput{
		{Folder: model.Folder{Name: "Debug"}, FolderPath: "专业课/多核架构及编程/代码/OpenCVTest/Debug"},
		{Folder: model.Folder{Name: "Release"}, FolderPath: "专业课/计算机网络/课程设计/Release"},
		{Folder: model.Folder{Name: "bin"}, FolderPath: "专业课/社会计算/实验作业/bin"},
		{Folder: model.Folder{Name: "obj"}, FolderPath: "专业课/社会计算/实验作业/obj"},
		{Folder: model.Folder{Name: ".vs"}, FolderPath: "专业课/数据库原理/.vs"},
	}

	for _, input := range cases {
		if doc := newTestProfileSearchDocumentBuilder(t).BuildFolder(input); doc != nil {
			t.Fatalf("BuildFolderSearchDocument(%q) returned document, want nil", input.FolderPath)
		}
	}
}

func TestBuildFolderSearchDocumentKeepsNormalCourseFolders(t *testing.T) {
	input := FolderSearchDocumentInput{
		Folder:     model.Folder{Name: "课程设计"},
		FolderPath: "专业课/计算机网络/课程设计",
	}

	if doc := newTestProfileSearchDocumentBuilder(t).BuildFolder(input); doc == nil {
		t.Fatal("BuildFolderSearchDocument() = nil, want document")
	}
}

func TestSearchTokenizerUsesCustomCourseDictionary(t *testing.T) {
	profile := mustLoadTestSemanticProfile(t)
	tokens := newSearchTokenizer(profile.CustomTokens).Tokens("概率论与数理统计B试题及详细答案.pdf")

	assertContains(t, tokens, "概率论")
	assertContains(t, tokens, "数理统计")
	assertContains(t, tokens, "试题")
	assertContains(t, tokens, "详细答案")
}

func TestLoadSemanticProfile(t *testing.T) {
	profile := mustLoadTestSemanticProfile(t)
	if len(profile.Categories) == 0 {
		t.Fatal("Categories is empty")
	}
	if len(profile.MaterialTypes) == 0 {
		t.Fatal("MaterialTypes is empty")
	}
	if len(profile.CustomTokens) == 0 {
		t.Fatal("CustomTokens is empty")
	}
}

func newTestProfileSearchDocumentBuilder(t *testing.T) *SearchDocumentBuilder {
	t.Helper()
	return NewProfileSearchDocumentBuilder(mustLoadTestSemanticProfile(t))
}

func mustLoadTestSemanticProfile(t *testing.T) *SemanticProfile {
	t.Helper()
	profile, err := LoadSemanticProfile(filepath.Join("..", "..", "config", "search_semantics.openwhu.json"))
	if err != nil {
		t.Fatalf("LoadSemanticProfile() error = %v", err)
	}
	return profile
}

func assertContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("%#v does not contain %q", values, want)
}
