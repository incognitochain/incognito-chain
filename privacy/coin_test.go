package privacy

import (
	"fmt"
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

	coin3Bytes := []byte{33, 3, 71, 166, 83, 226, 71, 95, 14, 188, 57, 177, 14, 85, 249, 136, 146, 169, 160, 86, 50, 207, 24, 120, 71, 251, 247, 227, 93, 147, 22, 190, 2, 80, 0, 32, 147, 64, 112, 6, 102, 82, 136,
		65, 199, 170, 115, 230, 118, 88, 199, 83, 112, 100, 141, 105, 0, 143, 141, 66, 128, 108, 255, 232, 4, 126, 88, 110, 0, 0, 1, 10, 0}
	coin3 := new(Coin)
	coin3.SetBytes(coin3Bytes)

	fmt.Printf("coin3 info: %+v", coin3)
	fmt.Printf("Public key: %v\n", coin3.PublicKey)

	assert.Equal(t, coin.CoinDetails.Randomness , coin2.CoinDetails.Randomness)
	assert.Equal(t, coin.CoinDetails.Value , coin2.CoinDetails.Value)
}



