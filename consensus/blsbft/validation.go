package blsbft

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/multisigschemes/bls"
)

type ValidationData struct {
	Producer       string
	ProducerSig    string
	ValidatiorsIdx []int
	AggSig         string
	BridgeSig      []string
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

func (e *BLSBFT) validatePreSignBlock(block common.BlockInterface, committee []string) error {
	if err := e.ValidateProducerSig(block.Hash(), block.GetValidationField()); err != nil {
		return err
	}
	// if err := e.ValidateProducerPosition(block); err != nil {
	// 	return err
	// }
	return nil
}

// func (e *BLSBFT) ValidateBlock(block common.BlockInterface) error {

// 	// 1. Verify producer's sig
// 	// 2. Verify Committee's sig
// 	// 3. Verify correct producer for blockHeight, round
// 	if err := e.ValidateProducerSig(block); err != nil {
// 		return err
// 	}
// 	if err := e.ValidateCommitteeSig(block); err != nil {
// 		return err
// 	}
// 	if err := e.ValidateProducerPosition(block); err != nil {
// 		return err
// 	}
// 	return nil
// }

func (e *BLSBFT) ValidateProducerPosition(block common.BlockInterface) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return err
	}
	committee := e.Chain.GetCommittee()
	producerPosition := (e.Chain.GetLastProposerIndex() + block.GetRound()) % e.Chain.GetCommitteeSize()
	tempProducer := committee[producerPosition]
	if strings.Compare(tempProducer, valData.Producer) != 0 {
		return errors.New("Producer should be should be :" + tempProducer)
	}

	return nil
}

func (e *BLSBFT) ValidateProducerSig(blockHash *common.Hash, validationData string) error {
	valData, err := DecodeValidationData(validationData)
	if err != nil {
		return err
	}
	if err := bls.ValidateSingleSig(blockHash, valData.ProducerSig, valData.Producer); err != nil {
		return err
	}
	return nil
}

func (e *BLSBFT) ValidateCommitteeSig(blockHash *common.Hash, committee []string, validationData string) error {
	valData, err := DecodeValidationData(validationData)
	if err != nil {
		return err
	}
	if err := bls.ValidateAggSig(blockHash, valData.AggSig, committee); err != nil {
		return err
	}
	return nil
}

func (e *BLSBFT) CreateValidationData(blockHash common.Hash, privateKey string, round int) ValidationData {
	var valData ValidationData
	return valData
}

func (e *BLSBFT) FinalizedValidationData(block common.BlockInterface, sigs []string) error {
	return nil
}
