package bolt

import (
	storm "github.com/asdine/storm/v3"

	"github.com/SlepoyShaman/FileStorage/backend/auth"
	"github.com/SlepoyShaman/FileStorage/backend/common/settings"
	"github.com/SlepoyShaman/FileStorage/backend/database/access"
	"github.com/SlepoyShaman/FileStorage/backend/database/share"
	"github.com/SlepoyShaman/FileStorage/backend/database/users"
)

// Storage is a storage powered by a Backend which makes the necessary
// verifications when fetching and saving data to ensure consistency.
type BoltStore struct {
	Users    *users.Storage
	Share    *share.Storage
	Auth     *auth.Storage
	Settings *settings.Storage
	Access   *access.Storage
}

// NewStorage creates a storage.Storage based on Bolt DB.
func NewStorage(db *storm.DB) (*BoltStore, error) {
	userStore := users.NewStorage(usersBackend{db: db})
	authStore, err := auth.NewStorage(authBackend{db: db}, userStore)
	if err != nil {
		return nil, err
	}
	return &BoltStore{
		Users:    userStore,
		Share:    share.NewStorage(shareBackend{db: db}, userStore),
		Auth:     authStore,
		Settings: settings.NewStorage(settingsBackend{db: db}),
		Access:   access.NewStorage(db, userStore),
	}, nil
}
