package lvdb

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/pkg/errors"
)

func getIssuingInfoKey(reqTxID common.Hash) []byte {
	key := append(reserveIssuingInfoPrefix, []byte(reqTxID.String())...)
	return key
}

func getIssuingInfoValue(
	amount uint64,
	instType string,
) []byte {
	values := []byte{}
	values = append(values, common.Uint64ToBytes(amount)...)
	values = append(values, []byte(instType)...)
	return values
}

func parseIssuingInfoValue(value []byte) (uint64, string, error) {
	if len(value) < 8 {
		return 0, "", errors.Errorf("Error parsing issuing info value: %x", value)
	}
	amount := common.BytesToUint64(value[:8])
	instType := string(value[8:])
	return amount, instType, nil
}

func (db *db) StoreIssuingInfo(
	reqTxID common.Hash,
	amount uint64,
	instType string,
) error {
	key := getIssuingInfoKey(reqTxID)
	value := getIssuingInfoValue(amount, instType)
	return db.Put(key, value)
}

func (db *db) GetIssuingInfo(reqTxID common.Hash) (uint64, string, error) {
	key := getIssuingInfoKey(reqTxID)
	value, err := db.Get(key)
	if err != nil {
		return 0, "", err
	}
	return parseIssuingInfoValue(value)
}
