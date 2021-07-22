package pdex

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
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

func (sp *stateProcessorV2) trade(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	pairs map[string]PoolPairState,
) (map[string]PoolPairState, error) {
	switch inst[1] {
	case metadataPdexv3.TradeAcceptedStatus:
		currentTrade := &instruction.Action{Content: metadataPdexv3.AcceptedTrade{}}
		err := currentTrade.FromStrings(inst)
		if err != nil {
			return pairs, err
		}
	case metadataPdexv3.TradeRefundedStatus:
		currentTrade := &instruction.Action{Content: metadataPdexv3.RefundedTrade{}}
		err := currentTrade.FromStrings(inst)
		if err != nil {
			return pairs, err
		}
	}
	// TODO : apply state changes
	return pairs, nil
}

func (sp *stateProcessorV2) addOrder(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	pairs map[string]PoolPairState,
) (map[string]PoolPairState, error) {
	switch inst[1] {
	case metadataPdexv3.OrderAcceptedStatus:
		currentTrade := &instruction.Action{Content: metadataPdexv3.AcceptedAddOrder{}}
		err := currentTrade.FromStrings(inst)
		if err != nil {
			return pairs, err
		}
	case metadataPdexv3.OrderRefundedStatus:
		currentTrade := &instruction.Action{Content: metadataPdexv3.RefundedAddOrder{}}
		err := currentTrade.FromStrings(inst)
		if err != nil {
			return pairs, err
		}
	}
	// TODO : apply state changes
	return pairs, nil
}
