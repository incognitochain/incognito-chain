package wallet

import (
	"log"
	"testing"
	"encoding/hex"
	"fmt"
)

func TestNewMasterkey(t *testing.T) {
	mnemonicGen := MnemonicGenerator{}
	entropy, err := mnemonicGen.NewEntropy(128)
	if err != nil {
		t.Error("Can not create entropy")
	}
	mnemonic, err := mnemonicGen.NewMnemonic(entropy)
	if err != nil {
		t.Error("Can not create mnemonic")
	}
	seed := mnemonicGen.NewSeed(mnemonic, "password")
	log.Print(hex.EncodeToString(seed))

	masterKey, _ := NewMasterKey(seed)
	//b58privateKey := masterKey.Base58CheckSerialize(true)

	b58publicKey := masterKey.Base58CheckSerialize(false)
	fmt.Printf("Base58Check encode of public key: %s\n", b58publicKey)
	fmt.Printf("Address of public key: %s\n", masterKey.ToAddress(false))

	//Private, _ := Base58CheckDeserialize(b58privateKey)
	//log.Print(hex.EncodeToString(Private.KeyPair.PrivateKey))

	Public, _ := Base58CheckDeserialize(b58publicKey)
	log.Print(hex.EncodeToString(Public.KeyPair.PublicKey))

	child0, _ := masterKey.NewChildKey(0)
	b58ChildpublicKey := child0.Base58CheckSerialize(false)
	fmt.Printf("Base58Check encode of public key: %s\n", b58ChildpublicKey)
	fmt.Printf("Address of public key: %s\n", child0.ToAddress(false))
}
