package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
)

type BeaconBlock struct {
	// AggregatedSig string  `json:"AggregatedSig"`
	// R             string  `json:"R"`
	// ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	// ProducerSig   string  `json:"ProducerSig"`

	ValidationData string `json:"ValidationData"`

	Body   BeaconBody
	Header BeaconHeader
}

func NewBeaconBlock() *BeaconBlock {
	return &BeaconBlock{}
}

func (beaconBlock *BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

func (beaconBlock *BeaconBlock) GetHeight() uint64 {
	return beaconBlock.Header.Height
}

// func (beaconBlock *BeaconBlock) GetProducerPubKey() string {
// 	return string(beaconBlock.Header.ProducerAddress.Pk)
// }

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBeaconBlock := &struct {
		ValidationData string `json:"ValidationData"`

		Header BeaconHeader
		Body   BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBeaconBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	// beaconBlock.AggregatedSig = tempBlk.AggregatedSig
	// beaconBlock.R = tempBlk.R
	// beaconBlock.ValidatorsIdx = tempBlk.ValidatorsIdx
	// beaconBlock.ProducerSig = tempBlk.ProducerSig
	beaconBlock.ValidationData = tempBeaconBlock.ValidationData
	beaconBlock.Header = tempBeaconBlock.Header
	beaconBlock.Body = tempBeaconBlock.Body
	return nil
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

func (beaconBlock BeaconBlock) GetInstructions() [][]string {
	return beaconBlock.Body.Instructions
}
