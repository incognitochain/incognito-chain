package jsonresult

import "github.com/big0t/constant-chain/wallet"

type GetAddressesByAccount struct {
	Addresses []wallet.KeySerializedData `json:"Addresses"`
}
