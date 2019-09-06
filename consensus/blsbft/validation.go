package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
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
		return nil, consensus.NewConsensusError(consensus.DecodeValidationDataError, err)
	}
	return &valData, nil
}

func EncodeValidationData(validationData ValidationData) (string, error) {
	result, err := json.Marshal(validationData)
	if err != nil {
		return "", consensus.NewConsensusError(consensus.EncodeValidationDataError, err)
	}
	return string(result), nil
}

func (e BLSBFT) CreateValidationData(block common.BlockInterface) ValidationData {
	var valData ValidationData
	selfPublicKey := e.UserKeySet.GetPublicKey()
	keyByte, _ := selfPublicKey.GetMiningKey(CONSENSUSNAME)
	valData.ProducerBLSSig, _ = e.UserKeySet.BLSSignData(block.Hash().GetBytes(), 0, []blsmultisig.PublicKey{keyByte})
	return valData
}

func (e BLSBFT) validatePreSignBlock(block common.BlockInterface) error {
	e.logger.Info("verifying block...")
	e.logger.Info("ValidateProducerPosition...")
	if err := e.ValidateProducerPosition(block, e.RoundData.LastProposerIndex, e.RoundData.CommitteeBLS.StringList); err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	e.logger.Info("ValidateProducerSig...")
	if err := e.ValidateProducerSig(block); err != nil {
		return consensus.NewConsensusError(consensus.ProducerSignatureError, err)
	}
	e.logger.Info("ValidatePreSignBlock...")
	if err := e.Chain.ValidatePreSignBlock(block); err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	e.logger.Info("done verify block...")
	return nil
}

func (e BLSBFT) ValidateProducerPosition(block common.BlockInterface, lastProposerIndex int, committeeBLSString []string) error {
	producerPosition := (lastProposerIndex + block.GetRound()) % len(committeeBLSString)
	tempProducer := committeeBLSString[producerPosition]
	if strings.Compare(tempProducer, block.GetProducer()) != 0 {
		return consensus.NewConsensusError(consensus.UnExpectedError, errors.New("Producer should be should be :"+tempProducer))
	}

	return nil
}

func (e BLSBFT) ValidateProducerSig(block common.BlockInterface) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}

	producerBase58 := block.GetProducer()
	producerBytes, _, err := base58.Base58Check{}.Decode(producerBase58)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	if err := validateSingleBLSSig(block.Hash(), valData.ProducerBLSSig, 0, []blsmultisig.PublicKey{producerBytes}); err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	return nil
}

func (e BLSBFT) ValidateCommitteeSig(block common.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	valData, err := DecodeValidationData(block.GetValidationField())
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range committee {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[CONSENSUSNAME])
	}
	if err := validateBLSSig(block.Hash(), valData.AggSig, valData.ValidatiorsIdx, committeeBLSKeys); err != nil {
		fmt.Println(block.Hash(), block.GetValidationField())
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	return nil
}

func (e BLSBFT) ValidateData(data []byte, sig string, publicKey string) error {
	sigByte, _, err := base58.Base58Check{}.Decode(sig)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	publicKeyByte, _, err := base58.Base58Check{}.Decode(publicKey)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	valid, err := blsmultisig.Verify(sigByte, data, []int{0}, []blsmultisig.PublicKey{publicKeyByte})
	if valid {
		return nil
	}
	return consensus.NewConsensusError(consensus.UnExpectedError, err)
}

func validateSingleBLSSig(
	dataHash *common.Hash,
	blsSig []byte,
	selfIdx int,
	committee []blsmultisig.PublicKey,
) error {
	result, err := blsmultisig.Verify(blsSig, dataHash.GetBytes(), []int{selfIdx}, committee)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	if !result {
		return consensus.NewConsensusError(consensus.UnExpectedError, errors.New("invalid BLS Signature"))
	}
	return nil
}

func validateSingleBriSig(
	dataHash *common.Hash,
	briSig []byte,
	candidate []byte,
) error {
	result, err := bridgesig.Verify(candidate, dataHash.GetBytes(), briSig)
	if err != nil {
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	if !result {
		return consensus.NewConsensusError(consensus.UnExpectedError, errors.New("invali BRI Signature"))
	}
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
		return consensus.NewConsensusError(consensus.UnExpectedError, err)
	}
	if !result {
		return consensus.NewConsensusError(consensus.UnExpectedError, errors.New("Invalid Signature!"))
	}
	return nil
}
