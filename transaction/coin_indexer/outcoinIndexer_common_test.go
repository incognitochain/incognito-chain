package coinIndexer

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"testing"
)

var (
	numTest = 100
)

func TestCoinIndexer_getIdxParamsForIndexing(t *testing.T) {
	for i := 0; i < numTest; i++ {
		fmt.Printf("TEST %v\n", i)
		//Prepare data
		scaleFactor := 50

		numWorkers := common.RandInt()%10 + 1
		fmt.Printf("totalNumWorkers: %v\n", numWorkers)
		ci := CoinIndexer{numWorkers: numWorkers}
		ci.IdxChan = make(chan IndexParam, scaleFactor*ci.numWorkers)
		ci.statusChan = make(chan JobStatus, scaleFactor*ci.numWorkers)
		ci.quitChan = make(chan bool)
		ci.idxQueue = make(map[byte][]IndexParam)
		ci.queueSize = 0
		common.MaxShardNumber = common.RandInt()%7 + 1
		fmt.Printf("#shards: %v\n", common.MaxShardNumber)
		for shardID := 0; shardID < common.MaxShardNumber; shardID++ {
			tmpIdxParams := make([]IndexParam, 0)
			r := common.RandInt() % (scaleFactor * numWorkers)
			for j := 0; j < r; j++ {
				tmpIdxParams = append(tmpIdxParams, IndexParam{})
			}

			ci.idxQueue[byte(shardID)] = tmpIdxParams
			ci.queueSize += r
			fmt.Printf("ShardID: %v, queueSize: %v\n", shardID, r)
		}
		fmt.Printf("totalQueueSize: %v\n", ci.queueSize)

		remainingWorkers := common.RandInt() % numWorkers
		fmt.Printf("remainingWorkers: %v\n", remainingWorkers)
		res := ci.getIdxParamsForIndexing(remainingWorkers)
		total := 0
		for _, num := range res {
			total += num
		}
		fmt.Printf("%v, #workers %v, #splits: %v\n", res, remainingWorkers, total)

		fmt.Printf("END TEST %v\n\n", i)
	}
}

func TestOTAKeyFromRaw(t *testing.T) {
	for i := 0; i < numTest; i++ {
		fmt.Printf("TEST %v\n", i)
		otaKey := key.GenerateOTAKey(common.RandBytes(32))

		rawOTABytes := OTAKeyToRaw(otaKey)
		newOTAKey := OTAKeyFromRaw(rawOTABytes)

		if !bytes.Equal(otaKey.GetPublicSpend().ToBytesS(), newOTAKey.GetPublicSpend().ToBytesS()) {
			panic(fmt.Errorf("public spending keys mismatch: %v != %v", otaKey.GetPublicSpend().String(), newOTAKey.GetPublicSpend().String()))
		}

		if !bytes.Equal(otaKey.GetOTASecretKey().ToBytesS(), newOTAKey.GetOTASecretKey().ToBytesS()) {
			panic(fmt.Errorf("OTA secret keys mismatch: %v != %v", otaKey.GetOTASecretKey().String(), newOTAKey.GetOTASecretKey().String()))
		}

		fmt.Printf("END TEST %v\n\n", i)
	}
}
