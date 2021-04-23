package blsmultisig

import (
	"log"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var cacher *cache.Cache
var totalCall = uint64(0)
var totalHitCache = uint64(0)
var totalCallG2 = uint64(0)
var totalHitCacheG2 = uint64(0)

func init() {
	cacher = cache.New(4*time.Hour, 4*time.Hour)
	go loop()
}

func loop() {
	for {
		log.Printf("Total cache hit/total call DecomG2 %v/%v=%v\n", totalHitCacheG2, totalCallG2, float64(totalHitCacheG2)/float64(totalCallG2))
		log.Printf("Total cache hit/total call AKGen %v/%v=%v\n", totalHitCache, totalCall, float64(totalHitCache)/float64(totalCall))
		time.Sleep(30 * time.Second)
	}
}
