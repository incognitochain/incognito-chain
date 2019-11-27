package rawdb

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/incdb"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func GetProducersBlackList(db incdb.Database, beaconHeight uint64) (map[string]uint8, error) {
	// key := producersBlackListPrefix
	beaconHeightBytes := []byte(fmt.Sprintf("%d", beaconHeight))
	key := append(producersBlackListPrefix, beaconHeightBytes...)
	producersBlackListBytes, err := db.Get(key)
	if err != nil && err != lvdberr.ErrNotFound {
		return nil, NewRawdbError(GetProducersBlackListError, err)
	}
	producersBlackList := make(map[string]uint8)
	if len(producersBlackListBytes) == 0 {
		return producersBlackList, nil
	}
	err = json.Unmarshal(producersBlackListBytes, &producersBlackList)
	if err != nil {
		return producersBlackList, NewRawdbError(JsonUnMarshalError, err)
	}
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
		return NewRawdbError(StoreProducersBlackListError, err)
	}
	return nil
}
