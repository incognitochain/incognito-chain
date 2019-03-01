package fromshardins

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

type NewDCBConstitutionIns struct {
	SubmitProposalInfo metadata.SubmitProposalInfo
	DCBParams          params.DCBParams
}

func (NewDCBConstitutionIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (NewDCBConstitutionIns) BuildTransaction(minerPrivateKey *privacy.SpendingKey, db database.DatabaseInterface) (metadata.Transaction, error) {
	panic("implement me")
}

func NewNewDCBConstitutionIns(submitProposalInfo metadata.SubmitProposalInfo, DCBParams params.DCBParams) *NewDCBConstitutionIns {
	return &NewDCBConstitutionIns{SubmitProposalInfo: submitProposalInfo, DCBParams: DCBParams}
}

type NewGOVConstitutionIns struct {
	SubmitProposalInfo metadata.SubmitProposalInfo
	GOVParams          params.GOVParams
}

func (NewGOVConstitutionIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

func (NewGOVConstitutionIns) BuildTransaction(minerPrivateKey *privacy.SpendingKey, db database.DatabaseInterface) (metadata.Transaction, error) {
	panic("implement me")
}

func NewNewGOVConstitutionIns(submitProposalInfo metadata.SubmitProposalInfo, GOVParams params.GOVParams) *NewGOVConstitutionIns {
	return &NewGOVConstitutionIns{SubmitProposalInfo: submitProposalInfo, GOVParams: GOVParams}
}
