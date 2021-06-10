package utils

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"testing"
)

func TestCoinIndexer_splitWorkers(t *testing.T) {
	for i := 0; i < 1000; i++ {
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
			r := common.RandInt() % (2*NumWorkers)
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
