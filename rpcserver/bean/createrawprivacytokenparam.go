package bean

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
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
	// all component
	arrayParams := common.InterfaceSlice(params)

	/****** START FEtch data from component *********/
	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKeySet, shardIDSender, err := GetKeySetFromPrivateKeyParams(senderKeyParam.(string))
	if err != nil {
		return nil, errors.New("invalid sender's key")
	}

	// param #2: list receiver
	receiversPaymentAddressStrParam := make(map[string]interface{})
	if arrayParams[1] != nil {
		receiversPaymentAddressStrParam = arrayParams[1].(map[string]interface{})
	}
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversPaymentAddressStrParam {
		keyWalletReceiver, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, errors.New("invalid receiver paymentaddress")
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	// param #3: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[2].(float64))

	// param #4: hasPrivacy flag for native coin
	hasPrivacyCoin := int(arrayParams[3].(float64)) > 0

	// param #5: token component
	tokenParamsRaw := arrayParams[4].(map[string]interface{})

	// param #6: hasPrivacyToken flag for token
	hasPrivacyToken := true
	if len(arrayParams) >= 6 {
		hasPrivacyToken = int(arrayParams[5].(float64)) > 0
	}

	// param#7: info (option)
	info := []byte{}
	if len(arrayParams) >= 7 {
		infoStr := arrayParams[6].(string)
		info = []byte(infoStr)
	}

	/****** END FEtch data from params *********/

	return &CreateRawPrivacyTokenTxParam{
		SenderKeySet:         senderKeySet,
		ShardIDSender:        shardIDSender,
		PaymentInfos:         paymentInfos,
		EstimateFeeCoinPerKb: int64(estimateFeeCoinPerKb),
		HasPrivacyCoin:       hasPrivacyCoin,
		Info:                 info,
		HasPrivacyToken:      hasPrivacyToken,
		TokenParamsRaw:       tokenParamsRaw,
	}, nil
}
