package lvdb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/database"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type CustodianState struct {
	IncognitoAddress string
	TotalCollateral  uint64
	FreeCollateral   uint64
	HoldingPubTokens map[string]uint64
	RemoteAddresses  map[string]string
}

type PortingRequest struct {
	UniquePortingID string
	TxReqID         string
	TokenID         string
	PorterAddress   string
	Amount          uint64
	Custodians      map[string]uint64
	PortingFee      uint64
}

type RedeemRequest struct {
	UniqueRedeemID        string
	TxReqID               string
	TokenID               string
	RedeemerAddress       string
	RedeemerRemoteAddress string
	Amount                uint64
	Custodians            map[string]uint64
	RedeemFee             uint64
}

func NewCustodianStateKey (beaconHeight uint64, custodianAddress string) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	key := append(CustodianStatePrefix, beaconHeightBytes...)
	key = append(key, []byte(custodianAddress)...)
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