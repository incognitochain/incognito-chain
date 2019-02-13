package lvdb

import (
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

const dividendReceiverValueSep = "-"

func getDividendReceiversKey(id uint64, forDCB bool) []byte {
	return []byte(string(dividendReceiversPrefix) + strconv.FormatUint(id, 10) + strconv.FormatBool(forDCB))
}

func getDividendReceiversValue(receivers []privacy.PaymentAddress, amounts []uint64) []byte {
	value := []string{}
	for i, receiver := range receivers {
		amount := amounts[i]
		s := strings.Join([]string{base58.Base58Check{}.Encode(receiver.Bytes(), 0x00), strconv.FormatUint(amount, 10)}, dividendReceiverValueSep)
		value = append(value, s)
	}
	return []byte(strings.Join(value, dividendReceiverValueSep))
}

func parseDividendReceiversValue(value []byte) ([]privacy.PaymentAddress, []uint64, error) {
	splits := strings.Split(string(value), dividendReceiverValueSep)
	if len(splits)%2 != 0 {
		return nil, nil, errors.Errorf("Invalid value stored as dividend receivers with length %d", len(splits))
	}

	receivers := []privacy.PaymentAddress{}
	amounts := []uint64{}
	for i := 0; i < len(splits); i += 2 {
		receiverInBytes, _, err := base58.Base58Check{}.Decode(splits[i])
		if err != nil {
			return nil, nil, err
		}
		receiver := &privacy.PaymentAddress{}
		receiver.SetBytes(receiverInBytes)
		receivers = append(receivers, *receiver)

		amount, err := strconv.ParseUint(splits[i+1], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		amounts = append(amounts, amount)
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
	return nil
}
