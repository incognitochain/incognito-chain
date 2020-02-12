package payment

import (
	Address "github.com/incognitochain/incognito-chain/anan/monero_address"
	C25519 "github.com/incognitochain/incognito-chain/privacy/curve25519"
)

type Payment struct {
	from  Address.MoneroAddress
	to    C25519.Key
	money int
}
