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

func NewPortingReqKey (beaconHeight uint64, portingID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalPortingRequestsPrefix, beaconHeightBytes...)
	key = append(key, []byte(portingID)...)
	return string(key)
}

func NewFinalExchangeRatesKey (beaconHeight uint64) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalFinalExchangeRatesPrefix, beaconHeightBytes...)
	return string(key)
}

func NewRedeemReqKey (beaconHeight uint64, redeemID string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalRedeemRequestsPrefix, beaconHeightBytes...)
	key = append(key, []byte(redeemID)...)
	return string(key)
}

func NewExchangeRatesRequestKey (beaconHeight uint64, txId string, lockTime string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalExchangeRatesPrefix, beaconHeightBytes...)
	key = append(key, []byte(lockTime)...)
	return string(key)
}

func NewCustodianDepositKey (beaconHeight uint64, custodianAddress string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(PortalCustodianDepositPrefix, beaconHeightBytes...)
	key = append(key, []byte(custodianAddress)...)
	return string(key)
}

func BuildCustodianDepositKey(
	prefix []byte,
	suffix []byte,
) []byte {
	return append(prefix, suffix...)
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

func (db *db) TrackCustodianDepositCollateral(prefix []byte, suffix []byte, content []byte) error {
	key := BuildCustodianDepositKey(prefix, suffix)
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

func (db *db) GetItemPortalByPrefix(prefix []byte) (byte, error) {
	itemRecord, dbErr := db.lvdb.Get(prefix, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return 0, database.NewDatabaseError(database.GetItemPortalByPrefixError, dbErr)
	}

	if len(itemRecord) == 0 {
		return 0, nil
	}

	return itemRecord[0], nil
}
