package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	for i := 0; ; i++ {
		burnPubKeyE := privacy.HashToPointFromIndex(int64(i), privacy.CStringBurnAddress)
		burnPubKey := burnPubKeyE.ToBytesS()
		if burnPubKey[len(burnPubKey)-1] == 0 {
			burnKey := wallet.KeyWallet{
				KeySet: incognitokey.KeySet{
					PaymentAddress: privacy.PaymentAddress{
						Pk: burnPubKey,
					},
				},
			}

			burnPaymentAddress := burnKey.Base58CheckSerialize(wallet.PaymentAddressType)
			fmt.Printf("Burning payment address : %s at index %d\n", burnPaymentAddress, i)

			fmt.Println("======================================")
			fmt.Printf("Burning public key bytes: %v\n", burnKey.KeySet.PaymentAddress.Pk)
			fmt.Println("======================================")
			break
		}
	}
}

// result
