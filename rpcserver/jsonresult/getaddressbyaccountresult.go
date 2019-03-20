package jsonresult

import "github.com/constant-money/constant-chain/wallet"

type GetAddressesByAccount struct {
	Addresses []wallet.KeySerializedData `json:"Addresses"`
}
