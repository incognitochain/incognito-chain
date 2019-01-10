package privacy

import (
	"fmt"
	"testing"
)

func TestEncryptionCoin(t *testing.T){
	//coin := new(OutputCoin)
	//coin.CoinDetails = new(Coin)
	//coin.CoinDetails.Randomness = RandInt()
	//coin.CoinDetails.Value = 10
	//
	//spendingKey := GenerateSpendingKey(new(big.Int).SetInt64(123).Bytes())
	//paymentAddress := GeneratePaymentAddress(spendingKey)
	//viewingKey := GenerateViewingKey(spendingKey)
	//
	//fmt.Printf("viewing key: %+v\n", viewingKey)
	//
	//coin.CoinDetails.PublicKey, _ = DecompressKey(paymentAddress.Pk)
	//
	//err := coin.Encrypt(paymentAddress.Tk)
	//if err != nil {
	//	Logger.Log.Error(err)
	//}
	//
	//coinByte := coin.Bytes()
	//
	//coin2 := new(OutputCoin)
	//err = coin2.SetBytes(coinByte)
	//if err != nil {
	//	Logger.Log.Error(err)
	//}
	//
	//coin.Decrypt(viewingKey)
	//
	//coin3Bytes := []byte{33, 3, 71, 166, 83, 226, 71, 95, 14, 188, 57, 177, 14, 85, 249, 136, 146, 169, 160, 86, 50, 207, 24, 120, 71, 251, 247, 227, 93, 147, 22, 190, 2, 80, 0, 32, 147, 64, 112, 6, 102, 82, 136,
	//	65, 199, 170, 115, 230, 118, 88, 199, 83, 112, 100, 141, 105, 0, 143, 141, 66, 128, 108, 255, 232, 4, 126, 88, 110, 0, 0, 1, 10, 0}
	//coin3 := new(Coin)
	//coin3.SetBytes(coin3Bytes)
	//
	//fmt.Printf("coin3 info: %+v\n", coin3)
	////fmt.Printf("Public key: %v\n", coin3.PublicKey.Compress())
	//fmt.Printf("input: %v\n", coin3.SNDerivator.Bytes())
	////fmt.Printf("Public key: %v\n", coin3.PublicKey)
	//
	//outCoin := new(OutputCoin)
	//outCoin.CoinDetails = new(Coin)
	//outCoin.CoinDetailsEncrypted = new(CoinDetailsEncrypted)
	//outCoin.CoinDetailsEncrypted.SetBytes([]byte{90, 75, 179, 132, 227, 135, 177, 187, 137, 58, 82, 190, 58, 177, 8, 61, 182, 52, 138, 157, 118, 19, 24, 203, 0, 145, 25, 111, 238, 103, 92, 209, 155, 129, 122, 161, 13, 114, 219, 3, 228, 247, 164, 252, 145, 249, 55, 161, 3, 101, 229, 42, 76, 11, 26, 35, 43, 144, 91, 193, 71, 62, 215, 238, 205, 76, 217, 202, 48, 36, 120, 244, 120, 91, 61, 210, 15, 99, 121, 93, 45, 2, 93, 57, 159, 34, 116, 196, 26, 134, 10, 195, 175, 91, 223, 98, 0, 207, 95, 154, 43, 116, 200, 255, 1, 205, 173, 0, 84, 125, 13, 181, 218, 81, 139, 31, 32, 157, 238, 97, 2, 175, 108, 132, 54, 12, 35, 241, 97, 43, 244})
	//fmt.Printf("Out coin set byte: %+v\n", outCoin.CoinDetailsEncrypted)
	//fmt.Printf("Out coin set byte randomness: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedRandomness))
	//fmt.Printf("Out coin set byte value: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedValue))
	//fmt.Printf("Out coin set byte sym key: %+v\n", len(outCoin.CoinDetailsEncrypted.EncryptedSymKey))
	//
	//outCoin.Decrypt(viewingKey)
	//
	//fmt.Printf("Out coin decrypted randomness: %+v\n", outCoin.CoinDetails.Randomness.Bytes())
	//fmt.Printf("Out coin decrypted value: %+v\n", outCoin.CoinDetails.Value)
	//
	//
	//assert.Equal(t, coin.CoinDetails.Randomness , coin2.CoinDetails.Randomness)
	//assert.Equal(t, coin.CoinDetails.Value , coin2.CoinDetails.Value)

	x := RandInt()
	fmt.Printf("secret: %v\n", x.Bytes())

	H := PedCom.G[0].ScalarMult(x)
	for i:=0; i< 1000000; i++{
		if PedCom.G[0].Hash(i).IsEqual(H){
			fmt.Printf("index is safe: %v\n", i)
			break
		}
	}
	fmt.Printf("Done")


}



