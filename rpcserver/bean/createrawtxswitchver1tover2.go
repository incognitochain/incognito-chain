package bean

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type CreateRawTxSwitchVer1ToVer2Param struct {
	SenderKeySet         *incognitokey.KeySet
	ShardIDSender        byte
	EstimateFeeCoinPerKb int64
	Info                 []byte
}

func NewCreateRawTxSwitchVer1ToVer2Param(params interface{}) (*CreateRawTxSwitchVer1ToVer2Param, error) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
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

	// param #2: estimation fee nano P per kb
	estimateFeeCoinPerKb, ok := arrayParams[1].(float64)
	if !ok {
		return nil, errors.New("estimate fee coin per kb is invalid")
	}

	// param #3: info (optional)
	info := []byte{}
	if len(arrayParams) > 2 {
		if arrayParams[2] != nil {
			infoStr, ok := arrayParams[2].(string)
			if !ok {
				return nil, errors.New("info is invalid")
			}
			info = []byte(infoStr)
		}

	}

	return &CreateRawTxSwitchVer1ToVer2Param{
		SenderKeySet:         senderKeySet,
		ShardIDSender:        shardIDSender,
		EstimateFeeCoinPerKb: int64(estimateFeeCoinPerKb),
		Info:                 info,
	}, nil
}
