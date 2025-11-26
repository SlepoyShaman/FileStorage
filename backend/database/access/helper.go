package access

import (
	"github.com/SlepoyShaman/FileStorage/backend/common/errors"
	"github.com/SlepoyShaman/FileStorage/backend/common/utils"
	"github.com/SlepoyShaman/FileStorage/backend/indexing"
	"github.com/SlepoyShaman/FileStorage/backend/indexing/iteminfo"
)

func (s *Storage) CheckChildItemAccess(response *iteminfo.ExtendedFileInfo, index *indexing.Index, username string) error {
	parentPath := index.MakeIndexPath(response.Path)

	// Проверка наличия хотя бы одного доступного элемента
	if !s.hasAnyAccessibleItems(response, index.Path, parentPath, username) {
		return errors.ErrAccessDenied
	}

	// Фильтрация элементов по правам доступа
	s.filterItemsByAccess(response, index.Path, parentPath, username)

	return nil
}

// Выделенный метод: проверка наличия доступных элементов
func (s *Storage) hasAnyAccessibleItems(response *iteminfo.ExtendedFileInfo, basePath, parentPath, username string) bool {
	allItemNames := s.collectAllItemNames(response)
	return s.HasAnyVisibleItems(basePath, parentPath, allItemNames, username)
}

// Выделенный метод: сбор всех имен элементов
func (s *Storage) collectAllItemNames(response *iteminfo.ExtendedFileInfo) []string {
	allItemNames := make([]string, 0, len(response.Folders)+len(response.Files))
	for _, folder := range response.Folders {
		allItemNames = append(allItemNames, folder.Name)
	}
	for _, file := range response.Files {
		allItemNames = append(allItemNames, file.Name)
	}
	return allItemNames
}

// Выделенный метод: фильтрация элементов по правам доступа
func (s *Storage) filterItemsByAccess(response *iteminfo.ExtendedFileInfo, basePath, parentPath, username string) {
	// Сохраняем оригинальные данные перед фильтрацией
	originalFolders := response.Folders
	originalFiles := response.Files

	// Сбрасываем и фильтруем элементы
	response.Folders = s.filterItems(originalFolders, basePath, parentPath, username)
	response.Files = s.filterItems(originalFiles, basePath, parentPath, username)
}

// Выделенный метод: универсальная фильтрация элементов
func (s *Storage) filterItems(items []iteminfo.ItemInfo, basePath, parentPath, username string) []iteminfo.ItemInfo {
	filtered := make([]iteminfo.ItemInfo, 0, len(items))

	for _, item := range items {
		indexPath := parentPath + item.Name
		if s.Permitted(basePath, indexPath, username) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

type FileOptionsExtended struct {
	utils.FileOptions
	Access *Storage
}
