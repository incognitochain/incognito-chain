package main

import (
	"fmt"

	"github.com/ninjadotorg/constant/utility/consensus/candidate"
	"github.com/ninjadotorg/constant/utility/generateKeys/generator"
)

func main() {
	res, err := candidate.AssignValidator(generator.PreSelectShardNodeTestnetSerializedPubkey, 4121500227)
	fmt.Printf("Result%+v\n error %+v\n", res, err)

	pendingValidator, currentValidator, err := candidate.SwapValidator(generator.PreSelectShardNodeTestnetSerializedPubkey, generator.PreSelectBeaconNodeTestnetSerializedPubkey, 1)
	fmt.Println(pendingValidator)
	fmt.Println("---------------------")
	fmt.Println(currentValidator)
	fmt.Println("---------------------")
	fmt.Println(err)
}
