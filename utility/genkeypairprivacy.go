package main

import (
	"fmt"
	"github.com/constant-money/constant-chain/common/base58"

	"github.com/constant-money/constant-chain/wallet"
)

func main() {
	mnemonicGen := wallet.MnemonicGenerator{}
	Entropy, _ := mnemonicGen.NewEntropy(128)
	Mnemonic, _ := mnemonicGen.NewMnemonic(Entropy)
	fmt.Printf("Mnemonic: %s\n", Mnemonic)
	Seed := mnemonicGen.NewSeed(Mnemonic, "123456")

	key, _ := wallet.NewMasterKey(Seed)
	var i int
	var k = 0
	for {
		//i++
		child, _ := key.NewChildKey(uint32(i))
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
		if true || child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1] == 0 {
			fmt.Printf("Acc %d:\n ", i)
			fmt.Printf("paymentAddress: %v\n", paymentAddress)
			fmt.Printf("spending key: %v\n", privAddr)
			fmt.Printf("pubkey: %v\n", base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, byte(0x00)))
			k++
			if k == 20 {
				break
			}
			//}
			i++
		}

		//112t8rqnMrtPkJ4YWzXfG82pd9vCe2jvWGxqwniPM5y4hnimki6LcVNfXxN911ViJS8arTozjH4rTpfaGo5i1KKcG1ayjiMsa4E3nABGAqQh
		//keyWallet, _ := wallet.Base58CheckDeserialize("112t8s2UkZEwS7JtqLHFruRrh4Drj53UzH4A6DrairctKutxVb8Vw2DMzxCReYsAZkXi9ycaSNRHEcB7TJaTwPhyPvqRzu5NnUgTMN9AEKwo")
		//keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)
		//fmt.Println("Pub-key byte array", keyWallet.KeySet.PaymentAddress.Pk)
		//fmt.Println(base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, byte(0x00)))
		//fmt.Printf("Pub-key : %+v \n", keyWallet.Base58CheckSerialize(wallet.PaymentAddressType))
	}
}
