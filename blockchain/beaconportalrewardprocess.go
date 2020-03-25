package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPortalReward(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	// unmarshal instructions content
	var actionData metadata.PortalRewardContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == "portalRewardInst" {
		// update reward amount for each custodian
		for _, receiver := range actionData.Rewards {
			cusStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, receiver.GetCustodianIncAddr())
			cusStateKeyStr := string(cusStateKey[:])
			custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
			if custodianState == nil {
				Logger.log.Errorf("[processPortalReward] Can not get custodian state with key %v", cusStateKey)
				continue
			}

			custodianState.SetRewardAmount(custodianState.GetRewardAmount() + receiver.GetAmount())
		}

		// store reward at beacon height into db
		err = statedb.StorePortalRewards(
			stateDB,
			beaconHeight + 1,
			actionData.Rewards,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	} else {
		Logger.log.Errorf("ERROR: Invalid status of instruction: %+v", reqStatus)
		return nil
	}

	return nil
}

func (blockchain *BlockChain) processPortalWithdrawReward(
	stateDB *statedb.StateDB,
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	// unmarshal instructions content
	var actionData metadata.PortalRequestWithdrawRewardContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v\n", err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalReqWithdrawRewardAcceptedChainStatus {
		// update reward amount of custodian
		cusStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, actionData.CustodianAddressStr)
		cusStateKeyStr := string(cusStateKey[:])
		custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
		if custodianState == nil {
			Logger.log.Errorf("[processPortalWithdrawReward] Can not get custodian state with key %v", cusStateKey)
			return nil
		}
		custodianState.SetRewardAmount( 0)

		// track request withdraw portal reward
		portalReqRewardStatus := metadata.PortalRequestWithdrawRewardStatus{
			Status: common.PortalReqWithdrawRewardAcceptedStatus,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RewardAmount: actionData.RewardAmount,
			TxReqID: actionData.TxReqID,
		}
		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
		err = statedb.StorePortalRequestWithdrawRewardStatus(
			stateDB,
			actionData.TxReqID.String(),
			portalReqRewardStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalReqUnlockCollateralRejectedChainStatus {
		// track request withdraw portal reward
		portalReqRewardStatus := metadata.PortalRequestWithdrawRewardStatus{
			Status: common.PortalReqWithdrawRewardRejectedStatus,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RewardAmount: actionData.RewardAmount,
			TxReqID: actionData.TxReqID,
		}
		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
		err = statedb.StorePortalRequestWithdrawRewardStatus(
			stateDB,
			actionData.TxReqID.String(),
			portalReqRewardStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	}

	return nil
}
