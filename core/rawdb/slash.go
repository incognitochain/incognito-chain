package rawdb

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func GetProducersBlackList(db incdb.Database, beaconHeight uint64) (map[string]uint8, error) {
	// key := producersBlackListPrefix
	beaconHeightBytes := []byte(fmt.Sprintf("%d", beaconHeight))
	key := append(producersBlackListPrefix, beaconHeightBytes...)
	producersBlackListBytes, dbErr := db.Get(key)
	if dbErr != nil && dbErr != lvdberr.ErrNotFound {
		return nil, incdb.NewDatabaseError(incdb.GetProducersBlackListError, dbErr)
	}
	producersBlackList := make(map[string]uint8)
	if len(producersBlackListBytes) == 0 {
		return producersBlackList, nil
	}
	err := json.Unmarshal(producersBlackListBytes, &producersBlackList)
	return producersBlackList, err
}

func StoreProducersBlackList(db incdb.Database, beaconHeight uint64, producersBlackList map[string]uint8) error {
	producersBlackListBytes, err := json.Marshal(producersBlackList)
	if err != nil {
		return err
	}
	// key := producersBlackListPrefix
	beaconHeightBytes := []byte(fmt.Sprintf("%d", beaconHeight))
	key := append(producersBlackListPrefix, beaconHeightBytes...)
	dbErr := db.Put(key, producersBlackListBytes)
	if dbErr != nil {
		return incdb.NewDatabaseError(incdb.StoreProducersBlackListError, errors.Wrap(dbErr, "db.lvdb.put"))
	}
	return nil
}
