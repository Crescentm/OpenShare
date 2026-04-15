package imports

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"openshare/backend/internal/model"
	"openshare/backend/pkg/identity"
)

type ManagedDirectoryRescanResult struct {
	RootPath       string `json:"root_path"`
	AddedFolders   int    `json:"added_folders"`
	AddedFiles     int    `json:"added_files"`
	UpdatedFolders int    `json:"updated_folders"`
	UpdatedFiles   int    `json:"updated_files"`
	DeletedFolders int    `json:"deleted_folders"`
	DeletedFiles   int    `json:"deleted_files"`
}

type ManagedRootRescanOutcome struct {
	FolderID       string `json:"folder_id"`
	RootPath       string `json:"root_path"`
	AddedFolders   int    `json:"added_folders"`
	AddedFiles     int    `json:"added_files"`
	UpdatedFolders int    `json:"updated_folders"`
	UpdatedFiles   int    `json:"updated_files"`
	DeletedFolders int    `json:"deleted_folders"`
	DeletedFiles   int    `json:"deleted_files"`
	Error          string `json:"error,omitempty"`
}

type managedRescanIndex struct {
	foldersByPath          map[string]ManagedSubtreeFolderRow
	childFoldersByParentID map[string]map[string]ManagedSubtreeFolderRow
	childFolderIDsByParent map[string][]string
	filesByFolderID        map[string]map[string]ManagedSubtreeFileRow
}

func (s *ImportService) RescanManagedRoots(ctx context.Context, operatorIP string) ([]ManagedRootRescanOutcome, error) {
	roots, err := s.repository.ListManagedRoots(ctx)
	if err != nil {
		return nil, fmt.Errorf("list managed roots: %w", err)
	}

	results := make([]ManagedRootRescanOutcome, 0, len(roots))
	for _, root := range roots {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		outcome := ManagedRootRescanOutcome{
			FolderID: root.ID,
			RootPath: normalizeOptionalPath(root.SourcePath),
		}
		result, rescanErr := s.rescanManagedDirectory(ctx, root.ID, "", operatorIP, nil, false)
		if rescanErr != nil {
			outcome.Error = rescanErr.Error()
			results = append(results, outcome)
			continue
		}

		outcome.RootPath = result.RootPath
		outcome.AddedFolders = result.AddedFolders
		outcome.AddedFiles = result.AddedFiles
		outcome.UpdatedFolders = result.UpdatedFolders
		outcome.UpdatedFiles = result.UpdatedFiles
		outcome.DeletedFolders = result.DeletedFolders
		outcome.DeletedFiles = result.DeletedFiles
		results = append(results, outcome)
	}

	return results, nil
}

func (s *ImportService) RescanManagedDirectory(ctx context.Context, folderID, adminID, operatorIP string) (*ManagedDirectoryRescanResult, error) {
	return s.rescanManagedDirectory(ctx, folderID, adminID, operatorIP, nil, false)
}

func (s *ImportService) RescanManagedPath(ctx context.Context, folderID, dirtyPath, operatorIP string) (*ManagedDirectoryRescanResult, error) {
	dirtyPath = normalizeRescanPath(dirtyPath)
	if dirtyPath == "." {
		dirtyPath = ""
	}
	dirtyPaths := make([]string, 0, 1)
	if dirtyPath != "" {
		dirtyPaths = append(dirtyPaths, dirtyPath)
	}
	return s.rescanManagedDirectory(ctx, folderID, "", operatorIP, dirtyPaths, false)
}

func (s *ImportService) AuditManagedDirectory(ctx context.Context, folderID, operatorIP string) (*ManagedDirectoryRescanResult, error) {
	return s.rescanManagedDirectory(ctx, folderID, "", operatorIP, nil, true)
}

func (s *ImportService) UpdateManagedRootSyncState(
	ctx context.Context,
	folderID string,
	state model.FolderSyncState,
	syncError string,
) error {
	if err := s.repository.UpdateFolderSyncState(
		ctx,
		strings.TrimSpace(folderID),
		string(state),
		strings.TrimSpace(syncError),
		s.nowFunc(),
	); err != nil {
		return fmt.Errorf("update managed root sync state: %w", err)
	}
	return nil
}

