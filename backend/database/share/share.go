package share

import (
	"sync"

	"github.com/SlepoyShaman/filebrowser/backend/database/users"
)

type CommonShare struct {
	DownloadsLimit           int                 `json:"downloadsLimit,omitempty"`
	ShareTheme               string              `json:"shareTheme,omitempty"`
	DisableAnonymous         bool                `json:"disableAnonymous,omitempty"`
	MaxBandwidth             int                 `json:"maxBandwidth,omitempty"`
	DisableThumbnails        bool                `json:"disableThumbnails,omitempty"`
	KeepAfterExpiration      bool                `json:"keepAfterExpiration,omitempty"`
	AllowedUsernames         []string            `json:"allowedUsernames,omitempty"`
	ThemeColor               string              `json:"themeColor,omitempty"`
	Banner                   string              `json:"banner,omitempty"`
	Title                    string              `json:"title,omitempty"`
	Description              string              `json:"description,omitempty"`
	Favicon                  string              `json:"favicon,omitempty"`
	QuickDownload            bool                `json:"quickDownload,omitempty"`
	HideNavButtons           bool                `json:"hideNavButtons,omitempty"`
	DisableSidebar           bool                `json:"disableSidebar"`
	Source                   string              `json:"source,omitempty"`
	Path                     string              `json:"path,omitempty"`
	DownloadURL              string              `json:"downloadURL,omitempty"`
	ShareURL                 string              `json:"shareURL,omitempty"`
	DisableShareCard         bool                `json:"disableShareCard,omitempty"`
	EnforceDarkLightMode     string              `json:"enforceDarkLightMode,omitempty"`
	ViewMode                 string              `json:"viewMode,omitempty"`
	EnableOnlyOffice         bool                `json:"enableOnlyOffice,omitempty"`
	ShareType                string              `json:"shareType"`
	PerUserDownloadLimit     bool                `json:"perUserDownloadLimit,omitempty"`
	ExtractEmbeddedSubtitles bool                `json:"extractEmbeddedSubtitles,omitempty"`
	AllowDelete              bool                `json:"allowDelete,omitempty"`
	AllowCreate              bool                `json:"allowCreate,omitempty"`
	AllowModify              bool                `json:"allowModify,omitempty"`
	DisableFileViewer        bool                `json:"disableFileViewer,omitempty"`
	DisableDownload          bool                `json:"disableDownload,omitempty"`
	AllowReplacements        bool                `json:"allowReplacements,omitempty"`
	SidebarLinks             []users.SidebarLink `json:"sidebarLinks"`
	HasPassword              bool                `json:"hasPassword,omitempty"`
}
type CreateBody struct {
	CommonShare
	Hash     string `json:"hash,omitempty"`
	Password string `json:"password"`
	Expires  string `json:"expires"`
	Unit     string `json:"unit"`
}

type Link struct {
	CommonShare
	Downloads     int            `json:"downloads"`
	Hash          string         `json:"hash" storm:"id,index"`
	UserID        uint           `json:"userID"`
	Expire        int64          `json:"expire"`
	PasswordHash  string         `json:"password_hash,omitempty"`
	Token         string         `json:"token,omitempty"`
	Mu            sync.Mutex     `json:"-"`
	UserDownloads map[string]int `json:"userDownloads,omitempty"`
	Version       int            `json:"version,omitempty"`
}
