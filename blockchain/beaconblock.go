package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
)

const (
	VERSION       = 1
	RANDOM_NUMBER = 3
)

type BeaconBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"R"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`

	ValidationData string `json:"ValidationData"`

	Body   BeaconBody
	Header BeaconHeader
}

func NewBeaconBlock() BeaconBlock {
	return BeaconBlock{}
}

func (beaconBlock *BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

func (beaconBlock *BeaconBlock) GetHeight() uint64 {
	return beaconBlock.Header.Height
}

func (beaconBlock *BeaconBlock) GetProducerPubKey() string {
	return string(beaconBlock.Header.ProducerAddress.Pk)
}

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string  `json:"AggregatedSig"`
		ValidatorsIdx [][]int `json:"ValidatorsIdx"`
		ProducerSig   string  `json:"ProducerSig"`
		R             string  `json:"R"`

		ValidationData string `json:"ValidationData"`

		Header BeaconHeader
		Body   BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	beaconBlock.AggregatedSig = tempBlk.AggregatedSig
	beaconBlock.R = tempBlk.R
	beaconBlock.ValidatorsIdx = tempBlk.ValidatorsIdx
	beaconBlock.ProducerSig = tempBlk.ProducerSig
	// blkBody := BeaconBody{}
	// err = blkBody.UnmarshalJSON(tempBlk.Body)
	// if err != nil {
	// 	return NewBlockChainError(UnmashallJsonBlockError, err)
	// }
	// blkHeader := BeaconHeader{}
	// err = blkBody.UnmarshalJSON(tempBlk.Header)
	// if err != nil {
	// 	return NewBlockChainError(UnmashallJsonBlockError, err)
	// }
	beaconBlock.Header = tempBlk.Header
	beaconBlock.Body = tempBlk.Body
	return nil
}

func (beaconBlock *BeaconBody) toString() string {
	res := ""
	for _, l := range beaconBlock.ShardState {
		for _, r := range l {
			res += strconv.Itoa(int(r.Height))
			res += r.Hash.String()
			crossShard, _ := json.Marshal(r.CrossShard)
			res += string(crossShard)

		}
	}

	for _, l := range beaconBlock.Instructions {
		for _, r := range l {
			res += r
		}
	}
	return res
}

func (beaconBody *BeaconBody) Hash() common.Hash {
	return common.HashH([]byte(beaconBody.toString()))
}

func (beaconHeader *BeaconHeader) toString() string {
	res := ""
	res += beaconHeader.ProducerAddress.String()
	res += fmt.Sprintf("%v", beaconHeader.Version)
	res += fmt.Sprintf("%v", beaconHeader.Height)
	res += fmt.Sprintf("%v", beaconHeader.Epoch)
	res += fmt.Sprintf("%v", beaconHeader.Round)
	res += fmt.Sprintf("%v", beaconHeader.Timestamp)
	res += beaconHeader.PrevBlockHash.String()
	res += beaconHeader.ValidatorsRoot.String()
	res += beaconHeader.BeaconCandidateRoot.String()
	res += beaconHeader.ShardCandidateRoot.String()
	res += beaconHeader.ShardValidatorsRoot.String()
	res += beaconHeader.ShardStateHash.String()
	res += beaconHeader.InstructionHash.String()
	return res
}

func (beaconBlock *BeaconHeader) MetaHash() common.Hash {
	return common.Keccak256([]byte(beaconBlock.toString()))
}

func (beaconBlock *BeaconHeader) Hash() common.Hash {
	// Block header of beacon uses Keccak256 as a hash func to check on Ethereum when relaying blocks
	blkMetaHash := beaconBlock.MetaHash()
	blkInstHash := beaconBlock.InstructionMerkleRoot
	combined := append(blkMetaHash[:], blkInstHash[:]...)
	return common.Keccak256(combined)
}

func (beaconBlock *BeaconBlock) AddValidationField(validateData string) error {
	beaconBlock.ValidationData = validateData
	return nil
}
func (beaconBlock *BeaconBlock) GetValidationField() string {
	return beaconBlock.ValidationData
}

func (beaconBlock BeaconBlock) GetRound() int {
	return beaconBlock.Header.Round
}
func (beaconBlock BeaconBlock) GetRoundKey() string {
	return fmt.Sprint(beaconBlock.Header.Height, "_", beaconBlock.Header.Round)
}
