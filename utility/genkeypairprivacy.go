//+build !test

package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"
	"io/ioutil"
	"os"
)

type AccountPub struct {
	PaymentAddress     string
	CommitteePublicKey string
}
type AccountAll struct {
	PrivateKey         string
	PaymentAddress     string
	CommitteePublicKey string
	IncognitoPublicKey string
	MiningKey          string
}

const MAX_SHARD = 1
const MIN_MEMBER = 40

func main() {
	mnemonicGen := wallet.MnemonicGenerator{}
	//Entropy, _ := mnemonicGen.NewEntropy(128)
	//Mnemonic, _ := mnemonicGen.NewMnemonic(Entropy)
	Mnemonic := ""
	fmt.Printf("Mnemonic: %s\n", Mnemonic)
	Seed := mnemonicGen.NewSeed(Mnemonic, "laskdflaks")

	key, _ := wallet.NewMasterKey(Seed)
	var i = 0
	var j = 0

	listAcc := make(map[int][]AccountPub)
	listAccAll := make(map[int][]AccountAll)
	listPrivAcc := make(map[int][]string)
	beaconAcc := []AccountPub{}
	beaconAccAll := []AccountAll{}
	beaconPriv := []string{}

	for j = 0; j < MAX_SHARD; j++ {
		listAcc[j] = []AccountPub{}
		listAccAll[j] = []AccountAll{}
	}

	for {
		child, _ := key.NewChildKey(uint32(i))
		i++
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
		pubKey := base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte)
		shardID := int(child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1])
		privateSeedBytes := common.HashB(common.HashB(child.KeySet.PrivateKey))
		privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
		committeePK, err := incognitokey.NewCommitteeKeyFromSeed(privateSeedBytes, child.KeySet.PaymentAddress.Pk)
		if err != nil {
			panic(err)
		}
		committeePKBytes, err := committeePK.Bytes()
		if err != nil {
			panic(err)
		}
		committeePublicKey := base58.Base58Check{}.Encode(committeePKBytes, common.ZeroByte)
		if shardID >= MAX_SHARD {
			continue
		}

		if len(listAcc[shardID]) < MIN_MEMBER {
			listAcc[shardID] = append(listAcc[shardID], AccountPub{paymentAddress, committeePublicKey})
			listAccAll[shardID] = append(listAccAll[shardID], AccountAll{privAddr, paymentAddress, committeePublicKey, pubKey, privateSeed})
			listPrivAcc[shardID] = append(listPrivAcc[shardID], privAddr)
		}

		shouldBreak := true
		for i, _ := range listAcc {
			if len(listAcc[i]) < MIN_MEMBER {
				shouldBreak = false
			}
		}

		if shouldBreak {
			break
		}
	}

	for {
		child, _ := key.NewChildKey(uint32(i))
		i++
		privAddr := child.Base58CheckSerialize(wallet.PriKeyType)
		paymentAddress := child.Base58CheckSerialize(wallet.PaymentAddressType)
		pubKey := base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte)
		shardID := int(child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1])
		privateSeedBytes := common.HashB(common.HashB(child.KeySet.PrivateKey))
		privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
		committeePK, err := incognitokey.NewCommitteeKeyFromSeed(privateSeedBytes, child.KeySet.PaymentAddress.Pk)
		if err != nil {
			panic(err)
		}
		committeePKBytes, err := committeePK.Bytes()
		if err != nil {
			panic(err)
		}
		committeePublicKey := base58.Base58Check{}.Encode(committeePKBytes, common.ZeroByte)
		if shardID != 0 {
			continue
		}

		if len(beaconAcc) < MIN_MEMBER {
			beaconAcc = append(beaconAcc, AccountPub{paymentAddress, committeePublicKey})
			beaconAccAll = append(beaconAccAll, AccountAll{privAddr, paymentAddress, committeePublicKey, pubKey, privateSeed})
			beaconPriv = append(beaconPriv, privAddr)
		} else {
			break
		}

	}

	data, _ := json.Marshal(struct {
		Shard  map[int][]AccountPub
		Beacon []AccountPub
	}{
		listAcc,
		beaconAcc,
	})

	os.Remove("keylist_stake.json")
	ioutil.WriteFile("keylist_stake.json", data, 0x766)

	data, _ = json.Marshal(struct {
		Shard  map[int][]AccountAll
		Beacon []AccountAll
	}{
		listAccAll,
		beaconAccAll,
	})

	os.Remove("key_all_stake.json")
	ioutil.WriteFile("key_all_stake.json", data, 0x766)

	data, _ = json.Marshal(struct {
		Shard  map[int][]string
		Beacon []string
	}{
		listPrivAcc,
		beaconPriv,
	})

	os.Remove("priv_stake.json")
	ioutil.WriteFile("priv_stake.json", data, 0x766)

}
