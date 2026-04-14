package service

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"openshare/backend/internal/model"
)

const (
	defaultImportSyncDebounce      = 1 * time.Second
	defaultImportSyncAuditInterval = 6 * time.Hour
	defaultImportSyncRefreshTick   = 1 * time.Minute
	defaultImportSyncDispatchTick  = 200 * time.Millisecond
)

type managedRootSyncRequest struct {
	RootID    string
	RootName  string
	RootPath  string
	Path      string
	ForceFull bool
	ReadyAt   time.Time
}

type managedRootSyncState struct {
	ID   string
	Name string
	Path string
}

type ImportSyncManager struct {
	importService *ImportService

	debounceInterval time.Duration
	auditInterval    time.Duration
	refreshInterval  time.Duration
	dispatchInterval time.Duration

	mu          sync.Mutex
	rootsByID   map[string]managedRootSyncState
	watchedDirs map[string]struct{}
	pending     map[string]managedRootSyncRequest

	watcher   *fsnotify.Watcher
	workCh    chan managedRootSyncRequest
	refreshCh chan struct{}
}

func NewImportSyncManager(importService *ImportService) *ImportSyncManager {
	return &ImportSyncManager{
		importService:    importService,
		debounceInterval: defaultImportSyncDebounce,
		auditInterval:    defaultImportSyncAuditInterval,
		refreshInterval:  defaultImportSyncRefreshTick,
		dispatchInterval: defaultImportSyncDispatchTick,
		rootsByID:        make(map[string]managedRootSyncState),
		watchedDirs:      make(map[string]struct{}),
		pending:          make(map[string]managedRootSyncRequest),
		workCh:           make(chan managedRootSyncRequest, 64),
		refreshCh:        make(chan struct{}, 1),
	}
}

func (m *ImportSyncManager) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	m.watcher = watcher

	if err := m.refreshManagedRoots(ctx, true); err != nil {
		_ = watcher.Close()
		return err
	}

	go m.watchEvents(ctx)
	go m.watchErrors(ctx)
	go m.dispatchLoop(ctx)
	go m.workerLoop(ctx)
	go m.auditLoop(ctx)
	go m.refreshLoop(ctx)
	go func() {
		<-ctx.Done()
		_ = watcher.Close()
	}()

	return nil
}

func (m *ImportSyncManager) NotifyManagedRootsChanged() {
	select {
	case m.refreshCh <- struct{}{}:
	default:
	}
}

func (m *ImportSyncManager) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(m.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-m.refreshCh:
		}

		if err := m.refreshManagedRoots(ctx, true); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("import sync refresh failed: %v", err)
		}
	}
}

func (m *ImportSyncManager) auditLoop(ctx context.Context) {
	ticker := time.NewTicker(m.auditInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		for _, root := range m.listRoots() {
			m.enqueue(root.ID, root.Name, root.Path, root.Path, true)
		}
	}
}

func (m *ImportSyncManager) dispatchLoop(ctx context.Context) {
	ticker := time.NewTicker(m.dispatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		now := time.Now()
		ready := m.popReady(now)
		for _, item := range ready {
			select {
			case <-ctx.Done():
				return
			case m.workCh <- item:
			}
		}
	}
}

func (m *ImportSyncManager) workerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case item := <-m.workCh:
			m.processRequest(ctx, item)
		}
	}
}

func (m *ImportSyncManager) watchEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			m.handleWatchEvent(ctx, event)
		}
	}
}

func (m *ImportSyncManager) watchErrors(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			m.handleWatchError(ctx, err)
		}
	}
}

func (m *ImportSyncManager) handleWatchEvent(_ context.Context, event fsnotify.Event) {
	root, ok := m.resolveRootForPath(event.Name)
	if !ok {
		return
	}

	if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		m.removeWatchedDirsUnder(event.Name)
	}

	enqueuePath := m.resolveAffectedDirectory(root.Path, event.Name)
	if enqueuePath == "" {
		enqueuePath = root.Path
	}

	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			if err := m.addRecursiveWatches(event.Name); err != nil {
				log.Printf("import sync add recursive watch failed for %s: %v", event.Name, err)
			}
			enqueuePath = m.resolveAffectedDirectory(root.Path, filepath.Dir(event.Name))
			if enqueuePath == "" {
				enqueuePath = root.Path
			}
		}
	}

	m.enqueue(root.ID, root.Name, root.Path, enqueuePath, false)
}

