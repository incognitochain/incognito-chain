package monero_address

import (
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

const KeyLength = C25519.KeyLength

type MoneroAddress struct {
	PrivateSpend C25519.Key
	PrivateView  C25519.Key
}

func GeneratePrivateKey(seed []byte) C25519.Key {
	bip32PrivateKey := C25519.HashToScalar(seed)
	privateKey := bip32PrivateKey.ToBytes()
	return privateKey
}

// Get Public Key from Private Key
func GetPublicKey(privateKey *C25519.Key) C25519.Key {
	publicKey := C25519.ScalarmultBase(privateKey)
	return publicKey.ToBytes()
}

func (p *MoneroAddress) GetPrivateSpend() C25519.Key {
	return p.PrivateSpend
}

func (p *MoneroAddress) GetPrivateView() C25519.Key {
	return p.PrivateView
}

func (p *MoneroAddress) GetPublicSpend() C25519.Key {
	return GetPublicKey(&p.PrivateSpend)
}

func (p *MoneroAddress) GetPublicView() C25519.Key {
	return GetPublicKey(&p.PrivateView)
}

func GenerateRandomAddress() *MoneroAddress {
	privateSpend := C25519.RandomScalar()
	privateView := C25519.RandomScalar()

	res := MoneroAddress{*privateSpend, *privateView}
	return &res
}
