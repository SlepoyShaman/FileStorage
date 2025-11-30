package cache_decorator

import (
	"path/filepath"
	"sync"
)

// FileService Интерфейс для сервиса работы с файлами
type FileService interface {
	GetFile(path string) (*File, error)
	ListFiles(dir string) ([]FileInfo, error)
	SaveFile(path string, content []byte) error
}

// CachingFileService Декоратор для кеширования
type CachingFileService struct {
	fileService FileService
	cache       map[string]interface{}
	mu          sync.RWMutex
}

func NewCachingFileService(fileService FileService) *CachingFileService {
	return &CachingFileService{
		fileService: fileService,
		cache:       make(map[string]interface{}),
	}
}

func (c *CachingFileService) GetFile(path string) (*File, error) {
	cacheKey := "file:" + path

	// Проверяем кеш
	c.mu.RLock()
	if cached, exists := c.cache[cacheKey]; exists {
		c.mu.RUnlock()
		return cached.(*File), nil
	}
	c.mu.RUnlock()

	// Получаем из основного сервиса
	file, err := c.fileService.GetFile(path)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кеш
	c.mu.Lock()
	c.cache[cacheKey] = file
	c.mu.Unlock()

	return file, nil
}

func (c *CachingFileService) ListFiles(dir string) ([]FileInfo, error) {
	cacheKey := "list:" + dir

	c.mu.RLock()
	if cached, exists := c.cache[cacheKey]; exists {
		c.mu.RUnlock()
		return cached.([]FileInfo), nil
	}
	c.mu.RUnlock()

	files, err := c.fileService.ListFiles(dir)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[cacheKey] = files
	c.mu.Unlock()

	return files, nil
}

func (c *CachingFileService) SaveFile(path string, content []byte) error {
	// При сохранении инвалидируем кеш
	c.mu.Lock()
	delete(c.cache, "file:"+path)
	delete(c.cache, "list:"+filepath.Dir(path))
	c.mu.Unlock()

	return c.fileService.SaveFile(path, content)
}
