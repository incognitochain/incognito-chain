package privacy

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"math/big"
	"testing"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.Log.Info("This runs before init()!")
	return
}()

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	m.Run()
}

/*
	Unit test for CommitAll Coin
*/

func TestCoinCommitAll(t *testing.T) {
	// init coin with fully fields
	// init public key
	//coin := new(Coin).Init()
	//seedKey := []byte{1, 2, 3}
	//privateKey := GeneratePrivateKey(seedKey)
	//publicKey := GeneratePublicKey(privateKey)
	//
	//// init other fields for coin
	//coin.publicKey.Decompress(publicKey)
	//coin.snDerivator = RandScalar()
	//coin.randomness = RandScalar()
	//coin.value = uint64(100)
	//coin.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.snDerivator)
	//coin.CommitAll()
	//coin.info = []byte("Incognito chain")
	//
	//expectedCm := coin.publicKey
	//expectedCm = expectedCm.Add(PedCom.G[PedersenValueIndex].ScalarMul(big.NewInt(int64(coin.value))))
	//expectedCm = expectedCm.Add(PedCom.G[PedersenSndIndex].ScalarMul(coin.snDerivator))
	//expectedCm = expectedCm.Add(PedCom.G[PedersenShardIDIndex].ScalarMul(big.NewInt(int64(common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())))))
	//expectedCm = expectedCm.Add(PedCom.G[PedersenRandomnessIndex].ScalarMul(coin.randomness))
	//
	//assert.Equal(t, expectedCm, coin.coinCommitment)

	for i:= 0; i<10000; i++{
		coin := new(Coin).Init()
		seedKey := RandBytes(3)
		privateKey := GeneratePrivateKey(seedKey)
		publicKey := GeneratePublicKey(privateKey)

		// init other fields for coin
		coin.publicKey.Decompress(publicKey)
		var r = rand.Reader
		coin.snDerivator = RandScalar(r)
		coin.randomness = RandScalar(r)
		coin.value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
		coin.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.snDerivator)
		coin.CommitAll()
		coin.info = []byte("Incognito chain")

		//expectedCm := coin.publicKey
		//expectedCm = expectedCm.Add(PedCom.G[PedersenValueIndex].ScalarMul(big.NewInt(int64(coin.value))))
		//expectedCm = expectedCm.Add(PedCom.G[PedersenSndIndex].ScalarMul(coin.snDerivator))
		//expectedCm = expectedCm.Add(PedCom.G[PedersenShardIDIndex].ScalarMul(big.NewInt(int64(common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())))))
		//expectedCm = expectedCm.Add(PedCom.G[PedersenRandomnessIndex].ScalarMul(coin.randomness))


		cmTmp := coin.GetPublicKey()
		cmTmp = cmTmp.Add(PedCom.G[PedersenValueIndex].ScalarMult(big.NewInt(int64(coin.GetValue()))))
		cmTmp = cmTmp.Add(PedCom.G[PedersenSndIndex].ScalarMult(coin.GetSNDerivator()))
		shardID := common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())
		cmTmp = cmTmp.Add(PedCom.G[PedersenShardIDIndex].ScalarMult(new(big.Int).SetBytes([]byte{shardID})))
		cmTmp = cmTmp.Add(PedCom.G[PedersenRandomnessIndex].ScalarMult(coin.GetRandomness()))

		res := cmTmp.IsEqual(coin.GetCoinCommitment())

		assert.Equal(t, true, res)
	}
}

