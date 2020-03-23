package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"errors"
)

type RemoteAddress struct {
	PTokenID string
	Address  string
}

type CustodianState struct {
	IncognitoAddress       string
	TotalCollateral        uint64            // prv
	FreeCollateral         uint64            // prv
	HoldingPubTokens       map[string]uint64 // tokenID : amount
	LockedAmountCollateral map[string]uint64 // tokenID : amount
	RemoteAddresses        []RemoteAddress   // tokenID : address
	RewardAmount           uint64            // reward in prv
}

type MatchingPortingCustodianDetail struct {
	IncAddress             string
	RemoteAddress          string
	Amount                 uint64
	LockedAmountCollateral uint64
	RemainCollateral       uint64
}

type MatchingRedeemCustodianDetail struct {
	IncAddress    string
	RemoteAddress string
	Amount        uint64
}

type PortingRequest struct {
	UniquePortingID string
	TxReqID         common.Hash
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      []*MatchingPortingCustodianDetail
	PortingFee      uint64
	Status          int
	BeaconHeight    uint64
}

type RedeemRequest struct {
	UniqueRedeemID        string
	TxReqID               common.Hash
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	RedeemAmount          uint64
	Custodians            []*MatchingRedeemCustodianDetail
	RedeemFee             uint64
	BeaconHeight          uint64
}

type ExchangeRatesRequest struct {
	SenderAddress string
	Rates         []*ExchangeRateInfo
}

type FinalExchangeRatesDetail struct {
	Amount uint64
}

type FinalExchangeRates struct {
	Rates map[string]FinalExchangeRatesDetail
}

type CustodianWithdrawRequest struct {
	PaymentAddress                string
	Amount                        uint64
	Status                        int
	RemainCustodianFreeCollateral uint64
}

type LiquidateTopPercentileExchangeRatesDetail struct {
	TPKey int
	TPValue                  int
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateTopPercentileExchangeRates struct {
	CustodianAddress 	string
	Status				byte
	Rates        		map[string]LiquidateTopPercentileExchangeRatesDetail //ptoken | detail
}

type LiquidateExchangeRatesDetail struct {
	HoldAmountFreeCollateral uint64
	HoldAmountPubToken       uint64
}

type LiquidateExchangeRates struct {
	Rates map[string]LiquidateExchangeRatesDetail //ptoken | detail
}

type RedeemLiquidateExchangeRates struct {
	TxReqID               common.Hash
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	RedeemAmount          uint64
	RedeemFee             uint64
	Status                byte
	TotalPTokenReceived	  uint64
}

type LiquidationCustodianDeposit struct {
	TxReqID common.Hash
	IncogAddressStr string
	PTokenId string
	DepositAmount uint64
	FreeCollateralSelected bool
	Status byte
}

type PortalRewardInfo struct {
	CustodianIncAddr string
	Amount           uint64
}

type ExchangeRateInfo struct {
	PTokenID string
	Rate     uint64
}

// GetPortalRecordByKey gets record by key from stateDB
func GetPortalRecordByKey(stateDB *StateDB, key []byte) ([]byte, error) {
	return []byte{}, nil
}

// GetPortalRecordsByPrefix gets records by prefix from stateDB
func GetPortalRecordsByPrefix(stateDB *StateDB, prefix []byte) ([][]byte, error) {
	return [][]byte{}, nil
}

// StorePortalRecord stores record by key into stateDB
func StorePortalRecord(stateDB *StateDB, key []byte, value []byte) error {
	return nil
}

//A
//Generate Key
// Portal
var (
	CustodianStatePrefix                = []byte("custodianstate-")
	PortalPortingRequestsPrefix         = []byte("portalportingrequest-")
	PortalPortingRequestsTxPrefix       = []byte("portalportingrequesttx-")
	PortalExchangeRatesPrefix           = []byte("portalexchangeratesrequest-")
	PortalFinalExchangeRatesPrefix      = []byte("portalfinalexchangerates-")
	PortalCustodianStatePrefix          = []byte("portalcustodianstate-")
	PortalCustodianDepositPrefix        = []byte("portalcustodiandeposit-")
	PortalWaitingPortingRequestsPrefix  = []byte("portalwaitingportingrequest-")
	PortalRequestPTokensPrefix          = []byte("portalrequestptokens-")
	PortalWaitingRedeemRequestsPrefix   = []byte("portalwaitingredeemrequest-")
	PortalRedeemRequestsPrefix          = []byte("portalredeemrequest-")
	PortalRedeemRequestsByTxReqIDPrefix = []byte("portalredeemrequestbytxid-")
	PortalRequestUnlockCollateralPrefix = []byte("portalrequestunlockcollateral-")
	PortalCustodianWithdrawPrefix 		= []byte("portalcustodianwithdraw-")

	// liquidation in portal
	PortalLiquidateCustodianPrefix = []byte("portalliquidatecustodian-")
	PortalLiquidateTopPercentileExchangeRatesPrefix = []byte("portalliquidatetoppercentileexchangerates-")
	PortalLiquidateExchangeRatesPrefix = []byte("portalliquidateexchangerates-")
	PortalLiquidationCustodianDepositPrefix = []byte("portalliquidationcustodiandepsit-")

	PortalExpiredPortingReqPrefix = []byte("portalexpiredportingreq-")

	// reward in portal
	PortalRewardByBeaconHeightPrefix  = []byte("portalreward-")
	PortalRequestWithdrawRewardPrefix = []byte("portalrequestwithdrawreward-")

	// Relaying
	RelayingBNBHeaderStatePrefix = []byte("relayingbnbheaderstate-")
	RelayingBNBHeaderChainPrefix = []byte("relayingbnbheaderchain-")
)

func NewCustodianWithdrawRequestKey(txHash string) string {
	key := append(PortalCustodianWithdrawPrefix, []byte(txHash)...)
	return string(key)
}

func NewCustodianStateKey(beaconHeight uint64, custodianAddress string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalCustodianStatePrefix, beaconHeightBytes...)
	key = append(key, []byte(custodianAddress)...)
	return string(key)
}

