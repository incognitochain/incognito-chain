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
	ProducerSig   string `json:"BlockProducerSignature"`
	Body          ShardBody
	Header        ShardHeader
}

type ShardToBeaconBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"BlockProducerSignature"`

	Header ShardHeader
	Params interface{}
}

type CrossShardBlock struct {
	AggregatedSig string `json:"AggregatedSig"`
	ValidatorsIdx []int  `json:"ValidatorsIdx"`
	ProducerSig   string `json:"BlockProducerSignature"`

	Header          ShardHeader
	MerklePathShard []common.Hash
	UTXOList        []privacy.Coin
}

//HashFinal creates a hash from block data that include AggregatedSig & ValidatorsIdx
func (self *ShardBlock) Hash() *common.Hash {
	record := common.EmptyString
	record += self.Header.Hash().String() + self.AggregatedSig + common.IntArrayToString(self.ValidatorsIdx, ",")
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (self *ShardBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string
		ValidatorsIdx []int
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
