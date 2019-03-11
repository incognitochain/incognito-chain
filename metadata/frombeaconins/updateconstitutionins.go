package frombeaconins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/blockchain/component"
	"strconv"
)

type UpdateDCBConstitutionIns struct {
	SubmitProposalInfo component.SubmitProposalInfo
	DCBParams          component.DCBParams
	Voter              component.Voter
}

func (updateDCBConstitutionIns *UpdateDCBConstitutionIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(updateDCBConstitutionIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.UpdateDCBConstitutionIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewUpdateDCBConstitutionIns(submitProposalInfo component.SubmitProposalInfo, DCBParams component.DCBParams, voter component.Voter) *UpdateDCBConstitutionIns {
	return &UpdateDCBConstitutionIns{SubmitProposalInfo: submitProposalInfo, DCBParams: DCBParams, Voter: voter}
}

func NewUpdateDCBConstitutionInsFromStr(inst []string) (*UpdateDCBConstitutionIns, error) {
	updateDCBConstitutionIns := &UpdateDCBConstitutionIns{}
	err := json.Unmarshal([]byte(inst[2]), updateDCBConstitutionIns)
	if err != nil {
		return nil, err
	}
	return updateDCBConstitutionIns, nil
}

type UpdateGOVConstitutionIns struct {
	SubmitProposalInfo component.SubmitProposalInfo
	GOVParams          component.GOVParams
	Voter              component.Voter
}

func (updateGOVConstitutionIns *UpdateGOVConstitutionIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(updateGOVConstitutionIns)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(component.UpdateGOVConstitutionIns),
		strconv.Itoa(-1),
		string(content),
	}, nil
}

func NewUpdateGOVConstitutionIns(submitProposalInfo component.SubmitProposalInfo, GOVParams component.GOVParams, voter component.Voter) *UpdateGOVConstitutionIns {
	return &UpdateGOVConstitutionIns{SubmitProposalInfo: submitProposalInfo, GOVParams: GOVParams, Voter: voter}
}

func NewUpdateGOVConstitutionInsFromStr(inst []string) (*UpdateGOVConstitutionIns, error) {
	updateGOVConstitutionIns := &UpdateGOVConstitutionIns{}
	err := json.Unmarshal([]byte(inst[2]), updateGOVConstitutionIns)
	if err != nil {
		return nil, err
	}
	return updateGOVConstitutionIns, nil
}
