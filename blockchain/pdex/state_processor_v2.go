package pdex

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateProcessorV2 struct {
	stateProcessorBase
}

func (sp *stateProcessorV2) addLiquidity(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	poolPairs map[string]PoolPairState,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
) (map[string]PoolPairState, map[string]Contribution, map[string]Contribution, error) {
	var err error
	switch inst[len(inst)-1] {
	case instruction.WaitingStatus:
		waitingContributions, err = sp.waitingContribution(stateDB, inst, waitingContributions, deletedWaitingContributions)
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
			stateDB, inst, beaconHeight, waitingContributions, deletedWaitingContributions, poolPairs)
		if err != nil {
			return poolPairs, waitingContributions, deletedWaitingContributions, err
		}
	case instruction.MatchAndReturnStatus:
		waitingContributions, deletedWaitingContributions, poolPairs, err = sp.matchAndReturnContribution(
			stateDB, inst, beaconHeight,
			waitingContributions, deletedWaitingContributions, poolPairs)
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
	deletedWaitingContributions map[string]Contribution,
) (map[string]Contribution, error) {
	waitingAddLiquidityInst := instruction.WaitingAddLiquidity{}
	err := waitingAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, err
	}
	metaData := waitingAddLiquidityInst.MetaData().(*metadataPdexv3.AddLiquidity)
	err = sp.verifyWaitingContribution(*metaData, waitingContributions, deletedWaitingContributions)
	if err != nil {
		return waitingContributions, err
	}
	contribution := NewContributionWithMetaData(
		*metaData, waitingAddLiquidityInst.TxReqID(), waitingAddLiquidityInst.ShardID(),
	)
	waitingContributions[metaData.PairHash()] = *contribution

	//TODO: Track status of contribution

	return waitingContributions, nil
}

func (sp *stateProcessorV2) verifyWaitingContribution(
	metaData metadataPdexv3.AddLiquidity,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
) error {
	_, found := waitingContributions[metaData.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list waitingContributions", metaData.PairHash())
		return err
	}
	_, found = deletedWaitingContributions[metaData.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list deletedWaitingContributions", metaData.PairHash())
		return err
	}
	return nil
}

