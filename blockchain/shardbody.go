package blockchain

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type ShardBody struct {
	Instructions    [][]string
	CrossOutputCoin map[byte][]CrossOutputCoin
	Transactions    []metadata.Transaction
}
type CrossOutputCoin struct {
	BlockHeight uint64
	BlockHash   common.Hash
	OutputCoin  []privacy.OutputCoin
}

func (shardBody *ShardBody) Hash() common.Hash {
	record := []byte{}
	for shardID, refs := range shardBody.CrossOutputCoin {
		record = append(record, shardID)
		for _, ref := range refs {
			record = append(record, []byte(strconv.Itoa(int(ref.BlockHeight)))...)
			record = append(record, ref.BlockHash.GetBytes()...)
			for _, coins := range ref.OutputCoin {
				record = append(record, coins.Bytes()...)
			}
		}
	}
	for _, tx := range shardBody.Transactions {
		record = append(record, tx.Hash().GetBytes()...)
	}
	return common.DoubleHashH(record)
}

/*
Customize UnmarshalJSON to parse list TxNormal
because we have many types of block, so we can need to customize data from marshal from json string to build a block
*/
func (shardBody *ShardBody) UnmarshalJSON(data []byte) error {
	type Alias ShardBody
	temp := &struct {
		Transactions []map[string]interface{}
		*Alias
	}{
		Alias: (*Alias)(shardBody),
	}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}

	// process tx from tx interface of temp
	for _, txTemp := range temp.Transactions {
		txTempJson, _ := json.MarshalIndent(txTemp, "", "\t")
		Logger.log.Debugf("Tx json data: ", string(txTempJson))

		var tx metadata.Transaction
		var parseErr error
		switch txTemp["Type"].(string) {
		case common.TxNormalType:
			{
				tx = &transaction.Tx{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		case common.TxSalaryType:
			{
				tx = &transaction.Tx{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		case common.TxCustomTokenType:
			{
				tx = &transaction.TxCustomToken{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		case common.TxCustomTokenPrivacyType:
			{
				tx = &transaction.TxCustomTokenPrivacy{}
				parseErr = json.Unmarshal(txTempJson, &tx)
			}
		default:
			{
				return NewBlockChainError(UnmashallJsonBlockError, errors.New("Can not parse a wrong tx"))
			}
		}

		if parseErr != nil {
			return NewBlockChainError(UnmashallJsonBlockError, parseErr)
		}
		/*meta, parseErr := metadata.ParseMetadata(txTemp["Metadata"])
		if parseErr != nil {
			return NewBlockChainError(UnmashallJsonBlockError, parseErr)
		}
		tx.SetMetadata(meta)*/
		shardBody.Transactions = append(shardBody.Transactions, tx)
	}

	return nil
}
func (shardBody *CrossOutputCoin) Hash() common.Hash {
	record := []byte{}
	record = append(record, shardBody.BlockHash.GetBytes()...)
	for _, coins := range shardBody.OutputCoin {
		record = append(record, coins.Bytes()...)
	}
	return common.DoubleHashH(record)
}
func (shardBody *ShardBody) CalcMerkleRootShard() *common.Hash {
	if common.SHARD_NUMBER == 1 {
		merkleRoot := common.HashH([]byte{})
		return &merkleRoot
	}
	var shardTxs = make(map[int][]*common.Hash)

	for _, tx := range shardBody.Transactions {
		shardID := int(tx.GetSenderAddrLastByte())
		shardTxs[shardID] = append(shardTxs[shardID], tx.Hash())
	}

	shardsHash := make([]*common.Hash, common.SHARD_NUMBER)
	for idx := range shardsHash {
		h := &common.Hash{}
		shardsHash[idx], _ = h.NewHashFromStr("")
	}

	for idx, shard := range shardTxs {
		txHashStrConcat := ""

		for _, tx := range shard {
			txHashStrConcat += tx.String()
		}

		h := &common.Hash{}
		hash, _ := h.NewHashFromStr(txHashStrConcat)

		shardsHash[idx] = hash
	}

	merkleRoots := Merkle{}.BuildMerkleTreeOfHashs(shardsHash)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	return merkleRoot
}

func (shardBody *ShardBody) CalcMerkleRootTx() *common.Hash {
	merkleRoots := Merkle{}.BuildMerkleTreeStore(shardBody.Transactions)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	return merkleRoot
}

func (shardBody *ShardBody) ExtractIncomingCrossShardMap() (map[byte][]common.Hash, error) {
	crossShardMap := make(map[byte][]common.Hash)
	for shardID, crossblocks := range shardBody.CrossOutputCoin {
		for _, crossblock := range crossblocks {
			crossShardMap[shardID] = append(crossShardMap[shardID], crossblock.BlockHash)
		}

	}
	return crossShardMap, nil
}

func (shardBody *ShardBody) ExtractOutgoingCrossShardMap() (map[byte][]common.Hash, error) {
	crossShardMap := make(map[byte][]common.Hash)
	// for _, crossblock := range shardBody.CrossOutputCoin {
	// 	crossShardMap[crossblock.ShardID] = append(crossShardMap[crossblock.ShardID], crossblock.BlockHash)
	// }
	return crossShardMap, nil
}
