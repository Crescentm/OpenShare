package search

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type SemanticProfile struct {
	Categories            map[string][]string `json:"categories" yaml:"categories"`
	MaterialTypes         map[string][]string `json:"material_types" yaml:"material_types"`
	CustomTokens          []string            `json:"custom_tokens" yaml:"custom_tokens"`
	IgnoredPathSegments   []string            `json:"ignored_path_segments" yaml:"ignored_path_segments"`
	IgnoredFileExtensions []string            `json:"ignored_file_extensions" yaml:"ignored_file_extensions"`
}

type ProfileSearchSemantics struct {
	*GenericSearchSemantics
	tokenizer             *searchTokenizer
	categories            searchAliasMatcher
	materialTypes         searchAliasMatcher
	ignoredPathSegments   map[string]struct{}
	ignoredFileExtensions map[string]struct{}
}

type searchAliasMatcher struct {
	exact    map[string]string
	contains []searchAlias
}

type searchAlias struct {
	alias     string
	canonical string
}

func LoadSemanticProfile(filename string) (*SemanticProfile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read semantic profile: %w", err)
	}

	return ParseSemanticProfileYAML(data)
}

func ParseSemanticProfileYAML(data []byte) (*SemanticProfile, error) {
	var profile SemanticProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parse semantic profile: %w", err)
	}
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	return &profile, nil
}

func NewProfileSearchDocumentBuilder(profile *SemanticProfile) *SearchDocumentBuilder {
	return NewSearchDocumentBuilder(NewProfileSearchSemantics(profile))
}

func NewProfileSearchSemantics(profile *SemanticProfile) *ProfileSearchSemantics {
	if profile == nil {
		profile = &SemanticProfile{}
	}

	return &ProfileSearchSemantics{
		GenericSearchSemantics: NewGenericSearchSemantics(),
		tokenizer:              newSearchTokenizer(profile.CustomTokens),
		categories:             newSearchAliasMatcher(profile.Categories),
		materialTypes:          newSearchAliasMatcher(profile.MaterialTypes),
		ignoredPathSegments:    normalizedSearchSet(profile.IgnoredPathSegments),
		ignoredFileExtensions:  normalizedSearchExtensionSet(profile.IgnoredFileExtensions),
	}
}

func (p *SemanticProfile) Validate() error {
	if p == nil {
		return nil
	}
	if err := validateSearchAliasGroups("categories", p.Categories); err != nil {
		return err
	}
	if err := validateSearchAliasGroups("material_types", p.MaterialTypes); err != nil {
		return err
	}
	return nil
}

func (s *ProfileSearchSemantics) ShouldSkipFile(input FileSearchDocumentInput) bool {
	if s.GenericSearchSemantics.ShouldSkipFile(input) {
		return true
	}
	if s.hasProfileIgnoredPathSegment(input.FolderPath) || s.hasProfileIgnoredPathSegment(input.File.Name) {
		return true
	}
	extension := normalizeSearchDocumentExtension(input.File.Extension)
	if extension == "" {
		extension = normalizeSearchDocumentExtension(path.Ext(input.File.Name))
	}
	_, ignored := s.ignoredFileExtensions[extension]
	return ignored
}

func (s *ProfileSearchSemantics) ShouldSkipFolder(input FolderSearchDocumentInput) bool {
	if s.GenericSearchSemantics.ShouldSkipFolder(input) {
		return true
	}
	folderPath := input.FolderPath
	if strings.TrimSpace(folderPath) == "" {
		folderPath = input.Folder.Name
	}
	return s.hasProfileIgnoredPathSegment(folderPath)
}

func (s *ProfileSearchSemantics) InferPathInfo(segments []string) SearchPathInfo {
	info := SearchPathInfo{}
	if len(segments) == 0 {
		return info
	}

	start := 0
	if category := s.categories.MatchExact(segments[0]); category != "" {
		info.Category = category
		start = 1
	}

	for i := start; i < len(segments); i++ {
		segment := strings.TrimSpace(segments[i])
		if segment == "" {
			continue
		}

		if materialType := s.materialTypes.Match(segment); materialType != "" {
			if info.MaterialType == "" {
				info.MaterialType = materialType
			}
			continue
		}
		if info.Course == "" {
			info.Course = segment
		}
	}

	return info
}

func (s *ProfileSearchSemantics) PathTokens(info SearchPathInfo, segments []string) []string {
	values := make([]string, 0, len(segments)+3)
	values = append(values, info.Category, info.Course, info.MaterialType)
	values = append(values, segments...)
	return s.tokenizer.Tokens(values...)
}

func (s *ProfileSearchSemantics) NameTokens(name string) []string {
	return s.tokenizer.NameTokens(name)
}

func (s *ProfileSearchSemantics) hasProfileIgnoredPathSegment(value string) bool {
	if len(s.ignoredPathSegments) == 0 {
		return false
	}
	for _, segment := range splitSearchDocumentPath(value) {
		_, ignored := s.ignoredPathSegments[normalizeSearchAliasKey(segment)]
		if ignored {
			return true
		}
	}
	return false
}

func newSearchAliasMatcher(groups map[string][]string) searchAliasMatcher {
	matcher := searchAliasMatcher{
		exact: make(map[string]string),
	}
	canonicals := make([]string, 0, len(groups))
	for canonical := range groups {
		canonicals = append(canonicals, canonical)
	}
	sort.Strings(canonicals)

	for _, canonical := range canonicals {
		aliases := groups[canonical]
		canonical = strings.TrimSpace(canonical)
		if canonical == "" {
			continue
		}

		allAliases := append([]string{canonical}, aliases...)
		for _, alias := range allAliases {
			key := normalizeSearchAliasKey(alias)
			if key == "" {
				continue
			}
			matcher.exact[key] = canonical
			matcher.contains = append(matcher.contains, searchAlias{
				alias:     key,
				canonical: canonical,
			})
		}
	}
	sort.SliceStable(matcher.contains, func(i, j int) bool {
		return len(matcher.contains[i].alias) > len(matcher.contains[j].alias)
	})
	return matcher
}

func (m searchAliasMatcher) Match(value string) string {
	if exact := m.MatchExact(value); exact != "" {
		return exact
	}

	value = normalizeSearchAliasKey(value)
	if value == "" {
		return ""
	}
	for _, alias := range m.contains {
		if strings.Contains(value, alias.alias) {
			return alias.canonical
		}
	}
	return ""
}

func (m searchAliasMatcher) MatchExact(value string) string {
	value = normalizeSearchAliasKey(value)
	if value == "" {
		return ""
	}
	return m.exact[value]
}

func validateSearchAliasGroups(name string, groups map[string][]string) error {
	for canonical, aliases := range groups {
		if strings.TrimSpace(canonical) == "" {
			return fmt.Errorf("semantic profile %s contains empty canonical value", name)
		}
		for _, alias := range aliases {
			if strings.TrimSpace(alias) == "" {
				return fmt.Errorf("semantic profile %s[%q] contains empty alias", name, canonical)
			}
		}
	}
	return nil
}

func normalizedSearchSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = normalizeSearchAliasKey(value)
		if value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

func normalizedSearchExtensionSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = normalizeSearchDocumentExtension(value)
		if value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

func normalizeSearchAliasKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
