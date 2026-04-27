package search

import (
	"path"
	"strings"
)

type GenericSearchSemantics struct {
	tokenizer *searchTokenizer
}

func NewGenericSearchSemantics() *GenericSearchSemantics {
	return &GenericSearchSemantics{
		tokenizer: newSearchTokenizer(nil),
	}
}

func (s *GenericSearchSemantics) ShouldSkipFile(input FileSearchDocumentInput) bool {
	file := input.File
	if hasIgnoredSearchPathSegment(input.FolderPath) {
		return true
	}
	if isIgnoredSearchName(file.Name) {
		return true
	}
	extension := normalizeSearchDocumentExtension(file.Extension)
	if extension == "" {
		extension = normalizeSearchDocumentExtension(path.Ext(file.Name))
	}
	return isIgnoredSearchFileExtension(extension)
}

func (s *GenericSearchSemantics) ShouldSkipFolder(input FolderSearchDocumentInput) bool {
	folderPath := input.FolderPath
	if strings.TrimSpace(folderPath) == "" {
		folderPath = input.Folder.Name
	}
	return hasIgnoredSearchPathSegment(folderPath)
}

func (s *GenericSearchSemantics) InferPathInfo(_ []string) SearchPathInfo {
	return SearchPathInfo{}
}

func (s *GenericSearchSemantics) PathTokens(info SearchPathInfo, segments []string) []string {
	values := make([]string, 0, len(segments)+3)
	values = append(values, info.Category, info.Course, info.MaterialType)
	values = append(values, segments...)
	return s.tokenizer.Tokens(values...)
}

func (s *GenericSearchSemantics) NameTokens(name string) []string {
	return s.tokenizer.NameTokens(name)
}
