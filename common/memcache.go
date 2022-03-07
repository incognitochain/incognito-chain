package common

import (
	"github.com/dgraph-io/ristretto"
)

type MemCache struct {
	cacher Cacher
}

type Cacher interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
}

func NewMemCache(cacher Cacher) *MemCache {
	return &MemCache{cacher: cacher}
}

func NewRistrettoMemCache(maxSize int64) (Cacher, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: CacheNumCounters,
		MaxCost:     maxSize,
		BufferItems: CacheBufferItems,
		Metrics:     true,
	})
	if err != nil {
		return nil, err
	}
	return cache, nil
}
