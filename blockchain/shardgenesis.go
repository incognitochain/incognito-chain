package blockchain

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/transaction"
)

func CreateShardGenesisBlock(
	version int,
	net uint16,
	genesisBlockTime string,
	icoParams GenesisParams,
) *ShardBlock {
	body := ShardBody{}
	layout := "2006-01-02T15:04:05.000Z"
	str := genesisBlockTime
	genesisTime, err := time.Parse(layout, str)
	if err != nil {
		fmt.Println(err)
	}
	header := ShardHeader{
		Timestamp:         genesisTime.Unix(),
		Version:           version,
		BeaconHeight:      1,
		Epoch:             1,
		Round:             1,
		Height:            1,
		PreviousBlockHash: common.Hash{},
	}

	for _, tx := range icoParams.InitialIncognito {
		testSalaryTX, err := transaction.NewTransactionFromJsonBytes([]byte(tx))
		if err != nil {
			panic("Something is wrong when NewTransactionFromJsonBytes")
		}
		body.Transactions = append(body.Transactions, testSalaryTX)
	}

	block := &ShardBlock{
		Body:   body,
		Header: header,
	}

	return block
}
