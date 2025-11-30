// file_service.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileInfo представляет информацию о файле/директории
type FileInfo struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"modTime"`
	IsDir   bool        `json:"isDir"`
}

// FileService предоставляет методы для работы с файловой системой
type FileService struct {
	basePath string
}

// NewFileService создает новый экземпляр FileService
func NewFileService(basePath string) (*FileService, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	// Проверяем, что базовый путь существует и это директория
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("base path does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("base path is not a directory")
	}

	return &FileService{basePath: absPath}, nil
}

// ListFiles возвращает список файлов и директорий в указанной директории
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
			continue // Пропускаем файлы с ошибками
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

// GetFile возвращает содержимое файла
func (fs *FileService) GetFile(relativePath string) ([]byte, error) {
	fullPath := fs.resolvePath(relativePath)

	// Проверяем, что это файл, а не директория
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

// SaveFile сохраняет содержимое в файл
func (fs *FileService) SaveFile(relativePath string, content []byte) error {
	fullPath := fs.resolvePath(relativePath)

	// Создаем директории, если их нет
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Создаем временный файл для атомарной записи
	tempPath := fullPath + ".tmp"
	if err := os.WriteFile(tempPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Атомарно заменяем старый файл новым
	if err := os.Rename(tempPath, fullPath); err != nil {
		// Пытаемся удалить временный файл в случае ошибки
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace file: %w", err)
	}

	return nil
}

// resolvePath преобразует относительный путь в абсолютный с проверкой безопасности
func (fs *FileService) resolvePath(relativePath string) string {
	// Очищаем путь от любых попыток выйти за пределы basePath
	cleanPath := filepath.Clean(relativePath)
	if cleanPath == ".." || len(cleanPath) >= 3 && cleanPath[0:3] == "../" {
		cleanPath = "."
	}

	fullPath := filepath.Join(fs.basePath, cleanPath)

	// Дополнительная проверка, что результат внутри basePath
	if !fs.isPathSafe(fullPath) {
		return fs.basePath // Возвращаем basePath в случае попытки обхода
	}

	return fullPath
}

// isPathSafe проверяет, что путь находится внутри basePath
func (fs *FileService) isPathSafe(path string) bool {
	rel, err := filepath.Rel(fs.basePath, path)
	if err != nil {
		return false
	}

	return rel != ".." && len(rel) >= 2 && rel[0:2] != ".."
}
