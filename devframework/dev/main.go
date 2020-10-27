package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	F "github.com/incognitochain/incognito-chain/devframework"
)

func main() {
	sim := F.NewStandaloneSimulation("sim1", F.Config{
		ShardNumber: 2,
	})
	sim.GenerateBlock().NextRound()
	acc1 := sim.NewAccountFromShard(1)
	_, err := sim.CreateTransaction(sim.IcoAccount, acc1, 1000)
	if err != nil {
		panic(err)
	}
	sim.GenerateBlock(F.Hook{
		//Create: func(chainID int, doCreate func(time time.Time) (common.BlockInterface, error)) {
		//
		//	doCreate(time.Now())
		//},
		//Validation: func(chainID int, block common.BlockInterface, doValidation func(common.BlockInterface) error) {
		//	fmt.Println("PreValidation block", 0)
		//	err := doValidation(block)
		//	fmt.Println("PostValidation block", 0, err)
		//},
		Insert: func(chainID int, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
			doInsert(block)
			if chainID == 0 {
				bl1, _ := sim.GetBalance(sim.IcoAccount)
				fmt.Println(bl1)
				bl2, _ := sim.GetBalance(acc1)
				fmt.Println(bl2)
				fmt.Printf("%+v", block.(*blockchain.ShardBlock).Body)
				sim.Pause()
			}

		},
	})
	sim.NextRound()

	for i := 0; i < 10; i++ {
		sim.GenerateBlock(F.Hook{
			Insert: func(chainID int, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
				if chainID == -1 {
					fmt.Printf("%+v %+v", block.(*blockchain.BeaconBlock).GetHeight(), block.(*blockchain.BeaconBlock).Body.ShardState)
					sim.Pause()
					doInsert(block)
					sim.Pause()
				} else {
					doInsert(block)
				}

			},
		}).NextRound()
	}

	balance, _ := sim.GetBalance(acc1)
	fmt.Printf("%+v", balance)
	sim.Pause()
}
