package file_service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileInfo struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"modTime"`
	IsDir   bool        `json:"isDir"`
}

type FileService struct {
	basePath string
	mu       sync.RWMutex

	metadataCache map[string]*metadataCacheEntry

	fileCache map[string]*fileCacheEntry
}

type metadataCacheEntry struct {
	files     []FileInfo
	timestamp time.Time
}

type fileCacheEntry struct {
	content   []byte
	timestamp time.Time
}

const (
	metadataCacheTTL = 5 * time.Minute
	fileCacheTTL     = 10 * time.Minute
	maxFileSize      = 1 * 1024 * 1024
	cleanupInterval  = 1 * time.Minute
)

func NewFileService(basePath string) (*FileService, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("base path does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("base path is not a directory")
	}

	fs := &FileService{
		basePath:      absPath,
		metadataCache: make(map[string]*metadataCacheEntry),
		fileCache:     make(map[string]*fileCacheEntry),
	}

	go fs.startCacheCleanup()

	return fs, nil
}

func (fs *FileService) ListFiles(relativePath string) ([]FileInfo, error) {
	cacheKey := fs.normalizePath(relativePath)

	fs.mu.RLock()
	if entry, exists := fs.metadataCache[cacheKey]; exists {
		if time.Since(entry.timestamp) < metadataCacheTTL {
			files := make([]FileInfo, len(entry.files))
			copy(files, entry.files)
			fs.mu.RUnlock()
			return files, nil
		}
	}
	fs.mu.RUnlock()

	fullPath := fs.resolvePath(relativePath)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(relativePath, entry.Name()),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		}
		files = append(files, fileInfo)
	}

	fs.mu.Lock()
	fs.metadataCache[cacheKey] = &metadataCacheEntry{
		files:     files,
		timestamp: time.Now(),
	}
	fs.mu.Unlock()

	return files, nil
}

func (fs *FileService) GetFile(relativePath string) ([]byte, error) {
	cacheKey := fs.normalizePath(relativePath)

	fs.mu.RLock()
	if entry, exists := fs.fileCache[cacheKey]; exists {
		if time.Since(entry.timestamp) < fileCacheTTL {
			content := make([]byte, len(entry.content))
			copy(content, entry.content)
			fs.mu.RUnlock()
			return content, nil
		}
	}
	fs.mu.RUnlock()

	fullPath := fs.resolvePath(relativePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(content) <= maxFileSize {
		fs.mu.Lock()
		fs.fileCache[cacheKey] = &fileCacheEntry{
			content:   content,
			timestamp: time.Now(),
		}
		fs.mu.Unlock()
	}

	return content, nil
}

func (fs *FileService) SaveFile(relativePath string, content []byte) error {
	fullPath := fs.resolvePath(relativePath)
	cacheKey := fs.normalizePath(relativePath)
	dirCacheKey := fs.normalizePath(filepath.Dir(relativePath))

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	tempPath := fullPath + ".tmp"
	if err := os.WriteFile(tempPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := os.Rename(tempPath, fullPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace file: %w", err)
	}

	fs.mu.Lock()
	delete(fs.fileCache, cacheKey)
	delete(fs.metadataCache, dirCacheKey)

	if len(content) <= maxFileSize {
		fs.fileCache[cacheKey] = &fileCacheEntry{
			content:   content,
			timestamp: time.Now(),
		}
	}
	fs.mu.Unlock()

	return nil
}

func (fs *FileService) clearCache() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.metadataCache = make(map[string]*metadataCacheEntry)
	fs.fileCache = make(map[string]*fileCacheEntry)
}

func (fs *FileService) startCacheCleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		fs.cleanupExpiredCache()
	}
}

func (fs *FileService) cleanupExpiredCache() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	now := time.Now()

	for key, entry := range fs.metadataCache {
		if now.Sub(entry.timestamp) > metadataCacheTTL {
			delete(fs.metadataCache, key)
		}
	}

	for key, entry := range fs.fileCache {
		if now.Sub(entry.timestamp) > fileCacheTTL {
			delete(fs.fileCache, key)
		}
	}
}

func (fs *FileService) normalizePath(path string) string {
	return filepath.Clean(path)
}
