package errors

import "errors"

var (
	ErrInvalidOption = errors.New("invalid option")
	ErrNotIndexed    = errors.New("directory or item excluded from indexing")
)
