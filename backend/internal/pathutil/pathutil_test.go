package pathutil

import (
	"path/filepath"
	"testing"
)

func TestWithinOrEqual(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "data", "root")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "root", path: root, want: true},
		{name: "child", path: filepath.Join(root, "course", "file.pdf"), want: true},
		{name: "sibling prefix", path: filepath.Join(string(filepath.Separator), "data", "root-other"), want: false},
		{name: "parent", path: filepath.Dir(root), want: false},
		{name: "empty", path: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithinOrEqual(tt.path, root); got != tt.want {
				t.Fatalf("WithinOrEqual(%q, %q) = %v, want %v", tt.path, root, got, tt.want)
			}
		})
	}
}

func TestWithin(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "data", "root")

	if Within(root, root) {
		t.Fatalf("Within should exclude the root itself")
	}
	if !Within(filepath.Join(root, "child"), root) {
		t.Fatalf("Within should include descendants")
	}
}
