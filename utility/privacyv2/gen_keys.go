package main

import (
	"fmt"
	"log"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	privateKey := "112t8roHikeAFyuBpdCU76kXurEqrC9VYWyRyfFb6PwX6nip9KGYbwpXL78H92mUoWK2GWkA2WysgXbHqwSxnC6XCkmtxBVb3zJeCXgfcYyL"
	paymentAddress := PrivateKeyToPaymentAddress(privateKey, -1)
	privateOTA := PrivateKeyToPrivateOTAKey(privateKey)
	fmt.Println(paymentAddress)
	fmt.Println(privateOTA)
}

// PrivateKeyToPaymentAddress returns the payment address for its private key corresponding to the key type.
// KeyType should be -1, 0, 1 where
//	- -1: payment address of version 2
//	- 0: payment address of version 1 with old encoding
//	- 1: payment address of version 1 with new encoding
func PrivateKeyToPaymentAddress(privateKey string, keyType int) string {
	keyWallet, _ := wallet.Base58CheckDeserialize(privateKey)
	err := keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		return ""
	}
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	switch keyType {
	case 0: //Old address, old encoding
		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, false)
		return addr
	case 1:
		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, true)
		return addr
	default:
		return paymentAddStr
	}
}

// PrivateKeyToPublicKey returns the public key of a private key.
//
// If the private key is invalid, it returns nil.
func PrivateKeyToPublicKey(privateKey string) []byte {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		return nil
	}
	return keyWallet.KeySet.PaymentAddress.Pk
}

// PrivateKeyToPrivateOTAKey returns the private OTA key of a private key.
//
// If the private key is invalid, it returns an empty string.
func PrivateKeyToPrivateOTAKey(privateKey string) string {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		log.Println("no private key found")
		return ""
	}
	return keyWallet.Base58CheckSerialize(wallet.OTAKeyType)
}

// PrivateKeyToReadonlyKey returns the readonly key of a private key.
//
// If the private key is invalid, it returns an empty string.
func PrivateKeyToReadonlyKey(privateKey string) string {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		log.Println("no private key found")
		return ""
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	return keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
}

// PrivateKeyToMiningKey returns the mining key of a private key.
func PrivateKeyToMiningKey(privateKey string) string {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return ""
	}
	miningKey := base58.Base58Check{}.Encode(common.HashB(common.HashB(keyWallet.KeySet.PrivateKey)), common.ZeroByte)
	return miningKey
}
