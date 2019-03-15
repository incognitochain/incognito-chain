package blockchain

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

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
type CrossTokenPrivacyData struct {
	BlockHeight      uint64
	BlockHash        common.Hash
	TokenPrivacyData []ContentCrossTokenPrivacyData
}
type CrossTransaction struct {
	BlockHeight      uint64
	BlockHash        common.Hash
	TokenPrivacyData []ContentCrossTokenPrivacyData
	OutputCoin       []privacy.OutputCoin
}
type ContentCrossTokenPrivacyData struct {
	OutputCoin     []privacy.OutputCoin
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}
type CrossShardTokenPrivacyMetaData struct {
	TokenID        common.Hash
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}

func (self *ContentCrossTokenPrivacyData) Bytes() []byte {
	res := []byte{}
	for _, item := range self.OutputCoin {
		res = append(res, item.Bytes()...)
	}
	res = append(res, self.PropertyID.GetBytes()...)
	res = append(res, []byte(self.PropertyName)...)
	res = append(res, []byte(self.PropertySymbol)...)
	typeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(typeBytes, uint32(self.Type))
	res = append(res, typeBytes...)
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(amountBytes, uint32(self.Amount))
	res = append(res, amountBytes...)
	if self.Mintable {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	return res
}
func (self *ContentCrossTokenPrivacyData) Hash() common.Hash {
	return common.HashH(self.Bytes())
}
func (shardBody *ShardBody) Hash() common.Hash {
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
func (crossOutputCoin *CrossOutputCoin) Hash() common.Hash {
	res := []byte{}
	res = append(res, crossOutputCoin.BlockHash.GetBytes()...)
	for _, coins := range crossOutputCoin.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	return common.HashH(res)
}
func (crossTransaction *CrossTransaction) Bytes() []byte {
	res := []byte{}
	res = append(res, crossTransaction.BlockHash.GetBytes()...)
	for _, coins := range crossTransaction.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	for _, coins := range crossTransaction.TokenPrivacyData {
		res = append(res, coins.Bytes()...)
	}
	return res
}
func (crossTransaction *CrossTransaction) Hash() common.Hash {
	return common.HashH(crossTransaction.Bytes())
}

/*
	Verify CrossShard Block
	- Agg Signature
	- MerklePath
*/
func (block *CrossShardBlock) VerifyCrossShardBlock(committees []string) error {
	if err := ValidateAggSignature(block.ValidatorsIdx, committees, block.AggregatedSig, block.R, block.Hash()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	if ok := VerifyCrossShardBlockUTXO(block, block.MerklePathShard); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	return nil
}
