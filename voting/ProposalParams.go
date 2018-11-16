package voting

import "github.com/ninjadotorg/constant/common"

type GovParams struct{

}

type DCBParams struct{

}

//xxx
func (DCBParams DCBParams) Hash() *common.Hash {

}
func (GovParams GovParams) Hash() *common.Hash {

}

//xxx
func (GovParams GovParams) Validate() bool {
	return true
}
func (DCBParams DCBParams) Validate() bool {
	return true
}
