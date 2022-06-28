package blsbft

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

/*
commit this propose block
- not yet commmit
- receive 2/3 vote
*/
func (a *actorV3) maybeCommit() {
	for _, proposeBlockInfo := range a.receiveBlockByHash {
		if a.currentTimeSlot == common.CalculateTimeSlot(proposeBlockInfo.block.GetProposeTime()) &&
			!proposeBlockInfo.IsCommitted {
			//no new pre vote
			if !proposeBlockInfo.HasNewVote {
				return
			}
			//has majority votes
			if proposeBlockInfo.ValidVotes > 2*len(proposeBlockInfo.SigningCommittees)/3 {
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

			previousProposeBlockInfo = a.validateVote(previousProposeBlockInfo)

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
