package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func getSaleDataKey(saleID []byte) []byte {
	key := append(crowdsalePrefix, saleID...)
	return key
}

func (db *db) StoreSaleData(saleID, data []byte) error {
	key := getSaleDataKey(saleID)
	return db.Put(key, data)
}

func (db *db) GetSaleData(saleID []byte) ([]byte, error) {
	key := getSaleDataKey(saleID)
	return db.Get(key)
}

func (db *db) GetAllSaleData() ([][]byte, error) {
	key := getSaleDataKey(make([]byte, 0))
	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	data := [][]byte{}
	for iter.Next() {
		value := iter.Value()
		d := make([]byte, len(value))
		copy(d, value)
		data = append(data, d)
	}
	return data, nil
}

func getDCBBondInfoKey(saleID []byte) []byte {
	key := append(dcbBondInfoPrefix, saleID...)
	return key
}

func getDCBBondInfoValue(amountAvail, cstPaid uint64) []byte {
	values := []byte{}
	values = append(values, common.Uint64ToBytes(amountAvail)...)
	values = append(values, common.Uint64ToBytes(cstPaid)...)
	return values
}

func parseDCBBondInfoValue(value []byte) (uint64, uint64) {
	if len(value) != 16 {
		return 0, 0
	}
	amountAvail := common.BytesToUint64(value[:8])
	cstPaid := common.BytesToUint64(value[8:])
	return amountAvail, cstPaid
}

func (db *db) StoreDCBBondInfo(bondID *common.Hash, amountAvail, cstPaid uint64) error {
	key := getDCBBondInfoKey(bondID[:])
	value := getDCBBondInfoValue(amountAvail, cstPaid)
	return db.Put(key, value)
}

func (db *db) GetDCBBondInfo(bondID *common.Hash) (uint64, uint64) {
	key := getDCBBondInfoKey(bondID[:])
	value, err := db.Get(key)
	if err != nil {
		return 0, 0 // Dummy amount to prevent divide by zero
	}
	return parseDCBBondInfoValue(value)
}
