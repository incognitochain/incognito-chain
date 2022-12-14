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
	"github.com/incognitochain/incognito-chain/multiview"
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

func ValidateCommitteeSignsMajority(signingInfo map[string]bool) bool {
	expectedVotes := 0
	gotVotes := 0
	expectedVotes = len(signingInfo)
	for _, v := range signingInfo {
		if v {
			gotVotes++
		}
	}
	return (gotVotes > 2*expectedVotes/3)
}

func ValidateCommitteeVotePowerMajority(prevView multiview.View, signingInfo map[string]bool) bool {
	reputation := prevView.GetReputation()
	expectedVotePower := uint64(0)
	gotVotePower := uint64(0)
	if len(reputation) == 0 {
		return ValidateCommitteeSignsMajority(signingInfo)
	}
	for pk, voted := range signingInfo {
		if power, ok := reputation[pk]; !ok {
			return false
		} else {
			expectedVotePower += power
			if voted {
				gotVotePower += power
			}
		}
	}
	return (gotVotePower > 2*expectedVotePower/3)
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

func ValidateCommitteeSigWithView(prevView multiview.View, block types.BlockInterface, committee []incognitokey.CommitteePublicKey, numFixNode int) error {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	committeeStr, err := incognitokey.CommitteeKeyListToString(committee)
	if err != nil {
		return err
	}
	signingInfo := map[string]bool{}
	for _, k := range committeeStr {
		signingInfo[k] = false
	}
	for _, idx := range valData.ValidatiorsIdx {
		signingInfo[committeeStr[idx]] = true
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

	valid := ValidateCommitteeVotePowerMajority(prevView, signingInfo)
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
