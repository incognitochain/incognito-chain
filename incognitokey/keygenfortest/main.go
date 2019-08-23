package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

type Key struct {
	Payment         string `json:"PaymentAddress"`
	CommitteePubKey string `json:"CommitteePublicKey"`
}

func (key *Key) NewFromSeed(seed []byte) {
	incKeySet := new(incognitokey.KeySet)
	incKeySet.GenerateKey(seed)
	wl := wallet.KeyWallet{}
	wl.KeySet = *incKeySet
	key.Payment = wl.Base58CheckSerialize(0x1)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(incKeySet.PrivateKey), incKeySet.PaymentAddress.Pk)
	key.CommitteePubKey, _ = committeeKey.ToBase58()
}

func NewKey(seed []byte) (*Key, string, string) {
	masterKey, _ := wallet.NewMasterKey(seed)
	pubKey := new(Key)
	pubKey.Payment = masterKey.Base58CheckSerialize(0x1)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(masterKey.KeySet.PrivateKey), masterKey.KeySet.PaymentAddress.Pk)
	pubKey.CommitteePubKey, _ = committeeKey.ToBase58()
	return pubKey, masterKey.Base58CheckSerialize(0x0), base58.Base58Check{}.Encode(common.HashB(masterKey.KeySet.PrivateKey), common.ZeroByte)
}

func NewKeyFromIncKey(key string) (*Key, string, string) {
	masterKey, _ := wallet.Base58CheckDeserialize(key)
	masterKey.KeySet.InitFromPrivateKey(&masterKey.KeySet.PrivateKey)
	pubKey := new(Key)
	pubKey.Payment = masterKey.Base58CheckSerialize(0x1)
	committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(masterKey.KeySet.PrivateKey), masterKey.KeySet.PaymentAddress.Pk)
	pubKey.CommitteePubKey, _ = committeeKey.ToBase58()
	return pubKey, masterKey.Base58CheckSerialize(0x0), base58.Base58Check{}.Encode(common.HashB(masterKey.KeySet.PrivateKey), common.ZeroByte)
}

type ShardPrivate struct {
	Seed map[int][][]byte
}

type BeaconPrivate struct {
	Seed [][]byte
}

type ShardPrivateKey struct {
	Pri map[int][]string
}

type BeaconPrivateKey struct {
	Pri []string
}

type ShardPrivateSeed struct {
	Pri map[int][]string
}

type BeaconPrivateSeed struct {
	Pri []string
}

type KeyList struct {
	Bc []Key         `json:"Beacon"`
	Sh map[int][]Key `json:"Shard"`
}

type KeyListFromPrivate struct {
	Bc []string         `json:"Beacon"`
	Sh map[int][]string `json:"Shard"`
}

func generateKeydotJsonFromGivenKeyList(filename string, numberOfShard, numberOfCandidate int) {
	keyListFromPrivate := KeyListFromPrivate{}
	flag := false
	if filename != "" {
		jsonFile, err := os.Open(filename)
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Successfully Opened users.json")
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		// var result map[string]interface{}
		json.Unmarshal([]byte(byteValue), &keyListFromPrivate)
		flag = true

	}
	keylist := KeyList{}
	beacon := BeaconPrivate{}
	shard := ShardPrivate{}
	beaconPri := BeaconPrivateKey{}
	shardPri := ShardPrivateKey{}
	beaconPriSeed := BeaconPrivateSeed{}
	shardPriSeed := ShardPrivateSeed{}
	for i := 0; i < numberOfCandidate; i++ {
		// key := Key{}
		seed := privacy.RandBytes(32)
		beacon.Seed = append(beacon.Seed, seed)
		//fmt.Sprintf("Beacon %v: %v\n", i, seed)
		if (flag) && (i < len(keyListFromPrivate.Bc)) {
			key, pri, priSeed := NewKeyFromIncKey(keyListFromPrivate.Bc[i])
			beaconPri.Pri = append(beaconPri.Pri, pri)
			beaconPriSeed.Pri = append(beaconPriSeed.Pri, priSeed)
			keylist.Bc = append(keylist.Bc, *key)
		} else {
			key, pri, priSeed := NewKey(seed)
			beaconPri.Pri = append(beaconPri.Pri, pri)
			beaconPriSeed.Pri = append(beaconPriSeed.Pri, priSeed)
			keylist.Bc = append(keylist.Bc, *key)
		}
	}
	keylist.Sh = map[int][]Key{}
	shard.Seed = map[int][][]byte{}
	shardPri.Pri = map[int][]string{}
	shardPriSeed.Pri = map[int][]string{}
	for j := 0; j < numberOfShard; j++ {
		for i := 0; i < numberOfCandidate; i++ {
			// key := Key{}
			seed := privacy.RandBytes(32)
			shard.Seed[j] = append(shard.Seed[j], seed)
			// fmt.Printf("Shard %v %v: %v\n", j, i, seed)
			// key.NewFromSeed(seed)
			if (flag) && (i < len(keyListFromPrivate.Sh[j])) {
				key, pri, priSeed := NewKeyFromIncKey(keyListFromPrivate.Sh[j][i])
				shardPri.Pri[j] = append(shardPri.Pri[j], pri)
				shardPriSeed.Pri[j] = append(shardPriSeed.Pri[j], priSeed)
				keylist.Sh[j] = append(keylist.Sh[j], *key)
			} else {
				key, pri, priSeed := NewKey(seed)
				shardPri.Pri[j] = append(shardPri.Pri[j], pri)
				shardPriSeed.Pri[j] = append(shardPriSeed.Pri[j], priSeed)
				keylist.Sh[j] = append(keylist.Sh[j], *key)
			}
		}
	}
	keylistJson, _ := json.Marshal(keylist)
	_ = ioutil.WriteFile("keylist.json", keylistJson, 0644)
	beaconJson, _ := json.Marshal(beacon)
	_ = ioutil.WriteFile("beaconseed.json", beaconJson, 0644)
	shardJson, _ := json.Marshal(beacon)
	_ = ioutil.WriteFile("shardseed.json", shardJson, 0644)
	beaconPriJson, _ := json.Marshal(beaconPri)
	_ = ioutil.WriteFile("beaconprivate.json", beaconPriJson, 0644)
	shardPriJson, _ := json.Marshal(shardPri)
	_ = ioutil.WriteFile("shardprivate.json", shardPriJson, 0644)
	beaconPriSeedJson, _ := json.Marshal(beaconPriSeed)
	_ = ioutil.WriteFile("beaconprivateseed.json", beaconPriSeedJson, 0644)
	shardPriSeedJson, _ := json.Marshal(shardPriSeed)
	_ = ioutil.WriteFile("shardprivateseed.json", shardPriSeedJson, 0644)
}

