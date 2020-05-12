package coin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"

	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

const (
	PedersenPrivateKeyIndex = operation.PedersenPrivateKeyIndex
	PedersenValueIndex      = operation.PedersenValueIndex
	PedersenSndIndex        = operation.PedersenSndIndex
	PedersenShardIDIndex    = operation.PedersenShardIDIndex
	PedersenRandomnessIndex = operation.PedersenRandomnessIndex
)

var PedCom operation.PedersenCommitment = operation.PedCom

func TestCoinV1CommitAllThenSwitchV2(t *testing.T) {
	coin := new(PlainCoinV1).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey, err := new(operation.Point).FromBytesS(key.GeneratePublicKey(privateKey))

	assert.Equal(t, nil, err)

	// init other fields for coin
	coin.SetPublicKey(publicKey)
	coin.SetSNDerivator(operation.RandomScalar())
	coin.SetRandomness(operation.RandomScalar())
	coin.SetValue(new(big.Int).SetBytes(common.RandBytes(2)).Uint64())
	coin.SetKeyImage(new(operation.Point).Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.GetSNDerivator()))
	coin.SetInfo([]byte("Incognito chain"))

	err = coin.CommitAll()
	assert.Equal(t, nil, err)

	allcm := coin.GetCommitment()
	cm := ParseCommitmentToV2WithCoin(coin)

	shardID, shardIDerr := coin.GetShardID()
	assert.Equal(t, nil, shardIDerr)

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
		coin := new(PlainCoinV1).Init()
		seedKey := operation.RandomScalar().ToBytesS()
		privateKey := key.GeneratePrivateKey(seedKey)
		publicKey := key.GeneratePublicKey(privateKey)

		// init other fields for coin
		coin.publicKey.FromBytesS(publicKey)

		coin.snDerivator = operation.RandomScalar()
		coin.randomness = operation.RandomScalar()
		coin.value = new(big.Int).SetBytes(common.RandBytes(2)).Uint64()
		coin.serialNumber = new(operation.Point).Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
		coin.CommitAll()
		coin.info = []byte("Incognito chain")

		cmTmp := coin.GetPublicKey()
		shardID, shardIDerr := coin.GetShardID()
		assert.Equal(t, nil, shardIDerr)

		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenValueIndex], new(operation.Scalar).FromUint64(uint64(coin.GetValue()))))
		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenSndIndex], coin.snDerivator))
		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenShardIDIndex], new(operation.Scalar).FromUint64(uint64(shardID))))
		cmTmp.Add(cmTmp, new(operation.Point).ScalarMult(PedCom.G[PedersenRandomnessIndex], coin.GetRandomness()))

		res := operation.IsPointEqual(cmTmp, coin.GetCommitment())
		assert.Equal(t, true, res)
	}
}

func TestCoinMarshalJSON(t *testing.T) {

	for i := 0; i < 3; i++ {
		coin := new(PlainCoinV1).Init()
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

		coin2 := new(PlainCoinV1)
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
		coin := new(PlainCoinV1).Init()
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
		coin2 := new(PlainCoinV1)
		err := coin2.SetBytes(coinBytes)

		assert.Equal(t, nil, err)
		assert.Equal(t, coin, coin2)
	}
}

func TestCoinBytesSetBytesWithMissingFields(t *testing.T) {
	for i := 0; i < 3; i++ {
		coin := new(PlainCoinV1).Init()
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
		coin2 := new(PlainCoinV1).Init()
		err := coin2.SetBytes(coinBytes)

		assert.Equal(t, nil, err)
		assert.Equal(t, coin, coin2)
	}
}

func TestCoinBytesSetBytesWithInvalidBytes(t *testing.T) {
	// init coin with fully fields
	// init public key
	coin := new(PlainCoinV1).Init()
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
	coin2 := new(PlainCoinV1).Init()
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
		coin := new(PlainCoinV1).Init()
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
		coin2 := new(PlainCoinV1)
		err := coin2.SetBytes(coinBytes)

		assert.Equal(t, nil, err)
		assert.Equal(t, coin, coin2)
	}
}

func TestInputCoinBytesSetBytesWithMissingFields(t *testing.T) {
	coin := new(PlainCoinV1).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)

	coin.publicKey.FromBytesS(publicKey)

	coin.snDerivator = operation.RandomScalar()
	coin.randomness = operation.RandomScalar()
	coin.value = uint64(100)
	coin.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
	coin.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// new coin object and set bytes from bytes array
	coin2 := new(PlainCoinV1).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestInputCoinBytesSetBytesWithInvalidBytes(t *testing.T) {
	coin := new(PlainCoinV1).Init()
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)
	publicKey := key.GeneratePublicKey(privateKey)

	coin.publicKey.FromBytesS(publicKey)

	coin.snDerivator = operation.RandomScalar()
	coin.randomness = operation.RandomScalar()
	coin.value = uint64(100)
	coin.serialNumber = PedCom.G[0].Derive(PedCom.G[0], new(operation.Scalar).FromBytesS(privateKey), coin.snDerivator)
	coin.info = []byte("Incognito chain")

	// convert coin object to bytes array
	coinBytes := coin.Bytes()
	assert.Greater(t, len(coinBytes), 0)

	// edit coinBytes
	coinBytes[len(coinBytes)-2] = byte(12)

	// new coin object and set bytes from bytes array
	coin2 := new(PlainCoinV1).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, coin, coin2)
}

