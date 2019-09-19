package lvdb

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/database"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func (db *db) GetProducersBlackList() (map[string]uint8, error) {
	key := producersBlackListPrefix
	producersBlackListBytes, dbErr := db.lvdb.Get(key, nil)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return nil, database.NewDatabaseError(database.GetProducersBlackListError, dbErr)
	}
	producersBlackList := make(map[string]uint8)
	if len(producersBlackListBytes) == 0 {
		return producersBlackList, nil
	}
	err := json.Unmarshal(producersBlackListBytes, &producersBlackList)
	return producersBlackList, err
}

func (db *db) StoreProducersBlackList(producersBlackList map[string]uint8) error {
	producersBlackListBytes, err := json.Marshal(producersBlackList)
	if err != nil {
		return err
	}
	key := producersBlackListPrefix
	dbErr := db.Put(key, producersBlackListBytes)
	if dbErr != nil {
		return database.NewDatabaseError(database.StoreProducersBlackListError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}
