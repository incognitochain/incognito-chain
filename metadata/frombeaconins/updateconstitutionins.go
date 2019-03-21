package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/privacy"
)

type UpdateDCBConstitutionIns struct {
	SubmitProposalInfo component.SubmitProposalInfo
	DCBParams          component.DCBParams
	Voters             []privacy.PaymentAddress
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

func NewUpdateDCBConstitutionIns(submitProposalInfo component.SubmitProposalInfo, DCBParams component.DCBParams, voters []privacy.PaymentAddress) *UpdateDCBConstitutionIns {
	return &UpdateDCBConstitutionIns{SubmitProposalInfo: submitProposalInfo, DCBParams: DCBParams, Voters: voters}
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
	Voters             []privacy.PaymentAddress
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

func NewUpdateGOVConstitutionIns(submitProposalInfo component.SubmitProposalInfo, GOVParams component.GOVParams, voters []privacy.PaymentAddress) *UpdateGOVConstitutionIns {
	return &UpdateGOVConstitutionIns{SubmitProposalInfo: submitProposalInfo, GOVParams: GOVParams, Voters: voters}
}

func NewUpdateGOVConstitutionInsFromStr(inst []string) (*UpdateGOVConstitutionIns, error) {
	updateGOVConstitutionIns := &UpdateGOVConstitutionIns{}
	err := json.Unmarshal([]byte(inst[2]), updateGOVConstitutionIns)
	if err != nil {
		return nil, err
	}
	return updateGOVConstitutionIns, nil
}
