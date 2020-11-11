package blsbft

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/consensustypes"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type vote struct {
	BLS          []byte
	BRI          []byte
	Confirmation []byte
}

type blockValidation interface {
	types.BlockInterface
	AddValidationField(validationData string)
}

func (e BLSBFT) CreateValidationData(block types.BlockInterface) consensustypes.ValidationData {
	var valData consensustypes.ValidationData
	// selfPublicKey := e.UserKeySet.GetPublicKey()
	// keyByte, _ := selfPublicKey.GetMiningKey(consensusName)
	// valData.ProducerBLSSig, _ = e.UserKeySet.BLSSignData(block.Hash().GetBytes(), 0, []blsmultisig.PublicKey{keyByte})
	valData.ProducerBLSSig, _ = e.UserKeySet.BriSignData(block.Hash().GetBytes()) //, 0, []blsmultisig.PublicKey{keyByte})
	return valData
}

func ValidateProducerSig(block types.BlockInterface) error {
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
	if err := validateSingleBriSig(block.Hash(), valData.ProducerBLSSig, producerKey.MiningPubKey[common.BridgeConsensus]); err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	//end := time.Now().Sub(start)
	//fmt.Printf("ConsLog just verify %v\n", end.Seconds())
	return nil
}

func ValidateCommitteeSig(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	committeeBLSKeys := []blsmultisig.PublicKey{}
	for _, member := range committee {
		committeeBLSKeys = append(committeeBLSKeys, member.MiningPubKey[consensusName])
	}

	if err := validateBLSSig(block.Hash(), valData.AggSig, valData.ValidatiorsIdx, committeeBLSKeys); err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	return nil
}

func (e BLSBFT) ValidateData(data []byte, sig string, publicKey string) error {
	sigByte, _, err := base58.Base58Check{}.Decode(sig)
	if err != nil {
		return NewConsensusError(UnExpectedError, err)
	}
	publicKeyByte := []byte(publicKey)
	// if err != nil {
	// 	return consensus.NewConsensusError(consensus.UnExpectedError, err)
	// }
	//fmt.Printf("ValidateData data %v, sig %v, publicKey %v\n", data, sig, publicKeyByte)
	dataHash := new(common.Hash)
	dataHash.NewHash(data)
	_, err = bridgesig.Verify(publicKeyByte, dataHash.GetBytes(), sigByte) //blsmultisig.Verify(sigByte, data, []int{0}, []blsmultisig.PublicKey{publicKeyByte})
	if err != nil {
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
