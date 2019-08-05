package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
)

type BeaconBlock struct {
	AggregatedSig string  `json:"AggregatedSig"`
	R             string  `json:"R"`
	ValidatorsIdx [][]int `json:"ValidatorsIdx"` //[0]: r | [1]:AggregatedSig
	ProducerSig   string  `json:"ProducerSig"`
	Header        BeaconHeader
	Body          BeaconBody
}

func NewBeaconBlock() *BeaconBlock {
	return &BeaconBlock{}
}

func (beaconBlock *BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBlk := &struct {
		AggregatedSig string  `json:"AggregatedSig"`
		ValidatorsIdx [][]int `json:"ValidatorsIdx"`
		ProducerSig   string  `json:"ProducerSig"`
		R             string  `json:"R"`
		Header        BeaconHeader
		Body          BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBlk)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	beaconBlock.AggregatedSig = tempBlk.AggregatedSig
	beaconBlock.R = tempBlk.R
	beaconBlock.ValidatorsIdx = tempBlk.ValidatorsIdx
	beaconBlock.ProducerSig = tempBlk.ProducerSig
	beaconBlock.Header = tempBlk.Header
	beaconBlock.Body = tempBlk.Body
	return nil
}
