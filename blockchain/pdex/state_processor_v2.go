package pdex

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
)

type stateProcessorV2 struct {
	stateProcessorBase
}

func (sp *stateProcessorV2) addLiquidity(
	insts [][]string,
) error {
	for _, inst := range insts {
		switch inst[1] {
		case instruction.WaitingStatus:
			waitingAddLiquidityInst := instruction.WaitingAddLiquidity{}
			err := waitingAddLiquidityInst.FromStringArr(inst)
			//TODO: Update state with current instruction
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sp *stateProcessorV2) modifyParams(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	params Params,
) (Params, error) {
	return params, nil
}
