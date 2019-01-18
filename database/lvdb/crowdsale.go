package lvdb

// TODO(@0xbunyip): implement

func (db *db) StoreCrowdsaleData(
	saleID []byte,
	endBlock uint64,
	buyingAsset []byte,
	amountBuying uint64,
	sellingAsset []byte,
	amountSelling uint64,
) error {
	return nil
}

func (db *db) GetCrowdsaleData(saleID []byte) (uint64, []byte, uint64, []byte, uint64, error) {
	return 0, nil, 0, nil, 0, nil
}

func (db *db) StoreCrowdsaleRequest(requestTxHash, saleID, pk, tk []byte) error {
	return nil
}

func (db *db) StoreCrowdsaleResponse(requestTxHash, responseTxHash []byte) error {
	return nil
}

func (db *db) GetCrowdsaleTxs(requestTxHash []byte) ([][]byte, error) {
	return nil, nil
}
