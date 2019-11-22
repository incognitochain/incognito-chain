package peerv2

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func ParseListenner(s, defaultIP string, defaultPort int) (string, int) {
	if s == "" {
		return defaultIP, defaultPort
	}
	splitStr := strings.Split(s, ":")
	if len(splitStr) > 1 {
		p, e := strconv.Atoi(splitStr[1])
		if e != nil {
			panic(e)
		}
		return splitStr[0], p
	}
	return splitStr[0], 0
}

func generateRand() []byte {
	res := make([]byte, 40)
	fmt.Println(time.Now().UnixNano())
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := 0; i < 40; i++ {
		rand := byte(rand.Intn(256))
		res[i] = rand
	}
	return res
}
