package coin

import "github.com/incognitochain/incognito-chain/privacy/key"

type Coin interface {
	Init() *Coin
	GetVersion() uint8
	Bytes() []byte
	SetBytes([]byte) error
	GetPubKeyLastByte() byte
	GetCoinValue(key.PrivateKey) uint64
}
