package mlsag

import (
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

type MlsagSignature struct {
	c         C25519.Key
	r         [][]C25519.Key
	keyImages []C25519.Key
}
