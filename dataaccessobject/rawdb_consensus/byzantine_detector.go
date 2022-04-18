package rawdb_consensus

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/incdb"
	"strings"
	"time"
)

type BlackListValidator struct {
	Error     string
	StartTime time.Time
	TTL       time.Duration
}

func NewBlackListValidator() *BlackListValidator {
	return &BlackListValidator{}
}

func StoreBlackListValidator(db incdb.KeyValueWriter, validator string, blackListValidator *BlackListValidator) error {

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

func GetAllBlackListValidator(db incdb.Database) (map[string]*BlackListValidator, error) {

	prefix := GetByzantineBlackListPrefix()

	it := db.NewIteratorWithPrefix(prefix)
	defer it.Release()

	blacklistValidators := make(map[string]*BlackListValidator)

	for it.Next() {
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		keys := strings.Split(string(key), string(splitter))
		validator := keys[1]

		data := make([]byte, len(it.Value()))
		copy(data, it.Value())
		blackListValidator := NewBlackListValidator()
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