func (m *ImportSyncManager) handleWatchError(ctx context.Context, watchErr error) {
	if watchErr == nil {
		return
	}

	for _, root := range m.listRoots() {
		if err := m.importService.UpdateManagedRootSyncState(
			ctx,
			root.ID,
			model.FolderSyncStateDirty,
			watchErr.Error(),
		); err != nil {
			log.Printf("import sync mark root dirty failed for %s: %v", root.Path, err)
		}
		m.enqueue(root.ID, root.Name, root.Path, root.Path, true)
	}
}

func (m *ImportSyncManager) processRequest(ctx context.Context, item managedRootSyncRequest) {
	var (
		result *ManagedDirectoryRescanResult
		err    error
	)

	switch {
	case item.ForceFull:
		result, err = m.importService.AuditManagedDirectory(ctx, item.RootID, "")
	case normalizeRescanPath(item.Path) != normalizeRescanPath(item.RootPath):
		result, err = m.importService.RescanManagedPath(ctx, item.RootID, item.Path, "")
	default:
		result, err = m.importService.RescanManagedDirectory(ctx, item.RootID, "", "")
	}
	if err != nil {
		log.Printf("import sync rescan failed for %s (%s): %v", item.RootName, item.RootPath, err)
		if updateErr := m.importService.UpdateManagedRootSyncState(
			ctx,
			item.RootID,
			model.FolderSyncStateError,
			err.Error(),
		); updateErr != nil {
			log.Printf("import sync mark root error failed for %s: %v", item.RootPath, updateErr)
		}
		return
	}

	log.Printf(
		"import sync synced %s (%s): +%d folders, +%d files, ~%d folders, ~%d files, -%d folders, -%d files",
		item.RootName,
		item.RootPath,
		result.AddedFolders,
		result.AddedFiles,
		result.UpdatedFolders,
		result.UpdatedFiles,
		result.DeletedFolders,
		result.DeletedFiles,
	)
}

func (m *ImportSyncManager) refreshManagedRoots(ctx context.Context, enqueueRoots bool) error {
	rows, err := m.importService.repository.ListManagedRoots(ctx)
	if err != nil {
		return err
	}

	nextRoots := make(map[string]managedRootSyncState, len(rows))
	newRoots := make([]managedRootSyncState, 0)

	m.mu.Lock()
	for _, row := range rows {
		root := managedRootSyncState{
			ID:   row.ID,
			Name: row.Name,
			Path: normalizeOptionalPath(row.SourcePath),
		}
		nextRoots[root.ID] = root
		if _, exists := m.rootsByID[root.ID]; !exists {
			newRoots = append(newRoots, root)
		}
	}
	m.rootsByID = nextRoots
	m.mu.Unlock()

	for _, root := range nextRoots {
		if root.Path == "" {
			continue
		}
		if err := m.addRecursiveWatches(root.Path); err != nil && !os.IsNotExist(err) {
			log.Printf("import sync add root watch failed for %s: %v", root.Path, err)
		}
	}

	m.pruneWatchesForActiveRoots()

	if enqueueRoots {
		for _, root := range newRoots {
			if root.Path == "" {
				continue
			}
			m.enqueue(root.ID, root.Name, root.Path, root.Path, false)
		}
	}

	return nil
}

func (m *ImportSyncManager) addRecursiveWatches(rootPath string) error {
	rootPath = normalizeRescanPath(rootPath)
	if rootPath == "" {
		return nil
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}

	return filepath.WalkDir(rootPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		return m.addWatch(path)
	})
}

