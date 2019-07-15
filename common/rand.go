package common

import (
	"time"
	"math/rand"
)

// RandInt returns a random int number using math/rand
func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

// RandInt64 returns a random int64 number using math/rand
func RandInt64() int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63()
}
