package lvdb

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func getCrowdsaleDataKey(saleID []byte) []byte {
	key := append(crowdsalePrefix, saleID...)
	return key
}

func getCrowdsaleDataValue(
	endBlock uint64,
	buyingAsset common.Hash,
	buyingAmount uint64,
	sellingAsset common.Hash,
	sellingAmount uint64,
) []byte {
	values := []byte{}
	values = append(values, common.Uint64ToBytes(endBlock)...)
	values = append(values, buyingAsset[:]...)
	values = append(values, common.Uint64ToBytes(buyingAmount)...)
	values = append(values, sellingAsset[:]...)
	values = append(values, common.Uint64ToBytes(sellingAmount)...)
	return values
}

func parseCrowdsaleDataValue(value []byte) (uint64, common.Hash, uint64, common.Hash, uint64, error) {
	if len(value) != 3*8+2*common.HashSize {
		return 0, common.Hash{}, 0, common.Hash{}, 0, errors.New("Length of crowdsale data is incorrect")
	}
	endBlock := common.BytesToUint64(value)
	buyingAsset, _ := common.NewHash(value[8 : 8+common.HashSize])
	buyingAmount := common.BytesToUint64(value[8+common.HashSize:])
	sellingAsset, _ := common.NewHash(value[16+common.HashSize : 16+2*common.HashSize])
	sellingAmount := common.BytesToUint64(value[16+2*common.HashSize:])
	return endBlock, *buyingAsset, buyingAmount, *sellingAsset, sellingAmount, nil
}

func (db *db) StoreCrowdsaleData(
	saleID []byte,
	endBlock uint64,
	buyingAsset common.Hash,
	buyingAmount uint64,
	sellingAsset common.Hash,
	sellingAmount uint64,
) error {
	if len(buyingAsset) != common.HashSize || len(sellingAsset) != common.HashSize {
		return errors.New("AssetID is not hash")
	}

	key := getCrowdsaleDataKey(saleID)
	value := getCrowdsaleDataValue(endBlock, buyingAsset, buyingAmount, sellingAsset, sellingAmount)
	return db.Put(key, value)
}

func (db *db) GetCrowdsaleData(saleID []byte) (uint64, common.Hash, uint64, common.Hash, uint64, error) {
	key := getCrowdsaleDataKey(saleID)
	value, err := db.Get(key)
	if err != nil {
		return 0, common.Hash{}, 0, common.Hash{}, 0, err
	}
	return parseCrowdsaleDataValue(value)
}

func (db *db) GetAllCrowdsales() ([]uint64, []common.Hash, []uint64, []common.Hash, []uint64, error) {
	saleID := []byte{} // Empty id to get all
	key := getCrowdsaleDataKey(saleID)

	endBlocks := []uint64{}
	buyingAssets := []common.Hash{}
	buyingAmounts := []uint64{}
	sellingAssets := []common.Hash{}
	sellingAmounts := []uint64{}

	iter := db.lvdb.NewIterator(util.BytesPrefix(key), nil)
	defer iter.Release()
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		endBlock, buyingAsset, buyingAmount, sellingAsset, sellingAmount, err := parseCrowdsaleDataValue(value)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		endBlocks = append(endBlocks, endBlock)
		buyingAssets = append(buyingAssets, buyingAsset)
		buyingAmounts = append(buyingAmounts, buyingAmount)
		sellingAssets = append(sellingAssets, sellingAsset)
		sellingAmounts = append(sellingAmounts, sellingAmount)
	}
	return endBlocks, buyingAssets, buyingAmounts, sellingAssets, sellingAmounts, nil
}
