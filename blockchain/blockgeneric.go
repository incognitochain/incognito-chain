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
	Height uint64 `json:"Height"`

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint64 on the wire and therefore is limited to 2106.
	Timestamp int64  `json:"Timestamp"`
	Epoch     uint64 `json:"Epoch"`
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

/*@Hung
type BeaconHeader struct {
	Version int 				`json:"Version"`
	ParentHash common.Hash 		`json:"ParentBlockHash"`
	Height uint64 				`json:"Height"`
	//epoch length should be config in consensus
	Epoch uint64				`json:"Epoch"`
	Timestamp int64 			`json:"Timestamp"`

	AggregatedSig string 		`json:"AggregatedSig"`
	ValidatorsIdx []int 		`json:"ValidatorsIdx"`
	ProducerSig   string 		`json:"BlockProducerSignature"`
	Type          string		`json:"TypeOf..."`

	Random		int64			`json:"RandomNumber"`
	//Validator list will be store in database/memory (locally)
	ValidatorsRoot common.Hash 	`json:"CurrentValidatorRootHash"`
	//Candidate = unassigned_validator list will be store in database/memory (locally)
	CandidateRoot common.Hash 	`json:"CandidateListRootHash"`
	// Store these two list make sure all node process the same data

	// each shard will have a list of blockHash
	// shardRoot is hash of all list
	shardRoot	common.Hash 	`json:"ShardListRootHash"`
	// hash of all parameters
	paramHash	common.Hash 	`json:"ParameterHash"`
}

type BeaconBlock struct {
	Header 			*BeaconHeader

	ShardBlocks  	[]*common.Hash
	Actions 		[]interface{}

	// size of block should be store
	size 		...(unknown type)

	// These fields are used to track if needed
	// inter-peer block relay.
	ReceivedAt   	time.Time
	ReceivedFrom 	interface{}
}

type ActionParams interface {
	//TODO
}
func (h *Header) Hash() common.Hash {
	//TODO
}

// define size in common
func (h *Header) Size() common.StorageSize {
	//TODO
}

func NewBlock(...) (*BeaconBlock, error) {
	//TODO
}

func NewBlockWithHeader(header *BeaconHeader) *BeaconBlock {
	//TODO
}

type ShardHeader struct {
	Version int 				`json:"Version"`
	ParentHash common.Hash 		`json:"ParentBlockHash"`

	Height uint64 				`json:"Height"`
	Epoch uint64				`json:"Epoch"`
	Timestamp int64 			`json:"Timestamp"`

	MerkleRoot      common.Hash	`json:"MerkleRoot"`
	MerkleRootShard common.Hash	`json:"MerkleRootShard"`

	ShardID         byte		`json:"ShardID"`

	AggregatedSig string 		`json:"AggregatedSig"`
	ValidatorsIdx []int 		`json:"ValidatorsIdx"`
	ProducerSig   string 		`json:"BlockProducerSignature"`
	Type          string		`json:"TypeOf..."`

	//Validator list will be store in database/memory (locally)
	ValidatorsRoot common.Hash 	`json:"CurrentValidatorRootHash"`
	//Candidate = pending validator list will be store in database/memory (locally)
	CandidateRoot common.Hash 	`json:"CandidateListRootHash"`
	// Store these two list make sure all node process the same data
}

type ShardBlock struct {
	Header 			*ShardHeader

	transactions	map[byte][]*Transaction

	Actions 		[]interface{}

	// cache
	// size of block should be store
	size 		...(unknown type)
	merklePath		map[byte][]byte

	// These fields are used to track if needed
	// inter-peer block relay.
	ReceivedAt   	time.Time
	ReceivedFrom 	interface{}
}
// add function process txstake transaction to output a list of candidate for beacon block

*/

/*@HUNG
type Blockchain Struct {

}
*/
