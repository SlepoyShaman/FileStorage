package share

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SlepoyShaman/filebrowser/backend/common/settings"
)

func (l *Link) IsSingleFileShare() bool {
	if l.Path == "" {
		return false
	}

	ext := filepath.Ext(l.Path)
	if ext != "" {
		return l.isFileOnFilesystem()
	}

	return !l.isDirectoryOnFilesystem()
}

func (l *Link) isFileOnFilesystem() bool {
	fullPath := l.Path
	if l.Source != "" {
		fullPath = filepath.Join(l.Source, l.Path)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return filepath.Ext(l.Path) != ""
	}

	return !info.IsDir()
}

func (l *Link) isDirectoryOnFilesystem() bool {
	fullPath := l.Path
	if l.Source != "" {
		fullPath = filepath.Join(l.Source, l.Path)
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return filepath.Ext(l.Path) == ""
	}

	return info.IsDir()
}

func (l *Link) IsExpired() bool {
	if l.Expire == 0 {
		return false
	}
	return false
}

func (l *Link) HasPassword() bool {
	return l.PasswordHash != ""
}

func (l *Link) IsPermanent() bool {
	return l.Expire == 0
}

func (l *Link) GetFileExtension() string {
	if l.Path == "" {
		return ""
	}
	return filepath.Ext(l.Path)
}

func (l *Link) GetFileName() string {
	if l.Path == "" {
		return ""
	}
	return filepath.Base(l.Path)
}

func (l *Link) InitUserDownloads() {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	if l.UserDownloads == nil {
		l.UserDownloads = make(map[string]int)
	}
}

func (l *Link) IncrementUserDownload(username string) {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	if l.UserDownloads == nil {
		l.UserDownloads = make(map[string]int)
	}
	l.UserDownloads[username]++
}

func (l *Link) GetUserDownloadCount(username string) int {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	if l.UserDownloads == nil {
		return 0
	}
	return l.UserDownloads[username]
}

func (l *Link) ResetDownloadCounts() {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.Downloads = 0
	l.UserDownloads = make(map[string]int)
}

func (l *Link) HasReachedUserLimit(username string) bool {
	if !l.PerUserDownloadLimit || l.DownloadsLimit == 0 {
		return false
	}
	count := l.GetUserDownloadCount(username)
	return count >= l.DownloadsLimit
}

func (l *Link) GetSourceName() (string, error) {
	sourceInfo, ok := settings.Config.Server.SourceMap[l.Source]
	if !ok {
		return "", fmt.Errorf("source not found")
	}
	return sourceInfo.Name, nil
}
