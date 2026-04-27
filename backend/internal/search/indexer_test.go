package search

import (
	"testing"

	"openshare/backend/internal/model"
)

func TestBuildSearchIndexFolderSnapshotsBuildsPathsAndRootIDs(t *testing.T) {
	rootID := "root-1"
	childID := "child-1"
	folders := []model.Folder{
		{ID: "leaf-1", ParentID: &childID, Name: "试卷"},
		{ID: rootID, Name: "专业课"},
		{ID: childID, ParentID: &rootID, Name: "数据结构"},
	}

	snapshots := buildSearchIndexFolderSnapshots(folders)
	leaf := snapshots["leaf-1"]

	if leaf.RootID != rootID {
		t.Fatalf("RootID = %q, want %q", leaf.RootID, rootID)
	}
	if leaf.Path != "专业课/数据结构/试卷" {
		t.Fatalf("Path = %q, want 专业课/数据结构/试卷", leaf.Path)
	}
}
