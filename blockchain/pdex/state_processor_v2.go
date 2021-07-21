package pdex

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateProcessorV2 struct {
	stateProcessorBase
}

func (sp *stateProcessorV2) addLiquidity(
	stateDB *statedb.StateDB,
	inst []string,
	poolPairs map[string]PoolPairState,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
) (map[string]PoolPairState, map[string]Contribution, map[string]Contribution, error) {
	var err error
	switch inst[len(inst)-1] {
	case instruction.WaitingStatus:
		waitingContributions, err = sp.waitingContribution(stateDB, inst, waitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case instruction.RefundStatus:
		waitingContributions, deletedWaitingContributions, err = sp.refundContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case instruction.MatchStatus:
		waitingContributions, deletedWaitingContributions, poolPairs, err = sp.matchContribution(
			stateDB, inst, waitingContributions, deletedWaitingContributions, poolPairs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case instruction.MatchAndReturnStatus:
		waitingContributions, poolPairs, err = sp.matchAndReturnContribution(stateDB, inst, waitingContributions, poolPairs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	}
	return poolPairs, waitingContributions, deletedWaitingContributions, nil
}

func (sp *stateProcessorV2) waitingContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]Contribution,
) (map[string]Contribution, error) {
	waitingAddLiquidityInst := instruction.WaitingAddLiquidity{}
	err := waitingAddLiquidityInst.FromStringArr(inst)
	if err != nil {
		return waitingContributions, err
	}
	metaData := waitingAddLiquidityInst.MetaData().(*metadataPdexV3.AddLiquidity)
	//TODO: Update state with current instruction
	contribution := NewContributionWithValue(
		metaData.PoolPairID(),
		metaData.ReceiveAddress(), metaData.RefundAddress(),
		metaData.TokenID(), waitingAddLiquidityInst.TxReqID(),
		metaData.TokenAmount(), metaData.Amplifier(),
		waitingAddLiquidityInst.ShardID(),
	)
	waitingContributions[metaData.PairHash()] = *contribution

	//TODO: Track status of contribution

	return waitingContributions, nil
}

func (sp *stateProcessorV2) refundContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
) (map[string]Contribution, map[string]Contribution, error) {
	refundAddLiquidityInst := instruction.RefundAddLiquidity{}
	err := refundAddLiquidityInst.FromStringArr(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, err
	}
	metaData := refundAddLiquidityInst.MetaData().(*metadataPdexV3.AddLiquidity)
	existingWaitingContribution, found := waitingContributions[metaData.PairHash()]
	if found {
		deletedWaitingContributions[metaData.PairHash()] = existingWaitingContribution
		delete(waitingContributions, metaData.PairHash())
	}
	//TODO: Track status of contribution

	return waitingContributions, deletedWaitingContributions, nil
}

func (sp *stateProcessorV2) matchContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
	poolPairs map[string]PoolPairState,
) (map[string]Contribution, map[string]Contribution, map[string]PoolPairState, error) {
	matchAddLiquidityInst := instruction.MatchAddLiquidity{}
	err := matchAddLiquidityInst.FromStringArr(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	metaData := matchAddLiquidityInst.MetaData().(*metadataPdexV3.AddLiquidity)
	existingWaitingContribution, found := waitingContributions[metaData.PairHash()]
	if !found {
		err := fmt.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", metaData.PairHash())
		Logger.log.Error(err)
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}

	incomingWaitingContribution := *NewContributionWithValue(
		metaData.PoolPairID(), metaData.ReceiveAddress(), metaData.RefundAddress(),
		metaData.TokenID(), matchAddLiquidityInst.TxReqID(),
		metaData.TokenAmount(), metaData.Amplifier(),
		matchAddLiquidityInst.ShardID(),
	)
	poolPair := initPoolPairState(existingWaitingContribution, incomingWaitingContribution)
	nfctID, err := poolPair.addShare(poolPair.token0RealAmount)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	if nfctID != matchAddLiquidityInst.NfctID() {
		err := fmt.Errorf("NfctID is invalid expect %s but get %s", nfctID, matchAddLiquidityInst.NfctID())
		Logger.log.Error(err)
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	//TODO: @tin add more conditions here
	deletedWaitingContributions[metaData.PairHash()] = existingWaitingContribution
	delete(waitingContributions, metaData.PairHash())

	//TODO: Track status of contribution
	return waitingContributions, deletedWaitingContributions, poolPairs, nil
}

func (sp *stateProcessorV2) matchAndReturnContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]Contribution,
	poolPairs map[string]PoolPairState,
) (map[string]Contribution, map[string]PoolPairState, error) {
	matchAndReturnAddLiquidity := instruction.MatchAndReturnAddLiquidity{}
	err := matchAndReturnAddLiquidity.FromStringArr(inst)
	if err != nil {
		return waitingContributions, poolPairs, err
	}
	metaData := matchAndReturnAddLiquidity.MetaData().(*metadataPdexV3.AddLiquidity)
	Logger.log.Info(metaData)

	//TODO: Track status of contribution
	return waitingContributions, poolPairs, nil
}

func (sp *stateProcessorV2) modifyParams(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	params Params,
) (Params, error) {
	return params, nil
}
