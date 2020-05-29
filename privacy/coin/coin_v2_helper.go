package coin

import (
	"fmt"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/wallet"
)

func NewCoinFromAmountAndReceiver(amount uint64, receiver key.PaymentAddress) (*CoinV2, error) {
	fmt.Println("Creating coins where amount =", amount, "and publickey =", receiver.Pk)
	paymentInfo := key.InitPaymentInfo(receiver, amount, []byte{})
	return NewCoinFromPaymentInfo(paymentInfo)
}

// This function create new coinv2 that has same shardID with targetShardID and with info of paymentInfo
func NewCoinFromPaymentInfo(info *key.PaymentInfo) (*CoinV2, error) {
	receiverPublicKey, err := new(operation.Point).FromBytesS(info.PaymentAddress.Pk)
	if err != nil {
		errStr := fmt.Sprintf("Cannot parse outputCoinV2 from PaymentInfo when parseByte PublicKey, error %v ", err)
		return nil, errors.New(errStr)
	}
	receiverPublicKeyBytes := receiverPublicKey.ToBytesS()
	targetShardID := common.GetShardIDFromLastByte(receiverPublicKeyBytes[len(receiverPublicKeyBytes)-1])

	c := new(CoinV2).Init()
	// Amount, Randomness, SharedRandom is transparency until we call concealData
	c.SetAmount(new(operation.Scalar).FromUint64(info.Amount))
	c.SetRandomness(operation.RandomScalar())
	c.SetSharedRandom(operation.RandomScalar()) // r
	c.SetTxRandomPoint(new(operation.Point).ScalarMultBase(c.GetSharedRandom())) // rG
	c.SetInfo(info.Message)
	c.SetCommitment(operation.PedCom.CommitAtIndex(c.GetAmount(), c.GetRandomness(), operation.PedersenValueIndex))

	// If this is going to burning address then dont need to create ota
	if wallet.IsPublicKeyBurningAddress(info.PaymentAddress.Pk) {
		c.SetIndex(0)
		publicKey, err := new(operation.Point).FromBytesS(info.PaymentAddress.Pk)
		if err != nil {
			panic("Something is wrong with info.paymentAddress.Pk, burning address should be a valid point")
		}
		c.SetPublicKey(publicKey)
		return c, nil
	}

	// Increase index until have the right shardID
	index := uint32(0)
	publicView := info.PaymentAddress.GetPublicView()
	publicSpend := info.PaymentAddress.GetPublicSpend()
	rK := new(operation.Point).ScalarMult(publicView, c.GetSharedRandom())
	for {
		index += 1
		c.SetIndex(index)

		// Get publickey
		hash := operation.HashToScalar(append(rK.ToBytesS(), common.Uint32ToBytes(index)...))
		HrKG := new(operation.Point).ScalarMultBase(hash)
		publicKey := new(operation.Point).Add(HrKG, publicSpend)
		c.SetPublicKey(publicKey)

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

func CoinV2ArrayToCoinArray(coinArray []*CoinV2) []Coin {
	res := make([]Coin, len(coinArray))
	for i := 0; i < len(coinArray); i += 1 {
		res[i] = coinArray[i]
	}
	return res
}