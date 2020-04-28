package coin

import (
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type Coin interface {
	GetVersion() uint8
	GetShardID() byte
	GetCommitment() *operation.Point
	GetInfo() []byte
	GetPublicKey() *operation.Point
	// GetCoinValue(key.PrivateKey) uint64
	GetKeyImage() *operation.Point
	GetSNDerivator() *operation.Scalar

	IsEncrypted() bool

	Bytes() []byte
	SetBytes([]byte) error
}

type PlainCoin interface {
	GetVersion() uint8
	GetShardID() uint8
	GetCommitment() *operation.Point
	GetInfo() []byte
	GetPublicKey() *operation.Point
	GetValue() uint64
	GetKeyImage() *operation.Point

	SetKeyImage(*operation.Point)
	SetPublicKey(*operation.Point)
	SetValue(uint64)
	SetInfo([]byte)

	IsEncrypted() bool
	ConcealData(additionalData interface{})

	Bytes() []byte
	SetBytes([]byte) error
}

// First byte should determine the version
func CreateCoinFromByte(b []byte) (Coin, error) {
	version := b[0]
	var c Coin
	if version == CoinVersion1 {
		c = new(CoinV1)
	} else if version == CoinVersion2 {
		c = new(CoinV2)
	}
	err := c.SetBytes(b)
	return c, err
}

func NewCoinFromVersion(version uint8) Coin {
	var c Coin
	if version == CoinVersion1 {
		pc := new(CoinV1)
		pc.Init()
		c = pc
	} else if version == CoinVersion2 {
		pc := new(CoinV2)
		pc.Init()
		c = pc
	}
	return c
}

func NewPlainCoinFromVersion(version uint8) PlainCoin {
	var c PlainCoin
	if version == CoinVersion1 {
		pc := new(PlainCoinV1)
		pc.Init()
		c = pc
	} else if version == CoinVersion2 {
		pc := new(CoinV2)
		pc.Init()
		c = pc
	}
	return c
}