func TestInputCoinBytesSetBytesWithEmptyBytes(t *testing.T) {
	// new coin object and set bytes from bytes array
	coin2 := new(PlainCoinV1).Init()
	err := coin2.SetBytes([]byte{})
	assert.Equal(t, errors.New("coinBytes is empty"), err)
}

/*
	Unit test for Bytes/SetBytes OutputCoin function
*/
func TestOutputCoinBytesSetBytes(t *testing.T) {
	coin := new(CoinV1).Init()
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
	coin2 := new(CoinV1)
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestOutputCoinBytesSetBytesWithMissingFields(t *testing.T) {
	coin := new(CoinV1).Init()
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
	coin2 := new(CoinV1).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.Equal(t, coin, coin2)
}

func TestOutputCoinBytesSetBytesWithInvalidBytes(t *testing.T) {
	coin := new(CoinV1).Init()
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
	coin2 := new(CoinV1).Init()
	err := coin2.SetBytes(coinBytes)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, coin, coin2)
}

func TestOutputCoinBytesSetBytesWithEmptyBytes(t *testing.T) {
	// new coin object and set bytes from bytes array
	coin2 := new(CoinV1).Init()
	err := coin2.SetBytes([]byte{})

	assert.Equal(t, errors.New("coinBytes is empty"), err)
}

func debugInterface(a interface{}) {
	d, _ := json.Marshal(a)
	fmt.Println(string(d))
}

/*
	Unit test for Encrypt/Decrypt OutputCoin
*/
func TestOutputCoinEncryptDecrypt(t *testing.T) {
	// prepare key
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)

	keySet := new(incognitokey.KeySet)
	err := keySet.InitFromPrivateKey(&privateKey)
	assert.Equal(t, nil, err)

	paymentAddress := key.GeneratePaymentAddress(privateKey)

	for i := 0; i < 3; i++ {
		// new output coin with value and randomness
		coin := new(CoinV1).Init()
		coin.CoinDetails.randomness = operation.RandomScalar()
		coin.CoinDetails.value = new(big.Int).SetBytes(common.RandBytes(2)).Uint64()
		coin.CoinDetails.publicKey.FromBytesS(paymentAddress.Pk)

		// encrypt output coins
		err := coin.Encrypt(paymentAddress.Tk)
		assert.Equal(t, (*errhandler.PrivacyError)(nil), err)

		// convert output coin to bytes array
		coinBytes := coin.Bytes()

		// create new output coin to test
		coin2 := new(CoinV1)
		err2 := coin2.SetBytes(coinBytes)
		assert.Equal(t, nil, err2)

		decrypted, err3 := coin2.Decrypt(keySet)
		assert.Equal(t, nil, err3)

		assert.Equal(t, coin.CoinDetails.randomness, decrypted.GetRandomness())
		assert.Equal(t, coin.CoinDetails.value, decrypted.GetValue())
	}
}

func TestOutputCoinEncryptDecryptWithUnmatchedKey(t *testing.T) {
	// prepare key
	seedKey := operation.RandomScalar().ToBytesS()
	privateKey := key.GeneratePrivateKey(seedKey)

	keySet := new(incognitokey.KeySet)
	err := keySet.InitFromPrivateKey(&privateKey)
	assert.Equal(t, nil, err)

	paymentAddress := key.GeneratePaymentAddress(privateKey)

	// new output coin with value and randomness
	coin := new(CoinV1).Init()
	coin.CoinDetails.randomness = operation.RandomScalar()
	coin.CoinDetails.value = new(big.Int).SetBytes(common.RandBytes(2)).Uint64()
	coin.CoinDetails.publicKey.FromBytesS(paymentAddress.Pk)

	// encrypt output coins
	err = coin.Encrypt(paymentAddress.Tk)
	assert.Equal(t, (*errhandler.PrivacyError)(nil), err)

	// convert output coin to bytes array
	coinBytes := coin.Bytes()

	// create new output coin to test
	coin2 := new(CoinV1)
	err2 := coin2.SetBytes(coinBytes)
	assert.Equal(t, nil, err2)

	// edit receiving key to be unmatched with transmission key
	keySet.ReadonlyKey.Rk[0] = 12
	decrypted, err3 := coin2.Decrypt(keySet)
	assert.Equal(t, nil, err3)
	assert.NotEqual(t, coin.CoinDetails.randomness, decrypted.GetRandomness())
	assert.NotEqual(t, coin.CoinDetails.value, decrypted.GetValue())
}
