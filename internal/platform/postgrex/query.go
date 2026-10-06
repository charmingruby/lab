package postgrex

import "time"

const (
	DefaultWriteTimeout = 6 * time.Second
	DefaultReadTimeout  = 4 * time.Second
)
