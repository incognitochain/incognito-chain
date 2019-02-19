package wallet

import (
	"encoding/hex"
	"fmt"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy"
	"log"
	"math/big"
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

func TestBase58SerialNumber(t *testing.T){

	spendingDecode, _ := Base58CheckDeserialize("112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV")
	privateKeyBN := new(big.Int).SetBytes(spendingDecode.KeySet.PrivateKey)

	publicKeyDecode, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu")
	fmt.Printf("Public key decode: %v\n", publicKeyDecode)

	cmDecode, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "15iYzoFTsoE2xkRe8cb2HbWeEBUoejZCBedV8e14xXiJjBPCHtX")
	fmt.Printf("cmDecode: %v\n", cmDecode)

	sndDecode, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "1gkrfYsYoria5TQMqYTzqnXutgdgqPvXwhMHk7q2UaZmjvkDnA")
	fmt.Printf("sndDecode: %v\n", sndDecode)

	randDecode, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "1M59rHeaXwDNfWoDTRh6QNPPUSLizfcrK14ic9cJTo2pjzxtLR")
	fmt.Printf("randDecode: %v\n", randDecode)

	sndBN := new(big.Int).SetBytes(sndDecode)

	serialNumber := privacy.PedCom.G[0].Derive(privateKeyBN, sndBN)

	fmt.Printf("serial number: %v\n", serialNumber.Compress())
	fmt.Printf("serial number str: %v\n\n", base58.Base58Check.Encode(base58.Base58Check{}, serialNumber.Compress(), 0))


	//fmt.Printf("Spending key: %x\n", privateKeyBN)
	//fmt.Printf("snd: %x\n", sndBN)



	//fmt.Printf("Spending key: %v\n", spendingDecode.KeySet.PrivateKey)


	publicKeyDecode2, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu")
	fmt.Printf("Public key decode: %v\n", publicKeyDecode2)

	cmDecode2, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "14twyd1G7ti8fohwiWeicRWq2PCZ4DfdVU8YKmv7tBgqJcSxV6n")
	fmt.Printf("cmDecode: %v\n", cmDecode2)

	sndDecode2, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "1yEnHGczkCsWfYQ8NHW8hE6iry7R6aKUvS8SxHxaXmdavpq2fP")
	fmt.Printf("sndDecode: %v\n", sndDecode2)

	randDecode2, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "1LPb7Q63pgxSjkDX6pnRijVX3nVGT3dZFcv39e26FRa7FourEz")
	fmt.Printf("randDecode: %v\n", randDecode2)

	sndBN2 := new(big.Int).SetBytes(sndDecode2)

	serialNumber2 := privacy.PedCom.G[0].Derive(privateKeyBN, sndBN2)

	fmt.Printf("serial number: %v\n", serialNumber2.Compress())
	fmt.Printf("serial number str: %v\n\n", base58.Base58Check.Encode(base58.Base58Check{}, serialNumber2.Compress(), 0))


	//publicKeyDecode3, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu")
	//fmt.Printf("Public key decode: %v\n", publicKeyDecode3)
	//
	//cmDecode3, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "17kkFSuLMKZzunuyE4PMfPE33h46xBb8YYXexs41GgAGdCSw15Q")
	//fmt.Printf("cmDecode: %v\n", cmDecode3)
	//
	//sndDecode3, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "1HL7HUaJhExapXS7T6boAtr2Y2YuxoEgY98kRL5QiSTHQ3H86p")
	//fmt.Printf("sndDecode: %v\n", sndDecode3)
	//
	//randDecode3, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "1sQnsntAyZvA2bV7fJXcJHbvJpyjcX9XhErhWT4HHvoQjLm296")
	//fmt.Printf("randDecode: %v\n", randDecode3)
	//
	//sndBN3 := new(big.Int).SetBytes(sndDecode3)
	//
	//serialNumber3 := privacy.PedCom.G[0].Derive(privateKeyBN, sndBN3)
	//
	//fmt.Printf("serial number: %v\n", serialNumber3.Compress())
	//fmt.Printf("serial number str: %v\n\n", base58.Base58Check.Encode(base58.Base58Check{}, serialNumber3.Compress(), 0))
	//
	//
	//publicKeyDecode4, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "177KNe6pRhi97hD9LqjUvGxLoNeKh9F5oSeh99V6Td2sQcm7qEu")
	//fmt.Printf("Public key decode: %v\n", publicKeyDecode4)
	//
	//cmDecode4, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "18kf2YBs4r4FQqam649shLEBXQMuLebW18tgLcJhHyy8kfvrw8N")
	//fmt.Printf("cmDecode: %v\n", cmDecode4)
	//
	//sndDecode4, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "1voL2ayHELoSEi8f5GoowXkjUQ1dEXNAeBKRZ7GYkAhr5HMAff")
	//fmt.Printf("sndDecode: %v\n", sndDecode4)
	//
	//randDecode4, _, _:= base58.Base58Check.Decode(base58.Base58Check{}, "12BsSFzKs7qdJfvwLe1puz3oyzido1Mwu4juBcWcpBNwk8t4nxS")
	//fmt.Printf("randDecode: %v\n", randDecode4)
	//
	//sndBN4 := new(big.Int).SetBytes(sndDecode4)
	//
	//serialNumber4 := privacy.PedCom.G[0].Derive(privateKeyBN, sndBN4)
	//
	//fmt.Printf("serial number: %v\n", serialNumber4.Compress())
	//fmt.Printf("serial number str: %v\n\n", base58.Base58Check.Encode(base58.Base58Check{}, serialNumber4.Compress(), 0))

}
