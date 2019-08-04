package blockchain

import (
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/transaction"
)

func CreateShardGenesisBlock(
	version int,
	icoParams GenesisParams,
) *ShardBlock {
	body := ShardBody{}
	header := ShardHeader{
		Timestamp:         time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:            1,
		Version:           version,
		PreviousBlockHash: common.Hash{},
		BeaconHeight:      1,
		Epoch:             1,
		Round:             1,
	}

	for _, tx := range icoParams.InitialIncognito {
		testSalaryTX := transaction.Tx{}
		testSalaryTX.UnmarshalJSON([]byte(tx))
		body.Transactions = append(body.Transactions, &testSalaryTX)
	}

	block := &ShardBlock{
		Body:   body,
		Header: header,
	}

	return block
}
