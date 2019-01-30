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

	//outCoin := new(OutputCoin)
	//outCoin.CoinDetails = new(Coin)
	//outCoin.CoinDetailsEncrypted = new(Ciphertext)
	//outCoin.CoinDetailsEncrypted.SetBytes([]byte{223, 198, 8, 136, 78, 57, 73, 132, 70, 137, 192, 152, 232, 168, 205, 40, 83, 235, 96, 71, 95, 216, 33, 195, 251, 59, 58, 26, 204, 71, 148, 31, 21, 47, 1, 168, 194, 71, 248, 147, 195, 146, 44, 151, 55, 116, 68, 95, 2, 50, 67, 64, 175, 218, 141, 69, 113, 134, 11, 130, 200, 66, 73, 237, 36, 91, 63, 149, 193, 47, 14, 42, 61, 122, 81, 183, 222, 92, 100, 125, 170, 3, 123, 246, 152, 211, 39, 110, 93, 59, 0, 95, 92, 44, 249, 195, 145, 1, 55, 136, 18, 233, 125, 7, 81, 37, 12, 177, 93, 35, 99, 241, 215, 199, 12, 114, 95, 83, 164, 137, 229, 9, 170, 21, 65, 111, 97, 26, 18, 17, 8})
	//fmt.Printf("Out coin set byte: %+v\n", outCoin.CoinDetailsEncrypted)
	////fmt.Printf("Out coin set byte randomness: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedRandomness))
	////fmt.Printf("Out coin set byte value: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedValue))
	////fmt.Printf("Out coin set byte sym key: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedSymKey))
	//
	//outCoin.Decrypt(viewingKey)




	assert.Equal(t, coin.CoinDetails.Randomness , coin2.CoinDetails.Randomness)
	assert.Equal(t, coin.CoinDetails.Value , coin2.CoinDetails.Value)
}



