package jsonresult

import "github.com/ninjadotorg/constant/wallet"

type GetAddressesByAccount struct {
	Addresses [] wallet.KeySerializedData `json:"Addresses"`
}
