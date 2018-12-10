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

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot common.Hash `json:"MerkleRoot"`

	//Block Height
	Height int32 `json:"Height"`

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint64 on the wire and therefore is limited to 2106.
	Timestamp int64 `json:"Timestamp"`
}

type BlockV2 struct {
	AggregatedSig []byte
	ValidatorsIdx []int
	ProducerSig   []byte
	Type          string

	Header BlockHeaderV2
	Body   BlockBodyV2
}

func (self *BlockV2) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig []byte
		ValidatorsIdx []int
		ProducerSig   []byte
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
		type AliasHeader BlockHeaderBeacon
		blkHeader := &AliasHeader{}
		err := json.Unmarshal(*tempBlk.Header, &blkHeader)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		var blkBody BlockBodyBeacon
		err = json.Unmarshal(*tempBlk.Body, &blkBody)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		self.Header = BlockHeaderBeacon{
			BlockHeaderGeneric: blkHeader.BlockHeaderGeneric,
			TestParam:          blkHeader.TestParam,
		}
		self.Body = blkBody
	case "shard":
		type AliasHeader BlockHeaderShard
		blkHeader := &AliasHeader{}
		err := json.Unmarshal(*tempBlk.Header, &blkHeader)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		var blkBody BlockBodyShard
		err = blkBody.UnmarshalJSON(*tempBlk.Body)
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		self.Header = BlockHeaderShard{
			BlockHeaderGeneric: blkHeader.BlockHeaderGeneric,
		}
		self.Body = blkBody
	default:
		return NewBlockChainError(UnmashallJsonBlockError, errors.New("Unknown block type "+self.Type))
	}
	return nil
}
