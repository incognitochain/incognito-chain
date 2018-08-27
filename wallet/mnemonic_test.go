package wallet

import (
	"log"
	"testing"
	"encoding/hex"
)

func TestNewEntropy(t *testing.T) {
	entropy, err := NewEntropy(128)
	if err != nil {
		t.Error("Can not create entropy")
	}
	log.Print(entropy)
}

func TestNewMnemonic(t *testing.T) {
	entropy, err := NewEntropy(128)
	if err != nil {
		t.Error("Can not create entropy")
	}
	mnemonic, err := NewMnemonic(entropy)
	if err != nil {
		t.Error("Can not create mnemonic")
	}
	log.Print(mnemonic)
}

func TestNewSeed(t *testing.T) {
	entropy, err := NewEntropy(128)
	if err != nil {
		t.Error("Can not create entropy")
	}
	mnemonic, err := NewMnemonic(entropy)
	if err != nil {
		t.Error("Can not create mnemonic")
	}
	seed := NewSeed(mnemonic, "password")
	log.Print(hex.EncodeToString(seed))
}
