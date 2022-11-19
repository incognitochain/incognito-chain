package blsbft

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type vote struct {
	BLS          []byte
	BRI          []byte
	Confirmation []byte
}

type BlockValidation interface {
	types.BlockInterface
	AddValidationField(validationData string)
}

func ValidateProducerSigV1(block types.BlockInterface) error {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}

	// producerBase58 := block.GetProducer()
	// producerBytes, _, err := base58.Base58Check{}.Decode(producerBase58)

	producerKey := incognitokey.CommitteePublicKey{}
	err = producerKey.FromBase58(block.GetProducer())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	//start := time.Now()
	if err := validateSingleBriSig(block.ProposeHash(), valData.ProducerBLSSig, producerKey.MiningPubKey[common.BridgeConsensus]); err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	//end := time.Now().Sub(start)
	//fmt.Printf("ConsLog just verify %v\n", end.Seconds())
	return nil
}

func ValidateProducerSigV2(block types.BlockInterface) error {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}

	producerKey := incognitokey.CommitteePublicKey{}
	err = producerKey.FromBase58(block.GetProposer())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	//start := time.Now()
	if err := validateSingleBriSig(block.ProposeHash(), valData.ProducerBLSSig, producerKey.MiningPubKey[common.BridgeConsensus]); err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	//end := time.Now().Sub(start)
	//fmt.Printf("ConsLog just verify %v\n", end.Seconds())
	return nil
}

func CheckValidationDataWithCommittee(valData *consensustypes.ValidationData, committee []incognitokey.CommitteePublicKey) bool {
	if len(committee) < 1 {
		return false
	}
	if len(valData.ValidatiorsIdx) < len(committee)*2/3+1 {
		return false
	}
	for i := 0; i < len(valData.ValidatiorsIdx)-1; i++ {
		if valData.ValidatiorsIdx[i] >= valData.ValidatiorsIdx[i+1] {
			return false
		}
	}
	return true
}

func ValidateCommitteeSig(block types.BlockInterface, committee []incognitokey.CommitteePublicKey, numFixNode int) error {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}

	reduceFixNodeVersion := GetLatestReduceFixNodeVersion()
	if reduceFixNodeVersion != 0 && block.GetVersion() >= reduceFixNodeVersion {
		cnt := 0
		for _, v := range valData.ValidatiorsIdx {
			if v < numFixNode {
				cnt++
			}
		}
		if cnt <= 2*numFixNode/3 {
			return errors.New("Not enough fix node votes")
		}
	}

	valid := CheckValidationDataWithCommittee(valData, committee)
	if !valid {
		committeeStr, _ := incognitokey.CommitteeKeyListToString(committee)
		return NewConsensusError(UnExpectedError, errors.New(fmt.Sprintf("This validation Idx %v is not valid with this committee %v", valData.ValidatiorsIdx, committeeStr)))
	}
	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range committee {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[consensusName])
	}

	if err := validateBLSSig(block.ProposeHash(), valData.AggSig, valData.ValidatiorsIdx, committeeBLSKeys); err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	return nil
}

func validateSingleBLSSig(
	dataHash *common.Hash,
	blsSig []byte,
	selfIdx int,
	committee []blsmultisig.PublicKey,
) error {
	//start := time.Now()
	result, err := blsmultisig.Verify(blsSig, dataHash.GetBytes(), []int{selfIdx}, committee)
	//end := time.Now().Sub(start)
	//fmt.Printf("ConsLog single verify %v\n", end.Seconds())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	if !result {
		return NewConsensusError(UnExpectedError, errors.New("invalid BLS Signature"))
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
		return NewConsensusError(UnExpectedError, err)
	}
	if !result {
		return NewConsensusError(UnExpectedError, errors.New("invalid BRI Signature"))
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
		return NewConsensusError(UnExpectedError, err)
	}
	if !result {
		return NewConsensusError(UnExpectedError, errors.New("Invalid Signature!"))
	}
	return nil
}
