package blockchain

import (
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/common"
)

type BlockHeaderV2 interface {
	Hash() common.Hash
	UnmarshalJSON([]byte) error
}

type BlockBodyV2 interface {
	Hash() common.Hash
	UnmarshalJSON([]byte) error
}

type BlockHeaderGeneric struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int `json:"Version"`

	// Hash of the previous block header in the block chain.
	PrevBlockHash common.Hash `json:"PrevBlockHash"`

	//Block Height
	Height int32 `json:"Height"`

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint64 on the wire and therefore is limited to 2106.
	Timestamp int64 `json:"Timestamp"`
}

type BlockV2 struct {
	AggregatedSig string // aggregated signature in base58
	ValidatorsIdx []int
	ProducerSig   string // block producer signature in base58
	Type          string

	Header BlockHeaderV2
	Body   BlockBodyV2
}

func (self *BlockV2) Hash() common.Hash {
	record := common.EmptyString
	record += self.Header.Hash().String() + string(self.AggregatedSig) + common.IntArrayToString(self.ValidatorsIdx, ",") + self.ProducerSig + self.Type

	return common.DoubleHashH([]byte(record))
}

func (self *BlockV2) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string
		ValidatorsIdx []int
		ProducerSig   string
		Type          string
		Header        *json.RawMessage
		Body          *json.RawMessage
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	self.Type = tempBlk.Type
	self.AggregatedSig = tempBlk.AggregatedSig
	self.ValidatorsIdx = tempBlk.ValidatorsIdx
	self.ProducerSig = tempBlk.ProducerSig

	switch self.Type {
	case "beacon":
		self.Header = &BeaconBlockHeader{}
		err := json.Unmarshal(*tempBlk.Header, self.Header)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}

		self.Body = &BeaconBlockBody{}
		err = json.Unmarshal(*tempBlk.Body, self.Body)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}

	case "shard":
		blkHeader := BlockHeaderShard{}
		err := blkHeader.UnmarshalJSON(*tempBlk.Header)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		blkBody := BlockBodyShard{}
		err = blkBody.UnmarshalJSON(*tempBlk.Body)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		self.Header = &BlockHeaderShard{
			BlockHeaderGeneric: blkHeader.BlockHeaderGeneric,
		}
		self.Body = &blkBody
	default:
		return NewBlockChainError(UnmashallJsonBlockError, errors.New("Unknown block type "+self.Type))
	}
	return nil
}
