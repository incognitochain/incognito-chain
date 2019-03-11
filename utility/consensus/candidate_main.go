package main

import (
	"fmt"

	"github.com/constant-money/constant-chain/blockchain"
	"github.com/constant-money/constant-chain/utility/consensus/candidate"
)

func main() {

	beaconBlock := blockchain.ChainTestParam.GenesisBlockBeacon
	beaconStakingCandidate, shardStakingCandidate := candidate.GetStakingCandidate(beaconBlock)

	fmt.Println(beaconStakingCandidate)
	fmt.Println(shardStakingCandidate)

	//res, err := candidate.AssignValidator(generator.PreSelectShardNodeTestnetSerializedPubkey, 4121500227)
	//fmt.Printf("Result%+v\n error %+v\n", res, err)
	//
	//pendingValidator, currentValidator, err := candidate.SwapValidator(generator.PreSelectShardNodeTestnetSerializedPubkey, generator.PreSelectBeaconNodeTestnetSerializedPubkey, 1)
	//fmt.Println(pendingValidator)
	//fmt.Println("---------------------")
	//fmt.Println(currentValidator)
	//fmt.Println("---------------------")
	//fmt.Println(err)
}
