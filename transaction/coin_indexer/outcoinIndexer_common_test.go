package coin_indexer

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"sync"
	"testing"
)

var (
	numTest = 100
)

func TestCoinIndexer_splitWorkers(t *testing.T) {
	for i := 0; i < numTest; i++ {
		fmt.Printf("TEST %v\n", i)
		//Prepare data
		ci := CoinIndexer{numWorkers: NumWorkers}
		ci.IdxChan = make(chan IndexParam, 2*ci.numWorkers)
		ci.statusChan = make(chan JobStatus, 2*ci.numWorkers)
		ci.quitChan = make(chan bool)
		ci.idxQueue = make(map[byte][]IndexParam)
		ci.queueSize = 0
		common.MaxShardNumber = common.RandInt() % 7 + 1
		fmt.Printf("#shards: %v\n", common.MaxShardNumber)
		for shardID := 0; shardID < common.MaxShardNumber; shardID++ {
			tmpIdxParams := make([]IndexParam, 0)
			r := common.RandInt() % (2* NumWorkers)
			for j := 0; j < r; j++ {
				tmpIdxParams = append(tmpIdxParams, IndexParam{})
			}

			ci.idxQueue[byte(shardID)] = tmpIdxParams
			ci.queueSize += r
			fmt.Printf("ShardID: %v, queueSize: %v\n", shardID, r)
		}
		fmt.Printf("totalQueueSize: %v\n", ci.queueSize)

		workersToBeSplit := common.RandInt() % (2 * NumWorkers)
		fmt.Printf("workersToBeSplit: %v\n", workersToBeSplit)
		res := ci.splitWorkers(workersToBeSplit)
		total := 0
		for _, num := range res {
			total += num
		}
		fmt.Printf("%v, #workers %v, #splits: %v\n", res, workersToBeSplit, total)

		fmt.Printf("END TEST %v\n\n", i)
	}
}

func TestCoinIndexer_cloneCachedCoins(t *testing.T) {
	for i := 0; i < numTest; i++ {
		fmt.Printf("===== TEST %v =====\n", i)
		//Prepare data
		ci := CoinIndexer{}
		ci.cachedCoinPubKeys = make(map[string]interface{})
		ci.mtx = new(sync.RWMutex)

		testSize := common.RandInt() % 10000
		for len(ci.cachedCoinPubKeys) < testSize {
			pubKeyStr := privacy.RandomPoint().String()
			ci.cachedCoinPubKeys[pubKeyStr] = true
		}

		clonedCache := ci.cloneCachedCoins()
		if len(clonedCache) != testSize {
			panic("")
		}

		for otaStr := range clonedCache {
			if _, ok := ci.cachedCoinPubKeys[otaStr]; !ok {
				panic(fmt.Errorf("key %v not found\n", otaStr))
			}
		}

		fmt.Printf("===== END TEST %v =====\n\n", i)
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