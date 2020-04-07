package coin

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/stretchr/testify/assert"
)

func TestCoinV1CommitAllThenSwitchV2(t *testing.T) {
	coin := new(CoinV1).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.publicKey.FromBytesS(publicKey)
	coin.snDerivator = operation.RandomScalar()
	coin.randomness = operation.RandomScalar()
	coin.value = new(big.Int).SetBytes(privacy_util.RandBytes(2)).Uint64()
	coin.serialNumber = new(operation.Point).Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
	coin.info = []byte("Incognito chain")

	err := coin.CommitAll()
	assert.Equal(t, nil, err)

	allcm := coin.GetCoinCommitment()
	cm := ParseCommitmentToV2WithCoin(coin)

	shardID := common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())
	allcm = ParseCommitmentToV2(
		allcm,
		coin.GetPublicKey(),
		coin.GetSNDerivator(),
		shardID,
	)

	b1 := allcm.ToBytesS()
	b2 := cm.ToBytesS()
	assert.Equal(t, true, bytes.Equal(b1, b2))
}

func TestCoinV1CommitAll(t *testing.T) {
	for i := 0; i < 3; i++ {
		coin := new(CoinV1).Init()
		seedKey := operation.RandomScalar().ToBytesS()
		privateKey := key.GeneratePrivateKey(seedKey)
		publicKey := key.GeneratePublicKey(privateKey)

		// init other fields for coin
		coin.publicKey.FromBytesS(publicKey)

		coin.snDerivator = operation.RandomScalar()
		coin.randomness = operation.RandomScalar()
		coin.value = new(big.Int).SetBytes(privacy_util.RandBytes(2)).Uint64()
		coin.serialNumber = new(operation.Point).Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
		coin.CommitAll()
		coin.info = []byte("Incognito chain")

		cmTmp := coin.GetPublicKey()
		shardID := common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())
		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenValueIndex], new(operation.Scalar).FromUint64(uint64(coin.GetValue()))))
		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenSndIndex], coin.snDerivator))
		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenShardIDIndex], new(operation.Scalar).FromUint64(uint64(shardID))))
		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenRandomnessIndex], coin.GetRandomness()))

		res := operation.IsPointEqual(cmTmp, coin.GetCoinCommitment())
		assert.Equal(t, true, res)
	}
}

func TestCoinMarshalJSON(t *testing.T) {

	for i := 0; i < 3; i++ {
		coin := new(CoinV1).Init()
		seedKey := operation.RandomScalar().ToBytesS()
		privateKey := key.GeneratePrivateKey(seedKey)
		publicKey := key.GeneratePublicKey(privateKey)

		// init other fields for coin
		coin.publicKey.FromBytesS(publicKey)
		coin.snDerivator = operation.RandomScalar()
		coin.randomness = operation.RandomScalar()
		coin.value = uint64(100)
		coin.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
		coin.CommitAll()
		coin.info = []byte("Incognito chain")

		bytesJSON, err := coin.MarshalJSON()
		assert.Equal(t, nil, err)

		coin2 := new(CoinV1)
		err2 := coin2.UnmarshalJSON(bytesJSON)
		assert.Equal(t, nil, err2)
		assert.Equal(t, coin, coin2)
	}
}

/*
	Unit test for Bytes/SetBytes Coin function
*/

func TestCoinBytesSetBytes(t *testing.T) {

	for i := 0; i < 3; i++ {
		coin := new(CoinV1).Init()
		seedKey := operation.RandomScalar().ToBytesS()
		privateKey := key.GeneratePrivateKey(seedKey)
		publicKey := key.GeneratePublicKey(privateKey)

		// init other fields for coin
		coin.publicKey.FromBytesS(publicKey)
		coin.snDerivator = operation.RandomScalar()
		coin.randomness = operation.RandomScalar()
		coin.value = uint64(100)
		coin.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
		coin.CommitAll()
		coin.info = []byte("Incognito chain")

		// convert coin object to bytes array
		coinBytes := coin.Bytes()

		assert.Greater(t, len(coinBytes), 0)

		// new coin object and set bytes from bytes array
		coin2 := new(CoinV1)
		err := coin2.SetBytes(coinBytes)

		assert.Equal(t, nil, err)
		assert.Equal(t, coin, coin2)
	}
}

