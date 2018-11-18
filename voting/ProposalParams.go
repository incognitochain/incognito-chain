package voting

import "github.com/ninjadotorg/constant/common"

type GOVParams struct{
	SalaryPerTx     uint64 // salary for each tx in block(mili constant)
	BasicSalary     uint64 // basic salary per block(mili constant)
	SellingBonds    *SellingBonds
}

type SellingBonds struct {
	BondsToSell    uint64
	BondPrice      uint64 // in Constant unit
	Maturity       uint32
	BuyBackPrice   uint64 // in Constant unit
	StartSellingAt uint32 // start selling bonds at block height
	SellingWithin  uint32 // selling bonds within n blocks
}

type DCBParams struct{

}

//xxx
func (DCBParams DCBParams) Hash() *common.Hash {

}
func (GOVParams GOVParams) Hash() *common.Hash {

}

//xxx
func (GOVParams GOVParams) Validate() bool {
	return true
}
func (DCBParams DCBParams) Validate() bool {
	return true
}
