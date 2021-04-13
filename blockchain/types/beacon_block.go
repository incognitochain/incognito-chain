package types

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
)

type BeaconBlock struct {
	ValidationData string `json:"ValidationData"`
	Body           BeaconBody
	Header         BeaconHeader
}

func (beaconBlock *BeaconBlock) GetVersion() int {
	return beaconBlock.Header.Version
}

func (beaconBlock *BeaconBlock) GetPrevHash() common.Hash {
	return beaconBlock.Header.PreviousBlockHash
}

func NewBeaconBlock() *BeaconBlock {
	return &BeaconBlock{}
}

func (beaconBlock *BeaconBlock) GetProposer() string {
	return beaconBlock.Header.Proposer
}

func (beaconBlock *BeaconBlock) GetProposeTime() int64 {
	return beaconBlock.Header.ProposeTime
}

func (beaconBlock *BeaconBlock) GetProduceTime() int64 {
	return beaconBlock.Header.Timestamp
}

func (beaconBlock BeaconBlock) Hash() *common.Hash {
	hash := beaconBlock.Header.Hash()
	return &hash
}

func (beaconBlock BeaconBlock) GetCurrentEpoch() uint64 {
	return beaconBlock.Header.Epoch
}

func (beaconBlock BeaconBlock) GetHeight() uint64 {
	return beaconBlock.Header.Height
}

func (beaconBlock BeaconBlock) GetShardID() int {
	return -1
}

func (beaconBlock BeaconBlock) CommitteeFromBlock() common.Hash {
	return common.Hash{}
}

func (beaconBlock *BeaconBlock) UnmarshalJSON(data []byte) error {
	tempBeaconBlock := &struct {
		ValidationData string `json:"ValidationData"`
		Header         BeaconHeader
		Body           BeaconBody
	}{}
	err := json.Unmarshal(data, &tempBeaconBlock)
	if err != nil {
		return err
	}
	beaconBlock.ValidationData = tempBeaconBlock.ValidationData
	beaconBlock.Header = tempBeaconBlock.Header
	beaconBlock.Body = tempBeaconBlock.Body
	return nil
}

func (beaconBlock *BeaconBlock) AddValidationField(validationData string) {
	beaconBlock.ValidationData = validationData
	return
}
func (beaconBlock BeaconBlock) GetValidationField() string {
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

func (beaconBlock BeaconBlock) GetProducer() string {
	return beaconBlock.Header.Producer
}

func (beaconBlock BeaconBlock) GetProducerPubKeyStr() string {
	return beaconBlock.Header.ProducerPubKeyStr
}

func (beaconBlock BeaconBlock) GetConsensusType() string {
	return beaconBlock.Header.ConsensusType
}
