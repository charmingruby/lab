package postgrex

import "errors"

var (
	ErrQueryPreparation       = errors.New("query preparation error")
	ErrConcurrentModification = errors.New("concurrent modification detected")
)
