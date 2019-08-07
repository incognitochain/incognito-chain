package blsbft

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/chain"
	"github.com/incognitochain/incognito-chain/consensus/multisigschemes/bls"
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

func (e *BLSBFT) validatePreSignBlock(block chain.BlockInterface, validationData string) error {
	confident, err := e.ValidateBlock(block, e.Chain)
	if confident != 2 {
		return err
	}
	return nil
}

func (e *BLSBFT) ValidateBlock(block chain.BlockInterface, chain chain.ChainInterface) (byte, error) {
	valData, error := DecodeValidationData(block.GetValidationField())

	// 1. Verify producer's sig
	// 2. Verify Committee's sig
	// 3. Verify correct producer for blockHeight, round
	blockHash := block.Hash()
	if err := bls.ValidateSingleSig(blockHash, valData.ProducerSig, valData.Producer); err != nil {
		return 0, err
	}
	committee := chain.GetCommittee()
	if err := bls.ValidateAggSig(block.Hash(), valData.AggSig, committee); err != nil {
		return 1, err
	}
	producerPosition := (chain.GetLastProposerIndex() + block.GetRound()) % chain.GetCommitteeSize()
	tempProducer := committee[producerPosition]
	if strings.Compare(tempProducer, valData.Producer) != 0 {
		return 2, NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}

	return 3, nil

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
