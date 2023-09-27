package presets

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrIDRequired     = errors.New("id required")
)
