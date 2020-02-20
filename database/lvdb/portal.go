package lvdb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	PortalTokenSymbolBTC = "BTC"
	PortalTokenSymbolBNB = "BNB"
	PortalTokenSymbolPRV = "PRV"
)

type CustodianState struct {
	IncognitoAddress string
	TotalCollateral  uint64			// prv
	FreeCollateral   uint64			// prv
	HoldingPubTokens map[string]uint64   	// tokenSymbol : amount
	LockedAmountCollateral map[string]uint64
	RemoteAddresses  map[string]string  	// tokenSymbol : address
}

type MatchingPortingCustodianDetail struct {
	RemoteAddress string
	Amount uint64
	LockedAmountCollateral uint64
}

type MatchingRedeemCustodianDetail struct {
	RemoteAddress string
	Amount uint64
	UnLockedAmountCollateral uint64
}


// todo: need to add beaconHeight when the porting request was created for checking timeout of porting request
type PortingRequest struct {
	UniquePortingID string
	TxReqID         common.Hash
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      map[string]MatchingPortingCustodianDetail			// key : incogAddress
	PortingFee      uint64
}

type RedeemRequest struct {
	UniqueRedeemID        string
	TxReqID               string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	RedeemAmount          uint64
	Custodians            map[string]MatchingRedeemCustodianDetail 	// key : incogAddress
	RedeemFee             uint64
}

type ExchangeRatesDetail struct {
	Amount uint64
}

type ExchangeRatesRequest struct {
	SenderAddress string
	Rates map[string]ExchangeRatesDetail
}

type FinalExchangeRatesDetail struct {
	Amount uint64
}

type FinalExchangeRates struct {
	Rates map[string]FinalExchangeRatesDetail
}

func NewCustodianStateKey (beaconHeight uint64, custodianAddress string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalCustodianStatePrefix, beaconHeightBytes...)
	key = append(key, []byte(custodianAddress)...)
	return string(key)
}

func NewPortingRequestKey (beaconHeight uint64, uniquePortingID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalPortingRequestsPrefix, []byte(uniquePortingID)...)
	key = append(key, beaconHeightBytes...)
	return string(key) //prefix + uniqueId + beaconHeight
}

func NewFinalExchangeRatesKey (beaconHeight uint64) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalFinalExchangeRatesPrefix, beaconHeightBytes...)
	return string(key)
}

func NewExchangeRatesRequestKey (beaconHeight uint64, txId string, lockTime string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalExchangeRatesPrefix, beaconHeightBytes...)
	key = append(key, []byte(txId)...)
	key = append(key, []byte(lockTime)...)

	return string(key)
}

func NewCustodianDepositKey (beaconHeight uint64, custodianAddress string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalCustodianDepositPrefix, beaconHeightBytes...)
	key = append(key, []byte(custodianAddress)...)
	return string(key)
}

func NewWaitingPortingReqKey (beaconHeight uint64, portingID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalWaitingPortingRequestsPrefix, beaconHeightBytes...)
	key = append(key, []byte(portingID)...)
	return string(key)
}

func NewWaitingRedeemReqKey (beaconHeight uint64, redeemID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalWaitingRedeemRequestsPrefix, beaconHeightBytes...)
	key = append(key, []byte(redeemID)...)
	return string(key)
}

// NewPortalReqPTokenKey creates key for tracking request pToken in portal
func NewPortalReqPTokenKey (beaconHeight uint64, portingID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalRequestPTokensPrefix, beaconHeightBytes...)
	key = append(key, []byte(portingID)...)
	return string(key)
}

func (db *db) GetAllRecordsPortalByPrefix(beaconHeight uint64, prefix []byte) ([][]byte, [][]byte, error) {
	keys := [][]byte{}
	values := [][]byte{}
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	prefixByBeaconHeight := append(prefix, beaconHeightBytes...)
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

func (db *db) TrackCustodianDepositCollateral(key []byte, content []byte) error {
	err := db.Put(key, content)
	if err != nil {
		return database.NewDatabaseError(database.TrackCustodianDepositError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) TrackReqPTokens(key []byte, content []byte) error {
	err := db.Put(key, content)
	if err != nil {
		return database.NewDatabaseError(database.TrackCustodianDepositError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
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

func (db *db) GetItemPortalByPrefix(prefix []byte) (byte, error) {
	itemRecord, dbErr := db.lvdb.Get(prefix, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return 0, database.NewDatabaseError(database.GetItemPortalByPrefixError, dbErr)
	}

	if len(itemRecord) == 0 {
		return 0, database.NewDatabaseError(database.GetItemPortalByPrefixNotFound, dbErr)
	}

	return itemRecord[0], nil
}

func (finalExchangeRates *FinalExchangeRates) ExchangePToken2PRVByTokenId(pTokenId string, value uint64) uint64 {
	switch pTokenId {
	case PortalTokenSymbolBTC:
		return finalExchangeRates.ExchangeBTC2PRV(value)
	case PortalTokenSymbolBNB:
		return finalExchangeRates.ExchangeBTC2PRV(value)
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

func (finalExchangeRates *FinalExchangeRates) ExchangeBTC2PRV(value uint64) uint64 {
	//get rate of BTC
	BTCRates := finalExchangeRates.Rates[PortalTokenSymbolBTC].Amount
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount
	//BTC -> USDT
	btc2usd := value * BTCRates

	//BTC -> PRV
	totalPRV := btc2usd / PRVRates
	//totalPRV = uint64(totalPRV)
	return totalPRV
}

func (finalExchangeRates *FinalExchangeRates) ExchangeBNB2PRV(value uint64) uint64 {
	//get rate of BTC
	BNBRates := finalExchangeRates.Rates[PortalTokenSymbolBNB].Amount
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount
	//BTC -> USDT
	bnb2usd := value * BNBRates

	//BTC -> PRV
	totalPRV := bnb2usd / PRVRates
	//totalPRV = uint64(totalPRV)
	return  totalPRV
}

func (finalExchangeRates *FinalExchangeRates) ExchangePRV2BTC(value uint64) uint64 {
	//get rate of BTC
	BTCRates := finalExchangeRates.Rates[PortalTokenSymbolBTC].Amount
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount
	//PRV -> USDT
	prv2usd := value * PRVRates

	//PRV -> BTC
	totalBTC := prv2usd / BTCRates
	return totalBTC
}

func (finalExchangeRates *FinalExchangeRates) ExchangePRV2BNB(value uint64) uint64 {
	//get rate of BTC
	BNBRates := finalExchangeRates.Rates[PortalTokenSymbolBNB].Amount
	PRVRates := finalExchangeRates.Rates[PortalTokenSymbolPRV].Amount
	//PRV -> USDT
	prv2usd := value * PRVRates

	//BNB -> PRV
	totalBNB := prv2usd / BNBRates
	return  totalBNB
}