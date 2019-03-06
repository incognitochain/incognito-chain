package fromshardins

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/blockchain/component"
)

type NewDCBConstitutionIns struct {
	SubmitProposalInfo component.SubmitProposalInfo
	DCBParams          component.DCBParams
	Voter              component.Voter
}

func NewNewDCBConstitutionIns(submitProposalInfo component.SubmitProposalInfo, DCBParams component.DCBParams, voter component.Voter) *NewDCBConstitutionIns {
	return &NewDCBConstitutionIns{SubmitProposalInfo: submitProposalInfo, DCBParams: DCBParams, Voter: voter}
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
	Voter              component.Voter
}

func NewNewGOVConstitutionIns(submitProposalInfo component.SubmitProposalInfo, GOVParams component.GOVParams, voter component.Voter) *NewGOVConstitutionIns {
	return &NewGOVConstitutionIns{SubmitProposalInfo: submitProposalInfo, GOVParams: GOVParams, Voter: voter}
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
