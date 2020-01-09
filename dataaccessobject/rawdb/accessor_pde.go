package rawdb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incdb"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func GetPDEPoolForPair(
	db incdb.Database,
	beaconHeight uint64,
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
) ([]byte, error) {
	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, tokenIDToBuyStr, tokenIDToSellStr)
	pdePoolForPairBytes, err := db.Get(pdePoolForPairKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return []byte{}, NewRawdbError(GetPDEPoolForPairKeyError, err)
	}
	return pdePoolForPairBytes, nil
}

func GetLatestPDEPoolForPair(
	db incdb.Database,
	tokenIDToBuyStr string,
	tokenIDToSellStr string,
) ([]byte, error) {
	iter := db.NewIteratorWithPrefix(PDEPoolPrefix)
	ok := iter.Last()
	if !ok {
		return []byte{}, nil
	}
	key := iter.Key()
	keyBytes := make([]byte, len(key))
	copy(keyBytes, key)
	iter.Release()
	err := iter.Error()
	if err != nil {
		return []byte{}, NewRawdbError(LvdbIteratorError, err)
	}
	parts := strings.Split(string(keyBytes), "-")
	if len(parts) <= 1 {
		return []byte{}, nil
	}
	beaconHeight, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return []byte{}, NewRawdbError(UnexpectedError, err)
	}
	pdePoolForPairKey := BuildPDEPoolForPairKey(beaconHeight, tokenIDToBuyStr, tokenIDToSellStr)
	pdePoolForPairBytes, err := db.Get(pdePoolForPairKey)
	if err != nil && err != lvdberr.ErrNotFound {
		return []byte{}, NewRawdbError(GetPDEPoolForPairKeyError, err)
	}
	return pdePoolForPairBytes, nil
}

func GetAllRecordsByPrefix(db incdb.Database, beaconHeight uint64, prefix []byte) ([][]byte, [][]byte, error) {
	keys := [][]byte{}
	values := [][]byte{}
	beaconHeightBytes := []byte(fmt.Sprintf("%d-", beaconHeight))
	prefixByBeaconHeight := append(prefix, beaconHeightBytes...)
	iter := db.NewIteratorWithPrefix(prefixByBeaconHeight)
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
		return keys, values, NewRawdbError(GetAllRecordsByPrefixError, err)
	}
	return keys, values, nil
}

func TrackPDEStatus(
	db incdb.Database,
	prefix []byte,
	suffix []byte,
	status byte,
) error {
	key := BuildPDEStatusKey(prefix, suffix)
	err := db.Put(key, []byte{status})
	if err != nil {
		return NewRawdbError(TrackPDEStatusError, err)
	}
	return nil
}

func GetPDEStatus(
	db incdb.Database,
	prefix []byte,
	suffix []byte,
) (byte, error) {
	key := BuildPDEStatusKey(prefix, suffix)
	pdeStatusBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return common.PDENotFoundStatus, NewRawdbError(GetPDEStatusError, dbErr)
	}
	if len(pdeStatusBytes) == 0 {
		return common.PDENotFoundStatus, nil
	}
	return pdeStatusBytes[0], nil
}

func TrackPDEContributionStatus(
	db incdb.Database,
	prefix []byte,
	suffix []byte,
	statusContent []byte,
) error {
	key := BuildPDEStatusKey(prefix, suffix)
	err := db.Put(key, statusContent)
	if err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func GetPDEContributionStatus(
	db incdb.Database,
	prefix []byte,
	suffix []byte,
) ([]byte, error) {
	key := BuildPDEStatusKey(prefix, suffix)
	pdeStatusContentBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return []byte{}, NewRawdbError(GetPDEStatusError, dbErr)
	}
	return pdeStatusContentBytes, nil
}
