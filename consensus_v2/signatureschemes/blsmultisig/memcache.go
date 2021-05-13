package blsmultisig

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

var cacher *cache.Cache

func init() {
	cacher = cache.New(4*time.Hour, 4*time.Hour)
}
