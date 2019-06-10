package main

import (
	"encoding/json"
	"fmt"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/wallet"
)

type AccountKey struct {
	PrivateKey string
	PaymentAdd string
	PubKey     string
}

func main() {
	mnemonicGen := wallet.MnemonicGenerator{}
	//Entropy, _ := mnemonicGen.NewEntropy(128)
	//Mnemonic, _ := mnemonicGen.NewMnemonic(Entropy)
	Mnemonic := ""
	fmt.Printf("Mnemonic: %s\n", Mnemonic)
	Seed := mnemonicGen.NewSeed(Mnemonic, "123456")

	key, _ := wallet.NewMasterKey(Seed)
	var i = 0
	var j byte = 0

	listAcc := make(map[byte][]AccountKey)

	for j = 0; j < 8; j++ {
		listAcc[j] = []AccountKey{}
	}

	for {
		child, _ := key.NewChildKey(uint32(i))
		i++
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
		pubKey := base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte)
		shardID := child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1]
		if shardID > 7 {
			continue
		}

		if len(listAcc[shardID]) < 4 {
			listAcc[shardID] = append(listAcc[shardID], AccountKey{privAddr, paymentAddress, pubKey})
		}

		shouldBreak := true
		for i, _ := range listAcc {
			if len(listAcc[i]) < 4 {
				shouldBreak = false
			}
		}

		if shouldBreak {
			break
		}
	}

	bs, _ := json.Marshal(listAcc)
	fmt.Println(string(bs))
}
