package lvdb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

type DividendReceivers struct {
	PaymentAddress []privacy.PaymentAddress
	Amounts        []uint64
}

func getDividendReceiversKey(id uint64, forDCB bool) []byte {
	return []byte(string(dividendReceiversPrefix) + strconv.FormatUint(id, 10) + strconv.FormatBool(forDCB))
}

func getDividendReceiversValue(receivers []privacy.PaymentAddress, amounts []uint64) []byte {
	dividendReceivers := &DividendReceivers{}
	for i, receiver := range receivers {
		amount := amounts[i]
		dividendReceivers.PaymentAddress = append(dividendReceivers.PaymentAddress, receiver)
		dividendReceivers.Amounts = append(dividendReceivers.Amounts, amount)
	}
	value, _ := json.Marshal(dividendReceivers)
	return value
}

func parseDividendReceiversValue(value []byte) ([]privacy.PaymentAddress, []uint64, error) {
	dividendReceivers := &DividendReceivers{}
	err := json.Unmarshal(value, dividendReceivers)
	if err != nil {
		return nil, nil, errors.Errorf("Invalid value stored as dividend receivers")
	}

	receivers := []privacy.PaymentAddress{}
	amounts := []uint64{}
	for i := 0; i < len(dividendReceivers.Amounts); i += 1 {
		receivers = append(receivers, dividendReceivers.PaymentAddress[i])
		amounts = append(amounts, dividendReceivers.Amounts[i])
	}
	return receivers, amounts, nil
}

func (db *db) GetDividendReceiversForID(id uint64, forDCB bool) ([]privacy.PaymentAddress, []uint64, bool, error) {
	key := getDividendReceiversKey(id, forDCB)
	if hasValue, err := db.HasValue(key); !hasValue {
		return nil, nil, hasValue, err
	}

	value, err := db.Get(key)
	if err != nil {
		return nil, nil, true, err
	}

	receivers, amounts, err := parseDividendReceiversValue(value)
	if err != nil {
		return nil, nil, false, err
	}
	return receivers, amounts, true, nil
}

func (db *db) StoreDividendReceiversForID(id uint64, forDCB bool, receivers []privacy.PaymentAddress, amounts []uint64) error {
	if len(receivers) != len(amounts) {
		return errors.Errorf("Different number of addresses and amounts: %d %d", len(receivers), len(amounts))
	}

	key := getDividendReceiversKey(id, forDCB)
	value := getDividendReceiversValue(receivers, amounts)
	if err := db.Put(key, value); err != nil {
		return err
	}
	fmt.Printf("[db] stored divReceiver: %d %t %d\n", id, forDCB, len(amounts))
	return nil
}
