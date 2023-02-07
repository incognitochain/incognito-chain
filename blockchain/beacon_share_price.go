package blockchain

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

//todo: big number calculation
func (bestView *BeaconBestState) CalculateDelegationSharePrice(bc *BlockChain, delegationReward uint64) ([][]string, error) {
	//check if start of epoch
	if bestView.GetHeight()%config.Param().EpochParam.NumberOfBlockInEpoch != 1 {
		return nil, errors.New("Not new epoch")
	}

	//get total value (Stake + delegation) of previous epoch
	beaconConsensusStateRootHash, err := bc.GetBeaconRootsHashFromBlockHeight(
		bestView.GetHeight() - config.Param().EpochParam.NumberOfBlockInEpoch - 1,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	stateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(bc.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	committeeData := statedb.GetCommitteeData(stateDB)
	committee := statedb.GetBeaconCommittee(stateDB)
	beaconCommitteeTotalStake := map[string]uint64{}
	beaconCommitteeReward := map[string]uint64{}
	totalAmount := uint64(0)
	for _, cpk := range committee {
		cpkStr, err := cpk.ToBase58()
		if err != nil {
			return nil, err
		}
		_, exist, _ := statedb.GetBeaconStakerInfo(stateDB, cpkStr)
		if !exist {
			return nil, fmt.Errorf("Cannot find cpk %v", cpkStr)
		}
		totalAmount += committeeData.BeginEpochInfo[cpkStr].DelegationAmount
		beaconCommitteeTotalStake[cpkStr] = committeeData.BeginEpochInfo[cpkStr].DelegationAmount
	}

	//if no delegation, then no distribute delegation reward
	if totalAmount == 0 {
		return nil, nil
	}

	for k, v := range beaconCommitteeTotalStake {
		beaconCommitteeReward[k] = delegationReward * v / totalAmount
	}

	//get reward with performance
	beaconConsensusStateRootHash, err = bc.GetBeaconRootsHashFromBlockHeight(
		bestView.GetHeight() - 1,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	stateDB, err = statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(bc.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	committeeData = statedb.GetCommitteeData(stateDB)
	for k, _ := range beaconCommitteeTotalStake {
		beaconCommitteeReward[k] = uint64(float64(beaconCommitteeReward[k]) *
			(float64(committeeData.LastEpochInfo[k].Performance) / float64(bestView.GetBeaconCommitteeState().(*committeestate.BeaconCommitteeStateV4).GetConfig().MAX_SCORE)))
	}

	//increase share price
	for cpkStr, _ := range beaconCommitteeTotalStake {
		price := GetBeaconSharePriceByEpoch(cpkStr, bestView.Epoch-1)
		price = price*beaconCommitteeReward[cpkStr] + beaconCommitteeTotalStake[cpkStr]/(beaconCommitteeTotalStake[cpkStr])
		//todo: create instruction
	}

}