func (s *ImportService) rescanManagedDirectory(
	ctx context.Context,
	folderID,
	adminID,
	operatorIP string,
	dirtyPaths []string,
	forceFull bool,
) (*ManagedDirectoryRescanResult, error) {
	rootFolder, rootPath, err := s.loadManagedRootFolder(ctx, folderID)
	if err != nil {
		return nil, err
	}

	folders, err := s.repository.ListManagedSubtreeFolders(ctx, rootFolder.ID)
	if err != nil {
		return nil, fmt.Errorf("list managed subtree folders: %w", err)
	}
	files, err := s.repository.ListManagedSubtreeFiles(ctx, rootFolder.ID)
	if err != nil {
		return nil, fmt.Errorf("list managed subtree files: %w", err)
	}

	index := buildManagedRescanIndex(folders, files)
	rootExisting, ok := index.foldersByPath[rootPath]
	if !ok {
		return nil, fmt.Errorf("managed root is missing from subtree index")
	}

	now := s.nowFunc()
	result := &ManagedDirectoryRescanResult{RootPath: rootPath}
	dirtyPaths = normalizeDirtyPaths(rootPath, dirtyPaths)

	addedFolders := make([]*model.Folder, 0)
	updatedFolders := make([]ManagedFolderUpdate, 0)
	addedFiles := make([]*model.File, 0)
	updatedFiles := make([]ManagedFileUpdate, 0)
	deletedFolderIDs := make(map[string]struct{})
	deletedFileIDs := make(map[string]struct{})

	var scanDirectory func(currentPath string, parentID *string, existing *ManagedSubtreeFolderRow, forceHere bool) (string, error)
	scanDirectory = func(
		currentPath string,
		parentID *string,
		existing *ManagedSubtreeFolderRow,
		forceHere bool,
	) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		currentPath = normalizeRescanPath(currentPath)
		info, err := os.Stat(currentPath)
		if err != nil {
			return "", fmt.Errorf("stat managed directory %s: %w", currentPath, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("managed path is not a directory: %s", currentPath)
		}

		dirMtime := info.ModTime().UTC().UnixNano()
		folderName := filepath.Base(currentPath)

		currentFolderID := ""
		isNewFolder := existing == nil
		if isNewFolder {
			currentFolderID, err = identity.NewID()
			if err != nil {
				return "", fmt.Errorf("generate rescanned folder id: %w", err)
			}
		} else {
			currentFolderID = existing.ID
		}

		shouldScan := forceFull || forceHere || hasDirtyPathWithin(currentPath, dirtyPaths) || isNewFolder
		if existing != nil && !shouldScan {
			shouldScan = existing.SyncState != string(model.FolderSyncStateClean) || existing.FsDirMtimeNs != dirMtime
		}
		if !shouldScan {
			return currentFolderID, nil
		}

		entries, err := s.storage.ReadDirectory(currentPath)
		if err != nil {
			return "", fmt.Errorf("read managed directory %s: %w", currentPath, err)
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir != entries[j].IsDir {
				return entries[i].IsDir
			}
			return entries[i].Name < entries[j].Name
		})

		existingChildFolders := cloneManagedChildFolderMap(index.childFoldersByParentID[currentFolderID])
		existingFiles := cloneManagedChildFileMap(index.filesByFolderID[currentFolderID])

		for _, entry := range entries {
			if entry.IsDir {
				childPath := normalizeRescanPath(entry.AbsolutePath)
				childRow, exists := existingChildFolders[entry.Name]
				if exists {
					delete(existingChildFolders, entry.Name)
					childForce := forceFull || hasDirtyPathWithin(childPath, dirtyPaths)
					childCopy := childRow
					if _, err := scanDirectory(childPath, stringPtr(currentFolderID), &childCopy, childForce); err != nil {
						return "", err
					}
					continue
				}

				if _, err := scanDirectory(childPath, stringPtr(currentFolderID), nil, true); err != nil {
					return "", err
				}
				continue
			}

			existingFile, exists := existingFiles[entry.Name]
			if exists {
				delete(existingFiles, entry.Name)
				fileChanged := existingFile.FolderID == nil ||
					strings.TrimSpace(*existingFile.FolderID) != currentFolderID ||
					existingFile.Name != entry.Name ||
					existingFile.Extension != entry.Extension ||
					existingFile.MimeType != entry.MimeType ||
					existingFile.Size != entry.Size ||
					existingFile.FsFileMtimeNs != entry.ModTimeUnixNano
				if fileChanged {
					verifiedAt := now
					updatedFiles = append(updatedFiles, ManagedFileUpdate{
						ID:             existingFile.ID,
						FolderID:       stringPtr(currentFolderID),
						Name:           entry.Name,
						Description:    existingFile.Description,
						Extension:      entry.Extension,
						MimeType:       entry.MimeType,
						Size:           entry.Size,
						FsFileMtimeNs:  entry.ModTimeUnixNano,
						LastVerifiedAt: &verifiedAt,
						TouchUpdatedAt: true,
					})
					result.UpdatedFiles++
				}
				continue
			}

			fileID, err := identity.NewID()
			if err != nil {
				return "", fmt.Errorf("generate rescanned file id: %w", err)
			}
			verifiedAt := now
			addedFiles = append(addedFiles, &model.File{
				ID:             fileID,
				FolderID:       stringPtr(currentFolderID),
				Name:           entry.Name,
				Description:    "",
				Extension:      entry.Extension,
				MimeType:       entry.MimeType,
				Size:           entry.Size,
				DownloadCount:  0,
				FsFileMtimeNs:  entry.ModTimeUnixNano,
				LastVerifiedAt: &verifiedAt,
				CreatedAt:      now,
				UpdatedAt:      now,
			})
			result.AddedFiles++
		}

		for _, missingFile := range existingFiles {
			deletedFileIDs[missingFile.ID] = struct{}{}
		}
		for _, missingFolder := range existingChildFolders {
			collectDeletedManagedSubtree(missingFolder.ID, index, deletedFolderIDs, deletedFileIDs)
		}

		scannedAt := now
		if isNewFolder {
			addedFolders = append(addedFolders, &model.Folder{
				ID:            currentFolderID,
				ParentID:      parentID,
				SourcePath:    stringPtr(currentPath),
				Name:          folderName,
				Description:   "",
				FsDirMtimeNs:  dirMtime,
				LastScannedAt: &scannedAt,
				SyncState:     string(model.FolderSyncStateClean),
				SyncError:     "",
				CreatedAt:     now,
				UpdatedAt:     now,
			})
			result.AddedFolders++
			return currentFolderID, nil
		}

		touchUpdatedAt := optionalStringValue(existing.ParentID) != optionalStringValue(parentID) ||
			existing.Name != folderName ||
			normalizeOptionalPath(existing.SourcePath) != currentPath
		updatedFolders = append(updatedFolders, ManagedFolderUpdate{
			ID:             existing.ID,
			ParentID:       parentID,
			Name:           folderName,
			SourcePath:     currentPath,
			FsDirMtimeNs:   dirMtime,
			LastScannedAt:  &scannedAt,
			SyncState:      string(model.FolderSyncStateClean),
			SyncError:      "",
			TouchUpdatedAt: touchUpdatedAt,
		})
		if touchUpdatedAt {
			result.UpdatedFolders++
		}
		return currentFolderID, nil
	}

	rootCopy := rootExisting
	if _, err := scanDirectory(rootPath, nil, &rootCopy, forceFull); err != nil {
		return nil, err
	}

	result.DeletedFolders = len(deletedFolderIDs)
	result.DeletedFiles = len(deletedFileIDs)

	deletedFolderIDList := mapKeys(deletedFolderIDs)
	deletedFileIDList := mapKeys(deletedFileIDs)
	sort.Strings(deletedFolderIDList)
	sort.Strings(deletedFileIDList)

	detail, _ := json.Marshal(result)
	if err := s.repository.ApplyRescanSync(ctx, RescanSyncInput{
		RootFolderID:     rootFolder.ID,
		OperatorID:       adminID,
		OperatorIP:       operatorIP,
		Detail:           string(detail),
		Now:              now,
		AddedFolders:     addedFolders,
		UpdatedFolders:   updatedFolders,
		DeletedFolderIDs: deletedFolderIDList,
		AddedFiles:       addedFiles,
		UpdatedFiles:     updatedFiles,
		DeletedFileIDs:   deletedFileIDList,
	}); err != nil {
		return nil, fmt.Errorf("apply managed directory rescan: %w", err)
	}

	return result, nil
}

