package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

//======================  Redeem  ======================
func GetWaitingRedeemRequests(stateDB *StateDB, beaconHeight uint64) (map[string]*WaitingRedeemRequest, error) {
	waitingRedeemRequests := stateDB.getAllWaitingRedeemRequest()
	return waitingRedeemRequests, nil
}

// StoreWaitingRedeemRequests stores waiting redeem requests at beaconHeight
func StoreWaitingRedeemRequests(
	stateDB *StateDB,
	beaconHeight uint64,
	waitingRedeemReqs map[string]*WaitingRedeemRequest) error {
	for _, waitingReq := range waitingRedeemReqs {
		key := GenerateWaitingRedeemRequestObjectKey(beaconHeight, waitingReq.uniqueRedeemID)
		value := NewWaitingRedeemRequestWithValue(
			waitingReq.uniqueRedeemID,
			waitingReq.tokenID,
			waitingReq.redeemerAddress,
			waitingReq.redeemerRemoteAddress,
			waitingReq.redeemAmount,
			waitingReq.custodians,
			waitingReq.redeemFee,
			waitingReq.beaconHeight,
			waitingReq.txReqID,
			)
		err := stateDB.SetStateObject(WaitingRedeemRequestObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreWaitingRedeemRequestError, err)
		}
	}

	return nil
}

func DeleteWaitingRedeemRequest(stateDB *StateDB, deletedWaitingRedeemRequests map[string]*WaitingRedeemRequest) {
	for key, _ := range deletedWaitingRedeemRequests {
		keyHash := common.Hash{}
		copy(keyHash[:], key)
		stateDB.MarkDeleteStateObject(WaitingRedeemRequestObjectType, keyHash)
	}
}

func StorePortalRedeemRequestStatus(stateDB *StateDB, redeemID string, statusContent []byte) error {
	statusType := PortalRedeemRequestStatusPrefix()
	statusSuffix := []byte(redeemID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalRedeemRequestStatusError, err)
	}

	return nil
}

func GetPortalRedeemRequestStatus(stateDB *StateDB, redeemID string) ([]byte, error) {
	statusType := PortalRedeemRequestStatusPrefix()
	statusSuffix := []byte(redeemID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalRedeemRequestStatusError, err)
	}

	return data, nil
}

func StorePortalRedeemRequestByTxIDStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalRedeemRequestStatusByTxReqIDPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalRedeemRequestByTxIDStatusError, err)
	}

	return nil
}

func GetPortalRedeemRequestByTxIDStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalRedeemRequestStatusByTxReqIDPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalRedeemRequestByTxIDStatusError, err)
	}

	return data, nil
}


//======================  Custodian pool  ======================
// getCustodianPoolState gets custodian pool state at beaconHeight
func GetCustodianPoolState(
	stateDB *StateDB,
	beaconHeight uint64,
) (map[string]*CustodianState, error) {
	waitingRedeemRequests := stateDB.getAllCustodianStatePool()
	return waitingRedeemRequests, nil
}

// StoreWaitingRedeemRequests stores waiting redeem requests at beaconHeight
func StoreCustodianState(
	stateDB *StateDB,
	beaconHeight uint64,
	custodians map[string]*CustodianState) error {
	for _, cus := range custodians {
		key := GenerateCustodianStateObjectKey(beaconHeight, cus.incognitoAddress)
		value := NewCustodianStateWithValue(
			cus.incognitoAddress,
			cus.totalCollateral,
			cus.freeCollateral,
			cus.holdingPubTokens,
			cus.lockedAmountCollateral,
			cus.remoteAddresses,
			cus.rewardAmount,
		)
		err := stateDB.SetStateObject(CustodianStateObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreCustodianStateError, err)
		}
	}

	return nil
}

func DeleteCustodianState(stateDB *StateDB, deletedCustodianStates map[string]*CustodianState) {
	for key, _ := range deletedCustodianStates {
		keyHash := common.Hash{}
		copy(keyHash[:], key)
		stateDB.MarkDeleteStateObject(CustodianStateObjectType, keyHash)
	}
}

func StoreCustodianDepositStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalCustodianDepositStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalCustodianDepositStatusError, err)
	}

	return nil
}

func GetCustodianDepositStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalCustodianDepositStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalCustodianDepositStatusError, err)
	}

	return data, nil
}

//======================  Exchange rate  ======================
func GetFinalExchangeRatesState(
	stateDB *StateDB,
	beaconHeight uint64,
) (map[string]*FinalExchangeRatesState, error) {
	finalExchangeRates := make(map[string]*FinalExchangeRatesState)

	allFinalExchangeRatesState := stateDB.getAllFinalExchangeRatesState()
	for _, item  := range allFinalExchangeRatesState {
		key := GenerateFinalExchangeRatesStateObjectKey(beaconHeight)
		value := NewFinalExchangeRatesStateWithValue(item.Rates())
		finalExchangeRates[key.String()] = value
	}
	return finalExchangeRates, nil
}

func StoreFinalExchangeRatesState(
	stateDB *StateDB,
	beaconHeight uint64,
	finalExchangeRatesState map[string]*FinalExchangeRatesState) error {
	for _, exchangeRates := range finalExchangeRatesState {
		key := GenerateFinalExchangeRatesStateObjectKey(beaconHeight)
		value := NewFinalExchangeRatesStateWithValue(exchangeRates.Rates())

		err := stateDB.SetStateObject(FinalExchangeRatesStateObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreFinalExchangeRatesStateError, err)
		}
	}
	return nil
}

func TrackExchangeRatesRequestStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePDEStatusObjectKey(statusType, statusSuffix)
	value := NewPDEStatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(PDEStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(TrackPDEStatusError, err)
	}
	return nil
}

func GetItemPortalByKey(key []byte) ([]byte, error) {
	/*itemRecord, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.GetItemPortalByKeyError, dbErr)
	}

	if itemRecord == nil {
		return nil, nil
	}

	return itemRecord, nil*/
	return  nil, nil
}

//======================  Liquidation  ======================
func StorePortalLiquidationCustodianRunAwayStatus(stateDB *StateDB, redeemID string, custodianIncognitoAddress string, statusContent []byte) error {
	statusType := PortalLiquidateCustodianRunAwayPrefix()
	statusSuffix := append([]byte(redeemID), []byte(custodianIncognitoAddress)...)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalLiquidationCustodianRunAwayStatusError, err)
	}

	return nil
}

func GetPortalLiquidationCustodianRunAwayStatus(stateDB *StateDB, redeemID string, custodianIncognitoAddress string,) ([]byte, error) {
	statusType := PortalLiquidateCustodianRunAwayPrefix()
	statusSuffix := append([]byte(redeemID), []byte(custodianIncognitoAddress)...)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalLiquidationCustodianRunAwayStatusError, err)
	}

	return data, nil
}

func StorePortalExpiredPortingRequestStatus(stateDB *StateDB, waitingPortingID string, statusContent []byte) error {
	statusType := PortalExpiredPortingReqPrefix()
	statusSuffix := []byte(waitingPortingID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalLiquidationCustodianRunAwayStatusError, err)
	}

	return nil
}

func GetPortalExpiredPortingRequestStatus(stateDB *StateDB, redeemID string, custodianIncognitoAddress string,) ([]byte, error) {
	statusType := PortalExpiredPortingReqPrefix()
	statusSuffix := []byte(waitingPortingID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalLiquidationCustodianRunAwayStatusError, err)
	}

	return data, nil
}

//======================  Porting  ======================

// getCustodianPoolState gets custodian pool state at beaconHeight
func GetWaitingPortingRequests(
	stateDB *StateDB,
	beaconHeight uint64,
) (map[string]*WaitingPortingRequest, error) {
	//todo:
	return nil, nil
}

// StoreWaitingRedeemRequests stores waiting redeem requests at beaconHeight
func StoreWaitingPortingRequests(
	stateDB *StateDB,
	beaconHeight uint64,
	portingReqs map[string]*WaitingPortingRequest) error {
	//todo:
	return nil
}

//======================  Portal status  ======================

func StorePortalStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePortalStatusObjectKey(statusType, statusSuffix)
	value := NewPortalStatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(PortalStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePortalStatusError, err)
	}
	return nil
}

func GetPortalStatus(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePortalStatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getPortalStatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalStatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPortalStatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}
	return s.statusContent, nil
}

func StoreRequestPTokenStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalRequestPTokenStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalCustodianDepositStatusError, err)
	}

	return nil
}

func GetRequestPTokenStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalRequestPTokenStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalCustodianDepositStatusError, err)
	}

	return data, nil
}


func StorePortalRequestUnlockCollateralStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalRequestUnlockCollateralStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalRequestUnlockCollateralStatusError, err)
	}

	return nil
}

func GetPortalRequestUnlockCollateralStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalRequestUnlockCollateralStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalRequestUnlockCollateralStatusError, err)
	}

	return data, nil
}

//======================  Portal reward  ======================
// getCustodianPoolState gets custodian pool state at beaconHeight
func GetPortalRewardsByBeaconHeight(
	stateDB *StateDB,
	beaconHeight uint64,
) ([]*PortalRewardInfo, error) {
	portalRewards := stateDB.getPortalRewards(beaconHeight)
	return portalRewards, nil
}

// StoreWaitingRedeemRequests stores waiting redeem requests at beaconHeight
func StorePortalRewards(
	stateDB *StateDB,
	beaconHeight uint64,
	portalRewardInfos []*PortalRewardInfo) error {
	for _, info := range portalRewardInfos {
		key := GeneratePortalRewardInfoObjectKey(beaconHeight, info.custodianIncAddr)
		value := NewPortalRewardInfoWithValue(
			info.custodianIncAddr,
			info.amount,
		)
		err := stateDB.SetStateObject(PortalRewardInfoObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePortalRewardError, err)
		}
	}

	return nil
}

func StorePortalRequestWithdrawRewardStatus(stateDB *StateDB, txID string, statusContent []byte) error {
	statusType := PortalRequestWithdrawRewardStatusPrefix()
	statusSuffix := []byte(txID)
	err := StorePortalStatus(stateDB, statusType, statusSuffix, statusContent)
	if err != nil {
		return NewStatedbError(StorePortalRequestWithdrawRewardStatusError, err)
	}

	return nil
}

func GetPortalRequestWithdrawRewardStatus(stateDB *StateDB, txID string) ([]byte, error) {
	statusType := PortalRequestWithdrawRewardStatusPrefix()
	statusSuffix := []byte(txID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalRequestWithdrawRewardStatusError, err)
	}

	return data, nil
}
