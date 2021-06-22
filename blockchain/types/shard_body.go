package types

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
)

type ShardBody struct {
	Instructions      [][]string
	CrossTransactions map[byte][]CrossTransaction //CrossOutputCoin from all other shard
	Transactions      []metadata.Transaction
}

/*
Customize UnmarshalJSON to parse list TxNormal
because we have many types of block, so we can need to customize data from marshal from json string to build a block
*/
func (shardBody *ShardBody) UnmarshalJSON(data []byte) error {
	type Alias ShardBody
	temp := &struct {
		Transactions []json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(shardBody),
	}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return fmt.Errorf("unmarshall Json Shard Block Is Failed. Error %v", err)
	}

	// process tx from tx interface of temp
	for _, txTemp := range temp.Transactions {
		// txTempJson, _ := json.MarshalIndent(txTemp, "", "\t")
		var tx metadata.Transaction
		var parseErr error
		var txChoice *transaction.TxChoice
		txChoice, parseErr = transaction.DeserializeTransactionJSON(txTemp)
		if parseErr != nil {
			return fmt.Errorf("unmarshall Json Shard Block Is Failed. Error %v", parseErr)
		}
		tx = txChoice.ToTx()
		if tx == nil {
			return fmt.Errorf("unmarshall Json Shard Block Is Failed. Corrupted TX")
		}

		shardBody.Transactions = append(shardBody.Transactions, tx)
	}
	return nil
}

func (shardBody ShardBody) Hash() common.Hash {
	res := []byte{}

	for _, item := range shardBody.Instructions {
		for _, l := range item {
			res = append(res, []byte(l)...)
		}
	}
	keys := []int{}
	for k := range shardBody.CrossTransactions {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range shardBody.CrossTransactions[byte(shardID)] {
			res = append(res, []byte(fmt.Sprintf("%v", value.BlockHeight))...)
			res = append(res, value.BlockHash.GetBytes()...)
			for _, coins := range value.OutputCoin {
				res = append(res, coins.Bytes()...)
			}
			for _, coins := range value.TokenPrivacyData {
				res = append(res, coins.Bytes()...)
			}
		}
	}
	for _, tx := range shardBody.Transactions {
		res = append(res, tx.Hash().GetBytes()...)
	}
	return common.HashH(res)
}

func (shardBody ShardBody) ExtractIncomingCrossShardMap() (map[byte][]common.Hash, error) {
	crossShardMap := make(map[byte][]common.Hash)
	for shardID, crossblocks := range shardBody.CrossTransactions {
		for _, crossblock := range crossblocks {
			crossShardMap[shardID] = append(crossShardMap[shardID], crossblock.BlockHash)
		}
	}
	return crossShardMap, nil
}

func (shardBody ShardBody) ExtractOutgoingCrossShardMap() (map[byte][]common.Hash, error) {
	crossShardMap := make(map[byte][]common.Hash)
	return crossShardMap, nil
}
