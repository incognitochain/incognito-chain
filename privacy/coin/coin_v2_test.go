package coin

import (
	"testing"

	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
)

func getRandomCoinV2() *Coin_v2 {
	c := new(Coin_v2)

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
		coinByBytes := new(Coin_v2)
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