func TestCoin2(t *testing.T){
	outCoin2 := new(OutputCoin)
	//data := []byte{0, 174, 33, 3, 60, 123, 206, 207, 7, 52, 248, 65, 70, 49, 30, 41, 32, 61, 234, 142, 11, 181, 170, 120, 127, 187, 113, 61, 104, 145, 81, 29, 206, 12, 226, 235, 33, 3, 250, 55, 13, 132, 208, 95, 63, 41, 182, 236, 75, 192, 191, 226, 65, 213, 63, 6, 21, 170, 176, 185, 244, 136, 30, 254, 135, 114, 220, 47, 246, 101, 32, 86, 91, 14, 170, 4, 145, 115, 68, 13, 234, 139, 17, 15, 64, 246, 149, 147, 185, 30, 118, 209, 93, 62, 153, 59, 6, 19, 151, 11, 5, 73, 0, 33, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 178, 231, 244, 64, 100, 215, 171, 192, 244, 124, 143, 18, 65, 36, 229, 173, 200, 165, 92, 178, 69, 146, 49, 132, 18, 137, 133, 105, 66, 168, 81, 181, 4, 59, 154, 202, 0, 0}
	data := []byte{0, 174, 33, 3, 60, 123, 206, 207, 7, 52, 248, 65, 70, 49, 30, 41, 32, 61, 234, 142, 11, 181, 170, 120, 127, 187, 113, 61, 104, 145, 81, 29, 206, 12, 226, 235, 33, 2, 48, 33, 142, 179, 236, 145, 252, 24, 55, 35, 136, 238, 241, 247, 74, 248, 61, 253, 181, 66, 8, 254, 27, 187, 155, 2, 230, 105, 22, 154, 72, 18, 32, 20, 11, 159, 16, 19, 37, 100, 86, 13, 227, 8, 236, 93, 32, 177, 171, 99, 4, 241, 120, 20, 14, 29, 197, 132, 209, 245, 95, 107, 225, 100, 0, 33, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 228, 12, 207, 86, 14, 216, 27, 216, 112, 222, 80, 217, 188, 231, 9, 224, 226, 195, 209, 105, 83, 63, 63, 239, 66, 237, 238, 17, 195, 181, 59, 219, 4, 59, 154, 202, 0, 0}
	outCoin2.SetBytes(data)
	fmt.Printf("outCoin bytes: %v\n", outCoin2.Bytes())
	fmt.Printf("Value: %v\n", outCoin2.CoinDetails.GetValue())
	fmt.Printf("SND: %v\n", outCoin2.CoinDetails.GetSNDerivator().Bytes())
	fmt.Printf("Randomness: %v\n", outCoin2.CoinDetails.GetRandomness().Bytes())
	fmt.Printf("PublicKey: %v\n", outCoin2.CoinDetails.GetPublicKey().Compress())
	fmt.Printf("PublicKey last byte: %v\n", outCoin2.CoinDetails.GetPubKeyLastByte())
	// right


	// please check coin commitment all
	cmTmp := outCoin2.CoinDetails.GetPublicKey()
	fmt.Printf("cmTmp 1: %v\n", cmTmp.Compress())
	cmTmp = cmTmp.Add(PedCom.G[PedersenValueIndex].ScalarMult(big.NewInt(int64(outCoin2.CoinDetails.GetValue()))))
	fmt.Printf("cmTmp 2: %v\n", cmTmp.Compress())
	cmTmp = cmTmp.Add(PedCom.G[PedersenSndIndex].ScalarMult(outCoin2.CoinDetails.GetSNDerivator()))
	fmt.Printf("cmTmp 3: %v\n", cmTmp.Compress())
	shardID := common.GetShardIDFromLastByte(outCoin2.CoinDetails.GetPubKeyLastByte())
	cmTmp = cmTmp.Add(PedCom.G[PedersenShardIDIndex].ScalarMult(new(big.Int).SetBytes([]byte{shardID})))
	fmt.Printf("cmTmp 4: %v\n", cmTmp.Compress())
	cmTmp = cmTmp.Add(PedCom.G[PedersenRandomnessIndex].ScalarMult(outCoin2.CoinDetails.GetRandomness()))
	fmt.Printf("cmTmp 5: %v\n", cmTmp.Compress())
	if !cmTmp.IsEqual(outCoin2.CoinDetails.GetCoinCommitment()) {
		Logger.Log.Errorf("Output coins %v commitment wrong!\n")
	}

	fmt.Printf("cmTmp: %v\n", cmTmp.Compress())

	//right commitment JS : 3, 112, 194, 190, 145, 26, 138, 190, 212, 123, 173, 209, 233, 112, 105, 51, 103, 139, 130, 171, 174, 230, 154, 192, 19, 219, 147, 224, 234, 47, 202, 68, 5


}

func TestBigInt(t *testing.T) {
	snd1 := new (big.Int).SetBytes([]byte{20, 11, 159, 16, 19, 37, 100, 86, 13, 227, 8, 236, 93, 32, 177, 171, 99, 4, 241, 120, 20, 14, 29, 197, 132, 209, 245, 95, 107, 225, 100})
	cm1 := PedCom.G[PedersenSndIndex].ScalarMult(snd1)
	fmt.Printf("cm1: %v\n", cm1.Compress())

	snd2 := new (big.Int).SetBytes([]byte{0, 20, 11, 159, 16, 19, 37, 100, 86, 13, 227, 8, 236, 93, 32, 177, 171, 99, 4, 241, 120, 20, 14, 29, 197, 132, 209, 245, 95, 107, 225, 100})
	cm2 := PedCom.G[PedersenSndIndex].ScalarMult(snd2)
	fmt.Printf("cm1: %v\n", cm2.Compress())
}
/*
	Unit test for MarshalJSON/UnmarshalJSON Coin
*/
func TestCoinMarshalJSON(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(Coin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.snDerivator = RandScalar(r)
	coin.randomness = RandScalar(r)
	coin.value = uint64(100)
	coin.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.snDerivator)
	coin.CommitAll()
	coin.info = []byte("Incognito chain")

	bytesJSON, err := coin.MarshalJSON()
	assert.Equal(t, nil, err)
	assert.Greater(t, len(bytesJSON), 0)

	coin2 := new(Coin)
	err2 := coin2.UnmarshalJSON(bytesJSON)
	assert.Equal(t, nil, err2)
	assert.Equal(t, coin, coin2)
}

