package blsbft

import (
	"math"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

/*
commit this propose block
- not yet commmit
- receive 2/3 vote
*/
func (a *actorV3) maybeCommit() {

	for _, proposeBlockInfo := range a.receiveBlockByHash {
		if proposeBlockInfo.block == nil {
			continue
		}
		previousView := a.chain.GetViewByHash(proposeBlockInfo.block.GetPrevHash())
		if previousView == nil {
			continue
		}
		if a.currentTimeSlot != previousView.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) ||
			proposeBlockInfo.IsCommitted {
			continue
		}
		if (proposeBlockInfo.block.GetVersion() >= int(config.Param().FeatureVersion[config.DELEGATION_REWARD])) && (a.chainID == common.BeaconChainID) {
			bcView, ok := previousView.(*blockchain.BeaconBestState)
			if !ok {
				a.logger.Error(errors.Errorf("Can not convert view %v to beacon best state", previousView.GetHash().String()))
				continue
			}
			beaconConsensusStateDB := bcView.GetBeaconConsensusStateDB()
			committeeKeyStringList, err := incognitokey.CommitteeKeyListToString(proposeBlockInfo.SigningCommittees)
			if err != nil {
				a.logger.Error(errors.Errorf("Can not convert committee key list to string, block %v", proposeBlockInfo.block.Hash().String()))
				continue
			}
			votingPower := make([]uint64, len(committeeKeyStringList))
			committeeData := statedb.GetCommitteeData(beaconConsensusStateDB)
			if committeeData == nil {
				a.logger.Error(errors.Errorf("Can not get committee data at beacon height %v, beacon view hash %v", previousView.GetBeaconHeight(), previousView.GetHash().String()))
				continue
			}
			for idx, cPK := range committeeKeyStringList {
				stakerInfo, has := committeeData.BeginEpochInfo[cPK]
				if !has {
					votingPower[idx] = 1
				} else {
					votingPower[idx] = uint64(math.Sqrt(float64(stakerInfo.Score)))
				}
			}
			if proposeBlockInfo.ValidateMajorityVotingPower(votingPower) && proposeBlockInfo.ValidateFixNodeMajority() {
				a.logger.Infof("Process Block With enough votes, %+v, has %+v, expect > %+v (from total %v)",
					proposeBlockInfo.block.FullHashString(), proposeBlockInfo.ValidVotes, 2*len(proposeBlockInfo.SigningCommittees)/3, len(proposeBlockInfo.SigningCommittees))
				a.commitBlock(proposeBlockInfo)
				proposeBlockInfo.IsCommitted = true
			}
		} else {
			//has majority votes
			if proposeBlockInfo.ValidVotes > 2*len(proposeBlockInfo.SigningCommittees)/3 && proposeBlockInfo.ValidateFixNodeMajority() {
				a.logger.Infof("Process Block With enough votes, %+v, has %+v, expect > %+v (from total %v)",
					proposeBlockInfo.block.FullHashString(), proposeBlockInfo.ValidVotes, 2*len(proposeBlockInfo.SigningCommittees)/3, len(proposeBlockInfo.SigningCommittees))
				a.commitBlock(proposeBlockInfo)
				proposeBlockInfo.IsCommitted = true
			}
		}
	}
}

func (a *actorV3) commitBlock(v *ProposeBlockInfo) error {
	validationData, err := a.createBLSAggregatedSignatures(v.SigningCommittees, v.block.ProposeHash(), v.block.GetValidationField(), v.Votes)
	if err != nil {
		return err
	}
	isInsertWithPreviousData := false
	v.block.(BlockValidation).AddValidationField(validationData)
	// validate and add previous block validation data
	previousBlock, _ := a.chain.GetBlockByHash(v.block.GetPrevHash())
	if previousBlock != nil {
		if previousProposeBlockInfo, ok := a.GetReceiveBlockByHash(previousBlock.ProposeHash().String()); ok &&
			previousProposeBlockInfo != nil && previousProposeBlockInfo.block != nil {

			a.validateVote(previousProposeBlockInfo)

			rawPreviousValidationData, err := a.createBLSAggregatedSignatures(
				previousProposeBlockInfo.SigningCommittees,
				previousProposeBlockInfo.block.ProposeHash(),
				previousProposeBlockInfo.block.GetValidationField(),
				previousProposeBlockInfo.Votes)
			if err != nil {
				a.logger.Error("Create BLS Aggregated Signature for previous block propose info, height ", previousProposeBlockInfo.block.GetHeight(), " error", err)
			} else {
				previousProposeBlockInfo.block.(BlockValidation).AddValidationField(rawPreviousValidationData)
				if err := a.chain.InsertAndBroadcastBlockWithPrevValidationData(v.block, rawPreviousValidationData); err != nil {
					return err
				}
				isInsertWithPreviousData = true
				previousValidationData, _ := consensustypes.DecodeValidationData(rawPreviousValidationData)
				a.logger.Infof("Block %+v broadcast with previous block %+v, previous block number of signatures %+v",
					v.block.GetHeight(), previousProposeBlockInfo.block.GetHeight(), len(previousValidationData.ValidatiorsIdx))
			}
		}
	} else {
		a.logger.Info("Cannot find block by hash", v.block.GetPrevHash().String())
	}

	if !isInsertWithPreviousData {
		if err := a.chain.InsertBlock(v.block, true); err != nil {
			return err
		}
	}
	loggedCommittee, _ := incognitokey.CommitteeKeyListToString(v.SigningCommittees)
	a.logger.Infof("Successfully Insert Block \n "+
		"ChainID %+v | Height %+v, Hash %+v, Version %+v \n"+
		"Committee %+v", a.chain, v.block.GetHeight(), v.block.FullHashString(), v.block.GetVersion(), loggedCommittee)
	return nil
}
