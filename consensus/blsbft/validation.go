package blsbft

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/multisigschemes/bls"
)

type blockValidation interface {
	common.BlockInterface
	AddValidationField(validationData string) error
}

type ValidationData struct {
	Producer       string
	ProducerBLSSig string
	ProducerBriSig string
	ValidatiorsIdx []int
	AggSig         string
	BridgeSig      []string
	// AgreeSigs         map[string][]string
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
	if err := e.ValidateProducerSig(block); err != nil {
		return err
	}
	if err := e.ValidateProducerPosition(block); err != nil {
		return err
	}
	if err := e.Chain.ValidatePreSignBlock(block); err != nil {
		return err
	}
	return nil
}

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

func (e *BLSBFT) ValidateProducerSig(block common.BlockInterface) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return err
	}
	if err := bls.ValidateSingleSig(block.Hash(), valData.ProducerSig, valData.Producer); err != nil {
		return err
	}
	return nil
}

func (e *BLSBFT) ValidateCommitteeSig(block common.BlockInterface, committee []string) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return err
	}
	if err := bls.ValidateAggSig(block.Hash(), valData.AggSig, committee); err != nil {
		return err
	}
	return nil
}

func (e *BLSBFT) CreateValidationData(blockHash *common.Hash) ValidationData {
	var valData ValidationData
	return valData
}

func (e *BLSBFT) FinalizedValidationData(block common.BlockInterface, sigs []string) error {
	return nil
}

func (e *BLSBFT) ValidateData(data []byte, sig string, publicKey string) error {
	return nil
}
