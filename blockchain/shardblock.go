package blockchain

import (
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/privacy"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
)

type ShardBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"r"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`
	Body          ShardBody
	Header        ShardHeader
}

type ShardToBeaconBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"r"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`

	Instructions [][]string
	Header       ShardHeader
}

type CrossShardBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"r"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`

	Header          ShardHeader
	MerklePathShard []common.Hash
	CrossOutputCoin []privacy.OutputCoin
}

func (shardBlock *CrossShardBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
	return &hash
}

func (shardBlock *ShardToBeaconBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
	return &hash
}

func (shardBlock *ShardBlock) Hash() *common.Hash {
	hash := shardBlock.Header.Hash()
	return &hash
}

func (shardBlock *ShardBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string
		R             string `json:"r"`
		ValidatorsIdx [][]int
		ProducerSig   string
		Header        ShardHeader
		Body          *json.RawMessage
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	shardBlock.AggregatedSig = tempBlk.AggregatedSig
	shardBlock.R = tempBlk.R
	shardBlock.ValidatorsIdx = tempBlk.ValidatorsIdx
	shardBlock.ProducerSig = tempBlk.ProducerSig

	blkBody := ShardBody{}
	err = blkBody.UnmarshalJSON(*tempBlk.Body)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	shardBlock.Header = tempBlk.Header

	shardBlock.Body = blkBody
	return nil
}

// /*
// AddTransaction adds a new transaction into block
// */
// // #1 - tx
func (shardBlock *ShardBlock) AddTransaction(tx metadata.Transaction) error {
	if shardBlock.Body.Transactions == nil {
		return NewBlockChainError(UnExpectedError, errors.New("not init tx arrays"))
	}
	shardBlock.Body.Transactions = append(shardBlock.Body.Transactions, tx)
	return nil
}
