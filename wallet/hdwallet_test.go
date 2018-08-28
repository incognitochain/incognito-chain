package wallet

import (
	"log"
	"testing"
	"encoding/hex"
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
	//b58privateKey := masterKey.B58Serialize(true)
	b58publicKey := masterKey.B58Serialize(false)

	//Private, _ := B58Deserialize(b58privateKey)
	//log.Print(hex.EncodeToString(Private.KeyPair.PrivateKey))

	Public, _ := B58Deserialize(b58publicKey)
	log.Print(hex.EncodeToString(Public.KeyPair.PublicKey))
}