func (m *ImportSyncManager) addWatch(path string) error {
	path = normalizeRescanPath(path)
	if path == "" {
		return nil
	}

	m.mu.Lock()
	if _, exists := m.watchedDirs[path]; exists {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	if err := m.watcher.Add(path); err != nil {
		return err
	}

	m.mu.Lock()
	m.watchedDirs[path] = struct{}{}
	m.mu.Unlock()
	return nil
}

func (m *ImportSyncManager) pruneWatchesForActiveRoots() {
	activeRoots := m.listRoots()

	m.mu.Lock()
	watched := make([]string, 0, len(m.watchedDirs))
	for path := range m.watchedDirs {
		watched = append(watched, path)
	}
	m.mu.Unlock()

	for _, path := range watched {
		if !pathBelongsToAnyRoot(path, activeRoots) {
			m.removeWatch(path)
		}
	}
}

func (m *ImportSyncManager) removeWatchedDirsUnder(prefix string) {
	prefix = normalizeRescanPath(prefix)
	if prefix == "" {
		return
	}

	m.mu.Lock()
	paths := make([]string, 0)
	for watched := range m.watchedDirs {
		if watched == prefix || strings.HasPrefix(watched, prefix+string(filepath.Separator)) {
			paths = append(paths, watched)
		}
	}
	m.mu.Unlock()

	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	for _, path := range paths {
		m.removeWatch(path)
	}
}

func (m *ImportSyncManager) removeWatch(path string) {
	path = normalizeRescanPath(path)
	if path == "" {
		return
	}
	_ = m.watcher.Remove(path)
	m.mu.Lock()
	delete(m.watchedDirs, path)
	m.mu.Unlock()
}

func (m *ImportSyncManager) enqueue(rootID, rootName, rootPath, path string, forceFull bool) {
	rootPath = normalizeRescanPath(rootPath)
	path = normalizeRescanPath(path)
	if path == "" || path == "." || !isPathWithinOrEqual(path, rootPath) {
		path = rootPath
	}

	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	for existingPath, request := range m.pending {
		if request.RootID != rootID {
			continue
		}
		switch {
		case existingPath == path:
			request.ForceFull = request.ForceFull || forceFull
			request.ReadyAt = now.Add(m.debounceInterval)
			m.pending[existingPath] = request
			return
		case isPathWithinOrEqual(path, existingPath):
			if forceFull {
				request.ForceFull = true
				request.ReadyAt = now.Add(m.debounceInterval)
				m.pending[existingPath] = request
			}
			return
		case isPathWithinOrEqual(existingPath, path):
			forceFull = forceFull || request.ForceFull
			delete(m.pending, existingPath)
		}
	}

	m.pending[path] = managedRootSyncRequest{
		RootID:    rootID,
		RootName:  rootName,
		RootPath:  rootPath,
		Path:      path,
		ForceFull: forceFull,
		ReadyAt:   now.Add(m.debounceInterval),
	}
}

func (m *ImportSyncManager) popReady(now time.Time) []managedRootSyncRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	ready := make([]managedRootSyncRequest, 0)
	for path, request := range m.pending {
		if request.ReadyAt.After(now) {
			continue
		}
		ready = append(ready, request)
		delete(m.pending, path)
	}
	sort.Slice(ready, func(i, j int) bool {
		if ready[i].RootPath != ready[j].RootPath {
			return ready[i].RootPath < ready[j].RootPath
		}
		return ready[i].Path < ready[j].Path
	})
	return ready
}

func (m *ImportSyncManager) listRoots() []managedRootSyncState {
	m.mu.Lock()
	defer m.mu.Unlock()

	roots := make([]managedRootSyncState, 0, len(m.rootsByID))
	for _, root := range m.rootsByID {
		roots = append(roots, root)
	}
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].Path != roots[j].Path {
			return roots[i].Path < roots[j].Path
		}
		return roots[i].ID < roots[j].ID
	})
	return roots
}

func (m *ImportSyncManager) resolveRootForPath(path string) (managedRootSyncState, bool) {
	path = normalizeRescanPath(path)
	if path == "" || path == "." {
		return managedRootSyncState{}, false
	}

	for _, root := range m.listRoots() {
		if root.Path == "" {
			continue
		}
		if isPathWithinOrEqual(path, root.Path) {
			return root, true
		}
	}
	return managedRootSyncState{}, false
}

func (m *ImportSyncManager) resolveAffectedDirectory(rootPath, rawPath string) string {
	path := normalizeRescanPath(rawPath)
	rootPath = normalizeRescanPath(rootPath)
	for path != "" && path != "." {
		info, err := os.Stat(path)
		if err == nil {
			if info.IsDir() {
				if isPathWithinOrEqual(path, rootPath) {
					return path
				}
				return rootPath
			}
			parent := filepath.Dir(path)
			if isPathWithinOrEqual(parent, rootPath) {
				return parent
			}
			return rootPath
		}

		next := filepath.Dir(path)
		if next == path {
			break
		}
		path = next
	}
	return rootPath
}

func pathBelongsToAnyRoot(path string, roots []managedRootSyncState) bool {
	for _, root := range roots {
		if root.Path != "" && isPathWithinOrEqual(path, root.Path) {
			return true
		}
	}
	return false
}

func isPathWithinOrEqual(path, root string) bool {
	path = normalizeRescanPath(path)
	root = normalizeRescanPath(root)
	if path == "" || root == "" {
		return false
	}
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
