package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPortalUnlockOverRateCollaterals(
	portalStateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	// parse instruction
	var unlockOverRateCollateralsContent metadata.PortalUnlockOverRateCollateralsContent
	err := json.Unmarshal([]byte(instructions[3]), &unlockOverRateCollateralsContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling content string of portal unlock over rate collaterals instruction: %+v", err)
		return nil
	}

	reqStatus := instructions[2]
	Logger.log.Infof("Portal unlock over rate collaterals, data input: %+v, status: %+v", unlockOverRateCollateralsContent, reqStatus)

	switch reqStatus {
	case common.PortalCusUnlockOverRateCollateralsAcceptedChainStatus:
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(unlockOverRateCollateralsContent.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		listTokensWithValue := unlockOverRateCollateralsContent.UnlockedAmounts
		unlockPrvAmount := listTokensWithValue[common.PRVIDStr]
		delete(listTokensWithValue, common.PRVIDStr)
		err = updateCustodianStateUnlockOverRateCollaterals(currentPortalState.CustodianPoolState[custodianStateKeyStr], unlockPrvAmount, listTokensWithValue, unlockOverRateCollateralsContent.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while updateCustodianStateUnlockOverRateCollaterals: %+v", err)
			return nil
		}

		//save db
		newUnlockOverRateCollaterals := metadata.NewUnlockOverRateCollateralsRequestStatus(
			common.PortalUnlockOverRateCollateralsAcceptedStatus,
			unlockOverRateCollateralsContent.CustodianAddressStr,
			unlockOverRateCollateralsContent.TokenID,
			unlockOverRateCollateralsContent.UnlockedAmounts,
		)

		newUnlockOverRateCollateralsStatusBytes, _ := json.Marshal(newUnlockOverRateCollaterals)
		err = statedb.StorePortalUnlockOverRateCollaterals(
			portalStateDB,
			unlockOverRateCollateralsContent.TxReqID.String(),
			newUnlockOverRateCollateralsStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("ERROR: Save UnlockOverRateCollaterals error: %+v", err)
			return nil
		}

	case common.PortalCusUnlockOverRateCollateralsRejectedChainStatus:
		//save db
		newUnlockOverRateCollaterals := metadata.NewUnlockOverRateCollateralsRequestStatus(
			common.PortalUnlockOverRateCollateralsRejectedStatus,
			unlockOverRateCollateralsContent.CustodianAddressStr,
			unlockOverRateCollateralsContent.TokenID,
			map[string]uint64{},
		)

		newUnlockOverRateCollateralsStatusBytes, _ := json.Marshal(newUnlockOverRateCollaterals)
		err = statedb.StorePortalUnlockOverRateCollaterals(
			portalStateDB,
			unlockOverRateCollateralsContent.TxReqID.String(),
			newUnlockOverRateCollateralsStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: Save UnlockOverRateCollaterals error: %+v", err)
			return nil
		}
	}

	return nil
}
