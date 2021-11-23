package rawdb_consensus

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incdb"
	"strings"
)

func StoreBlackListValidator(db incdb.KeyValueWriter, validator string, blackListValidator *consensustypes.BlackListValidator) error {

	key := GetByzantineBlackListKey(validator)

	value, err := json.Marshal(blackListValidator)
	if err != nil {
		return err
	}

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func GetAllBlackListValidator(db incdb.Database) (map[string]*consensustypes.BlackListValidator, error) {

	prefix := GetByzantineBlackListPrefix()

	it := db.NewIteratorWithPrefix(prefix)
	defer it.Release()

	blacklistValidators := make(map[string]*consensustypes.BlackListValidator)

	for it.Next() {
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		keys := strings.Split(string(key), string(splitter))
		validator := keys[1]

		data := make([]byte, len(it.Value()))
		copy(data, it.Value())
		blackListValidator := consensustypes.NewBlackListValidator()
		if err := json.Unmarshal(data, blackListValidator); err != nil {
			return nil, err
		}

		blacklistValidators[validator] = blackListValidator
	}

	return blacklistValidators, nil
}

func DeleteBlackListValidator(db incdb.KeyValueWriter, validator string) error {
	key := GetByzantineBlackListKey(validator)
	return db.Delete(key)
}

func GetBlackListValidator(db incdb.KeyValueReader, validator string) (bool, *consensustypes.BlackListValidator, error) {

	key := GetByzantineBlackListKey(validator)
	has, err := db.Has(key)
	if err != nil {
		return false, nil, err
	}

	if !has {
		return false, nil, nil
	}

	data, err := db.Get(key)
	if err != nil {
		return false, nil, err
	}
	blackListValidator := consensustypes.NewBlackListValidator()

	if err := json.Unmarshal(data, blackListValidator); err != nil {
		return false, nil, err
	}

	return true, blackListValidator, nil
}