func (sp *stateProcessorV2) refundContribution(
	stateDB *statedb.StateDB,
	inst []string,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
) (map[string]Contribution, map[string]Contribution, error) {
	refundAddLiquidityInst := instruction.RefundAddLiquidity{}
	err := refundAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, err
	}
	metaData := refundAddLiquidityInst.MetaData().(*metadataPdexv3.AddLiquidity)
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
	beaconHeight uint64,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
	poolPairs map[string]PoolPairState,
) (map[string]Contribution, map[string]Contribution, map[string]PoolPairState, error) {
	matchAddLiquidityInst := instruction.MatchAddLiquidity{}
	err := matchAddLiquidityInst.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	metaData := matchAddLiquidityInst.MetaData().(*metadataPdexv3.AddLiquidity)
	existingWaitingContribution, found := waitingContributions[metaData.PairHash()]
	if !found {
		err := fmt.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", metaData.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	_, found = deletedWaitingContributions[metaData.PairHash()]
	if found {
		err := fmt.Errorf("Pair Hash %v has been existed in list deletedWaitingContributions", metaData.PairHash())
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}

	incomingWaitingContribution := *NewContributionWithMetaData(
		*metaData, matchAddLiquidityInst.TxReqID(), matchAddLiquidityInst.ShardID(),
	)
	poolPair := initPoolPairState(existingWaitingContribution, incomingWaitingContribution)
	poolPairID := generatePoolPairKey(
		existingWaitingContribution.tokenID,
		incomingWaitingContribution.tokenID,
		existingWaitingContribution.txReqID,
	)
	nfctID := poolPair.addShare(poolPairID, poolPair.token0RealAmount, beaconHeight)
	if nfctID != matchAddLiquidityInst.NfctID() {
		err := fmt.Errorf("NfctID is invalid expect %s but get %s", nfctID, matchAddLiquidityInst.NfctID())
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	if poolPairID != matchAddLiquidityInst.NewPoolPairID() {
		err := fmt.Errorf("PoolPairID is invalid expect %s but get %s", poolPairID, matchAddLiquidityInst.NewPoolPairID())
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	poolPairs[poolPairID] = *poolPair
	deletedWaitingContributions[metaData.PairHash()] = existingWaitingContribution
	delete(waitingContributions, metaData.PairHash())

	//TODO: Track status of contribution
	return waitingContributions, deletedWaitingContributions, poolPairs, nil
}

func (sp *stateProcessorV2) matchAndReturnContribution(
	stateDB *statedb.StateDB,
	inst []string,
	beaconHeight uint64,
	waitingContributions map[string]Contribution,
	deletedWaitingContributions map[string]Contribution,
	poolPairs map[string]PoolPairState,
) (map[string]Contribution, map[string]Contribution, map[string]PoolPairState, error) {
	matchAndReturnAddLiquidity := instruction.MatchAndReturnAddLiquidity{}
	err := matchAndReturnAddLiquidity.FromStringSlice(inst)
	if err != nil {
		return waitingContributions, deletedWaitingContributions, poolPairs, err
	}
	metaData := matchAndReturnAddLiquidity.MetaData().(*metadataPdexv3.AddLiquidity)
	waitingContribution, found := waitingContributions[metaData.PairHash()]
	if found {
		incomingWaitingContribution := NewContributionWithMetaData(
			*metaData, matchAndReturnAddLiquidity.TxReqID(), matchAndReturnAddLiquidity.ShardID(),
		)
		if incomingWaitingContribution.poolPairID != waitingContribution.poolPairID {
			err := fmt.Errorf("Expect poolPairID %v but get %v", waitingContribution.poolPairID, incomingWaitingContribution.poolPairID)
			return waitingContributions, deletedWaitingContributions, poolPairs, err
		}
		poolPair := poolPairs[waitingContribution.poolPairID]
		shareAmount := uint64(0)
		if waitingContribution.tokenID == incomingWaitingContribution.tokenID {
			shareAmount = poolPair.updateReserveAndShares(
				waitingContribution.tokenID, matchAndReturnAddLiquidity.ExistedTokenID(),
				waitingContribution.tokenAmount-matchAndReturnAddLiquidity.ReturnAmount(),
				matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
			)
		} else {
			shareAmount = poolPair.updateReserveAndShares(
				waitingContribution.tokenID, metaData.TokenID(),
				matchAndReturnAddLiquidity.ExistedTokenActualAmount(),
				metaData.TokenAmount()-matchAndReturnAddLiquidity.ReturnAmount(),
			)
		}
		nfctID := poolPair.addShare(waitingContribution.poolPairID, shareAmount, beaconHeight)
		//TODO: After release beacon recompute for contributions amount
		if err != nil {
			return waitingContributions, deletedWaitingContributions, poolPairs, err
		}
		if nfctID != matchAndReturnAddLiquidity.NfctID() {
			err := fmt.Errorf("NfctID is invalid expect %s but get %s", nfctID, matchAndReturnAddLiquidity.NfctID())
			return waitingContributions, deletedWaitingContributions, poolPairs, err
		}
		deletedWaitingContributions[metaData.PairHash()] = waitingContribution
		delete(waitingContributions, metaData.PairHash())
	}

	//TODO: Track status of contribution
	return waitingContributions, deletedWaitingContributions, poolPairs, nil
}

func (sp *stateProcessorV2) modifyParams(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	params Params,
) (Params, error) {
	if len(inst) != 4 {
		msg := fmt.Sprintf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		Logger.log.Errorf(msg)
		return params, errors.New(msg)
	}

	// unmarshal instructions content
	var actionData metadataPdexv3.ParamsModifyingContent
	err := json.Unmarshal([]byte(inst[3]), &actionData)
	if err != nil {
		msg := fmt.Sprintf("Could not unmarshal instruction content %v - Error: %v\n", inst[3], err)
		Logger.log.Errorf(msg)
		return params, err
	}

	modifyingStatus := inst[2]
	var reqTrackStatus int
	if modifyingStatus == metadataPdexv3.RequestAcceptedChainStatus {
		params = Params(actionData.Content)
		reqTrackStatus = metadataPdexv3.ParamsModifyingSuccessStatus
	} else {
		reqTrackStatus = metadataPdexv3.ParamsModifyingFailedStatus
	}

	modifyingReqStatus := metadataPdexv3.ParamsModifyingRequestStatus{
		Status:       reqTrackStatus,
		Pdexv3Params: metadataPdexv3.Pdexv3Params(actionData.Content),
	}
	modifyingReqStatusBytes, _ := json.Marshal(modifyingReqStatus)
	err = statedb.TrackPdexv3Status(
		stateDB,
		statedb.Pdexv3ParamsModifyingStatusPrefix(),
		[]byte(actionData.TxReqID.String()),
		modifyingReqStatusBytes,
	)
	if err != nil {
		Logger.log.Errorf("PDex Params Modifying: An error occurred while tracking shielding request tx - Error: %v", err)
	}

	return params, nil
}
