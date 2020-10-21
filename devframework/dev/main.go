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
			fmt.Println("PostCreate block", chainID, blk, err)
		},
		Validation: func(chainID int, chain interface{}, block interface{}, doValidation func(interface{}) error) {
			fmt.Println("PreValidation block", chainID)
			err := doValidation(block)
			fmt.Println("PostValidation block", chainID, err)
		},
		Insert: func(chainID int, chain interface{}, block interface{}, doInsert func(interface{}) error) {
			fmt.Println("PreInsert block", chainID)
			err := doInsert(block)
			fmt.Println("PostInsert block", chainID, err)
		},
	})

	sim.Pause()
}
