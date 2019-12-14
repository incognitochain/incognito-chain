package rawdb

import (
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreBytesValueWithTemporaryID(db incdb.Database, tempID []byte, value []byte) error {
	key := append(temporaryIDPrefix, tempID...)
	err := db.Put(key, value)
	if err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func GetBytesValueWithTemporaryID(db incdb.Database, tempID []byte) ([]byte, error) {
	key := append(temporaryIDPrefix, tempID...)
	value, err := db.Get(key)
	if err != nil {
		return []byte{}, NewRawdbError(LvdbGetError, err)
	}
	return value, nil
}

func HasBytesValueWithTemporaryID(db incdb.Database, tempID []byte) (bool, error) {
	key := append(temporaryIDPrefix, tempID...)
	value, err := db.Has(key)
	if err != nil {
		return false, NewRawdbError(LvdbHasError, err)
	}
	return value, nil
}

func DeleteBytesValueWithTemporaryID(db incdb.Database, tempID []byte) error {
	key := append(temporaryIDPrefix, tempID...)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(LvdbGetError, err)
	}
	return nil
}
