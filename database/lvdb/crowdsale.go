package lvdb

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func getCrowdsaleDataKey(saleID []byte) []byte {
	key := append(crowdsalePrefix, saleID...)
	return key
}

func parseCrowdsaleDataKey(key []byte) []byte {
	saleID := make([]byte, len(key)-len(crowdsalePrefix))
	copy(saleID, key[len(crowdsalePrefix):])
	return saleID
}

func getCrowdsaleDataValue(
	proposalTxHash common.Hash,
	buyingAmount uint64,
	sellingAmount uint64,
) []byte {
	values := []byte{}
	values = append(values, proposalTxHash[:]...)
	values = append(values, common.Uint64ToBytes(buyingAmount)...)
	values = append(values, common.Uint64ToBytes(sellingAmount)...)
	return values
}

func parseCrowdsaleDataValue(value []byte) (common.Hash, uint64, uint64, error) {
	if len(value) != common.HashSize+2*8 {
		return common.Hash{}, 0, 0, errors.New("Length of crowdsale data is incorrect")
	}
	proposalTxHash, _ := common.NewHash(value[:common.HashSize])
	buyingAmount := common.BytesToUint64(value[common.HashSize : common.HashSize+8])
	sellingAmount := common.BytesToUint64(value[common.HashSize+8:])
	return *proposalTxHash, buyingAmount, sellingAmount, nil
}

func (db *db) StoreCrowdsaleData(
	saleID []byte,
	proposalTxHash common.Hash,
	buyingAmount uint64,
	sellingAmount uint64,
) error {
	key := getCrowdsaleDataKey(saleID)
	value := getCrowdsaleDataValue(proposalTxHash, buyingAmount, sellingAmount)
	fmt.Printf("[db] storecsdata key/value: \n%x\n%x\n", key, value)
	return db.Put(key, value)
}

func (db *db) GetCrowdsaleData(saleID []byte) (common.Hash, uint64, uint64, error) {
	key := getCrowdsaleDataKey(saleID)
	value, err := db.Get(key)
	if err != nil {
		return common.Hash{}, 0, 0, err
	}
	return parseCrowdsaleDataValue(value)
}

func (db *db) GetAllCrowdsales() ([][]byte, []common.Hash, []uint64, []uint64, error) {
	saleID := []byte{} // Empty id to get all
	key := getCrowdsaleDataKey(saleID)
	fmt.Printf("[db] get key: %x\n", key)

	saleIDs := [][]byte{}
	proposalTxHashes := []common.Hash{}
	buyingAmounts := []uint64{}
	sellingAmounts := []uint64{}

	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	defer iter.Release()
	for iter.Next() {
		fmt.Printf("[db] found key/value: \n%x\n%x\n", iter.Key(), iter.Value())
		saleID := parseCrowdsaleDataKey(iter.Key())
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		proposalTxHash, buyingAmount, sellingAmount, err := parseCrowdsaleDataValue(value)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		saleIDs = append(saleIDs, saleID)
		proposalTxHashes = append(proposalTxHashes, proposalTxHash)
		buyingAmounts = append(buyingAmounts, buyingAmount)
		sellingAmounts = append(sellingAmounts, sellingAmount)
	}
	return saleIDs, proposalTxHashes, buyingAmounts, sellingAmounts, nil
}
