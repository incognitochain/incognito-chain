package statedb

import (
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


