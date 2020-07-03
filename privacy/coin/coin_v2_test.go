package coin

import (
	"testing"
	"github.com/incognitochain/incognito-chain/incognitokey"

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
	c.txRandom = NewTxRandom()
	c.publicKey = operation.RandomPoint()
	c.commitment = operation.RandomPoint()
	c.info = []byte{1, 2, 3, 4, 5}
	return c
}

func TestCoinV2BytesAndSetBytes(t *testing.T) {
	for i := 0; i < 5; i += 1 {
// test byte-marshalling of random plain coins
		coin := getRandomCoinV2()
		b := coin.Bytes()
		coinByBytes := new(CoinV2).Init()
		err := coinByBytes.SetBytes(b)
		assert.Equal(t, nil, err, "Set Bytes should not have any error")
		assert.Equal(t, coin.GetVersion(), coinByBytes.GetVersion(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.GetRandomness().ToBytesS(), coinByBytes.GetRandomness().ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.amount.ToBytesS(), coinByBytes.amount.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.GetTxRandomPoint().ToBytesS(), coinByBytes.GetTxRandomPoint().ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.publicKey.ToBytesS(), coinByBytes.publicKey.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.commitment.ToBytesS(), coinByBytes.commitment.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.GetIndex(), coinByBytes.GetIndex(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.info, coinByBytes.info, "FromBytes then SetBytes should be equal")

// test byte-marshalling of concealed coins
		privateKey := key.GeneratePrivateKey([]byte{byte(i)})
		keyset := new(incognitokey.KeySet)
		err = keyset.InitFromPrivateKey(&privateKey)
		assert.Equal(t, nil, err)
		paymentInfo := key.InitPaymentInfo(keyset.PaymentAddress, 3000, []byte{})
		coin,err = NewCoinFromPaymentInfo(paymentInfo)
		coin.ConcealData(keyset.PaymentAddress.GetPublicView())
		b = coin.Bytes()
		coinByBytes = new(CoinV2).Init()
		err = coinByBytes.SetBytes(b)
		assert.Equal(t, nil, err, "Set Bytes should not have any error")
		assert.Equal(t, coin.GetVersion(), coinByBytes.GetVersion(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.GetRandomness().ToBytesS(), coinByBytes.GetRandomness().ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.amount.ToBytesS(), coinByBytes.amount.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.GetTxRandomPoint().ToBytesS(), coinByBytes.GetTxRandomPoint().ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.publicKey.ToBytesS(), coinByBytes.publicKey.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.commitment.ToBytesS(), coinByBytes.commitment.ToBytesS(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.GetIndex(), coinByBytes.GetIndex(), "FromBytes then SetBytes should be equal")
		assert.Equal(t, coin.info, coinByBytes.info, "FromBytes then SetBytes should be equal")
	}
}

func TestCoinV2CreateCoinAndDecrypt(t *testing.T) {
	for i := 0; i < 20; i += 1 {
		privateKey := key.GeneratePrivateKey([]byte{byte(i)})
		keyset := new(incognitokey.KeySet)
		err := keyset.InitFromPrivateKey(&privateKey)
		assert.Equal(t, nil, err)

		r := common.RandBytes(8)
		val, errB := common.BytesToUint64(r)
		assert.Equal(t, nil, errB)

		paymentInfo := key.InitPaymentInfo(keyset.PaymentAddress, val, []byte{})

		c, err := NewCoinFromPaymentInfo(paymentInfo)
		assert.Equal(t, val, c.GetValue())
		assert.Equal(t, nil, err)
		assert.Equal(t, false, c.IsEncrypted())

		// Conceal
		c.ConcealData(keyset.PaymentAddress.GetPublicView())
		assert.Equal(t, true, c.IsEncrypted())
		assert.Equal(t, true, c.GetSharedRandom() == nil)
		assert.NotEqual(t, val, c.GetValue())

		// ensure tampered coins fail to decrypt
		testCoinV2ConcealedTampered(c,keyset,t)

		var pc PlainCoin
		pc, err = c.Decrypt(keyset)
		assert.Equal(t, nil, err)
		assert.Equal(t, false, pc.IsEncrypted())
		assert.Equal(t, val, c.GetValue())
	}
}

func testCoinV2ConcealedTampered(c *CoinV2, ks *incognitokey.KeySet, t *testing.T){
	saved := c.GetAmount()
	c.SetAmount(operation.RandomScalar())
	_, err := c.Decrypt(ks)
	assert.NotEqual(t, nil, err)
	
	// fmt.Println(err)
	c.SetAmount(saved)

	saved = c.GetRandomness()
	c.SetRandomness(operation.RandomScalar())
	_, err = c.Decrypt(ks)
	assert.NotEqual(t, nil, err)
	// fmt.Println(err)
	c.SetRandomness(saved)
}

func TestTxRandomGroup(t *testing.T) {
	for i := 0; i < 5; i += 1 {
		group := NewTxRandom()

		r := operation.RandomPoint()
		i := uint32(common.RandInt() & ((1 << 32) - 1))
		group.SetTxRandomPoint(r)
		group.SetIndex(i)
		assert.Equal(t, true, operation.IsPointEqual(group.GetTxRandomPoint(), r))
		assert.Equal(t, i, group.GetIndex())

		b := group.Bytes()
		var group2 TxRandom
		err := group2.SetBytes(b)
		assert.Equal(t, nil, err)
		assert.Equal(t, true, operation.IsPointEqual(group.GetTxRandomPoint(), group2.GetTxRandomPoint()))
		assert.Equal(t, i, group2.GetIndex())
		assert.Equal(t, group.GetTxRandomPoint(), group2.GetTxRandomPoint())
	}
}
