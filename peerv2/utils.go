package peerv2

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/pkg/errors"
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

func batchingBlkForSync(
	batchlen int,
	info syncBlkInfo,
) []syncBlkInfo {
	res := []syncBlkInfo{}
	if info.byHash {
		rawBatches := batchingBlkHashesForSync(batchlen, info.hashes)
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
		rawBatches := batchingBlkHeightsForSync(batchlen, info.heights)
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
		rawBatches := batchingRangeBlkForSync(uint64(batchlen), info.from, info.to)
		for i := 0; i < len(rawBatches)-1; i++ {
			res = append(res, syncBlkInfo{
				byHash:        info.byHash,
				bySpecHeights: info.bySpecHeights,
				from:          rawBatches[i],
				to:            rawBatches[i+1],
				heights:       info.heights,
				hashes:        info.hashes,
			})
		}
		return res
	}
}

func batchingBlkHeightsForSync(
	batchlen int,
	height []uint64,
) [][]uint64 {
	res := [][]uint64{}
	i := 0
	for ; i+batchlen < len(height); i += batchlen {
		res = append(res, height[i:i+batchlen])
	}
	if i < len(height) {
		res = append(res, height[i:])
	}
	return res
}

func batchingBlkHashesForSync(
	batchlen int,
	hashesBytes [][]byte,
) [][][]byte {
	res := [][][]byte{}
	i := 0
	for ; i+batchlen < len(hashesBytes); i += batchlen {
		res = append(res, hashesBytes[i:i+batchlen])
	}
	if i < len(hashesBytes) {
		res = append(res, hashesBytes[i:])
	}
	return res
}

func batchingRangeBlkForSync(
	batchlen uint64,
	from uint64,
	to uint64,
) []uint64 {
	res := []uint64{}
	if from == to {
		res = append(res, from)
	}
	for i := from; i < to; i += batchlen {
		res = append(res, i)
	}
	res = append(res, to)
	return res
}

func getAvgRTT(
	ctx context.Context,
	ts <-chan ping.Result,
) (
	time.Duration,
	error,
) {
	s := time.Duration(0)
	for i := 0; i < PingTimes; i++ {
		select {
		case res := <-ts:
			if res.Error != nil {
				return 0, res.Error
			}
			s += res.RTT
		case <-time.After(PingTimeout):
			return 0, errors.Errorf("Ping timeout!! %v", ctx.Value("UUID"))
		}
	}
	return s / PingTimes, nil
}
