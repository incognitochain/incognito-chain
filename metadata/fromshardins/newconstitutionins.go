package fromshardins

import (
	"encoding/json"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/privacy"
)

type NewDCBConstitutionIns struct {
	SubmitProposalInfo component.SubmitProposalInfo
	DCBParams          component.DCBParams
	Voters             []privacy.PaymentAddress
}

func NewNewDCBConstitutionIns(submitProposalInfo component.SubmitProposalInfo, DCBParams component.DCBParams, voters []privacy.PaymentAddress) *NewDCBConstitutionIns {
	return &NewDCBConstitutionIns{SubmitProposalInfo: submitProposalInfo, DCBParams: DCBParams, Voters: voters}
}

func NewNewDCBConstitutionInsFromStr(inst string) (*NewDCBConstitutionIns, error) {
	newDCBConstitutionIns := &NewDCBConstitutionIns{}
	err := json.Unmarshal([]byte(inst), newDCBConstitutionIns)
	if err != nil {
		return nil, err
	}
	return newDCBConstitutionIns, nil
}

func (NewDCBConstitutionIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}

type NewGOVConstitutionIns struct {
	SubmitProposalInfo component.SubmitProposalInfo
	GOVParams          component.GOVParams
	Voters             []privacy.PaymentAddress
}

func NewNewGOVConstitutionIns(submitProposalInfo component.SubmitProposalInfo, GOVParams component.GOVParams, voters []privacy.PaymentAddress) *NewGOVConstitutionIns {
	return &NewGOVConstitutionIns{SubmitProposalInfo: submitProposalInfo, GOVParams: GOVParams, Voters: voters}
}

func NewNewGOVConstitutionInsFromStr(inst string) (*NewGOVConstitutionIns, error) {
	newGOVConstitutionIns := &NewGOVConstitutionIns{}
	err := json.Unmarshal([]byte(inst), newGOVConstitutionIns)
	if err != nil {
		return nil, err
	}
	return newGOVConstitutionIns, nil
}

func (NewGOVConstitutionIns) GetStringFormat() ([]string, error) {
	panic("implement me")
}
