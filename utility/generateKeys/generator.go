package main

import (
	"github.com/constant-money/constant-chain/utility/generateKeys/generator"
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

type KeyPairs struct {
	KeyPair []KeyPair
}

func main() {
	// fmt.Println(generator.GenerateAddress(generator.PreSelectShardNodeTestnet))
	// inst := [][]string{}
	// build validator beacon
	// strBeacon := []string{"assign"}
	// strBeacon = append(strBeacon, generator.PreSelectBeaconNodeTestnetSerializedPubkey...)
	// strBeacon = append(strBeacon, "beacon")

	// strShard := []string{"assign"}
	// strShard = append(strShard, generator.PreSelectShardNodeTestnetSerializedPubkey...)
	// strShard = append(strShard, "shard")
	// inst = append(inst, strBeacon)
	// inst = append(inst, strShard)
	// fmt.Println(inst)
	// privateKeys, pubAddresses, _ := generator.GenerateAddressByte(generator.GenerateKeyPair())
	// // fmt.Println(res)
	// keyPairs := KeyPairs{}
	// for index, _ := range privateKeys {
	// 	keyPair := KeyPair{PrivateKey: privateKeys[index], PublicKey: pubAddresses[index]}
	// 	keyPairs.KeyPair = append(keyPairs.KeyPair, keyPair)
	// }
	// json, err := json.Marshal(keyPairs)
	// err = ioutil.WriteFile("output.txt", json, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	generator.GenerateAddressByShard(1)
}
