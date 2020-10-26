package testsuite

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	F "github.com/incognitochain/incognito-chain/devframework"
	"testing"
	"time"
)

func Test_SendTX(t *testing.T) {
	sim := F.NewStandaloneSimulation("sim1", F.Config{
		ShardNumber: 2,
	})

	sim.Pause()
	sim.GenerateBlock(0, &F.Hook{
		Insert: func(chain interface{}, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
			doInsert(block)
			fmt.Println(block.GetProduceTime())
		},
	}, true)

	_ = sim.GenerateBlock(0, &F.Hook{
		Create: func(chain interface{}, doCreate func(time time.Time) (common.BlockInterface, error)) {
			fmt.Println("PreCreate block", 0)
			createTxs := []F.GenerateTxParam{}
			createTxs = append(createTxs, F.GenerateTxParam{
				SenderPrK: "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or",
				Receivers: map[string]int{
					"12RqRRt4q6nVis3bfVVf7L4fGquHU8KReA4ZWjjs43kH3VfWRFfGHwZjexHBgSkWMrAQrmq5CeWPkjD1hYt4KousUY7GUWn3Cg3iyJk": 100000,
					"12RquWY3vpaSPMtAQEozAB1pgbJJnnphzhJTux2VGaX5eHBxYGKcUTYEqJqQdAUsjzr8cpNQRnSTygnduxBpBvrqH1XthdrJMxCQyaC": 100000,
				},
			})

			err := sim.GenerateTxs(createTxs)
			if err != nil {
				panic(err)
			}
			blk, err := doCreate(time.Now())
			fmt.Println("PostCreate block", 0, blk, err)
		},
		Validation: func(chain interface{}, block common.BlockInterface, doValidation func(common.BlockInterface) error) {
			fmt.Println("PreValidation block", 0)
			err := doValidation(block)
			fmt.Println("PostValidation block", 0, err)
		},
		Insert: func(chain interface{}, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
			fmt.Println("PreInsert block", 0)
			err := doInsert(block)
			fmt.Println("PostInsert block", 0, err)
			bl1, err := sim.GetBalance("112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or")
			fmt.Println(bl1)
			bl2, err := sim.GetBalance("112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5JqxDAX6sjMoUT6")
			fmt.Println(bl2)
			bl3, err := sim.GetBalance("112t8rnY3WLfkE9MsKyW9s3Z5qGnPgCkeutTXJzcT5KJgAMS3vgTL9YbaJ7wyc52CzMnrj8QtwHuCpDzo47PV1qCnrui2dfJzKpuYJ3H6fa9")
			fmt.Println(bl3)
			fmt.Println("xxxxxx", block.GetProduceTime())
		},
	}, true)

	//fmt.Printf("%+v", cBlk[0])
	sim.Pause()
}
