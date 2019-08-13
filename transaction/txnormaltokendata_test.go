package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateCustomTokenReceiverArray(t *testing.T) {
	data := make(map[string]interface{})
	data["1Uv3BkYiWy9Mjt1yBa4dXBYKo3az22TeCVEpeXN93ieJ8qhrTDuUZBzsPZWjjP2AeRQnjw1y18iFPHTRuAqqufwVC1vNUAWs4wHFbbWC2"] = 10.0
	data["1Uv2oZbQ2Zs1qGtNBfWmLUm862wJyxw8Br7r1CysKmC4MMyv9ySR7urcaEjABH9NK2zfei8s8iGepP7ZAmyh4ew9JT9mQy7RpGF8W9w6C"] = 20.0
	result, voutsAmount, err := CreateCustomTokenReceiverArray(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(30), uint64(voutsAmount))
	assert.Equal(t, 2, len(result))
}

func TestTxTokenData(t *testing.T) {
	paymentAddress, _ := wallet.Base58CheckDeserialize("1Uv3BkYiWy9Mjt1yBa4dXBYKo3az22TeCVEpeXN93ieJ8qhrTDuUZBzsPZWjjP2AeRQnjw1y18iFPHTRuAqqufwVC1vNUAWs4wHFbbWC2")
	data := TxTokenData{
		PropertyID:     common.PRVCoinID,
		PropertyName:   "PRV Coin",
		PropertySymbol: "PRV",
		Amount:         10,
		Vouts:          []TxTokenVout{{PaymentAddress: paymentAddress.KeySet.PaymentAddress, Value: 10, index: 3}},
		Mintable:       false,
		Type:           0,
		Vins:           []TxTokenVin{{PaymentAddress: paymentAddress.KeySet.PaymentAddress, TxCustomTokenID: common.PRVCoinID, VoutIndex: 1}},
	}
	str := data.JSONString()
	assert.NotEqual(t, "", str)
	hash, err := data.Hash()
	assert.Equal(t, nil, err)
	assert.Equal(t, 32, len(hash))

	hash = data.Vouts[0].Hash()
	assert.Equal(t, 32, len(hash))
	i := data.Vouts[0].GetIndex()
	assert.Equal(t, 3, i)
	jsonStri := data.Vouts[0].JSONString()
	assert.NotEqual(t, "", jsonStri)
	data.Vouts[0].SetIndex(2)
	assert.Equal(t, 2, data.Vouts[0].index)
	data.Vouts[0].SetTxCustomTokenID(common.Hash{})
	assert.Equal(t, common.Hash{}, data.Vouts[0].GetTxCustomTokenID())

	hash = data.Vins[0].Hash()
	assert.Equal(t, 32, len(hash))
	jsonStri = data.Vins[0].JSONString()
	assert.NotEqual(t, "", jsonStri)
}

func TestCustomTokenParamTx(t *testing.T) {
	param := CustomTokenParamTx{}
	param.SetVinsAmount(10)
	assert.Equal(t, uint64(10), param.vinsAmount)
	param.SetVins([]TxTokenVin{{TxCustomTokenID: common.PRVCoinID, VoutIndex: 1}})
	assert.Equal(t, 1, len(param.vins))
}
