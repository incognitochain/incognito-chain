package privacy

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init()!")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestCoin(t *testing.T) {
	privateKey := GeneratePrivateKey(new(big.Int).SetInt64(123).Bytes())
	paymentAddress := GeneratePaymentAddress(privateKey)
	viewingKey := GenerateViewingKey(privateKey)

	coin := new(OutputCoin)
	coin.CoinDetails = new(Coin)
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = 10
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

	assert.Equal(t, coin.CoinDetails.Randomness, coin2.CoinDetails.Randomness)
	assert.Equal(t, coin.CoinDetails.Value, coin2.CoinDetails.Value)

	// test for JS
	hCoin := new(Coin)
	hCoin.SetBytes([]byte{33, 3, 71, 166, 83, 226, 71, 95, 14, 188, 57, 177, 14, 85, 249, 136, 146, 169, 160, 86, 50, 207, 24, 120, 71, 251, 247, 227, 93, 147, 22, 190, 2, 80, 33, 2, 182, 112, 40, 248, 74, 48, 122, 127, 238, 105, 148, 211, 170, 98, 165, 26, 249, 83, 116, 115, 27, 244, 132, 202, 217, 1, 55, 249, 75, 35, 244, 76, 32, 76, 68, 103, 85, 160, 100, 31, 182, 137, 58, 230, 23, 114, 193, 200, 6, 239, 127, 71, 231, 116, 40, 98, 203, 94, 63, 8, 201, 225, 166, 127, 69, 33, 3, 52, 31, 171, 251, 4, 27, 117, 71, 234, 164, 94, 158, 28, 77, 255, 22, 236, 62, 61, 162, 118, 72, 43, 157, 77, 131, 31, 122, 103, 177, 192, 39, 32, 118, 48, 107, 67, 179, 21, 83, 87, 36, 163, 20, 54, 148, 60, 254, 13, 20, 249, 88, 61, 96, 194, 84, 182, 197, 121, 77, 119, 20, 164, 186, 174, 1, 10, 0})
	fmt.Printf("*********** hCoin info *******\n")
	fmt.Printf("hCoin.Public key : %v\n", hCoin.PublicKey.Compress())
	fmt.Printf("hCoin.SerialNumber : %v\n", hCoin.SerialNumber.Compress())
	fmt.Printf("hCoin.CoinCommitment : %v\n", hCoin.CoinCommitment.Compress())
	fmt.Printf("hCoin.SNDerivator : %v\n", hCoin.SNDerivator.Bytes())
	fmt.Printf("hCoin.Randomness : %v\n", hCoin.Randomness.Bytes())
	fmt.Printf("hCoin.Value : %v\n", hCoin.Value)

	houtCoin := new(OutputCoin)
	err = houtCoin.SetBytes([]byte{115, 3, 114, 229, 30, 30, 7, 124, 128, 126, 156, 178, 159, 6, 79, 14, 206, 103, 116, 69, 132, 190, 238, 4, 130, 119, 70, 191, 61, 6, 82, 71, 108, 193, 2, 7, 53, 57, 178, 89, 183, 150, 176, 35, 23, 152, 50, 222, 42, 209, 194, 34, 211, 13, 161, 163, 6, 127, 174, 62, 88, 98, 38, 103, 96, 218, 23, 140, 228, 137, 255, 248, 143, 145, 218, 46, 246, 216, 253, 171, 231, 134, 67, 112, 99, 50, 111, 58, 203, 65, 199, 79, 141, 45, 126, 101, 160, 42, 31, 197, 109, 51, 95, 152, 84, 91, 252, 171, 100, 3, 214, 62, 145, 83, 237, 2, 171, 33, 3, 71, 166, 83, 226, 71, 95, 14, 188, 57, 177, 14, 85, 249, 136, 146, 169, 160, 86, 50, 207, 24, 120, 71, 251, 247, 227, 93, 147, 22, 190, 2, 80, 33, 2, 44, 79, 100, 248, 68, 98, 176, 129, 195, 54, 128, 23, 194, 92, 227, 73, 147, 75, 184, 0, 115, 10, 208, 93, 89, 214, 95, 188, 227, 63, 91, 236, 32, 200, 246, 88, 6, 36, 28, 226, 130, 166, 147, 58, 169, 16, 67, 114, 186, 185, 46, 137, 185, 96, 239, 121, 79, 97, 252, 7, 83, 243, 78, 191, 27, 33, 2, 235, 32, 11, 106, 166, 17, 58, 220, 2, 135, 170, 142, 246, 27, 56, 139, 254, 236, 219, 32, 165, 228, 161, 120, 82, 97, 163, 38, 164, 45, 93, 62, 32, 65, 227, 148, 142, 178, 96, 105, 34, 82, 85, 97, 51, 244, 164, 162, 224, 29, 222, 145, 112, 66, 142, 175, 128, 35, 142, 227, 143, 108, 91, 35, 46, 1, 10, 0})
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
	}

	houtCoin.Decrypt(viewingKey)
	fmt.Printf("*********** hOutCoin info *******\n")
	fmt.Printf("hCoin.Public key : %v\n", houtCoin.CoinDetails.PublicKey.Compress())
	fmt.Printf("hCoin.SerialNumber : %v\n", houtCoin.CoinDetails.SerialNumber.Compress())
	fmt.Printf("hCoin.CoinCommitment : %v\n", houtCoin.CoinDetails.CoinCommitment.Compress())
	fmt.Printf("hCoin.SNDerivator : %v\n", houtCoin.CoinDetails.SNDerivator.Bytes())
	fmt.Printf("hCoin.Randomness : %v\n", houtCoin.CoinDetails.Randomness.Bytes())
	fmt.Printf("hCoin.Value : %v\n", houtCoin.CoinDetails.Value)
}

func TestEncryptCoin(t *testing.T) {
	// prepare key
	privateKey := GeneratePrivateKey(new(big.Int).SetInt64(123).Bytes())
	paymentAddress := GeneratePaymentAddress(privateKey)
	viewingKey := GenerateViewingKey(privateKey)

	for i := 0; i < 100000; i++ {
		fmt.Printf("\n\n i: %v\n", i)
		// new output coin with value and randomness
		coin := new(OutputCoin)
		coin.CoinDetails = new(Coin)
		coin.CoinDetails.Randomness = RandScalar()
		coin.CoinDetails.Value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
		coin.CoinDetails.PublicKey = new(EllipticPoint)
		err := coin.CoinDetails.PublicKey.Decompress(paymentAddress.Pk)
		if err != nil {
			Logger.Log.Error(err)
		}

		// encrypt output coins
		err = coin.Encrypt(paymentAddress.Tk)
		if err.(*PrivacyError) != nil {
			Logger.Log.Error(err)
		}

		// convert output coin to bytes array
		coinByte := coin.Bytes()

		// create new output coin to test
		coin2 := new(OutputCoin)
		err = coin2.SetBytes(coinByte)
		if err != nil {
			Logger.Log.Error(err)
		}

		err = coin2.Decrypt(viewingKey)
		if err.(*PrivacyError) != nil {
			Logger.Log.Error(err)
		}

		assert.Equal(t, coin.CoinDetails.Randomness, coin2.CoinDetails.Randomness)
		assert.Equal(t, coin.CoinDetails.Value, coin2.CoinDetails.Value)
	}
}
