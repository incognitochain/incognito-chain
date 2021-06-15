package types

import (
	"encoding/json"
	"errors"
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
		Transactions []map[string]*json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(shardBody),
	}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	// process tx from tx interface of temp
	for _, txTemp := range temp.Transactions {
		txTempJson, _ := json.MarshalIndent(txTemp, "", "\t")
		var tx metadata.Transaction
		var parseErr error
		txType := ""
		err = json.Unmarshal(*txTemp["Type"], &txType)
		if err != nil {
			return err
		}
		switch txType {
		case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
			{
				tx = &transaction.Tx{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		case common.TxCustomTokenPrivacyType:
			{
				tx = &transaction.TxCustomTokenPrivacy{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		default:
			{
				return errors.New("can not parse a wrong tx")
			}
		}
		if parseErr != nil {
			return parseErr
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