func (s *ImportService) loadManagedRootFolder(ctx context.Context, folderID string) (*model.Folder, string, error) {
	rootFolder, err := s.repository.FindFolderByID(ctx, strings.TrimSpace(folderID))
	if err != nil {
		return nil, "", fmt.Errorf("find managed root: %w", err)
	}
	if rootFolder == nil {
		return nil, "", ErrFolderTreeNotFound
	}
	if rootFolder.ParentID != nil {
		return nil, "", ErrManagedRootRequired
	}
	if rootFolder.SourcePath == nil || strings.TrimSpace(*rootFolder.SourcePath) == "" {
		return nil, "", &ManagedDirectoryUnavailableError{}
	}

	rootPath := filepath.Clean(strings.TrimSpace(*rootFolder.SourcePath))
	info, err := os.Stat(rootPath)
	if err != nil || !info.IsDir() {
		return nil, "", &ManagedDirectoryUnavailableError{Path: rootPath}
	}
	return rootFolder, rootPath, nil
}

func buildManagedRescanIndex(
	folders []ManagedSubtreeFolderRow,
	files []ManagedSubtreeFileRow,
) managedRescanIndex {
	index := managedRescanIndex{
		foldersByPath:          make(map[string]ManagedSubtreeFolderRow, len(folders)),
		childFoldersByParentID: make(map[string]map[string]ManagedSubtreeFolderRow),
		childFolderIDsByParent: make(map[string][]string),
		filesByFolderID:        make(map[string]map[string]ManagedSubtreeFileRow),
	}

	for _, folder := range folders {
		path := normalizeOptionalPath(folder.SourcePath)
		if path != "" {
			index.foldersByPath[path] = folder
		}
		if folder.ParentID != nil {
			parentID := strings.TrimSpace(*folder.ParentID)
			if _, ok := index.childFoldersByParentID[parentID]; !ok {
				index.childFoldersByParentID[parentID] = make(map[string]ManagedSubtreeFolderRow)
			}
			index.childFoldersByParentID[parentID][folder.Name] = folder
			index.childFolderIDsByParent[parentID] = append(index.childFolderIDsByParent[parentID], folder.ID)
		}
	}

	for _, file := range files {
		folderID := optionalStringValue(file.FolderID)
		if folderID == "" {
			continue
		}
		if _, ok := index.filesByFolderID[folderID]; !ok {
			index.filesByFolderID[folderID] = make(map[string]ManagedSubtreeFileRow)
		}
		index.filesByFolderID[folderID][file.Name] = file
	}

	return index
}

