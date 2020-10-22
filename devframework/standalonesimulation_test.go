package devframework

import (
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func Test_SendTX(t *testing.T) {
	sim := NewStandaloneSimulation(Config{
		ShardNumber: 2,
	})

	sim.Pause()

	sim.GenerateBlock(-1, &Hook{
		Create: func(chain interface{}, doCreate func() (common.BlockInterface, error)) {
			fmt.Println("PreCreate block", -1)
			blk, err := doCreate()
			fmt.Println("PostCreate block", -1, blk, err)
		},
		Validation: func(chain interface{}, block common.BlockInterface, doValidation func(common.BlockInterface) error) {
			fmt.Println("PreValidation block", -1)
			err := doValidation(block)
			fmt.Println("PostValidation block", -1, err)
		},
		Insert: func(chain interface{}, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
			fmt.Println("PreInsert block", -1)
			err := doInsert(block)
			fmt.Println("PostInsert block", -1, err)
		},
	})

	sim.Pause()
}
