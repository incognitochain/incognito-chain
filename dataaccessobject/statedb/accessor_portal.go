package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
)

//======================  Redeem  ======================
func GetWaitingRedeemRequests(stateDB *StateDB, beaconHeight uint64) (map[string]*WaitingRedeemRequest, error) {
	waitingRedeemRequests := stateDB.getAllWaitingRedeemRequest(beaconHeight)
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
	if err != nil && err.(*StatedbError).GetErrorCode() != ErrCodeMessage[GetPortalStatusNotFoundError].Code {
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
	waitingRedeemRequests := stateDB.getAllCustodianStatePool(beaconHeight)
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

func GetOneCustodian(stateDB *StateDB, beaconHeight uint64, custodianAddress string) (*CustodianState, error) {
	key := GenerateCustodianStateObjectKey(beaconHeight, custodianAddress)
	custodianState, has, err := stateDB.getCustodianByKey(key)
	if err != nil {
		return nil, NewStatedbError(GetPortalStatusError, err)
	}
	if !has {
		return nil, NewStatedbError(GetPortalStatusError, fmt.Errorf("key with beacon height %+v, custodian address %+v not found", beaconHeight, custodianAddress))
	}

	return custodianState, nil
}

//======================  Exchange rate  ======================
func GetFinalExchangeRatesState(
	stateDB *StateDB,
	beaconHeight uint64,
) (map[string]*FinalExchangeRatesState, error) {
	finalExchangeRates := stateDB.getFinalExchangeRatesState(beaconHeight)
	return finalExchangeRates, nil
}

func GetFinalExchangeRatesByKey(stateDB *StateDB, beaconHeight uint64) (*FinalExchangeRatesState, error) {
	key := GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
	finalExchangeRates, has, err := stateDB.getFinalExchangeRatesByKey(key)
	if err != nil {
		return nil, NewStatedbError(GetPortalFinalExchangeRatesStateError, err)
	}
	if !has {
		return nil, NewStatedbError(GetPortalFinalExchangeRatesStateError, fmt.Errorf("key with beacon height %+v not found", beaconHeight))
	}

	return finalExchangeRates, nil
}

func StoreBulkFinalExchangeRatesState(
	stateDB *StateDB,
	beaconHeight uint64,
	finalExchangeRatesState map[string]*FinalExchangeRatesState) error {
	for _, exchangeRates := range finalExchangeRatesState {
		key := GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
		err := stateDB.SetStateObject(PortalFinalExchangeRatesStateObjectType, key, exchangeRates)
		if err != nil {
			return NewStatedbError(StoreFinalExchangeRatesStateError, err)
		}
	}
	return nil
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

func GetPortalLiquidationCustodianRunAwayStatus(stateDB *StateDB, redeemID string, custodianIncognitoAddress string) ([]byte, error) {
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

func GetPortalExpiredPortingRequestStatus(stateDB *StateDB, waitingPortingID string) ([]byte, error) {
	statusType := PortalExpiredPortingReqPrefix()
	statusSuffix := []byte(waitingPortingID)
	data, err := GetPortalStatus(stateDB, statusType, statusSuffix)
	if err != nil {
		return []byte{}, NewStatedbError(GetPortalLiquidationCustodianRunAwayStatusError, err)
	}

	return data, nil
}

func GetLiquidateExchangeRatesPool(
	stateDB *StateDB,
	beaconHeight uint64,
) (map[string]*LiquidateExchangeRatesPool, error) {
	liquidateExchangeRates := stateDB.getLiquidateExchangeRatesPool(beaconHeight)
	return liquidateExchangeRates, nil
}

func StoreBulkLiquidateExchangeRatesPool(
	stateDB *StateDB,
	beaconHeight uint64,
	liquidateExchangeRates map[string]*LiquidateExchangeRatesPool,
) error {
	for _, value := range liquidateExchangeRates {
		key := GeneratePortalLiquidateExchangeRatesPoolObjectKey(beaconHeight)
		err := stateDB.SetStateObject(PortalLiquidationExchangeRatesPoolObjectType, key, value)
		if err != nil {
			return NewStatedbError(StoreLiquidateExchangeRatesPoolError, err)
		}
	}
	return nil
}

func GetLiquidateExchangeRatesPoolByKey(stateDB *StateDB, beaconHeight uint64) (*LiquidateExchangeRatesPool, error) {
	key := GeneratePortalLiquidateExchangeRatesPoolObjectKey(beaconHeight)
	liquidateExchangeRates, has, err := stateDB.getLiquidateExchangeRatesPoolByKey(key)
	if err != nil {
		return nil, NewStatedbError(GetPortalLiquidationExchangeRatesPoolError, err)
	}
	if !has {
		return nil, NewStatedbError(GetPortalLiquidationExchangeRatesPoolError, fmt.Errorf("key with beacon height %+v not found", beaconHeight))
	}

	return liquidateExchangeRates, nil
}

//======================  Porting  ======================
func TrackPortalStateStatusMultiple(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte, beaconHeight uint64) error {
	key := GeneratePortalStatusObjectKey(statusType, statusSuffix)
	value := NewPortalStatusStateWithValue(statusType, statusSuffix, statusContent)

	dataaccessobject.Logger.Log.Infof("TrackPortalStateStatusMultiple [beaconHeight: %v] statusType: %+v, statusSuffix: %+v, value: %+v", beaconHeight, string(statusType), string(statusSuffix), value.ToString())

	err := stateDB.SetStateObject(PortalStatusObjectType, key, value)

	var errType int
	switch string(statusType) {
	case string(PortalLiquidationTpExchangeRatesStatusPrefix()):
		errType = StoreLiquidateTopPercentileExchangeRatesError
	case string(PortalLiquidationRedeemRequestStatusPrefix()):
		errType = StoreRedeemLiquidationExchangeRatesError
	case string(PortalLiquidationCustodianDepositStatusPrefix()):
		errType = StoreLiquidationCustodianDepositError
	case string(PortalPortingRequestStatusPrefix()):
		errType = StorePortalStatusError
	case string(PortalPortingRequestTxStatusPrefix()):
		errType = StorePortalTxStatusError
	case string(PortalExchangeRatesRequestStatusPrefix()):
		errType = StorePortalExchangeRatesStatusError
	case string(PortalCustodianWithdrawStatusPrefix()):
		errType = StorePortalCustodianWithdrawRequestStatusError
	default:
		errType = StorePortalStatusError
	}

	if err != nil {
		return NewStatedbError(errType, err)
	}

	return nil
}

func GetPortalStateStatusMultiple(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePortalStatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getPortalStatusByKey(key)

	var errType int
	switch string(statusType) {
	case string(PortalPortingRequestStatusPrefix()):
		errType = GetPortingRequestStatusError
	case string(PortalPortingRequestTxStatusPrefix()):
		errType = GetPortingRequestTxStatusError
	case string(PortalLiquidationTpExchangeRatesStatusPrefix()):
		errType = GetLiquidationTopPercentileExchangeRatesStatusError
	case string(PortalCustodianWithdrawStatusPrefix()):
		errType = GetPortalCustodianWithdrawStatusError
	default:
		errType = StorePortalStatusError
	}

	if err != nil {
		return []byte{}, NewStatedbError(errType, err)
	}

	if !has {
		return []byte{}, NewStatedbError(errType, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
	}

	return s.statusContent, nil
}

func IsPortingRequestIdExist(stateDB *StateDB, statusSuffix []byte) (bool, error) {
	key := GeneratePortalStatusObjectKey(PortalPortingRequestStatusPrefix(), statusSuffix)
	_, has, err := stateDB.getPortalStatusByKey(key)

	if err != nil {
		return false, NewStatedbError(GetPortingRequestStatusError, err)
	}

	if !has {
		return false, nil
	}

	return true, nil
}

//====================== Waiting Porting  ======================
// getCustodianPoolState gets custodian pool state at beaconHeight
func GetWaitingPortingRequests(
	stateDB *StateDB,
	beaconHeight uint64,
) (map[string]*WaitingPortingRequest, error) {
	waitingPortingRequestList := stateDB.getWaitingPortingRequests(beaconHeight)
	return waitingPortingRequestList, nil
}

// StoreWaitingRedeemRequests stores waiting redeem requests at beaconHeight
func StoreBulkWaitingPortingRequests(
	stateDB *StateDB,
	beaconHeight uint64,
	waitingPortingRequest map[string]*WaitingPortingRequest) error {
	for _, items := range waitingPortingRequest {
		key := GeneratePortalWaitingPortingRequestObjectKey(beaconHeight, items.UniquePortingID())
		err := stateDB.SetStateObject(PortalWaitingPortingRequestObjectType, key, items)
		if err != nil {
			return NewStatedbError(StoreWaitingPortingRequestError, err)
		}
	}
	return nil
}

func StoreWaitingPortingRequests(stateDB *StateDB, beaconHeight uint64, portingRequestId string, statusContent *WaitingPortingRequest) error {
	key := GeneratePortalWaitingPortingRequestObjectKey(beaconHeight, portingRequestId)
	err := stateDB.SetStateObject(PortalWaitingPortingRequestObjectType, key, statusContent)
	if err != nil {
		return NewStatedbError(StoreWaitingPortingRequestError, err)
	}

	return nil
}

func DeleteWaitingPortingRequest(stateDB *StateDB, deletedWaitingPortingRequests map[string]*WaitingPortingRequest) {
	for key, _ := range deletedWaitingPortingRequests {
		keyHash := common.Hash{}
		copy(keyHash[:], key)
		stateDB.MarkDeleteStateObject(PortalWaitingPortingRequestObjectType, keyHash)
	}
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
		return []byte{}, NewStatedbError(GetPortalStatusNotFoundError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), string(statusSuffix)))
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
// GetPortalRewardsByBeaconHeight gets portal reward state at beaconHeight
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
			info.rewards,
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

func GetLockedCollateralStateByBeaconHeight(
	stateDB *StateDB,
	beaconHeight uint64,
) (*LockedCollateralState, error) {
	lockedCollateralState, _, err := stateDB.getLockedCollateralState(beaconHeight)
	if err != nil {
		return nil, NewStatedbError(GetLockedCollateralStateError, err)
	}
	return lockedCollateralState, nil
}

// StoreWaitingRedeemRequests stores waiting redeem requests at beaconHeight
func StoreLockedCollateralState(
	stateDB *StateDB,
	beaconHeight uint64,
	lockedCollateralState *LockedCollateralState) error {
	key := GenerateLockedCollateralStateObjectKey(beaconHeight)
	err := stateDB.SetStateObject(LockedCollateralStateObjectType, key, lockedCollateralState)
	if err != nil {
		return NewStatedbError(StorePortalRewardError, err)
	}

	return nil
}

//======================  Feature reward  ======================
func StoreRewardFeatureState(
	stateDB *StateDB,
	featureName string,
	rewardInfo []*RewardInfoDetail,
	epoch uint64) error {
	key := GenerateRewardFeatureStateObjectKey(featureName, epoch)
	value := NewRewardFeatureStateWithValue(rewardInfo)

	err := stateDB.SetStateObject(RewardFeatureStateObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreRewardFeatureError, err)
	}

	return nil
}

func GetRewardFeatureAmountByTokenID(
	stateDB *StateDB,
	tokenID string,
	epoch uint64) (uint64, error) {

	totalAmount := uint64(0)
	// reset for portal reward
	allRewardFeature, err := GetAllRewardFeatureState(stateDB, epoch)
	if err != nil {
		return uint64(0), NewStatedbError(GetRewardFeatureAmountByTokenIDError, err)
	}
	totalRewards := allRewardFeature.GetTotalRewards()
	for i := 0; i < len(totalRewards); i++ {
		if totalRewards[i].GetTokenID() == tokenID {
			totalAmount = totalRewards[i].GetAmount()
			break
		}
	}

	return totalAmount, nil
}

func GetRewardFeatureStateByFeatureName(
	stateDB *StateDB,
	featureName string,
	epoch uint64) (*RewardFeatureState, error) {
	result, _, err := stateDB.getFeatureRewardByFeatureName(featureName, epoch)
	if err != nil {
		return nil, NewStatedbError(GetRewardFeatureError, err)
	}

	return result, nil
}

func GetAllRewardFeatureState(
	stateDB *StateDB, epoch uint64) (*RewardFeatureState, error) {
	result, _, err := stateDB.getAllFeatureRewards(epoch)
	if err != nil {
		return nil, NewStatedbError(GetAllRewardFeatureError, err)
	}

	return result, nil
}
