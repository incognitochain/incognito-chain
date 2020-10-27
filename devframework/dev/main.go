package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	F "github.com/incognitochain/incognito-chain/devframework"
	"time"
)

func main() {
	sim := F.NewStandaloneSimulation("sim1", F.Config{
		ShardNumber: 2,
	})

	sim.GenerateBlock().NextRound()
	sim.Pause()
	acc1 := sim.NewAccount()
	acc2 := sim.NewAccount()
	sim.ApplyChain(0).GenerateBlock(F.Hook{
		Create: func(chainID int, doCreate func(time time.Time) (common.BlockInterface, error)) {
			fmt.Println("PreCreate", chainID)
			sim.Pause()
			_, err := sim.CreateTransaction(sim.IcoAccount, acc1, 1000, acc2, 1e9)
			if err != nil {
				panic(err)
			}
			blk, err := doCreate(time.Now())
			fmt.Println("PostCreate block", 0, blk, err)
		},
		Validation: func(chainID int, block common.BlockInterface, doValidation func(common.BlockInterface) error) {
			fmt.Println("PreValidation block", 0)
			err := doValidation(block)
			fmt.Println("PostValidation block", 0, err)
		},
		Insert: func(chainID int, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
			fmt.Println("PreInsert block", 0)
			err := doInsert(block)
			fmt.Println("PostInsert block", 0, err)
			bl1, err := sim.GetBalance(sim.IcoAccount)
			fmt.Println(bl1)
			bl2, err := sim.GetBalance(acc1)
			fmt.Println(bl2)
			bl3, err := sim.GetBalance(acc2)
			fmt.Println(bl3)
			fmt.Println("xxxxxx", block.GetProduceTime())
		},
	})
	sim.NextRound()

	//fmt.Printf("%+v", cBlk[0])
	sim.Pause()
}
