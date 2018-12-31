package privacy

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestEncryptionCoin(t *testing.T){
	coin := new(OutputCoin)
	coin.CoinDetails = new(Coin)
	coin.CoinDetails.Randomness = RandInt()
	coin.CoinDetails.Value = 10

	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	paymentAddress := GeneratePaymentAddress(spendingKey)
	viewingKey := GenerateViewingKey(spendingKey)

	coin.CoinDetails.PublicKey, _ = DecompressKey(paymentAddress.Pk)

	err := coin.Encrypt(paymentAddress.Tk)
	if err != nil {
		Logger.Log.Error(err)
	}

	coinByte := coin.Bytes()

	coin2 := new(OutputCoin)
	err = coin2.SetBytes(coinByte)
	if err != nil {
		Logger.Log.Error(err)
	}

	coin.Decrypt(viewingKey)

	assert.Equal(t, coin.CoinDetails.Randomness , coin2.CoinDetails.Randomness)
	assert.Equal(t, coin.CoinDetails.Value , coin2.CoinDetails.Value)
}



