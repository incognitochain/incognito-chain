package lvdb

import "github.com/ninjadotorg/constant/common"

// TODO(@0xbunyip): implement

func (db *db) StoreCrowdsaleData(
	saleID []byte,
	endBlock uint64,
	buyingAsset common.Hash,
	buyingAmount uint64,
	sellingAsset common.Hash,
	sellingAmount uint64,
) error {
	return nil
}

func (db *db) GetCrowdsaleData(saleID []byte) (uint64, common.Hash, uint64, common.Hash, uint64, error) {
	return 0, common.Hash{}, 0, common.Hash{}, 0, nil
}

func (db *db) StoreCrowdsaleRequest(requestTxHash, saleID, pk, tk []byte) error {
	return nil
}

func (db *db) GetCrowdsaleTxs(requestTxHash []byte) ([][]byte, error) {
	return nil, nil
}
