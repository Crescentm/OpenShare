package search

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-ego/gse"
)

const searchCustomTokenFrequency = "100000"

var (
	searchTokenSplitter = regexp.MustCompile(`[[:space:]/\\_\-+.,;:()[\]{}【】（）《》<>，。；：、！!？?&|"'“”‘’]+`)
)

type searchTokenizer struct {
	segmenter *gse.Segmenter
}

func newSearchTokenizer(customTokens []string) *searchTokenizer {
	if len(customTokens) == 0 {
		return &searchTokenizer{}
	}

	seg := gse.Segmenter{
		AlphaNum:   true,
		NotLoadHMM: true,
	}
	err := seg.LoadDictStr(searchCustomDictionary(customTokens))
	tokenizer := &searchTokenizer{}
	if err == nil {
		tokenizer.segmenter = &seg
	}
	return tokenizer
}

func searchCustomDictionary(tokens []string) string {
	var builder strings.Builder
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		builder.WriteString(token)
		builder.WriteByte(' ')
		builder.WriteString(searchCustomTokenFrequency)
		builder.WriteByte(' ')
		builder.WriteString("n")
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (t *searchTokenizer) Tokens(values ...string) []string {
	collector := newSearchTokenCollector()
	for _, value := range values {
		t.addTokens(collector, value)
	}
	return collector.Values()
}

func (t *searchTokenizer) NameTokens(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return []string{}
	}
	title := strings.TrimSuffix(name, filepath.Ext(name))
	return t.Tokens(title, name)
}

func (t *searchTokenizer) addTokens(collector *searchTokenCollector, value string) {
	value = normalizeSearchTokenText(value)
	if value == "" {
		return
	}

	for _, part := range searchTokenSplitter.Split(value, -1) {
		collector.Add(part)
	}

	if t != nil && t.segmenter != nil {
		for _, token := range t.segmenter.CutSearch(value, false) {
			collector.Add(token)
		}
		return
	}

	for _, token := range searchTokenSplitter.Split(value, -1) {
		collector.Add(token)
	}
}

type searchTokenCollector struct {
	values []string
	seen   map[string]struct{}
}

func newSearchTokenCollector() *searchTokenCollector {
	return &searchTokenCollector{
		values: make([]string, 0, 16),
		seen:   make(map[string]struct{}, 16),
	}
}

func (c *searchTokenCollector) Add(value string) {
	value = normalizeSearchToken(value)
	if !isUsefulSearchToken(value) {
		return
	}
	if _, exists := c.seen[value]; exists {
		return
	}
	c.seen[value] = struct{}{}
	c.values = append(c.values, value)
}

func (c *searchTokenCollector) Values() []string {
	if len(c.values) == 0 {
		return []string{}
	}
	return c.values
}

func normalizeSearchTokenText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, filepath.Ext(value))
	return value
}

func normalizeSearchToken(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'“”‘’`)
	return strings.ToLower(value)
}

func isUsefulSearchToken(value string) bool {
	if value == "" {
		return false
	}
	if searchStopTokens[value] {
		return false
	}
	if isNumericSearchToken(value) {
		return len(value) >= 2
	}
	if containsASCIIAlpha(value) {
		return len(value) >= 2
	}
	return utf8.RuneCountInString(value) >= 2
}

func isNumericSearchToken(value string) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func containsASCIIAlpha(value string) bool {
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

var searchStopTokens = map[string]bool{
	"and": true,
	"for": true,
	"the": true,
	"与":   true,
	"及":   true,
	"和":   true,
	"的":   true,
}
