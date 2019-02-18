package wallet

import (
	"encoding/hex"
	"log"
	"testing"
)

func TestNewEntropy(t *testing.T) {
	mnemonic := MnemonicGenerator{}
	entropy, err := mnemonic.NewEntropy(128)
	if err != nil {
		t.Error("Can not create Entropy")
	}
	log.Print(entropy)
}

func TestNewMnemonic(t *testing.T) {
	mnemonicGen := MnemonicGenerator{}
	entropy, err := mnemonicGen.NewEntropy(128)
	if err != nil {
		t.Error("Can not create Entropy")
	}
	mnemonic, err := mnemonicGen.NewMnemonic(entropy)
	if err != nil {
		t.Error("Can not create Mnemonic")
	}
	log.Print(mnemonic)
}

func TestNewSeed(t *testing.T) {
	mnemonicGen := MnemonicGenerator{}
	entropy, err := mnemonicGen.NewEntropy(128)
	if err != nil {
		t.Error("Can not create Entropy")
	}
	mnemonic, err := mnemonicGen.NewMnemonic(entropy)
	if err != nil {
		t.Error("Can not create Mnemonic")
	}
	seed := mnemonicGen.NewSeed(mnemonic, "password")
	log.Print(hex.EncodeToString(seed))
}
