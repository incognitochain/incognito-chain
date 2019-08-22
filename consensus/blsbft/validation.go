package blsbft

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/blsmultisig"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type vote struct {
	BLS []byte
	BRI []byte
}

type blockValidation interface {
	common.BlockInterface
	AddValidationField(validationData string) error
}

type ValidationData struct {
	ProducerBLSSig []byte
	ProducerBriSig []byte
	ValidatiorsIdx []int
	AggSig         []byte
	BridgeSig      [][]byte
}

func DecodeValidationData(data string) (*ValidationData, error) {
	var valData ValidationData
	err := json.Unmarshal([]byte(data), &valData)
	if err != nil {
		return nil, err
	}
	return &valData, nil
}

func EncodeValidationData(validationData ValidationData) (string, error) {
	result, err := json.Marshal(validationData)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (e *BLSBFT) validatePreSignBlock(block common.BlockInterface, committee []incognitokey.CommitteePubKey) error {
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

func validateSingleBLSSig(
	dataHash *common.Hash,
	blsSig []byte,
	selfIdx int,
	committee []blsmultisig.PublicKey,
) error {
	result, err := blsmultisig.Verify(blsSig, dataHash.GetBytes(), []int{selfIdx}, committee)
	if err != nil {
		return err
	}
	if !result {
		return errors.New("invalid Signature")
	}
	return nil
}

func validateSingleBriSig(
	dataHash *common.Hash,
	aggSig []byte,
) error {
	return nil
}

func validateBLSSig(
	dataHash *common.Hash,
	aggSig []byte,
	validatorIdx []int,
	committee []blsmultisig.PublicKey,
) error {
	result, err := blsmultisig.Verify(aggSig, dataHash.GetBytes(), validatorIdx, committee)
	if err != nil {
		return err
	}
	if !result {
		return errors.New("Invalid Signature!")
	}
	return nil
}

func (e *BLSBFT) ValidateProducerPosition(block common.BlockInterface) error {
	committee := e.Chain.GetCommittee()
	producerPosition := (e.Chain.GetLastProposerIndex() + block.GetRound()) % e.Chain.GetCommitteeSize()
	tempProducer := committee[producerPosition].GetMiningKeyBase58(CONSENSUSNAME)
	if strings.Compare(tempProducer, block.GetProducer()) != 0 {
		return errors.New("Producer should be should be :" + tempProducer)
	}

	return nil
}

func (e *BLSBFT) ValidateProducerSig(block common.BlockInterface) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return err
	}

	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range e.Chain.GetCommittee() {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[CONSENSUSNAME])
	}

	if err := validateSingleBLSSig(block.Hash(), valData.ProducerBLSSig, e.Chain.GetPubKeyCommitteeIndex(block.GetProducer()), committeeBLSKeys); err != nil {
		return err
	}
	return nil
}

func (e *BLSBFT) ValidateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePubKey) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return err
	}
	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range committee {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[CONSENSUSNAME])
	}
	if err := validateBLSSig(block.Hash(), valData.AggSig, valData.ValidatiorsIdx, committeeBLSKeys); err != nil {
		return err
	}
	return nil
}

func (e *BLSBFT) CreateValidationData(blockHash *common.Hash) ValidationData {
	var valData ValidationData
	selfPublicKey := e.UserKeySet.GetPublicKey()
	selfIdx := e.Chain.GetPubKeyCommitteeIndex(selfPublicKey.GetMiningKeyBase58(CONSENSUSNAME))
	committeeKeys := []blsmultisig.PublicKey{}
	for _, key := range e.Chain.GetCommittee() {
		keyByte, _ := key.GetMiningKey(CONSENSUSNAME)
		committeeKeys = append(committeeKeys, keyByte)
	}

	e.UserKeySet.BLSSignData(blockHash.GetBytes(), selfIdx, committeeKeys)

	return valData
}

func (e *BLSBFT) ValidateData(data []byte, sig string, publicKey string) error {
	return nil
}
