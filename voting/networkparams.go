package voting

import "github.com/ninjadotorg/constant/common"

// Todo: @0xjackalope, @0xbunyip Check logic in Hash and Validate and rpcfunction because other will change params struct without modified these function
type GOVVotingParams struct {
	SalaryPerTx  uint64 // salary for each tx in block(mili constant)
	BasicSalary  uint64 // basic salary per block(mili constant)
	TxFee        uint64
	SellingBonds *SellingBonds
	RefundInfo   *RefundInfo
}

type SellingBonds struct {
	BondsToSell    uint64
	BondPrice      uint64 // in Constant unit
	Maturity       uint32
	BuyBackPrice   uint64 // in Constant unit
	StartSellingAt uint32 // start selling bonds at block height
	SellingWithin  uint32 // selling bonds within n blocks
}

type SaleData struct {
	SaleID []byte // Unique id of the crowdsale to store in db

	BuyingAsset  []byte
	SellingAsset []byte
	EndBlock     int32
}

type RefundInfo struct {
	ThresholdToLargeTx uint64
	RefundAmount       uint64
}

type DCBVotingParams struct {
	*SaleData
}

//xxx
func (DCBParams DCBVotingParams) Hash() *common.Hash {
	record := ""
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (GOVParams GOVVotingParams) Hash() *common.Hash {
	record := string(GOVParams.SalaryPerTx)
	record += string(GOVParams.BasicSalary)
	record += string(common.ToBytes(GOVParams.SellingBonds.Hash()))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (SellingBonds SellingBonds) Hash() *common.Hash {
	record := string(SellingBonds.BondsToSell)
	record += string(SellingBonds.BondPrice)
	record += string(SellingBonds.Maturity)
	record += string(SellingBonds.BuyBackPrice)
	record += string(SellingBonds.StartSellingAt)
	record += string(SellingBonds.SellingWithin)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

//xxx
func (GOVParams GOVVotingParams) Validate() bool {
	return true
}
func (DCBParams DCBVotingParams) Validate() bool {
	return true
}

func (DCBVotingParams DCBVotingParams) ValidateSanityData() bool {
	// Todo: @0xbunyip
	return true
}

func (GOVVotingParams GOVVotingParams) ValidateSanityData() bool {
	// Todo: @0xankylosaurus
	return true
}
