package pathutil

import (
	"path/filepath"
	"strings"
)

func Clean(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func Within(path, root string) bool {
	path = Clean(path)
	root = Clean(root)
	if path == "" || root == "" || path == root {
		return false
	}

	relativePath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relativePath != ".." &&
		!strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}

func WithinOrEqual(path, root string) bool {
	path = Clean(path)
	root = Clean(root)
	if path == "" || root == "" {
		return false
	}
	if path == root {
		return true
	}
	return Within(path, root)
}
