package portalprocess
//
//TODO:
//import (
//	"encoding/json"
//	"github.com/incognitochain/incognito-chain/common"
//	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
//	"github.com/incognitochain/incognito-chain/metadata"
//	"github.com/incognitochain/incognito-chain/portal"
//	metadata2 "github.com/incognitochain/incognito-chain/portal/metadata"
//	"strconv"
//)
//
//func (blockchain *BlockChain) processPortalReward(
//	stateDB *statedb.StateDB,
//	beaconHeight uint64, instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//
//	// unmarshal instructions content
//	var actionData metadata.PortalRewardContent
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
//		return nil
//	}
//
//	reqStatus := instructions[2]
//	metaType, _ := strconv.Atoi(instructions[0])
//	if reqStatus == "portalRewardInst" {
//		// update reward amount for custodian
//		UpdateCustodianRewards(currentPortalState, actionData.Rewards)
//
//		// at the end of epoch
//		if (beaconHeight+1)%blockchain.config.ChainParams.Epoch == 1 {
//			currentPortalState.LockedCollateralForRewards.Reset()
//		}
//
//		// update locked collateral for rewards base on holding public tokens
//		if metaType == metadata.PortalRewardMetaV3 {
//			UpdateLockedCollateralForRewardsV3(currentPortalState, portalParams)
//		} else if metaType == metadata.PortalRewardMeta {
//			UpdateLockedCollateralForRewards(currentPortalState, portalParams)
//		}
//
//		// store reward at beacon height into db
//		err = statedb.StorePortalRewards(
//			stateDB,
//			beaconHeight+1,
//			actionData.Rewards,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
//			return nil
//		}
//	} else {
//		Logger.log.Errorf("ERROR: Invalid status of instruction: %+v", reqStatus)
//		return nil
//	}
//
//	return nil
//}
//
//func (blockchain *BlockChain) processPortalWithdrawReward(
//	stateDB *statedb.StateDB,
//	beaconHeight uint64, instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//
//	// unmarshal instructions content
//	var actionData metadata2.PortalRequestWithdrawRewardContent
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
//		return nil
//	}
//
//	reqStatus := instructions[2]
//	if reqStatus == common.PortalReqWithdrawRewardAcceptedChainStatus {
//		// update reward amount of custodian
//		cusStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianAddressStr)
//		cusStateKeyStr := cusStateKey.String()
//		custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
//		if custodianState == nil {
//			Logger.log.Errorf("[processPortalWithdrawReward] Can not get custodian state with key %v", cusStateKey)
//			return nil
//		}
//		updatedRewardAmount := custodianState.GetRewardAmount()
//		updatedRewardAmount[actionData.TokenID.String()] = 0
//		currentPortalState.CustodianPoolState[cusStateKeyStr].SetRewardAmount(updatedRewardAmount)
//
//		// track request withdraw portal reward
//		portalReqRewardStatus := metadata2.PortalRequestWithdrawRewardStatus{
//			Status:              common.PortalReqWithdrawRewardAcceptedStatus,
//			CustodianAddressStr: actionData.CustodianAddressStr,
//			TokenID:             actionData.TokenID,
//			RewardAmount:        actionData.RewardAmount,
//			TxReqID:             actionData.TxReqID,
//		}
//		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
//		err = statedb.StorePortalRequestWithdrawRewardStatus(
//			stateDB,
//			actionData.TxReqID.String(),
//			portalReqRewardStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
//			return nil
//		}
//
//	} else if reqStatus == common.PortalReqUnlockCollateralRejectedChainStatus {
//		// track request withdraw portal reward
//		portalReqRewardStatus := metadata2.PortalRequestWithdrawRewardStatus{
//			Status:              common.PortalReqWithdrawRewardRejectedStatus,
//			CustodianAddressStr: actionData.CustodianAddressStr,
//			TokenID:             actionData.TokenID,
//			RewardAmount:        actionData.RewardAmount,
//			TxReqID:             actionData.TxReqID,
//		}
//		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
//		err = statedb.StorePortalRequestWithdrawRewardStatus(
//			stateDB,
//			actionData.TxReqID.String(),
//			portalReqRewardStatusBytes,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
//			return nil
//		}
//	}
//
//	return nil
//}
//
//func (blockchain *BlockChain) processPortalTotalCustodianReward(
//	stateDB *statedb.StateDB,
//	beaconHeight uint64, instructions []string,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams) error {
//
//	// unmarshal instructions content
//	var actionData metadata.PortalTotalCustodianReward
//	err := json.Unmarshal([]byte(instructions[3]), &actionData)
//	if err != nil {
//		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
//		return nil
//	}
//
//	reqStatus := instructions[2]
//	if reqStatus == "portalTotalRewardInst" {
//		epoch := beaconHeight / blockchain.config.ChainParams.Epoch
//		// store total custodian reward into db
//		err = statedb.StoreRewardFeatureState(
//			stateDB,
//			statedb.PortalRewardName,
//			actionData.Rewards,
//			epoch,
//		)
//		if err != nil {
//			Logger.log.Errorf("ERROR: an error occured while storing total custodian reward: %+v", err)
//			return nil
//		}
//	} else {
//		Logger.log.Errorf("ERROR: Invalid status of instruction: %+v", reqStatus)
//		return nil
//	}
//
//	return nil
//}
