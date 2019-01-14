package blockchain

import (
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/privacy"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
)

type ShardBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"ProducerSig"`
	Body          ShardBody
	Header        ShardHeader
}

type ShardToBeaconBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"ProducerSig"`

	Header ShardHeader
}

type CrossShardBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"ProducerSig"`

	Header          ShardHeader
	MerklePathShard []common.Hash
	CrossOutputCoin []privacy.OutputCoin
}

func (self *CrossShardBlock) Hash() *common.Hash {
	hash := self.Header.Hash()
	return &hash
}

func (self *ShardBlock) Hash() *common.Hash {
	hash := self.Header.Hash()
	return &hash
}

func (self *ShardBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string
		ValidatorsIdx []int
		ProducerSig   string
		Header        ShardHeader
		Body          *json.RawMessage
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self.AggregatedSig = tempBlk.AggregatedSig
	self.ValidatorsIdx = tempBlk.ValidatorsIdx

	blkBody := ShardBody{}
	err = blkBody.UnmarshalJSON(*tempBlk.Body)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self.Header = tempBlk.Header

	self.Body = blkBody
	return nil
}

// /*
// AddTransaction adds a new transaction into block
// */
// // #1 - tx
func (self *ShardBlock) AddTransaction(tx metadata.Transaction) error {
	if self.Body.Transactions == nil {
		return NewBlockChainError(UnExpectedError, errors.New("Not init tx arrays"))
	}
	self.Body.Transactions = append(self.Body.Transactions, tx)
	return nil
}
