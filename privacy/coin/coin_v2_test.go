package coin

import (
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
	"testing"
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

func parseCoinBasedOnPaymentInfo(info *privacy.PaymentInfo, targetShardID byte, index uint8) (*coin.CoinV2, error) {
	c := new(coin.CoinV2)
	c.SetVersion(2)
	c.SetIndex(index)
	c.SetInfo(info.Message)

	for true {
		// Mask and Amount will temporary visible by everyone, until after we done proving things, then will hide it.
		r := operation.RandomScalar()
		c.SetMask(r)
		c.SetAmount(new(operation.Scalar).FromUint64(info.Amount))
		c.SetCommitment(operation.PedCom.CommitAtIndex(c.GetAmount(), r, operation.PedersenValueIndex))
		c.SetPublicKey(coin.ParseOnetimeAddress(
			info.PaymentAddress.GetPublicSpend(),
			info.PaymentAddress.GetPublicView(),
			r,
			index,
		))
		c.SetTxRandom(new(operation.Point).ScalarMultBase(r)) // rG

		currentShardID, err := c.GetShardID()
		if err != nil {
			Logger.Log.Errorf("Cannot get shardID of newly created coin with err %v", err)
			return nil, err
		}
		if currentShardID != targetShardID {
			continue
		} else {
			break
		}
	}

	return c, nil
}

func TestCoinV2()