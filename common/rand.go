package common

import (
	"time"
	"math/rand"
)

func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

func RandInt64() int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63()
}
