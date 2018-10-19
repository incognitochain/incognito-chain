package jsonresult

import "github.com/ninjadotorg/cash/wallet"

type GetAddressesByAccount struct {
	Addresses [] wallet.KeySerializedData `json:"Addresses"`
}
