package peerv2

import (
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
		if splitStr[0] == "" {
			return defaultIP, p
		}
		return splitStr[0], p
	}
	return splitStr[0], 0
}

func generateRand() []byte {
	res := make([]byte, 40)
	Logger.Info(time.Now().UnixNano())
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := 0; i < 40; i++ {
		rand := byte(rand.Intn(256))
		res[i] = rand
	}
	return res
}

// GetCommitteeIDOfTopic handle error later TODO handle error pls
func GetCommitteeIDOfTopic(topic string) int {
	topicElements := strings.Split(topic, "-")
	if len(topicElements) == 0 {
		return -1
	}
	if topicElements[1] == "" {
		return -1
	}
	cID, _ := strconv.Atoi(topicElements[1])
	return cID
}

func BatchingBlkForSync(
	batchlen int,
	info syncBlkInfo,
) []syncBlkInfo {
	res := []syncBlkInfo{}
	if info.byHash {
		rawBatches := BatchingBlkHashesForSync(batchlen, info.hashes)
		for _, rawBatch := range rawBatches {
			res = append(res, syncBlkInfo{
				byHash:        info.byHash,
				bySpecHeights: info.bySpecHeights,
				from:          info.from,
				to:            info.to,
				heights:       info.heights,
				hashes:        rawBatch,
			})
		}
		return res
	}
	if info.bySpecHeights {
		rawBatches := BatchingBlkHeightsForSync(batchlen, info.heights)
		for _, rawBatch := range rawBatches {
			res = append(res, syncBlkInfo{
				byHash:        info.byHash,
				bySpecHeights: info.bySpecHeights,
				from:          info.from,
				to:            info.to,
				heights:       rawBatch,
				hashes:        info.hashes,
			})
		}
		return res
	} else {
		rawBatches := BatchingRangeBlkForSync(uint64(batchlen), info.from, info.to)
		res = append(res, syncBlkInfo{
			byHash:        info.byHash,
			bySpecHeights: info.bySpecHeights,
			from:          info.from,
			to:            rawBatches[0],
			heights:       info.heights,
			hashes:        info.hashes,
		})
		for i := 0; i < len(rawBatches)-1; i++ {
			res = append(res, syncBlkInfo{
				byHash:        info.byHash,
				bySpecHeights: info.bySpecHeights,
				from:          rawBatches[i] + 1,
				to:            rawBatches[i+1],
				heights:       info.heights,
				hashes:        info.hashes,
			})
		}
		return res
	}
}

func BatchingBlkHeightsForSync(
	batchlen int,
	height []uint64,
) [][]uint64 {
	res := [][]uint64{}
	i := 1
	for ; i <= len(height)/batchlen; i++ {
		res = append(res, height[(i-1)*batchlen:i*batchlen])
	}
	if len(height)%batchlen != 0 {
		res = append(res, height[(i-1)*batchlen:])
	}
	return res
}

func BatchingBlkHashesForSync(
	batchlen int,
	hashesBytes [][]byte,
) [][][]byte {
	res := [][][]byte{}
	i := 1
	for ; i <= len(hashesBytes)/batchlen; i++ {
		res = append(res, hashesBytes[(i-1)*batchlen:i*batchlen])
	}
	if len(hashesBytes)%batchlen != 0 {
		res = append(res, hashesBytes[(i-1)*batchlen:])
	}
	return res
}

func BatchingRangeBlkForSync(
	batchlen uint64,
	from uint64,
	to uint64,
) []uint64 {
	res := []uint64{}
	i := uint64(1)
	for ; i <= (to-from+1)/batchlen; i++ {
		res = append(res, from+i*batchlen-1)
	}
	if ((from + (i-1)*batchlen - 1) < to) || (from == to) {
		res = append(res, to)
	}
	return res
}
