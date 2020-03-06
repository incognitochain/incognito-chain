package lvdb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math"
)

const (
	PortalTokenSymbolBTC = "BTC"
	PortalTokenSymbolBNB = "BNB"
	PortalTokenSymbolPRV = "PRV"
)

type CustodianState struct {
	IncognitoAddress       string
	TotalCollateral        uint64            // prv
	FreeCollateral         uint64            // prv
	HoldingPubTokens       map[string]uint64 // tokenSymbol : amount
	LockedAmountCollateral map[string]uint64
	RemoteAddresses        map[string]string // tokenSymbol : address
}

type MatchingPortingCustodianDetail struct {
	RemoteAddress          string
	Amount                 uint64
	LockedAmountCollateral uint64
	RemainCollateral       uint64
}

type MatchingRedeemCustodianDetail struct {
	RemoteAddress            string
	Amount                   uint64
}

type PortingRequest struct {
	UniquePortingID string
	TxReqID         common.Hash
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      map[string]MatchingPortingCustodianDetail // key : incogAddress
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
	Custodians            map[string]*MatchingRedeemCustodianDetail // key : incogAddress
	RedeemFee             uint64
	BeaconHeight          uint64
}

type ExchangeRatesRequest struct {
	SenderAddress string
	Rates         map[string]uint64
}

type FinalExchangeRatesDetail struct {
	Amount uint64
}

type FinalExchangeRates struct {
	Rates map[string]FinalExchangeRatesDetail
}

type CustodianWithdrawRequest struct {
	PaymentAddress string
	Amount uint64
	Status int
	RemainCustodianFreeCollateral uint64
}

