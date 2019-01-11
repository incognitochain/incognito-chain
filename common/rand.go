package common

import (
	"time"
	"math/rand"
)

func RandInt() int {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
	return rand.Int()
}

func RandInt64() int64 {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
	return rand.Int63()
}
