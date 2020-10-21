package devframework

import (
	"fmt"
	"testing"
)

func Test_SendTX(t *testing.T) {
	sim := NewStandaloneSimulation(Config{
		Mode:        MODE_STANDALONE,
		ShardNumber: 2,
	})

	sim.Pause()

	sim.GenerateBlock(&Hook{
		Create: func(chainID int, chain interface{}, doCreate func() (interface{}, error)) {
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
