package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
)

type BeaconBlock struct {
	AggregatedSig   string  `json:"AggregatedSig"`
	R               string  `json:"R"`
	ValidatorsIndex [][]int `json:"ValidatorsIndex"` //[0]: r | [1]:AggregatedSig
	ProducerSig     string  `json:"ProducerSig"`
	Header          BeaconHeader
	Body            BeaconBody
}

func NewBeaconBlock() *BeaconBlock {
	return &BeaconBlock{}
}

func (beaconBlock *BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBeaconBlock := &struct {
		AggregatedSig   string  `json:"AggregatedSig"`
		ValidatorsIndex [][]int `json:"ValidatorsIndex"`
		ProducerSig     string  `json:"ProducerSig"`
		R               string  `json:"R"`
		Header          BeaconHeader
		Body            BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBeaconBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	beaconBlock.AggregatedSig = tempBeaconBlock.AggregatedSig
	beaconBlock.R = tempBeaconBlock.R
	beaconBlock.ValidatorsIndex = tempBeaconBlock.ValidatorsIndex
	beaconBlock.ProducerSig = tempBeaconBlock.ProducerSig
	beaconBlock.Header = tempBeaconBlock.Header
	beaconBlock.Body = tempBeaconBlock.Body
	return nil
}
