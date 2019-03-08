package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type ShardBody struct {
	Instructions [][]string
	//CrossOutputCoin from all other shard
	CrossOutputCoin map[byte][]CrossOutputCoin
	Transactions    []metadata.Transaction
}
type CrossOutputCoin struct {
	BlockHeight uint64
	BlockHash   common.Hash
	OutputCoin  []privacy.OutputCoin
}
type CrossTxTokenData struct {
	BlockHeight uint64
	BlockHash   common.Hash
	TxTokenData []transaction.TxTokenData
}

func (shardBody *ShardBody) Hash() common.Hash {
	res := []byte{}

	for _, item := range shardBody.Instructions {
		for _, l := range item {
			res = append(res, []byte(l)...)
		}
	}
	keys := []int{}
	for k := range shardBody.CrossOutputCoin {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range shardBody.CrossOutputCoin[byte(shardID)] {
			res = append(res, []byte(fmt.Sprintf("%v", value.BlockHeight))...)
			res = append(res, value.BlockHash.GetBytes()...)
			for _, coins := range value.OutputCoin {
				res = append(res, coins.Bytes()...)
			}

		}
	}
	for _, tx := range shardBody.Transactions {
		res = append(res, tx.Hash().GetBytes()...)
	}
	return common.HashH(res)
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
				return NewBlockChainError(UnmashallJsonBlockError, errors.New("can not parse a wrong tx"))
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

/*
- Concatenate all transaction in one shard as a string
- Then each shard producer a string value include all transactions within this block
- For each string value: Convert string value to hash value
- So if we have 256 shard, we will have 256 leaf value for merkle tree
- Make merkle root from these value
*/

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

// func (shardBody *ShardBody) CalcMerkleRootShard(activeShards int) *common.Hash {
// 	if activeShards == 1 {
// 		merkleRoot := common.HashH([]byte{})
// 		return &merkleRoot
// 	}
// 	// fmt.Println("Shard Body/CalcMerkleRootShard ================== 1")
// 	var shardTxs = make(map[int][]*common.Hash)
// 	// Init shard Txs
// 	for shardID := 0; shardID < activeShards; shardID++ {
// 		shardTxs[shardID] = []*common.Hash{}
// 	}
// 	for _, tx := range shardBody.Transactions {
// 		shardID := int(tx.GetSenderAddrLastByte())
// 		shardTxs[shardID] = append(shardTxs[shardID], tx.Hash())
// 	}
// 	// fmt.Println(shardTxs)
// 	// fmt.Println("Shard Body/CalcMerkleRootShard ================== 2")
// 	shardsHash := make([]*common.Hash, activeShards)
// 	// for idx := range shardsHash {
// 	// 	fmt.Println("idx", idx)
// 	// 	h := &common.Hash{}
// 	// 	shardsHash[idx], _ = h.NewHashFromStr("")
// 	// }
// 	// fmt.Println(shardsHash)
// 	for shardID := 0; shardID < activeShards; shardID++ {
// 		txHashStrConcat := ""
// 		for _, tx := range shardTxs[shardID] {
// 			txHashStrConcat += tx.String()
// 		}
// 		// fmt.Printf("txHashStrConcat for ShardID %+v is %+v \n", shardID, txHashStrConcat)
// 		txHashStrConcatHash := common.HashH([]byte(txHashStrConcat))
// 		// fmt.Printf("txHashStrConcatHash for ShardID %+v is %+v \n", shardID, txHashStrConcatHash)
// 		// h := &common.Hash{}
// 		// hash, _ := h.NewHash(txHashStrConcatHash[:32])
// 		// fmt.Printf("Hash of txHashStrConcat for ShardID %+v is %+v \n", shardID, hash)
// 		shardsHash[shardID] = &txHashStrConcatHash
// 	}
// 	// fmt.Println("Shard Body/CalcMerkleRootShard ================== 3")
// 	// fmt.Println(shardsHash)
// 	// for idx, shard := range shardTxs {
// 	// 	fmt.Println("idx", idx)
// 	// 	txHashStrConcat := ""

// 	// 	for _, tx := range shard {
// 	// 		txHashStrConcat += tx.String()
// 	// 	}

// 	// 	h := &common.Hash{}
// 	// 	hash, _ := h.NewHashFromStr(txHashStrConcat)

// 	// 	shardsHash[idx] = hash
// 	// 	fmt.Println(shardsHash)
// 	// }
// 	// fmt.Println("Shard Body/CalcMerkleRootShard ================== 4")
// 	merkleRoots := Merkle{}.BuildMerkleTreeOfHashs(shardsHash)
// 	// fmt.Println(merkleRoots)
// 	// fmt.Println("Shard Body/CalcMerkleRootShard ================== 5")
// 	merkleRoot := merkleRoots[len(merkleRoots)-1]
// 	// fmt.Println(merkleRoot)
// 	// fmt.Println("Shard Body/CalcMerkleRootShard ================== 6")
// 	return merkleRoot
// }