/*
	Unit test for Bytes/SetBytes Coin function
*/

func TestCoinBytesSetBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(Coin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.snDerivator = RandScalar(r)
	coin.randomness = RandScalar(r)
	coin.value = uint64(100)
	coin.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.snDerivator)
	coin.CommitAll()
	coin.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()

	assert.Greater(t, len(coinBytes), 0)

	// new coin object and set bytes from bytes array
	coin2 := new(Coin)
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestCoinBytesSetBytesWithMissingFields(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(Coin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number, coin commitment)
	coin.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.snDerivator = RandScalar(r)
	coin.randomness = RandScalar(r)
	coin.value = uint64(100)
	//coin.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.SNDerivator)
	//coin.CommitAll()
	coin.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()

	assert.Greater(t, len(coinBytes), 0)

	// new coin object and set bytes from bytes array
	coin2 := new(Coin).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestCoinBytesSetBytesWithInvalidBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(Coin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.snDerivator = RandScalar(r)
	coin.randomness = RandScalar(r)
	coin.value = uint64(100)
	coin.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.snDerivator)
	coin.CommitAll()
	coin.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes)-1] = byte(123)

	// new coin object and set bytes from bytes array
	coin2 := new(Coin).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, coin, coin2)
}

func TestCoinBytesSetBytesWithEmptyBytes(t *testing.T) {
	// new coin object and set bytes from bytes array
	coin2 := new(Coin).Init()
	err := coin2.SetBytes([]byte{})

	assert.Equal(t, errors.New("coinBytes is empty"), err)
}

/*
	Unit test for HashH Coin function
*/

func TestCoinHashH(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(Coin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.snDerivator = RandScalar(r)
	coin.randomness = RandScalar(r)
	coin.value = uint64(100)
	coin.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.snDerivator)
	coin.CommitAll()
	coin.info = []byte("Incognito chain")

	hash := coin.HashH()
	assert.Equal(t, common.HashSize, len(hash[:]))
}

/*
	Unit test for Bytes/SetBytes InputCoin function
*/

func TestInputCoinBytesSetBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(InputCoin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.CoinDetails.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.CoinDetails.snDerivator = RandScalar(r)
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.snDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()

	assert.Greater(t, len(coinBytes), 0)

	// new coin object and set bytes from bytes array
	coin2 := new(InputCoin)
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestInputCoinBytesSetBytesWithMissingFields(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(InputCoin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number, coin commitment)
	coin.CoinDetails.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.CoinDetails.snDerivator = RandScalar(r)
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.info = []byte("Incognito chain")
	coin.CoinDetails.serialNumber = nil
	coin.CoinDetails.coinCommitment = nil

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// new coin object and set bytes from bytes array
	coin2 := new(InputCoin).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestInputCoinBytesSetBytesWithInvalidBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(InputCoin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.CoinDetails.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.CoinDetails.snDerivator = RandScalar(r)
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.snDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes)-1] = byte(123)

	// new coin object and set bytes from bytes array
	coin2 := new(InputCoin).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, coin, coin2)
}

func TestInputCoinBytesSetBytesWithEmptyBytes(t *testing.T) {
	// new coin object and set bytes from bytes array
	coin2 := new(InputCoin).Init()
	err := coin2.SetBytes([]byte{})

	assert.Equal(t, errors.New("coinBytes is empty"), err)
}

/*
	Unit test for Bytes/SetBytes OutputCoin function
*/

func TestOutputCoinBytesSetBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(OutputCoin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)
	//receivingKey := GenerateReceivingKey(privateKey)
	paymentAddr := GeneratePaymentAddress(privateKey)

	// init other fields for coin
	coin.CoinDetails.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.CoinDetails.snDerivator = RandScalar(r)
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.snDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")
	coin.Encrypt(paymentAddr.Tk)

	// convert coin object to bytes array
	coinBytes := coin.Bytes()

	assert.Greater(t, len(coinBytes), 0)

	// new coin object and set bytes from bytes array
	coin2 := new(OutputCoin)
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestOutputCoinBytesSetBytesWithMissingFields(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(OutputCoin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number, coin commitment, CoinDetailsEncrypted)
	coin.CoinDetails.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.CoinDetails.snDerivator = RandScalar(r)
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.info = []byte("Incognito chain")
	coin.CoinDetails.serialNumber = nil
	coin.CoinDetails.coinCommitment = nil

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// new coin object and set bytes from bytes array
	coin2 := new(OutputCoin).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestOutputCoinBytesSetBytesWithInvalidBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(OutputCoin).Init()
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.CoinDetails.publicKey.Decompress(publicKey)
	var r = rand.Reader
	coin.CoinDetails.snDerivator = RandScalar(r)
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.serialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.snDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes)-1] = byte(123)

	// new coin object and set bytes from bytes array
	coin2 := new(OutputCoin).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, coin, coin2)
}

func TestOutputCoinBytesSetBytesWithEmptyBytes(t *testing.T) {
	// new coin object and set bytes from bytes array
	coin2 := new(OutputCoin).Init()
	err := coin2.SetBytes([]byte{})

	assert.Equal(t, errors.New("coinBytes is empty"), err)
}

/*
	Unit test for Encrypt/Decrypt OutputCoin
*/
func TestOutputCoinEncryptDecrypt(t *testing.T) {
	// prepare key
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	paymentAddress := GeneratePaymentAddress(privateKey)
	viewingKey := GenerateViewingKey(privateKey)

	var r = rand.Reader
	for i := 0; i < 100; i++ {
		// new output coin with value and randomness
		coin := new(OutputCoin).Init()
		coin.CoinDetails.randomness = RandScalar(r)
		coin.CoinDetails.value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
		coin.CoinDetails.publicKey.Decompress(paymentAddress.Pk)

		// encrypt output coins
		err := coin.Encrypt(paymentAddress.Tk)
		assert.Equal(t, (*PrivacyError)(nil), err)

		// convert output coin to bytes array
		coinBytes := coin.Bytes()

		// create new output coin to test
		coin2 := new(OutputCoin)
		err2 := coin2.SetBytes(coinBytes)
		assert.Equal(t, nil, err2)

		err3 := coin2.Decrypt(viewingKey)
		assert.Equal(t, (*PrivacyError)(nil), err3)

		assert.Equal(t, coin.CoinDetails.randomness, coin2.CoinDetails.randomness)
		assert.Equal(t, coin.CoinDetails.value, coin2.CoinDetails.value)
	}
}

func TestOutputCoinEncryptDecryptWithUnmatchedKey(t *testing.T) {
	// prepare key
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	paymentAddress := GeneratePaymentAddress(privateKey)
	viewingKey := GenerateViewingKey(privateKey)

	// new output coin with value and randomness
	coin := new(OutputCoin).Init()
	var r = rand.Reader
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
	coin.CoinDetails.publicKey.Decompress(paymentAddress.Pk)

	// encrypt output coins
	err := coin.Encrypt(paymentAddress.Tk)
	assert.Equal(t, (*PrivacyError)(nil), err)

	// convert output coin to bytes array
	coinBytes := coin.Bytes()

	// create new output coin to test
	coin2 := new(OutputCoin)
	err2 := coin2.SetBytes(coinBytes)
	assert.Equal(t, nil, err2)

	// edit receiving key to be unmatched with transmission key
	viewingKey.Rk[len(viewingKey.Rk)-1] = 123

	err3 := coin2.Decrypt(viewingKey)
	assert.Equal(t, (*PrivacyError)(nil), err3)
	assert.NotEqual(t, coin.CoinDetails.randomness, coin2.CoinDetails.randomness)
	assert.NotEqual(t, coin.CoinDetails.value, coin2.CoinDetails.value)
}

func TestOutputCoinEncryptWithInvalidKey(t *testing.T) {
	// prepare key
	seedKey := []byte{1, 2, 3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// new output coin with value and randomness
	coin := new(OutputCoin).Init()
	var r = rand.Reader
	coin.CoinDetails.randomness = RandScalar(r)
	coin.CoinDetails.value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
	coin.CoinDetails.publicKey.Decompress(publicKey)

	dataKey := [][]byte{
		{1, 2, 3}, // 3 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}, // 33 bytes, but not is on curve P256
	}

	for _, item := range dataKey {
		err := coin.Encrypt(item)
		assert.Equal(t, ErrCodeMessage[DecompressTransmissionKeyErr].Code, err.GetCode())
	}
}
