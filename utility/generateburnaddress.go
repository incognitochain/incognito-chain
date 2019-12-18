package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	for j := 0; j < 1000; j++ {
		for i := 0; ; i++ {
			burnPubKeyE := privacy.HashToPointFromIndex(int64(i), privacy.CStringBurnAddress)
			burnPubKey := burnPubKeyE.ToBytesS()
			if burnPubKey[len(burnPubKey) - 1] == 0 {
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

}

// result
// Burning payment address : 15pGE5XZc7gsHSLgQxgZRVt7UsXBodiyTYe6xtvXipTFQPMpdQqtXhkzriL at index 68
// ======================================
// Burning public key bytes: [127 76 149 36 97 166 59 24 204 39 108 209 42 199 106 173 88 95 221 184 142 215 198 51 10 150 125 89 73 86 24 0]
// ======================================