func NewPortingRequestKey(uniquePortingID string) string {
	key := append(PortalPortingRequestsPrefix, []byte(uniquePortingID)...)
	return string(key) //prefix + uniqueId
}

func NewPortingRequestTxKey(txReqID string) string {
	key := append(PortalPortingRequestsTxPrefix, []byte(txReqID)...)
	return string(key) //prefix + txHash
}

func NewFinalExchangeRatesKey(beaconHeight uint64) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalFinalExchangeRatesPrefix, beaconHeightBytes...)
	key = append(key, []byte("portal")...)
	return string(key)
}

func NewExchangeRatesRequestKey(beaconHeight uint64, txId string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalExchangeRatesPrefix, beaconHeightBytes...)
	key = append(key, []byte(txId)...)
	return string(key)
}

func NewCustodianDepositKey(txID string) string {
	key := append(PortalCustodianDepositPrefix, []byte(txID)...)
	return string(key)
}

func NewWaitingPortingReqKey(beaconHeight uint64, portingID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalWaitingPortingRequestsPrefix, beaconHeightBytes...)
	key = append(key, []byte(portingID)...)
	return string(key)
}

// NewPortalReqPTokenKey creates key for tracking request pToken in portal
func NewPortalReqPTokenKey(txReqStr string) string {
	key := append(PortalRequestPTokensPrefix, []byte(txReqStr)...)
	return string(key)
}


//Custodian

// GetRemoteAddressByTokenID returns remote address for tokenID
func GetRemoteAddressByTokenID(addresses []RemoteAddress, tokenID string) (string, error) {
	for _, addr := range addresses {
		if addr.PTokenID == tokenID {
			return addr.Address, nil
		}
	}

	return "", errors.New("Can not found address with tokenID")
}

// GetCustodianDepositCollateralStatus returns custodian deposit status with deposit TxID
func GetCustodianDepositCollateralStatus(stateDB *StateDB, txIDStr string) ([]byte, error) {
	key := NewCustodianDepositKey(txIDStr)
	custodianDepositStatusBytes, err :=  GetPortalRecordByKey(stateDB, []byte(key))
	if err != nil  {
		return nil, NewStatedbError(GetCustodianDepositStatusError, err)
	}

	return custodianDepositStatusBytes, err
}

// GetReqPTokenStatusByTxReqID returns request ptoken status with txReqID
func GetReqPTokenStatusByTxReqID(stateDB *StateDB, txReqID string) ([]byte, error) {
	key := append(PortalRequestPTokensPrefix, []byte(txReqID)...)
	items, err := GetPortalRecordByKey(stateDB, []byte(key))
	if err != nil {
		return nil, NewStatedbError(GetReqPTokenStatusError, err)
	}

	return items, err
}

//Porting request
// StorePortingRequestItem store status of porting request by portingID
func StorePortingRequestItem(stateDB *StateDB, keyId []byte, content interface{}) error {
	contributionBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	err = StorePortalRecord(stateDB, keyId, contributionBytes )
	if err != nil {
		return NewStatedbError(StorePortingRequestStateError, err)
	}

	return nil
}

//Redeem

//Withdraw


//Liquidation

// B