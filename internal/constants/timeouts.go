package constants

import "time"

const (
	DefaultContextTimeout = 5 * time.Second
	CacheExpiryTime       = 5 * time.Minute
	DBOperationTimeout    = 3 * time.Second
)
