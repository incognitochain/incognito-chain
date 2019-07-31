package blsbft

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/chain"
)

type ValidationData struct {
	Producer       string
	ProducerSig    string
	ValidatiorsIdx []int
	AggSig         string
}

func DecodeValidationData(data string) (*ValidationData, error) {
	var valData ValidationData
	err := json.Unmarshal([]byte(data), &valData)
	if err != nil {
		return nil, err
	}
	return &valData, nil
}

func EncodeValidationData(validationData ValidationData) ([]byte, error) {
	return json.Marshal(validationData)
}

func (e *BLSBFT) validatePreSignBlock(blockHash common.Hash, validationData string) error {
	valData, err := DecodeValidationData(validationData)
	return nil
}

func (e *BLSBFT) ValidateBlock(block chain.BlockInterface) error {
	valData, error := DecodeValidationData(block.GetValidationField())
	return nil

}

func (e *BLSBFT) CreateValidationData(blockHash common.Hash, privateKey string, round int) ValidationData {
	var valData ValidationData
	return valData
}

func (e *BLSBFT) FinalizedValidationData(block chain.BlockInterface, sigs []string) error {
	return nil
}

func (e *BLSBFT) ValidateProducerSig(blockHash common.Hash, validationData string) error {

	return nil
}

func (e BLSBFT) NewInstance() chain.ConsensusInterface {
	var newInstance BLSBFT
	return &newInstance
}
