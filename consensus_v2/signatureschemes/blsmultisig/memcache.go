package blsmultisig

import (
	lru "github.com/hashicorp/golang-lru"
)

var Cacher *lru.Cache

func init() {
	Cacher, _ = lru.New(10000)
}
