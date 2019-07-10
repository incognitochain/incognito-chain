package transaction

import (
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTxTokenPrivacyData(t *testing.T) {
	data := TxTokenPrivacyData{
		TxNormal: Tx{Proof: &zkp.PaymentProof{
			OutputCoins: []*privacy.OutputCoin{{
				CoinDetails: &privacy.Coin{
					PublicKey: new(privacy.EllipticPoint).Zero(),
					Value:     10,
				},
			}},
			InputCoins: []*privacy.InputCoin{{
				CoinDetails: &privacy.Coin{
					PublicKey: new(privacy.EllipticPoint).Zero(),
					Value:     10,
				},
			}},
		}},
	}
	hash, _ := data.Hash()
	assert.Equal(t, 32, len(hash))
	str := data.JSONString()
	assert.NotEqual(t, "", str)
}

func TestCreateCustomTokenPrivacyReceiverArray(t *testing.T) {
	data := make(map[string]interface{})
	data["1Uv3BkYiWy9Mjt1yBa4dXBYKo3az22TeCVEpeXN93ieJ8qhrTDuUZBzsPZWjjP2AeRQnjw1y18iFPHTRuAqqufwVC1vNUAWs4wHFbbWC2"] = 10.0
	data["1Uv2oZbQ2Zs1qGtNBfWmLUm862wJyxw8Br7r1CysKmC4MMyv9ySR7urcaEjABH9NK2zfei8s8iGepP7ZAmyh4ew9JT9mQy7RpGF8W9w6C"] = 20.0
	result, voutsAmount := CreateCustomTokenPrivacyReceiverArray(data)
	assert.Equal(t, uint64(30), uint64(voutsAmount))
	assert.Equal(t, 2, len(result))
}
