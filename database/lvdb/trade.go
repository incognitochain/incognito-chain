package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/pkg/errors"
)

func getTradeActivationKey(tradeID []byte) []byte {
	key := append(tradeActivationPrefix, tradeID...)
	return key
}

func getTradeActivationValue(
	bondID *common.Hash,
	buy bool,
	activated bool,
	amount uint64,
) []byte {
	values := []byte{}
	values = append(values, bondID[:]...)
	m := map[bool]byte{false: byte(0), true: byte(1)}
	values = append(values, m[buy])
	values = append(values, m[activated])
	values = append(values, common.Uint64ToBytes(amount)...)
	return values
}

func parseTradeActivationValue(value []byte) (*common.Hash, bool, bool, uint64, error) {
	if len(value) != common.HashSize+10 {
		return nil, false, false, 0, errors.Errorf("invalid trade activation value")
	}
	bondID := &common.Hash{}
	err := bondID.SetBytes(value[:common.HashSize])
	if err != nil {
		return nil, false, false, 0, errors.Errorf("invalid trade activation bondID")
	}
	buy := false
	if value[common.HashSize] > 0 {
		buy = true
	}
	activated := false
	if value[common.HashSize+1] > 0 {
		activated = true
	}
	amount := common.BytesToUint64(value[common.HashSize+2:])
	return bondID, buy, activated, amount, nil
}

func (db *db) StoreTradeActivation(
	tradeID []byte,
	bondID *common.Hash,
	buy bool,
	activated bool,
	amount uint64,
) error {
	key := getTradeActivationKey(tradeID)
	value := getTradeActivationValue(bondID, buy, activated, amount)
	return db.Put(key, value)
}

func (db *db) GetTradeActivation(tradeID []byte) (*common.Hash, bool, bool, uint64, error) {
	key := getTradeActivationKey(tradeID)
	value, err := db.Get(key)
	if err != nil {
		return nil, false, false, 0, err
	}
	return parseTradeActivationValue(value)
}
