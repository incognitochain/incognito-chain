package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPortalReward(
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	db := blockchain.GetDatabase()

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
		for incAddrCus, amount := range actionData.Receivers {
			cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, incAddrCus)
			custodianState := currentPortalState.CustodianPoolState[cusStateKey]
			if custodianState == nil {
				Logger.log.Errorf("[processPortalReward] Can not get custodian state with key %v", cusStateKey)
				continue
			}

			custodianState.RewardAmount += amount
		}

		// store reward at beacon height into db
		portalRewardKey := lvdb.NewPortalRewardKey(beaconHeight + 1)
		err = db.StorePortalRewardByBeaconHeight(
			[]byte(portalRewardKey),
			[]byte(instructions[3]),
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
	beaconHeight uint64, instructions []string,
	currentPortalState *CurrentPortalState) error {

	db := blockchain.GetDatabase()

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
		cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, actionData.CustodianAddressStr)
		custodianState := currentPortalState.CustodianPoolState[cusStateKey]
		if custodianState == nil {
			Logger.log.Errorf("[processPortalWithdrawReward] Can not get custodian state with key %v", cusStateKey)
			return nil
		}
		custodianState.RewardAmount = 0

		// track request withdraw portal reward
		portalRewardKey := lvdb.NewPortalReqWithdrawRewardKey(beaconHeight + 1, actionData.CustodianAddressStr)
		portalReqRewardStatus := metadata.PortalRequestWithdrawRewardStatus{
			Status: common.PortalReqWithdrawRewardAcceptedStatus,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RewardAmount: actionData.RewardAmount,
			TxReqID: actionData.TxReqID,

		}
		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
		err = db.TrackPortalReqWithdrawReward(
			[]byte(portalRewardKey),
			portalReqRewardStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalReqUnlockCollateralRejectedChainStatus {
		// track request withdraw portal reward
		portalRewardKey := lvdb.NewPortalReqWithdrawRewardKey(beaconHeight + 1, actionData.CustodianAddressStr)
		portalReqRewardStatus := metadata.PortalRequestWithdrawRewardStatus{
			Status: common.PortalReqWithdrawRewardRejectedStatus,
			CustodianAddressStr: actionData.CustodianAddressStr,
			TxReqID: actionData.TxReqID,
		}
		portalReqRewardStatusBytes, _ := json.Marshal(portalReqRewardStatus)
		err = db.TrackPortalReqWithdrawReward(
			[]byte(portalRewardKey),
			portalReqRewardStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	}

	return nil
}
