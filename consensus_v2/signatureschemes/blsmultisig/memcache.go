package blsmultisig

import (
	lru "github.com/hashicorp/golang-lru"
)

var cacher *lru.Cache

func init() {
	cacher, _ = lru.New(1000)
}
