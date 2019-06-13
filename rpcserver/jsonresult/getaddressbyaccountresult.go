package jsonresult

import "github.com/incognitochain/incognito-chain/wallet"

type GetAddressesByAccount struct {
	Addresses []wallet.KeySerializedData `json:"Addresses"`
}
