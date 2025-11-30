package file_service

import (
	"fmt"
	"os"
	"path/filepath"
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
}

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

	return &FileService{basePath: absPath}, nil
}

func (fs *FileService) ListFiles(relativePath string) ([]FileInfo, error) {
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

	return files, nil
}

func (fs *FileService) GetFile(relativePath string) ([]byte, error) {
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

	return content, nil
}

func (fs *FileService) SaveFile(relativePath string, content []byte) error {
	fullPath := fs.resolvePath(relativePath)

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

	return nil
}

func (fs *FileService) resolvePath(relativePath string) string {
	cleanPath := filepath.Clean(relativePath)
	if cleanPath == ".." || len(cleanPath) >= 3 && cleanPath[0:3] == "../" {
		cleanPath = "."
	}

	fullPath := filepath.Join(fs.basePath, cleanPath)

	if !fs.isPathSafe(fullPath) {
		return fs.basePath
	}

	return fullPath
}

func (fs *FileService) isPathSafe(path string) bool {
	rel, err := filepath.Rel(fs.basePath, path)
	if err != nil {
		return false
	}

	return rel != ".." && len(rel) >= 2 && rel[0:2] != ".."
}
