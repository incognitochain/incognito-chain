package main

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/privacy"

	"github.com/ninjadotorg/constant/wallet"
)

func main() {

	a, _ := wallet.Base58CheckDeserialize("1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba")
	k1 := base58.Base58Check{}.Encode(a.KeySet.PaymentAddress.Pk, 0x00)
	_ = k1

	burnPubKeyE := privacy.PedCom.G[0].Hash(1000000)
	burnPubKey := burnPubKeyE.Compress()
	burnKey := wallet.KeyWallet{
		KeySet: cashec.KeySet{
			PaymentAddress: privacy.PaymentAddress{
				Pk: burnPubKey,
			},
		},
	}
	burnPaymentAddress := burnKey.Base58CheckSerialize(wallet.PaymentAddressType)
	fmt.Printf("Burn payment address : %s \n", burnPaymentAddress)

	keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	fmt.Println("======================================")
	fmt.Println(keyWalletBurningAdd.KeySet.PaymentAddress.Pk)
	fmt.Println("======================================")
	// mnemonicGen := wallet.MnemonicGenerator{}
	// Entropy, _ := mnemonicGen.NewEntropy(128)
	// Mnemonic, _ := mnemonicGen.NewMnemonic(Entropy)
	// fmt.Printf("Mnemonic: %s\n", Mnemonic)
	// Seed := mnemonicGen.NewSeed(Mnemonic, "123456")

	// key, _ := wallet.NewMasterKey(Seed)
	// var i int
	// var k = 0
	// for {
	// 	i++
	// 	child, _ := key.NewChildKey(uint32(i))
	// 	privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
	// 	paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
	// 	if child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1] == 0 {
	// 		fmt.Printf("Acc %d:\n ", i)
	// 		fmt.Printf("paymentAddress: %v\n", paymentAddress)
	// 		fmt.Printf("privateKey: %v\n", privAddr)
	// 		k++
	// 		if k == 3 {
	// 			break
	// 		}
	// 	}
	// 	i++
	// }

	//112t8rqnMrtPkJ4YWzXfG82pd9vCe2jvWGxqwniPM5y4hnimki6LcVNfXxN911ViJS8arTozjH4rTpfaGo5i1KKcG1ayjiMsa4E3nABGAqQh
	keyWallet, _ := wallet.Base58CheckDeserialize("112t8s2UkZEwS7JtqLHFruRrh4Drj53UzH4A6DrairctKutxVb8Vw2DMzxCReYsAZkXi9ycaSNRHEcB7TJaTwPhyPvqRzu5NnUgTMN9AEKwo")
	keyWallet.KeySet.ImportFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	fmt.Println("Pub-key byte array", keyWallet.KeySet.PaymentAddress.Pk)
	fmt.Println(base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, byte(0x00)))
	fmt.Printf("Pub-key : %+v \n", keyWallet.Base58CheckSerialize(wallet.PaymentAddressType))
}
