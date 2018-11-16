package voting

import "github.com/ninjadotorg/constant/common"

type GovProposalData struct {
	GovParams GovParams
	Explaination      string
}

type DCBProposalData struct {
	DCBParams DCBParams
	Explaination      string
}

func (DCBProposalData DCBProposalData) Validate() bool {
	return DCBProposalData.DCBParams.Validate() && ValidateExplaination(DCBProposalData.Explaination)
}

func (GovProposalData GovProposalData) Validate() bool {
	return GovProposalData.GovParams.Validate() && ValidateExplaination(GovProposalData.Explaination)
}

func ValidateExplaination(explaination string) bool {
	return true
}

func (data DCBProposalData) Hash() *common.Hash {
	record := string(common.ToBytes(data.DCBParams.Hash()))
	record += data.Explaination
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (data GovProposalData) Hash() *common.Hash {
	record := string(common.ToBytes(data.GovParams.Hash()))
	record += data.Explaination
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

