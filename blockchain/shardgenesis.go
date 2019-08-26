package blockchain

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/transaction"
)

func CreateShardGenesisBlock(
	version int,
	icoParams GenesisParams,
) *ShardBlock {
	body := ShardBody{}
	layout := "2006-01-02T15:04:05.000Z"
	str := "2018-08-01T00:00:00.000Z"
	genesisTime, err := time.Parse(layout, str)
	if err != nil {
		fmt.Println(err)
	}
	header := ShardHeader{
		Timestamp:         genesisTime.Unix(),
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