func TestCoinBytesSetBytesWithMissingFields(t *testing.T) {
	for i := 0; i < 3; i++ {
		coin := new(CoinV1).Init()
		seedKey := operation.RandomScalar().ToBytesS()
		privateKey := key.GeneratePrivateKey(seedKey)
		publicKey := key.GeneratePublicKey(privateKey)

		// init other fields for coin
		coin.publicKey.FromBytesS(publicKey)
		coin.snDerivator = operation.RandomScalar()
		coin.randomness = operation.RandomScalar()
		coin.value = uint64(100)
		coin.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
		//coin.CommitAll()
		coin.info = []byte("Incognito chain")

		// convert coin object to bytes array
		coinBytes := coin.Bytes()

		assert.Greater(t, len(coinBytes), 0)

		// new coin object and set bytes from bytes array
		coin2 := new(CoinV1).Init()
		err := coin2.SetBytes(coinBytes)

		assert.Equal(t, nil, err)
		assert.Equal(t, coin, coin2)
	}
}

func TestCoinBytesSetBytesWithInvalidBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(CoinV1).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)

	// init other fields for coin
	coin.publicKey.FromBytesS(publicKey)
	coin.snDerivator = operation.RandomScalar()
	coin.randomness = operation.RandomScalar()
	coin.value = uint64(100)
	coin.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
	coin.CommitAll()
	coin.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes)-2] = byte(12)

	// new coin object and set bytes from bytes array
	coin2 := new(CoinV1).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, coin, coin2)
}

func TestCoinBytesSetBytesWithEmptyBytes(t *testing.T) {
	// new coin object and set bytes from bytes array
	coin2 := new(CoinV1).Init()
	err := coin2.SetBytes([]byte{})

	assert.Equal(t, errors.New("coinBytes is empty"), err)
}

/*
	Unit test for Bytes/SetBytes InputCoin function
*/

func TestInputCoinBytesSetBytes(t *testing.T) {
	for i := 0; i < 3; i++ {
		coin := new(InputCoin).Init()
		seedKey := operation.RandomScalar().ToBytesS()
		privateKey := key.GeneratePrivateKey(seedKey)
		publicKey := key.GeneratePublicKey(privateKey)

		// init other fields for coin
		coin.CoinDetails.publicKey.FromBytesS(publicKey)

		coin.CoinDetails.snDerivator = operation.RandomScalar()
		coin.CoinDetails.randomness = operation.RandomScalar()
		coin.CoinDetails.value = uint64(100)
		coin.CoinDetails.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.CoinDetails.snDerivator)
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
}

