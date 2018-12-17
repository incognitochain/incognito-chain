package main

import (
	"fmt"

	"github.com/ninjadotorg/constant/utility/generateKeys/generator"
)

func main() {
	fmt.Println(generator.GenerateAddress(generator.PreSelectShardNodeTestnet))
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
}
