package main

import (
	"fmt"
	"os"

	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	privateKey := os.Args[1]
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		panic(err)
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		panic("no private key found")
	}
	fmt.Println("ota", keyWallet.Base58CheckSerialize(wallet.OTAKeyType))
	fmt.Println("pri", keyWallet.Base58CheckSerialize(wallet.PriKeyType))
	fmt.Println("paymentaddress", keyWallet.Base58CheckSerialize(wallet.PaymentAddressType))
	fmt.Println("readonly key", keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType))
}
