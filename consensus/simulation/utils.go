package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbftv2"
	"math"
	"time"
)

var START_TIME = time.Now().Unix()

func failOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func NewEmptyBlock() common.BlockInterface {
	return &blockchain.ShardBlock{}
}

func NewBlock(height uint64, time int64, producer string, prev common.Hash) common.BlockInterface {
	return &blockchain.ShardBlock{
		Header: blockchain.ShardHeader{
			Version:           1,
			Height:            height,
			Round:             1,
			Epoch:             1,
			Timestamp:         time,
			PreviousBlockHash: prev,
			Producer:          producer,
		},
		Body: blockchain.ShardBody{},
	}
}

func GetTimeSlot(t int64) int64 {
	fmt.Println(t+1, START_TIME)
	return int64(math.Ceil(float64(t+1-START_TIME) / blsbftv2.TIMESLOT))
}
func NextTimeSlot(t int64) int64 {
	return t + blsbftv2.TIMESLOT
}
