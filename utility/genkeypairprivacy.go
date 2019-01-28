package main

import (
	"fmt"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/privacy"

	"github.com/ninjadotorg/constant/wallet"
)

func main() {

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
	fmt.Printf("Burn payment address : %s", burnPaymentAddress)

	mnemonicGen := wallet.MnemonicGenerator{}
	Entropy, _ := mnemonicGen.NewEntropy(128)
	Mnemonic, _ := mnemonicGen.NewMnemonic(Entropy)
	Seed := mnemonicGen.NewSeed(Mnemonic, "123456")

	key, _ := wallet.NewMasterKey(Seed)
	fmt.Printf("PubKey: %v\n\n", *key)
	fmt.Printf("PubKey: %x\n\n", *key)

	/*pubAddr := key.Base58CheckSerialize(wallet.PaymentAddressType)
	privAddr := key.Base58CheckSerialize(wallet.PriKeyType)
	readAddr := key.Base58CheckSerialize(wallet.ReadonlyKeyType)
	fmt.Printf("pubAddr: %v\n", pubAddr)
	fmt.Printf("pubAddr: %x\n\n", pubAddr)
	fmt.Printf("privAddr: %v\n", privAddr)
	fmt.Printf("privAddr: %x\n\n", privAddr)
	fmt.Printf("readAddr: %v\n", readAddr)
	fmt.Printf("readAddr: %x\n\n", readAddr)*/

	for i := 0; i < 30; i++ {
		child, _ := key.NewChildKey(uint32(i))
		//pubAddr := child.Base58CheckSerialize(wallet.PaymentAddressType)
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		//readAddr := child.Base58CheckSerialize(wallet.ReadonlyKeyType)
		fmt.Printf("Acc %d:\n ", i)
		/*fmt.Printf("pubAddr: %v\n", pubAddr)
		fmt.Printf("pubAddr: %x\n\n", pubAddr)*/
		fmt.Printf("privAddr: %v\n", privAddr)
		fmt.Printf("privAddr: %x\n", privAddr)
		/*fmt.Printf("readAddr: %v\n", readAddr)
		fmt.Printf("readAddr: %x\n\n", readAddr)*/
	}
}
