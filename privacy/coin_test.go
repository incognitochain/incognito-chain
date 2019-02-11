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
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = 10

	spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	paymentAddress := GeneratePaymentAddress(spendingKey)
	viewingKey := GenerateViewingKey(spendingKey)

	fmt.Printf("viewing key: %+v\n", viewingKey)

	coin.CoinDetails.PublicKey = new(EllipticPoint)
	err := coin.CoinDetails.PublicKey.Decompress(paymentAddress.Pk)
	if err != nil {
		Logger.Log.Error(err)
	}

	err = coin.Encrypt(paymentAddress.Tk)
	if err != nil {
		Logger.Log.Error(err)
	}

	coinByte := coin.Bytes()

	coin2 := new(OutputCoin)
	err = coin2.SetBytes(coinByte)
	if err != nil {
		Logger.Log.Error(err)
	}

	err = coin2.Decrypt(viewingKey)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
	}

	fmt.Printf("Out coin decrypted randomness: %+v\n", coin2.CoinDetails.Randomness.Bytes())
	fmt.Printf("Out coin decrypted value: %+v\n", coin2.CoinDetails.Value)


	//coin3Bytes := []byte{33, 3, 71, 166, 83, 226, 71, 95, 14, 188, 57, 177, 14, 85, 249, 136, 146, 169, 160, 86, 50, 207, 24, 120, 71, 251, 247, 227, 93, 147, 22, 190, 2, 80, 0, 32, 147, 64, 112, 6, 102, 82, 136,
	//	65, 199, 170, 115, 230, 118, 88, 199, 83, 112, 100, 141, 105, 0, 143, 141, 66, 128, 108, 255, 232, 4, 126, 88, 110, 0, 0, 1, 10, 0}
	//coin3 := new(Coin)
	//coin3.SetBytes(coin3Bytes)
	//
	//fmt.Printf("coin3 info: %+v\n", coin3)
	////fmt.Printf("Public key: %v\n", coin3.PublicKey.Compress())
	//fmt.Printf("input: %v\n", coin3.SNDerivator.Bytes())
	////fmt.Printf("Public key: %v\n", coin3.PublicKey)

	outCoin := new(OutputCoin)
	outCoin.CoinDetails = new(Coin)
	outCoin.CoinDetailsEncrypted = new(Ciphertext)
	outCoin.CoinDetailsEncrypted.SetBytes([]byte{2, 131, 97, 41, 144, 247, 106, 3, 226, 241, 47, 187, 150, 165, 59, 10, 227, 28, 185, 225, 90, 168, 114, 15, 47, 68, 43, 50, 97, 192, 228, 245, 248, 2, 20, 32, 124, 237, 116, 13, 204, 165, 88, 71, 244, 236, 240, 252, 180, 59, 28, 125, 60, 83, 10, 130, 164, 148, 229, 217, 133, 66, 237, 224, 158, 29, 131, 185, 142, 253, 236, 201, 14, 248, 104, 214, 180, 90, 184, 68, 217, 247, 119, 202, 142, 18, 76, 34, 180, 35, 8, 138, 76, 191, 65, 26, 252, 143, 170, 226, 218, 169, 114, 229, 214, 32, 156, 82, 99, 123, 120, 105, 218, 252, 87})
	fmt.Printf("Out coin set byte: %+v\n", outCoin.CoinDetailsEncrypted)
	//fmt.Printf("Out coin set byte randomness: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedRandomness))
	//fmt.Printf("Out coin set byte value: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedValue))
	//fmt.Printf("Out coin set byte sym key: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedSymKey))

	outCoin.Decrypt(viewingKey)

	fmt.Printf("Outcoin value: %v\n", outCoin.CoinDetails.Value)
	fmt.Printf("Outcoin randomness: %v\n", outCoin.CoinDetails.Randomness.Bytes())



	assert.Equal(t, coin.CoinDetails.Randomness , coin2.CoinDetails.Randomness)
	assert.Equal(t, coin.CoinDetails.Value , coin2.CoinDetails.Value)
}