func TestInputCoinBytesSetBytesWithMissingFields(t *testing.T) {
	coin := new(InputCoin).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)

	coin.CoinDetails.publicKey.FromBytesS(publicKey)

	coin.CoinDetails.snDerivator = operation.RandomScalar()
	coin.CoinDetails.randomness = operation.RandomScalar()
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.CoinDetails.snDerivator)
	//coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")

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
	coin := new(InputCoin).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)

	coin.CoinDetails.publicKey.FromBytesS(publicKey)

	coin.CoinDetails.snDerivator = operation.RandomScalar()
	coin.CoinDetails.randomness = operation.RandomScalar()
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.CoinDetails.snDerivator)
	//coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes)-2] = byte(12)

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
	coin := new(OutputCoin).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)
	paymentAddr := key.GeneratePaymentAddress(privateKey)

	coin.CoinDetails.publicKey.FromBytesS(publicKey)

	coin.CoinDetails.snDerivator = operation.RandomScalar()
	coin.CoinDetails.randomness = operation.RandomScalar()
	coin.CoinDetails.value = uint64(100)
	coin.CoinDetails.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.CoinDetails.snDerivator)
	//coin.CoinDetails.CommitAll()
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
	coin := new(OutputCoin).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)
	paymentAddr := key.GeneratePaymentAddress(privateKey)

	coin.CoinDetails.publicKey.FromBytesS(publicKey)

	coin.CoinDetails.snDerivator = operation.RandomScalar()
	coin.CoinDetails.randomness = operation.RandomScalar()
	coin.CoinDetails.value = uint64(100)
	//coin.CoinDetails.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(Scalar).FromBytes(SliceToArray(privateKey)), coin.CoinDetails.snDerivator)
	//coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")
	coin.Encrypt(paymentAddr.Tk)

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
	coin := new(OutputCoin).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)
	paymentAddr := key.GeneratePaymentAddress(privateKey)

	coin.CoinDetails.publicKey.FromBytesS(publicKey)

	coin.CoinDetails.snDerivator = operation.RandomScalar()
	coin.CoinDetails.randomness = operation.RandomScalar()
	coin.CoinDetails.value = uint64(100)
	//coin.CoinDetails.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(Scalar).FromBytes(SliceToArray(privateKey)), coin.CoinDetails.snDerivator)
	//coin.CoinDetails.CommitAll()
	coin.CoinDetails.info = []byte("Incognito chain")
	coin.Encrypt(paymentAddr.Tk)

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes)-2] = byte(12)

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
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	paymentAddress := key.GeneratePaymentAddress(privateKey)
	viewingKey := key.GenerateViewingKey(privateKey)

	for i := 0; i < 100; i++ {
		// new output coin with value and randomness
		coin := new(OutputCoin).Init()
		coin.CoinDetails.randomness = operation.RandomScalar()
		coin.CoinDetails.value = new(big.Int).SetBytes(privacy_util.RandBytes(2)).Uint64()
		coin.CoinDetails.publicKey.FromBytesS(paymentAddress.Pk)

		// encrypt output coins
		err := coin.Encrypt(paymentAddress.Tk)
		assert.Equal(t, (*errhandler.PrivacyError)(nil), err)

		// convert output coin to bytes array
		coinBytes := coin.Bytes()

		// create new output coin to test
		coin2 := new(OutputCoin)
		err2 := coin2.SetBytes(coinBytes)
		assert.Equal(t, nil, err2)

		err3 := coin2.Decrypt(viewingKey)
		assert.Equal(t, (*errhandler.PrivacyError)(nil), err3)

		assert.Equal(t, coin.CoinDetails.randomness, coin2.CoinDetails.randomness)
		assert.Equal(t, coin.CoinDetails.value, coin2.CoinDetails.value)
	}
}

func TestOutputCoinEncryptDecryptWithUnmatchedKey(t *testing.T) {
	// prepare key
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	paymentAddress := key.GeneratePaymentAddress(privateKey)
	viewingKey := key.GenerateViewingKey(privateKey)

	// new output coin with value and randomness
	coin := new(OutputCoin).Init()
	coin.CoinDetails.randomness = operation.RandomScalar()
	coin.CoinDetails.value = new(big.Int).SetBytes(privacy_util.RandBytes(2)).Uint64()
	coin.CoinDetails.publicKey.FromBytesS(paymentAddress.Pk)

	// encrypt output coins
	err := coin.Encrypt(paymentAddress.Tk)
	assert.Equal(t, (*errhandler.PrivacyError)(nil), err)

	// convert output coin to bytes array
	coinBytes := coin.Bytes()

	// create new output coin to test
	coin2 := new(OutputCoin)
	err2 := coin2.SetBytes(coinBytes)
	assert.Equal(t, nil, err2)

	// edit receiving key to be unmatched with transmission key
	viewingKey.Rk[0] = 12
	err3 := coin2.Decrypt(viewingKey)
	assert.Equal(t, (*errhandler.PrivacyError)(nil), err3)
	assert.NotEqual(t, coin.CoinDetails.randomness, coin2.CoinDetails.randomness)
	assert.NotEqual(t, coin.CoinDetails.value, coin2.CoinDetails.value)
}
