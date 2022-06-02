package blsmultisig

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

var Cacher *cache.Cache

func init() {
	Cacher = cache.New(4*time.Hour, 4*time.Hour)
}
