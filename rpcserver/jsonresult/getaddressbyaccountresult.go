package jsonresult

import "github.com/ninjadotorg/cash-prototype/wallet"

type GetAddressesByAccount struct {
	Addresses [] wallet.KeySerializedData `json:"Addresses"`
}