func collectDeletedManagedSubtree(
	folderID string,
	index managedRescanIndex,
	deletedFolderIDs map[string]struct{},
	deletedFileIDs map[string]struct{},
) {
	if _, exists := deletedFolderIDs[folderID]; exists {
		return
	}
	deletedFolderIDs[folderID] = struct{}{}

	for _, file := range index.filesByFolderID[folderID] {
		deletedFileIDs[file.ID] = struct{}{}
	}
	for _, childID := range index.childFolderIDsByParent[folderID] {
		collectDeletedManagedSubtree(childID, index, deletedFolderIDs, deletedFileIDs)
	}
}

func normalizeDirtyPaths(rootPath string, paths []string) []string {
	rootPath = normalizeRescanPath(rootPath)
	if rootPath == "" {
		return nil
	}

	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		candidate := normalizeRescanPath(path)
		if candidate == "" || candidate == "." {
			continue
		}
		if candidate == rootPath || strings.HasPrefix(candidate, rootPath+string(filepath.Separator)) {
			normalized = append(normalized, candidate)
		}
	}
	sort.Strings(normalized)
	return normalized
}

func hasDirtyPathWithin(currentPath string, dirtyPaths []string) bool {
	currentPath = normalizeRescanPath(currentPath)
	for _, dirtyPath := range dirtyPaths {
		if dirtyPath == currentPath || strings.HasPrefix(dirtyPath, currentPath+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func cloneManagedChildFolderMap(
	source map[string]ManagedSubtreeFolderRow,
) map[string]ManagedSubtreeFolderRow {
	if len(source) == 0 {
		return map[string]ManagedSubtreeFolderRow{}
	}
	cloned := make(map[string]ManagedSubtreeFolderRow, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func cloneManagedChildFileMap(
	source map[string]ManagedSubtreeFileRow,
) map[string]ManagedSubtreeFileRow {
	if len(source) == 0 {
		return map[string]ManagedSubtreeFileRow{}
	}
	cloned := make(map[string]ManagedSubtreeFileRow, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func mapKeys(source map[string]struct{}) []string {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	return keys
}
