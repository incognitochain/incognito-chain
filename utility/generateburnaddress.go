package main

import (
	"fmt"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
)

func main() {
	temp := 0
	for i := 0; ; i++ {
		burnPubKeyE := privacy.PedCom.G[0].Hash(uint64(i))
		burnPubKey := burnPubKeyE.Compress()
		if burnPubKey[len(burnPubKey)-1] == 0 {
			burnKey := wallet.KeyWallet{
				KeySet: cashec.KeySet{
					PaymentAddress: privacy.PaymentAddress{
						Pk: burnPubKey,
					},
				},
			}
			burnPaymentAddress := burnKey.Base58CheckSerialize(wallet.PaymentAddressType)
			fmt.Printf("Special payment address : %s %d\n", burnPaymentAddress, i)
			keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(burnPaymentAddress)
			fmt.Println("======================================")
			fmt.Println(keyWalletBurningAdd.KeySet.PaymentAddress.Pk)
			fmt.Println("======================================")
			temp += 1
			if temp == 3 {
				goto Out
			}
		}
	}
Out:
	fmt.Println("Finished")
}
