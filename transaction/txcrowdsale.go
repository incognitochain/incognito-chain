package transaction

type SaleData struct {
	SaleID []byte // Unique id of the crowdsale to store in db
	BondID []byte // in case either base or quote asset is bond

	BuyingAsset  string
	SellingAsset string
	Price        uint64
}
