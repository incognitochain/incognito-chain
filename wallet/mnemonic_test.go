package wallet

import (
	"log"
	"testing"
	"encoding/hex"
)

func TestNewEntropy(t *testing.T) {
	mnemonic := MnemonicGenerator{}
	entropy, err := mnemonic.NewEntropy(128)
	if err != nil {
		t.Error("Can not create entropy")
	}
	log.Print(entropy)
}

func TestNewMnemonic(t *testing.T) {
	mnemonicGen := MnemonicGenerator{}
	entropy, err := mnemonicGen.NewEntropy(128)
	if err != nil {
		t.Error("Can not create entropy")
	}
	mnemonic, err := mnemonicGen.NewMnemonic(entropy)
	if err != nil {
		t.Error("Can not create mnemonic")
	}
	log.Print(mnemonic)
}

func TestNewSeed(t *testing.T) {
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
}
