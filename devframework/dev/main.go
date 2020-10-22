package main

import (
	"fmt"
	F "github.com/incognitochain/incognito-chain/devframework"
)

func main() {
	sim := F.NewStandaloneSimulation("sim", F.Config{
		ShardNumber: 2,
	})

	sim.Pause()

	sim.GenerateBlock(&F.Hook{
		Create: func(chainID int, chain interface{}, doCreate func() (blk interface{}, err error)) {
			fmt.Println("PreCreate block", chainID)
			blk, err := doCreate()
			fmt.Printf("PostCreate block %+v %v\n", blk, err)
		},
		Validation: func(chainID int, chain interface{}, block interface{}, doValidation func(interface{}) error) {
			fmt.Println("PreValidation block", chainID)
			err := doValidation(block)
			fmt.Printf("PostValidation %v\n", err)
		},
		Insert: func(chainID int, chain interface{}, block interface{}, doInsert func(interface{}) error) {
			fmt.Println("PreInsert block", chainID)
			err := doInsert(block)
			fmt.Printf("PostInsert %v\n", err)
		},
	})

	sim.Pause()
}