func generateKeydotJson(numberOfShard, numberOfCandidate int) {
	keylist := KeyList{}
	beacon := BeaconPrivate{}
	shard := ShardPrivate{}
	beaconPri := BeaconPrivateKey{}
	shardPri := ShardPrivateKey{}
	beaconPriSeed := BeaconPrivateSeed{}
	shardPriSeed := ShardPrivateSeed{}
	for i := 0; i < numberOfCandidate; i++ {
		// key := Key{}
		seed := privacy.RandBytes(32)
		beacon.Seed = append(beacon.Seed, seed)
		//fmt.Sprintf("Beacon %v: %v\n", i, seed)
		key, pri, priSeed := NewKey(seed)
		beaconPri.Pri = append(beaconPri.Pri, pri)
		beaconPriSeed.Pri = append(beaconPriSeed.Pri, priSeed)
		keylist.Bc = append(keylist.Bc, *key)
	}
	keylist.Sh = map[int][]Key{}
	shard.Seed = map[int][][]byte{}
	shardPri.Pri = map[int][]string{}
	shardPriSeed.Pri = map[int][]string{}
	for j := 0; j < numberOfShard; j++ {
		for i := 0; i < numberOfCandidate; i++ {
			// key := Key{}
			seed := privacy.RandBytes(32)
			shard.Seed[j] = append(shard.Seed[j], seed)
			// fmt.Printf("Shard %v %v: %v\n", j, i, seed)
			// key.NewFromSeed(seed)
			key, pri, priSeed := NewKey(seed)
			shardPri.Pri[j] = append(shardPri.Pri[j], pri)
			shardPriSeed.Pri[j] = append(shardPriSeed.Pri[j], priSeed)
			keylist.Sh[j] = append(keylist.Sh[j], *key)
		}
	}
	keylistJson, _ := json.Marshal(keylist)
	_ = ioutil.WriteFile("keylist.json", keylistJson, 0644)
	beaconJson, _ := json.Marshal(beacon)
	_ = ioutil.WriteFile("beaconseed.json", beaconJson, 0644)
	shardJson, _ := json.Marshal(beacon)
	_ = ioutil.WriteFile("shardseed.json", shardJson, 0644)
	// inc private key
	beaconPriJson, _ := json.Marshal(beaconPri)
	_ = ioutil.WriteFile("beaconprivate.json", beaconPriJson, 0644)
	shardPriJson, _ := json.Marshal(shardPri)
	_ = ioutil.WriteFile("shardprivate.json", shardPriJson, 0644)
	// bls mining keys
	beaconPriSeedJson, _ := json.Marshal(beaconPriSeed)
	_ = ioutil.WriteFile("beaconprivateseed.json", beaconPriSeedJson, 0644)
	shardPriSeedJson, _ := json.Marshal(shardPriSeed)
	_ = ioutil.WriteFile("shardprivateseed.json", shardPriSeedJson, 0644)
}

func main() {
	// generateKeydotJson(64, 64)
	generateKeydotJsonFromGivenKeyList("private_key_testnet.json", 256, 20)
}
