package coin

import (
	"errors"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"testing"

	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

func getRandomCoinV2() *CoinV2 {
	c := new(CoinV2)

	c.version = uint8(2)
	c.mask = operation.RandomScalar()
	c.amount = operation.RandomScalar()
	c.txRandom = operation.RandomPoint()
	c.publicKey = operation.RandomPoint()
	c.commitment = operation.RandomPoint()
	c.index = uint8(0)
	c.info = []byte{1, 2, 3, 4, 5}
	return c
}

func TestCoinV2BytesAndSetBytes(t *testing.T) {
	for i := 0; i < 5; i += 1 {
		coin := getRandomCoinV2()
		b := coin.Bytes()
		coinByBytes := new(CoinV2)
		err := coinByBytes.SetBytes(b)
		assert.Equal(t, nil, err, "Set Bytes should not have any error")
		assert.Equal(t, coin.version, coinByBytes.version, "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.mask.ToBytesS(), coinByBytes.mask.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.amount.ToBytesS(), coinByBytes.amount.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.txRandom.ToBytesS(), coinByBytes.txRandom.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.publicKey.ToBytesS(), coinByBytes.publicKey.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.commitment.ToBytesS(), coinByBytes.commitment.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.index, coinByBytes.index, "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.info, coinByBytes.info, "FromBytes then SetBytes should be equal")
	}
}

func createNewCoin(amount uint64, paymentAddress key.PaymentAddress, targetShardID uint8, index uint8) (*CoinV2, error) {
	if targetShardID >= common.MaxShardNumber {
		return nil, errors.New("Cannot create new coin with targetShardID, targetShardID is larger than max shard number")
	}
	c := new(CoinV2)
	c.SetVersion(2)
	c.SetIndex(index)
	c.SetInfo([]byte{})
	for true {
		// Mask and Amount will temporary visible by everyone, until after we done proving things, then will hide it.
		r := operation.RandomScalar()
		c.SetMask(r)
		c.SetAmount(new(operation.Scalar).FromUint64(amount))
		c.SetCommitment(operation.PedCom.CommitAtIndex(c.GetAmount(), r, operation.PedersenValueIndex))
		c.SetPublicKey(ParseOnetimeAddress(
			paymentAddress.GetPublicSpend(),
			paymentAddress.GetPublicView(),
			r,
			index,
		))
		c.SetTxRandom(new(operation.Point).ScalarMultBase(r)) // rG

		currentShardID, err := c.GetShardID()
		if err != nil {
			return nil, err
		}
		if currentShardID == targetShardID {
			break
		}
	}
	return c, nil
}

func TestCoinV2CreateCoinAndDecrypt(t *testing.T) {
	for i := 0; i < 20; i += 1 {
		privateKey := key.GeneratePrivateKey([]byte{1})
		keyset := new(incognitokey.KeySet)
		err := keyset.InitFromPrivateKey(&privateKey)
		assert.Equal(t, nil, err)

		r := common.RandBytes(8)
		val, errB := common.BytesToUint64(r)
		assert.Equal(t, nil, errB)

		c, errNewCoin := createNewCoin(val, keyset.PaymentAddress, 1, 0)
		assert.NotEqual(t, nil, errNewCoin)

		c, err = createNewCoin(val, keyset.PaymentAddress, 0, 0)
		assert.Equal(t, val, c.GetValue())
		assert.Equal(t, nil, err)
		assert.Equal(t, false, c.IsEncrypted())

		// Conceal
		c.ConcealData(keyset.PaymentAddress.GetPublicView())
		assert.Equal(t, true, c.IsEncrypted())
		assert.NotEqual(t, val, c.GetValue())

		var pc PlainCoin
		pc, err = c.Decrypt(keyset)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, pc.IsEncrypted())
		assert.Equal(t, val, c.GetValue())
	}
}
