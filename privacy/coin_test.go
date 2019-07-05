package privacy

import (
	"errors"
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
	coin := new(Coin).Init()
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.PublicKey.Decompress(publicKey)
	coin.SNDerivator = RandScalar()
	coin.Randomness = RandScalar()
	coin.Value = uint64(100)
	coin.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.SNDerivator)
	coin.CommitAll()
	coin.Info = []byte("Incognito chain")

	expectedCm := coin.PublicKey
	expectedCm = expectedCm.Add(PedCom.G[VALUE].ScalarMult(big.NewInt(int64(coin.Value))))
	expectedCm = expectedCm.Add(PedCom.G[SND].ScalarMult(coin.SNDerivator))
	expectedCm = expectedCm.Add(PedCom.G[SHARDID].ScalarMult(big.NewInt(int64(common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())))))
	expectedCm = expectedCm.Add(PedCom.G[RAND].ScalarMult(coin.Randomness))

	assert.Equal(t, expectedCm, coin.CoinCommitment)
}

/*
	Unit test for MarshalJSON/UnmarshalJSON Coin
 */
func TestCoinMarshalJSON(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(Coin).Init()
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.PublicKey.Decompress(publicKey)
	coin.SNDerivator = RandScalar()
	coin.Randomness = RandScalar()
	coin.Value = uint64(100)
	coin.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.SNDerivator)
	coin.CommitAll()
	coin.Info = []byte("Incognito chain")

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.PublicKey.Decompress(publicKey)
	coin.SNDerivator = RandScalar()
	coin.Randomness = RandScalar()
	coin.Value = uint64(100)
	coin.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.SNDerivator)
	coin.CommitAll()
	coin.Info = []byte("Incognito chain")

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number, coin commitment)
	coin.PublicKey.Decompress(publicKey)
	coin.SNDerivator = RandScalar()
	coin.Randomness = RandScalar()
	coin.Value = uint64(100)
	//coin.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.SNDerivator)
	//coin.CommitAll()
	coin.Info = []byte("Incognito chain")

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.PublicKey.Decompress(publicKey)
	coin.SNDerivator = RandScalar()
	coin.Randomness = RandScalar()
	coin.Value = uint64(100)
	coin.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.SNDerivator)
	coin.CommitAll()
	coin.Info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes) - 1] = byte(123)

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.PublicKey.Decompress(publicKey)
	coin.SNDerivator = RandScalar()
	coin.Randomness = RandScalar()
	coin.Value = uint64(100)
	coin.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.SNDerivator)
	coin.CommitAll()
	coin.Info = []byte("Incognito chain")

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.CoinDetails.PublicKey.Decompress(publicKey)
	coin.CoinDetails.SNDerivator = RandScalar()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = uint64(100)
	coin.CoinDetails.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.SNDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.Info = []byte("Incognito chain")

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number, coin commitment)
	coin.CoinDetails.PublicKey.Decompress(publicKey)
	coin.CoinDetails.SNDerivator = RandScalar()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = uint64(100)
	coin.CoinDetails.Info = []byte("Incognito chain")
	coin.CoinDetails.SerialNumber = nil
	coin.CoinDetails.CoinCommitment = nil

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.CoinDetails.PublicKey.Decompress(publicKey)
	coin.CoinDetails.SNDerivator = RandScalar()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = uint64(100)
	coin.CoinDetails.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.SNDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.Info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes) - 1] = byte(123)

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)
	//receivingKey := GenerateReceivingKey(privateKey)
	paymentAddr := GeneratePaymentAddress(privateKey)

	// init other fields for coin
	coin.CoinDetails.PublicKey.Decompress(publicKey)
	coin.CoinDetails.SNDerivator = RandScalar()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = uint64(100)
	coin.CoinDetails.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.SNDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.Info = []byte("Incognito chain")
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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number, coin commitment, CoinDetailsEncrypted)
	coin.CoinDetails.PublicKey.Decompress(publicKey)
	coin.CoinDetails.SNDerivator = RandScalar()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = uint64(100)
	coin.CoinDetails.Info = []byte("Incognito chain")
	coin.CoinDetails.SerialNumber = nil
	coin.CoinDetails.CoinCommitment = nil

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// init other fields for coin (exclude serial number)
	coin.CoinDetails.PublicKey.Decompress(publicKey)
	coin.CoinDetails.SNDerivator = RandScalar()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = uint64(100)
	coin.CoinDetails.SerialNumber = PedCom.G[0].Derive(new(big.Int).SetBytes(privateKey), coin.CoinDetails.SNDerivator)
	coin.CoinDetails.CommitAll()
	coin.CoinDetails.Info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes) - 1] = byte(123)

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
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	paymentAddress := GeneratePaymentAddress(privateKey)
	viewingKey := GenerateViewingKey(privateKey)

	for i := 0; i < 100; i++ {
		// new output coin with value and randomness
		coin := new(OutputCoin).Init()
		coin.CoinDetails.Randomness = RandScalar()
		coin.CoinDetails.Value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
		coin.CoinDetails.PublicKey.Decompress(paymentAddress.Pk)

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

		assert.Equal(t, coin.CoinDetails.Randomness, coin2.CoinDetails.Randomness)
		assert.Equal(t, coin.CoinDetails.Value, coin2.CoinDetails.Value)
	}
}

func TestOutputCoinEncryptDecryptWithUnmatchedKey(t *testing.T) {
	// prepare key
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	paymentAddress := GeneratePaymentAddress(privateKey)
	viewingKey := GenerateViewingKey(privateKey)

	// new output coin with value and randomness
	coin := new(OutputCoin).Init()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
	coin.CoinDetails.PublicKey.Decompress(paymentAddress.Pk)

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
	viewingKey.Rk[len(viewingKey.Rk) -1] = 123

	err3 := coin2.Decrypt(viewingKey)
	assert.Equal(t, (*PrivacyError)(nil), err3)
	assert.NotEqual(t, coin.CoinDetails.Randomness, coin2.CoinDetails.Randomness)
	assert.NotEqual(t, coin.CoinDetails.Value, coin2.CoinDetails.Value)
}

func TestOutputCoinEncryptWithInvalidKey(t *testing.T) {
	// prepare key
	seedKey := []byte{1,2,3}
	privateKey := GeneratePrivateKey(seedKey)
	publicKey := GeneratePublicKey(privateKey)

	// new output coin with value and randomness
	coin := new(OutputCoin).Init()
	coin.CoinDetails.Randomness = RandScalar()
	coin.CoinDetails.Value = new(big.Int).SetBytes(RandBytes(2)).Uint64()
	coin.CoinDetails.PublicKey.Decompress(publicKey)

	dataKey := [][]byte{
		{1,2,3},		// 3 bytes
		{16, 223, 34, 4, 35, 63, 73, 48, 69, 10, 11, 182, 183, 144, 150, 160, 17, 183, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33}, // 33 bytes, but not is on curve P256
	}

	for _, item := range dataKey {
		err := coin.Encrypt(item)
		assert.Equal(t, ErrCodeMessage[DecompressTransmissionKeyErr].code, err.GetCode())
	}
}

