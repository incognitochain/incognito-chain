package lvdb

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/database"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/tendermint/tendermint/types"
)

type BNBHeader struct {
	Header *types.Header   		`json:"Header"`
	LastCommit *types.Commit	`json:"LastCommit"`
}

func NewBNBHeaderRelayingStateKey(beaconHeight uint64) string {
	beaconHeightBytes := []byte(fmt.Sprintf("%d", beaconHeight))
	key := append(RelayingBNBHeaderStatePrefix, beaconHeightBytes...)
	return string(key) //prefix + beaconHeight
}

func NewRelayingBNBHeaderChainKey(blockHeight uint64) string {
	blockHeightBytes := []byte(fmt.Sprintf("%d", blockHeight))
	key := append(RelayingBNBHeaderChainPrefix, blockHeightBytes...)
	return string(key) //prefix + blockHeight
}


func(db*db) GetItemByKey(key []byte) ([]byte, error){
	valueBytes, err := db.lvdb.Get([]byte(key), nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.UnexpectedError, err)
	}

	return valueBytes, nil
}

func (db *db) StoreRelayingBNBHeaderChain(blockHeight uint64, header []byte) error {
	key := NewRelayingBNBHeaderChainKey(blockHeight)

	err := db.Put([]byte(key), header)
	if err != nil {
		return database.NewDatabaseError(database.StoreRelayingBNBHeaderError, err)
	}

	return nil
}

func (db *db) GetRelayingBNBHeaderChain(blockHeight uint64) ([]byte, error) {
	key := NewRelayingBNBHeaderChainKey(blockHeight)

	data, err := db.lvdb.Get([]byte(key), nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.GetRelayingBNBHeaderError, err)
	}

	return data, nil
}

func (db *db) GetBNBDataHashByBlockHeight(blockHeight uint64) ([]byte, error) {
	key := NewRelayingBNBHeaderChainKey(blockHeight)

	data, err := db.lvdb.Get([]byte(key), nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.GetRelayingBNBHeaderError, err)
	}

	var bnbHeader types.Header
	err = json.Unmarshal(data, &bnbHeader)
	if err != nil {
		return nil, database.NewDatabaseError(database.GetRelayingBNBHeaderError, err)
	}

	return bnbHeader.DataHash, nil
}
