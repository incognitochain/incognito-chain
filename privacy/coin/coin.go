package coin

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type Coin interface {
	GetVersion() uint8
	GetShardID() (uint8, error)
	GetCommitment() *operation.Point
	GetInfo() []byte
	GetPublicKey() *operation.Point
	GetKeyImage() *operation.Point
	GetSNDerivator() *operation.Scalar
	GetValue() uint64
	GetRandomness() *operation.Scalar
	GetIndex() uint32
	IsEncrypted() bool

	// DecryptOutputCoinByKey process outputcoin to get outputcoin data which relate to keyset
	// Param keyset: (private key, payment address, read only key)
	// in case private key: return unspent outputcoin tx
	// in case read only key: return all outputcoin tx with amount value
	// in case payment address: return all outputcoin tx with no amount value
	Decrypt(*incognitokey.KeySet) (PlainCoin, error)

	Bytes() []byte
	SetBytes([]byte) error

	CheckCoinValid(key.PaymentAddress, []byte, uint64) bool
}

type PlainCoin interface {
	// Overide
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error

	GetVersion() uint8
	GetShardID() (uint8, error)
	GetIndex() uint32
	GetCommitment() *operation.Point
	GetInfo() []byte
	GetPublicKey() *operation.Point
	GetValue() uint64
	GetKeyImage() *operation.Point
	GetRandomness() *operation.Scalar
	GetSNDerivator() *operation.Scalar

	SetKeyImage(*operation.Point)
	SetPublicKey(*operation.Point)
	SetCommitment(*operation.Point)
	SetInfo([]byte)
	SetValue(uint64)
	SetRandomness(*operation.Scalar)

	// ParseKeyImage as Mlsag specification
	ParseKeyImageWithPrivateKey(key.PrivateKey) (*operation.Point, error)
	ParsePrivateKeyOfCoin(key.PrivateKey) (*operation.Scalar, error)

	IsEncrypted() bool
	ConcealData(additionalData interface{})

	Bytes() []byte
	SetBytes([]byte) error
}

func NewPlainCoinFromByte(b []byte) (PlainCoin, error) {
	version := b[0]
	var c PlainCoin
	if version == CoinVersion2 {
		c = new(CoinV2)
	} else {
		c = new(PlainCoinV1)
	}
	err := c.SetBytes(b)
	return c, err
}

// First byte should determine the version
func NewCoinFromByte(b []byte) (Coin, error) {
	version := b[0]

	if version == CoinVersion2 {
		c := new(CoinV2)
		err := c.SetBytes(b)
		//fmt.Println(string(b))
		//fmt.Println(string(b))
		//fmt.Println(string(b))
		//fmt.Println(err)
		return c, err
	} else {
		//fmt.Println(b)
		//fmt.Println(b)
		//fmt.Println(b)
		//fmt.Println(string(b))
		//panic("OK")
		c := new(CoinV1)
		if err := json.Unmarshal(b, &c); err != nil {
			err = c.SetBytes(b)
			return c, err
		}
		return c, nil
	}

}

// Check whether the utxo is from this address
func IsCoinBelongToViewKey(coin Coin, viewKey key.ViewingKey) bool {
	if coin.GetVersion() == 1 {
		return operation.IsPointEqual(viewKey.GetPublicSpend(), coin.GetPublicKey())
	} else if coin.GetVersion() == 2 {
		c, err := coin.(*CoinV2)
		if err == false {
			return false
		}
		rK := new(operation.Point).ScalarMult(c.GetTxRandomPoint(), viewKey.GetPrivateView())

		hashed := operation.HashToScalar(
			append(rK.ToBytesS(), common.Uint32ToBytes(c.GetIndex())...),
		)
		HnG := new(operation.Point).ScalarMultBase(hashed)
		KCheck := new(operation.Point).Sub(c.GetPublicKey(), HnG)

		return operation.IsPointEqual(KCheck, viewKey.GetPublicSpend())
	} else {
		return false
	}
}

func ParseCoinsFromBytes(data []json.RawMessage) ([]Coin, error) {
	coinList := make([]Coin, len(data))
	for i := 0; i < len(data); i++ {
		if coin, err := NewCoinFromByte(data[i]); err != nil {
			return nil, err
		} else {
			coinList[i] = coin
		}
	}
	return coinList, nil
}
