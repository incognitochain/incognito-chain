package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

// StoreBeaconBlock store block hash => block value
func StoreFlatFileIndexByBlockHash(db incdb.KeyValueWriter, hash common.Hash, index uint64) error {
	keyHash := GetBlockHashToFFIndexKey(hash)
	buf := common.Uint64ToBytes(index)
	if err := db.Put(keyHash, buf); err != nil {
		return NewRawdbError(StoreFFIndexError, err)
	}
	return nil
}

func GetFlatFileIndexByBlockHash(db incdb.KeyValueReader, hash common.Hash) (uint64, error) {
	keyHash := GetBlockHashToFFIndexKey(hash)
	if data, err := db.Get(keyHash); err != nil {
		return 0, NewRawdbError(StoreFFIndexError, err)
	} else {
		return common.BytesToUint64(data)
	}

}
