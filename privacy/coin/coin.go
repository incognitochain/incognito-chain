package coin

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
)

type Coin interface {
	GetVersion() uint8
	GetCommitment() *operation.Point
	GetInfo() []byte
	GetPublicKey() *operation.Point
	GetKeyImage() *operation.Point
	GetValue() uint64
	GetRandomness() *operation.Scalar
	GetShardID() (uint8, error)
	GetSNDerivator() *operation.Scalar
	GetCoinDetailEncrypted() []byte
	IsEncrypted() bool
	GetTxRandom() *TxRandom
	GetSharedRandom() *operation.Scalar
	GetSharedConcealRandom() *operation.Scalar
	GetAssetTag() *operation.Point
	GetCoinID() [operation.Ed25519KeySize]byte
	GetOTATag() *uint8

	// DecryptOutputCoinByKey process outputcoin to get outputcoin data which relate to keyset
	// Param keyset: (private key, payment address, read only key)
	// in case private key: return unspent outputcoin tx
	// in case read only key: return all outputcoin tx with amount value
	// in case payment address: return all outputcoin tx with no amount value
	Decrypt(*incognitokey.KeySet) (PlainCoin, error)

	Bytes() []byte
	SetBytes([]byte) error

	CheckCoinValid(key.PaymentAddress, []byte, uint64) bool
	DoesCoinBelongToKeySet(keySet *incognitokey.KeySet) (bool, *operation.Point)
}

type PlainCoin interface {
	// Overide
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error

	GetVersion() uint8
	GetCommitment() *operation.Point
	GetInfo() []byte
	GetPublicKey() *operation.Point
	GetValue() uint64
	GetKeyImage() *operation.Point
	GetRandomness() *operation.Scalar
	GetShardID() (uint8, error)
	GetSNDerivator() *operation.Scalar
	GetCoinDetailEncrypted() []byte
	IsEncrypted() bool
	GetTxRandom() *TxRandom
	GetSharedRandom() *operation.Scalar
	GetSharedConcealRandom() *operation.Scalar
	GetAssetTag() *operation.Point
	GetOTATag() *uint8

	SetKeyImage(*operation.Point)
	SetPublicKey(*operation.Point)
	SetCommitment(*operation.Point)
	SetInfo([]byte)
	SetValue(uint64)
	SetRandomness(*operation.Scalar)

	// ParseKeyImage as Mlsag specification
	ParseKeyImageWithPrivateKey(key.PrivateKey) (*operation.Point, error)
	ParsePrivateKeyOfCoin(key.PrivateKey) (*operation.Scalar, error)

	ConcealOutputCoin(additionalData *operation.Point) error

	Bytes() []byte
	SetBytes([]byte) error
}

func NewPlainCoinFromByte(b []byte) (PlainCoin, error) {
	version := byte(CoinVersion2)
	if len(b) >= 1 {
		version = b[0]
	}
	var c PlainCoin
	if version == CoinVersion2 || version == CoinVersion3 {
		c = new(CoinV2)
	} else {
		c = new(PlainCoinV1)
	}
	err := c.SetBytes(b)
	return c, err
}

// First byte should determine the version or json marshal "34"
func NewCoinFromByte(b []byte) (Coin, error) {
	coinV1 := new(CoinV1)
	coinV2 := new(CoinV2)
	if errV2 := json.Unmarshal(b, coinV2); errV2 != nil {
		if errV1 := json.Unmarshal(b, coinV1); errV1 != nil {
			version := b[0]
			if version == CoinVersion2 {
				err := coinV2.SetBytes(b)
				return coinV2, err
			}
			err := coinV1.SetBytes(b)
			return coinV1, err
		}
		return coinV1, nil
	}
	return coinV2, nil
}

func ParseCoinsFromBytes(data []json.RawMessage) ([]Coin, error) {
	coinList := make([]Coin, len(data))
	for i := 0; i < len(data); i++ {
		coin, err := NewCoinFromByte(data[i])
		if err != nil {
			return nil, err
		}
		coinList[i] = coin
	}
	return coinList, nil
}
