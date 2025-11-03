package bolt

import (
	"github.com/SlepoyShaman/FileStorage/backend/auth"
	"github.com/SlepoyShaman/FileStorage/backend/common/errors"
	storm "github.com/asdine/storm/v3"
)

type authBackend struct {
	db *storm.DB
}

func (s authBackend) Get(t string) (auth.Auther, error) {
	var auther auth.Auther
	switch t {
	case "password":
		auther = &auth.JSONAuth{}
	case "proxy":
		auther = &auth.ProxyAuth{}
	case "noauth":
		auther = &auth.NoAuth{}
	default:
		return nil, errors.ErrInvalidAuthMethod
	}
	return auther, get(s.db, "auther", auther)
}

func (s authBackend) Save(a auth.Auther) error {
	return Save(s.db, "auther", a)
}
