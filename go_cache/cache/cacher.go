package cache

import "time"

// time.Duration is for TTL(Time to Live) which determines how long to cache
type Cacher interface {
	Set([]byte, []byte, time.Duration) error
	Get([]byte) ([]byte, error)
	Has([]byte) bool
	Delete([]byte) error
}
