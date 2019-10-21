package bean

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type CreateRawPrivacyTokenTxParam struct {
	SenderKeySet         *incognitokey.KeySet
	ShardIDSender        byte
	PaymentInfos         []*privacy.PaymentInfo
	EstimateFeeCoinPerKb int64
	HasPrivacyCoin       bool
	TokenParamsRaw       map[string]interface{}
	TokenParams          map[string]interface{}
	HasPrivacyToken      bool
	Info                 []byte
}

func NewCreateRawPrivacyTokenTxParam(params interface{}) (*CreateRawPrivacyTokenTxParam, error) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 5 {
		return nil, errors.New("not enough param")
	}

	// create basic param for tx
	txparam, err := NewCreateRawTxParam(params)
	if err != nil {
		return nil, err
	}

	// param #5: token component
	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok  {
		return nil, errors.New("token param is invalid")
	}

	// param #7: hasPrivacyToken flag for token
	hasPrivacyToken := true
	if len(arrayParams) >= 7 {
		hasPrivacyTokenParam, ok := arrayParams[6].(float64)
		if !ok  {
			return nil, errors.New("has privacy for token param is invalid")
		}
		hasPrivacyToken = int(hasPrivacyTokenParam) > 0
	}

	/****** END FEtch data from params *********/

	return &CreateRawPrivacyTokenTxParam{
		SenderKeySet:         txparam.SenderKeySet,
		ShardIDSender:        txparam.ShardIDSender,
		PaymentInfos:         txparam.PaymentInfos,
		EstimateFeeCoinPerKb: int64(txparam.EstimateFeeCoinPerKb),
		HasPrivacyCoin:       txparam.HasPrivacyCoin,
		Info:                 txparam.Info,
		HasPrivacyToken:      hasPrivacyToken,
		TokenParamsRaw:       tokenParamsRaw,
	}, nil
}
