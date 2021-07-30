package common

import (
	"math/rand"
	"time"
)

// RandInt returns a random int number using math/rand
func RandInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

func RandUint64() uint64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint64()
}

// RandIntInterval returns a random int in range [L; R]
func RandIntInterval(L, R int) int {
	length := R - L + 1
	r := RandInt() % length
	return L + r
}

// RandInt64 returns a random int64 number using math/rand
func RandInt64() int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63()
}
