package main

import (
	"fmt"

	"github.com/ninjadotorg/cash-prototype/wallet"
)

func main5() {
	mnemonicGen := wallet.MnemonicGenerator{}
	Entropy, _ := mnemonicGen.NewEntropy(128)
	Mnemonic, _ := mnemonicGen.NewMnemonic(Entropy)
	Seed := mnemonicGen.NewSeed(Mnemonic, "autonomous")
	// fmt.Println("Seed: ", Seed)
	// fmt.Printf("Seed: %x\n\n", Seed)

	// keyPair, _ := (&cashec.KeyPair{}).GenerateKey(Seed)
	// fmt.Printf("Keypair: %v\n", *keyPair)
	// fmt.Printf("Keypair: %x\n\n", *keyPair)

	key, _ := wallet.NewMasterKey(Seed)
	fmt.Printf("Key: %v\n\n", *key)
	fmt.Printf("Key: %x\n\n", *key)

	pubAddr := key.Base58CheckSerialize(wallet.PubKeyType)
	privAddr := key.Base58CheckSerialize(wallet.PriKeyType)
	fmt.Printf("pubAddr: %v\n", pubAddr)
	fmt.Printf("pubAddr: %x\n\n", pubAddr)
	fmt.Printf("privAddr: %v\n", privAddr)
	fmt.Printf("privAddr: %x\n\n", privAddr)

	newKey, _ := wallet.Base58CheckDeserialize(pubAddr)
	fmt.Printf("NewKey: %v\n", *newKey)
	fmt.Printf("NewKey: %x\n\n", *newKey)

	newKeyPriv, _ := wallet.Base58CheckDeserialize(privAddr)
	fmt.Printf("NewKeyPriv: %v\n", *newKeyPriv)
	fmt.Printf("NewKeyPriv: %x\n", *newKeyPriv)
}
