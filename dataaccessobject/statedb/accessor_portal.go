package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
)

//// prefix key for portal
//var (
//	CustodianStatePrefix                = []byte("custodianstate-")
//	PortalPortingRequestsPrefix         = []byte("portalportingrequest-")
//	PortalPortingRequestsTxPrefix       = []byte("portalportingrequesttx-")
//	PortalExchangeRatesPrefix           = []byte("portalexchangeratesrequest-")
//	PortalFinalExchangeRatesPrefix      = []byte("portalfinalexchangerates-")
//	PortalCustodianStatePrefix          = []byte("portalcustodianstate-")
//	PortalCustodianDepositPrefix        = []byte("portalcustodiandeposit-")
//	PortalWaitingPortingRequestsPrefix  = []byte("portalwaitingportingrequest-")
//	PortalRequestPTokensPrefix          = []byte("portalrequestptokens-")
//	PortalWaitingRedeemRequestsPrefix   = []byte("portalwaitingredeemrequest-")
//	PortalRedeemRequestsPrefix          = []byte("portalredeemrequest-")
//	PortalRedeemRequestsByTxReqIDPrefix = []byte("portalredeemrequestbytxid-")
//	PortalRequestUnlockCollateralPrefix = []byte("portalrequestunlockcollateral-")
//	PortalCustodianWithdrawPrefix       = []byte("portalcustodianwithdraw-")
//
//	// liquidation in portal
//	PortalLiquidateCustodianPrefix                  = []byte("portalliquidatecustodian-")
//	PortalLiquidateTopPercentileExchangeRatesPrefix = []byte("portalliquidatetoppercentileexchangerates-")
//	PortalLiquidateExchangeRatesPrefix              = []byte("portalliquidateexchangerates-")
//	PortalLiquidationCustodianDepositPrefix         = []byte("portalliquidationcustodiandepsit-")
//
//	PortalExpiredPortingReqPrefix = []byte("portalexpiredportingreq-")
//
//	// reward in portal
//	PortalRewardByBeaconHeightPrefix  = []byte("portalreward-")
//	PortalRequestWithdrawRewardPrefix = []byte("portalrequestwithdrawreward-")
//
//	// Relaying
//	RelayingBNBHeaderStatePrefix = []byte("relayingbnbheaderstate-")
//	RelayingBNBHeaderChainPrefix = []byte("relayingbnbheaderchain-")
//)
//
//type RemoteAddress struct {
//	PTokenID string
//	Address  string
//}
//
//type CustodianState struct {
//	IncognitoAddress       string
//	TotalCollateral        uint64            // prv
//	FreeCollateral         uint64            // prv
//	HoldingPubTokens       map[string]uint64 // tokenID : amount
//	LockedAmountCollateral map[string]uint64 // tokenID : amount
//	RemoteAddresses        []RemoteAddress   // tokenID : address
//	RewardAmount           uint64            // reward in prv
//}
//
//type MatchingPortingCustodianDetail struct {
//	IncAddress             string
//	RemoteAddress          string
//	Amount                 uint64
//	LockedAmountCollateral uint64
//	RemainCollateral       uint64
//}
//
//type MatchingRedeemCustodianDetail struct {
//	IncAddress    string
//	RemoteAddress string
//	Amount        uint64
//}
//
//type PortingRequest struct {
//	UniquePortingID string
//	TxReqID         common.Hash
//	TokenID         string
//	PorterAddress   string
//	Amount          uint64
//	Custodians      []*MatchingPortingCustodianDetail
//	PortingFee      uint64
//	Status          int
//	BeaconHeight    uint64
//}
//
//type RedeemRequest struct {
//	UniqueRedeemID        string
//	TxReqID               common.Hash
//	TokenID               string
//	RedeemerAddress       string
//	RedeemerRemoteAddress string
//	RedeemAmount          uint64
//	Custodians            []*MatchingRedeemCustodianDetail
//	RedeemFee             uint64
//	BeaconHeight          uint64
//}
//
//type ExchangeRatesRequest struct {
//	SenderAddress string
//	Rates         []*ExchangeRateInfo
//}
//
//type FinalExchangeRatesDetail struct {
//	Amount uint64
//}
//
//type FinalExchangeRates struct {
//	Rates map[string]FinalExchangeRatesDetail
//}
//
//type CustodianWithdrawRequest struct {
//	PaymentAddress                string
//	Amount                        uint64
//	Status                        int
//	RemainCustodianFreeCollateral uint64
//}
//
//type LiquidateTopPercentileExchangeRatesDetail struct {
//	TPKey                    int
//	TPValue                  int
//	HoldAmountFreeCollateral uint64
//	HoldAmountPubToken       uint64
//}
//
//type LiquidateTopPercentileExchangeRates struct {
//	CustodianAddress string
//	Status           byte
//	Rates            map[string]LiquidateTopPercentileExchangeRatesDetail //ptoken | detail
//}
//
//type LiquidateExchangeRatesDetail struct {
//	HoldAmountFreeCollateral uint64
//	HoldAmountPubToken       uint64
//}
//
//type LiquidateExchangeRates struct {
//	Rates map[string]LiquidateExchangeRatesDetail //ptoken | detail
//}
//
//type RedeemLiquidateExchangeRates struct {
//	TxReqID               common.Hash
//	TokenID               string
//	RedeemerAddress       string
//	RedeemerRemoteAddress string
//	RedeemAmount          uint64
//	RedeemFee             uint64
//	Status                byte
//	TotalPTokenReceived   uint64
//}
//
//type LiquidationCustodianDeposit struct {
//	TxReqID                common.Hash
//	IncogAddressStr        string
//	PTokenId               string
//	DepositAmount          uint64
//	FreeCollateralSelected bool
//	Status                 byte
//}
//
//type PortalRewardInfo struct {
//	CustodianIncAddr string
//	Amount           uint64
//}
//
//type ExchangeRateInfo struct {
//	PTokenID string
//	Rate     uint64
//}


//// GetCustodianDepositCollateralStatus returns custodian deposit status with deposit TxID
//func GetCustodianDepositCollateralStatus(stateDB *StateDB, txIDStr string) ([]byte, error) {
//	key := NewCustodianDepositKey(txIDStr)
//	custodianDepositStatusBytes, err :=  GetPortalRecordByKey(stateDB, []byte(key))
//	if err != nil  {
//		return nil, NewStatedbError(GetCustodianDepositStatusError, err)
//	}
//
//	return custodianDepositStatusBytes, err
//}
//
//// GetReqPTokenStatusByTxReqID returns request ptoken status with txReqID
//func GetReqPTokenStatusByTxReqID(stateDB *StateDB, txReqID string) ([]byte, error) {
//	key := append(PortalRequestPTokensPrefix, []byte(txReqID)...)
//	items, err := GetPortalRecordByKey(stateDB, []byte(key))
//	if err != nil {
//		return nil, NewStatedbError(GetReqPTokenStatusError, err)
//	}
//
//	return items, err
//}
//
////Porting request
//// StorePortingRequestItem store status of porting request by portingID
//func StorePortingRequestItem(stateDB *StateDB, keyId []byte, content interface{}) error {
//	contributionBytes, err := json.Marshal(content)
//	if err != nil {
//		return err
//	}
//
//	err = StorePortalRecord(stateDB, keyId, contributionBytes )
//	if err != nil {
//		return NewStatedbError(StorePortingRequestStateError, err)
//	}
//
//	return nil
//}




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
//todo: replace method getFinalExchangeRates (at: blockchain/portalutils.go)
func GetFinalExchangeRates(
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


//Withdraw



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


