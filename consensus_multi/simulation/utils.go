package main

import (
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
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