func NewCustodianWithdrawRequestTxStateKey(txHash string) string {
	key := append(PortalCustodianWithdrawTxPrefix, []byte(txHash)...)
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

func NewExchangeRatesRequestKey (beaconHeight uint64, txId string) string {
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

func (db *db) GetAllRecordsPortalByPrefix(beaconHeight uint64, prefix []byte) ([][]byte, [][]byte, error) {
	keys := [][]byte{}
	values := [][]byte{}

	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	prefixByBeaconHeight := append(prefix, beaconHeightBytes...)

	//prefixByBeaconHeight:  prefix-beaconHeight-

	iter := db.lvdb.NewIterator(util.BytesPrefix(prefixByBeaconHeight), nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		keyBytes := make([]byte, len(key))
		valueBytes := make([]byte, len(value))
		copy(keyBytes, key)
		copy(valueBytes, value)
		keys = append(keys, keyBytes)
		values = append(values, valueBytes)
	}
	iter.Release()
	err := iter.Error()
	if err != nil && err != lvdberr.ErrNotFound {
		return keys, values, database.NewDatabaseError(database.GetAllRecordsByPrefixError, err)
	}
	return keys, values, nil
}

func (db *db) GetAllRecordsPortalByPrefixWithoutBeaconHeight(key []byte) ([][]byte, [][]byte, error) {
	keys := [][]byte{}
	values := [][]byte{}

	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		keyBytes := make([]byte, len(key))
		valueBytes := make([]byte, len(value))
		copy(keyBytes, key)
		copy(valueBytes, value)
		keys = append(keys, keyBytes)
		values = append(values, valueBytes)
	}
	iter.Release()
	err := iter.Error()

	if err != nil && err != lvdberr.ErrNotFound {
		return keys, values, database.NewDatabaseError(database.GetAllRecordsByPrefixError, err)
	}

	return keys, values, nil
}

func (db *db) TrackCustodianDepositCollateral(key []byte, content []byte) error {
	err := db.Put(key, content)
	if err != nil {
		return database.NewDatabaseError(database.TrackCustodianDepositError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

// GetCustodianDepositCollateralStatus returns custodian deposit status with deposit txid
func (db *db) GetCustodianDepositCollateralStatus(txIDStr string) ([]byte, error) {
	key := NewCustodianDepositKey(txIDStr)
	custodianDepositStatusBytes, err := db.lvdb.Get([]byte(key), nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.GetCustodianDepositStatusError, err)
	}

	return custodianDepositStatusBytes, err
}

func (db *db) TrackReqPTokens(key []byte, content []byte) error {
	err := db.Put(key, content)
	if err != nil {
		return database.NewDatabaseError(database.TrackCustodianDepositError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

// GetReqPTokenStatusByTxReqID returns request ptoken status with  txReqID
func (db *db) GetReqPTokenStatusByTxReqID(txReqID string) ([]byte, error) {
	key := append(PortalRequestPTokensPrefix, []byte(txReqID)...)

	reqPTokenStatusBytes, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.GetReqPTokenStatusError, err)
	}

	return reqPTokenStatusBytes, err
}

func (db *db) StorePortingRequestItem(keyId []byte, content interface{}) error {
	contributionBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	err = db.Put(keyId, contributionBytes)
	if err != nil {
		return database.NewDatabaseError(database.StorePortingRequestStateError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}

func (db *db) StoreExchangeRatesRequestItem(keyId []byte, content interface{}) error {
	contributionBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	err = db.Put(keyId, contributionBytes)
	if err != nil {
		return database.NewDatabaseError(database.StoreExchangeRatesRequestStateError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}

func (db *db) StoreFinalExchangeRatesItem(keyId []byte, content interface{}) error {
	contributionBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	err = db.Put(keyId, contributionBytes)
	if err != nil {
		return database.NewDatabaseError(database.StoreFinalExchangeRatesStateError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}

func (db *db) GetItemPortalByKey(key []byte) ([]byte, error) {
	itemRecord, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.GetItemPortalByKeyError, dbErr)
	}

	if itemRecord == nil {
		return nil, nil
	}

	return itemRecord, nil
}

func (db *db) GetPortingRequestStatusByPortingID(portingID string) (int, error) {
	key := NewPortingRequestKey(portingID)
	portingRequest, err := db.GetItemPortalByKey([]byte(key))

	if err != nil {
		return 0, err
	}

	var portingRequestResult PortingRequest

	if portingRequest == nil {
		return 0, nil
	}

	//get value via idx
	err = json.Unmarshal(portingRequest, &portingRequestResult)
	if err != nil {
		return 0, err
	}

	return portingRequestResult.Status, nil
}

func (db *db) UpdatePortingRequestStatus(portingID string, newStatus int) error {
	key := NewPortingRequestKey(portingID)
	portingRequest, err := db.GetItemPortalByKey([]byte(key))

	if err != nil {
		return err
	}

	var portingRequestResult PortingRequest

	if portingRequest == nil {
		return nil
	}

	//get value via idx
	err = json.Unmarshal(portingRequest, &portingRequestResult)
	if err != nil {
		return err
	}

	portingRequestResult.Status = newStatus

	//save porting request
	err = db.StorePortingRequestItem([]byte(key), portingRequestResult)
	if err != nil {
		return err
	}

	return nil
}

func (finalExchangeRates FinalExchangeRates) ExchangePToken2PRVByTokenId(pTokenId string, value uint64) uint64 {
	switch pTokenId {
	case PortalTokenSymbolBTC:
		return finalExchangeRates.ExchangeBTC2PRV(value)
	case PortalTokenSymbolBNB:
		return finalExchangeRates.ExchangeBNB2PRV(value)
	}

	return 0
}

func (finalExchangeRates *FinalExchangeRates) ExchangePRV2PTokenByTokenId(pTokenId string, value uint64) uint64 {
	switch pTokenId {
	case PortalTokenSymbolBTC:
		return finalExchangeRates.ExchangePRV2BTC(value)
	case PortalTokenSymbolBNB:
		return finalExchangeRates.ExchangePRV2BNB(value)
	}

	return 0
}

func (finalExchangeRates *FinalExchangeRates) convert(value uint64, ratesFrom uint64, RatesTo uint64) uint64 {
	//convert to pusdt
	total := (value * ratesFrom) / uint64(math.Pow10(9)) //value of nanno

	//pusdt -> new coin
	result := (total * uint64(math.Pow10(9))) / RatesTo

	return result

}

func (finalExchangeRates *FinalExchangeRates) ExchangeBTC2PRV(value uint64) uint64 {
	//input : nano
	BTCRates := finalExchangeRates.Rates[PortalTokenSymbolBTC].Amount //return nano pUSDT
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount //return nano pUSDT
	valueExchange := finalExchangeRates.convert(value, BTCRates, PRVRates)

	database.Logger.Log.Infof("================ Convert, BTC %d 2 PRV with BTCRates %d PRVRates %d , result %d", value, BTCRates, PRVRates, valueExchange)

	//nano
	return valueExchange
}

func (finalExchangeRates *FinalExchangeRates) ExchangeBNB2PRV(value uint64) uint64 {
	BNBRates := finalExchangeRates.Rates[PortalTokenSymbolBNB].Amount
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount

	valueExchange := finalExchangeRates.convert(value, BNBRates, PRVRates)

	database.Logger.Log.Infof("================ Convert, BNB %v 2 PRV with BNBRates %v PRVRates %v, result %v", value, BNBRates, PRVRates, valueExchange)

	return valueExchange
}

func (finalExchangeRates *FinalExchangeRates) ExchangePRV2BTC(value uint64) uint64 {
	//input nano
	BTCRates := finalExchangeRates.Rates[PortalTokenSymbolBTC].Amount //return nano pUSDT
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount //return nano pUSDT

	valueExchange := finalExchangeRates.convert(value, PRVRates, BTCRates)

	database.Logger.Log.Infof("================ Convert, PRV %v 2 BTC with BTCRates %v PRVRates %v, result %v", value, BTCRates, PRVRates, valueExchange)

	return valueExchange
}

func (finalExchangeRates *FinalExchangeRates) ExchangePRV2BNB(value uint64) uint64 {
	BNBRates := finalExchangeRates.Rates[PortalTokenSymbolBNB].Amount
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount

	valueExchange := finalExchangeRates.convert(value, PRVRates, BNBRates)

	database.Logger.Log.Infof("================ Convert, PRV %v 2 BNB with BNBRates %v PRVRates %v, result %v", value, BNBRates, PRVRates, valueExchange)
	return valueExchange
}

// ======= REDEEM =======
// NewWaitingRedeemReqKey creates key for storing waiting redeems list in portal
func NewWaitingRedeemReqKey(beaconHeight uint64, redeemID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalWaitingRedeemRequestsPrefix, beaconHeightBytes...)
	key = append(key, []byte(redeemID)...)
	return string(key)
}

// NewRedeemReqKey creates key for tracking redeems status in portal
func NewRedeemReqKey(redeemID string) string {
	key := append(PortalRedeemRequestsPrefix, []byte(redeemID)...)
	return string(key)
}

// NewRedeemReqKey creates key for tracking redeems status in portal
func NewTrackRedeemReqByTxReqIDKey(txID string) string {
	key := append(PortalRedeemRequestsByTxReqIDPrefix, []byte(txID)...)
	return string(key)
}

// StoreRedeemRequest stores status of redeem request by redeemID
func (db *db) StoreRedeemRequest(key []byte, value []byte) error {
	err := db.Put(key, value)
	if err != nil {
		return database.NewDatabaseError(database.StoreRedeemRequestError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetRedeemRequestByRedeemID(redeemID string) ([]byte, error) {
	key := NewRedeemReqKey(redeemID)
	return db.GetItemPortalByKey([]byte(key))
}

// TrackRedeemRequestByTxReqID tracks status of redeem request by txReqID
func (db *db) TrackRedeemRequestByTxReqID(key []byte, value []byte) error {
	err := db.Put(key, value)
	if err != nil {
		return database.NewDatabaseError(database.TrackRedeemReqByTxReqIDError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) StoreCustodianWithdrawRequest(key []byte, content interface{}) error  {
	contributionBytes, err := json.Marshal(content)
	if err != nil {
		return err
	}

	err = db.Put(key, contributionBytes)
	if err != nil {
		return database.NewDatabaseError(database.StorePortalCustodianWithdrawRequestStateError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}


