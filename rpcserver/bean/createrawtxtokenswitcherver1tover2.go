package bean

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type CreateRawTxTokenSwitchVer1ToVer2Param struct {
	SenderKeySet         *incognitokey.KeySet
	ShardIDSender        byte
	TokenID				 *common.Hash
	EstimateFeeCoinPerKb int64
	Info                 []byte
}

func NewCreateRawPrivacyTokenTxConversionVer1To2Param(params interface{}) (*CreateRawTxTokenSwitchVer1ToVer2Param, error) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 3 {
		return nil, errors.New("not enough param")
	}

	// param #1: private key of sender
	senderKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, errors.New("sender private key is invalid")
	}
	senderKeySet, shardIDSender, err := GetKeySetFromPrivateKeyParams(senderKeyParam)
	if err != nil {
		return nil, err
	}

	// params #2: tokenID
	tokenID, ok := arrayParams[1].(string)
	if !ok {
		return nil, errors.New("tokenID is invalid")
	}
	tokenHash, err := common.TokenStringToHash(tokenID)
	if err != nil {
		return nil, errors.New("cannot parsetokenID to hash")
	}

	// param #3: estimation fee nano P per kb
	estimateFeeCoinPerKb, ok := arrayParams[2].(float64)
	if !ok {
		return nil, errors.New("estimate fee coin per kb is invalid")
	}

	// param #4: info (optional)
	info := []byte{}
	if len(arrayParams) > 3 {
		if arrayParams[3] != nil {
			infoStr, ok := arrayParams[3].(string)
			if !ok {
				return nil, errors.New("info is invalid")
			}
			info = []byte(infoStr)
		}
	}

	return &CreateRawTxTokenSwitchVer1ToVer2Param{
		SenderKeySet:         senderKeySet,
		ShardIDSender:        shardIDSender,
		TokenID: 			  tokenHash,
		EstimateFeeCoinPerKb: int64(estimateFeeCoinPerKb),
		Info:                 info,
	}, nil
}