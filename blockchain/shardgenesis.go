package blockchain

import (
	"time"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

func CreateShardGenesisBlock(
	version int,
	icoParams GenesisParams,
) *ShardBlock {
	body := ShardBody{}
	header := ShardHeader{
		Timestamp:       time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:          1,
		Version:         version,
		PrevBlockHash:   common.Hash{},
		BeaconHeight:    1,
		Epoch:           1,
		Round:           1,
		ProducerAddress: privacy.PaymentAddress{},
	}
	
	for _, tx := range icoParams.InitialConstant {
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
