package blockchain

import (
	"bytes"
	"encoding/json"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

const (
	dataSep  = "-"
	valueSep = "_"
)

//// Crowdsale bond
type CrowdsalePaymentInstruction struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	AssetID        common.Hash

	// Data for updating crowdsale on beacon component
	SaleID     []byte
	SentAmount uint64
	UpdateSale bool
}

func (inst *CrowdsalePaymentInstruction) String() (string, error) {
	data, err := json.Marshal(inst)
	return string(data), err
}

func ParseCrowdsalePaymentInstruction(data string) (*CrowdsalePaymentInstruction, error) {
	inst := &CrowdsalePaymentInstruction{}
	err := json.Unmarshal([]byte(data), inst)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func (inst *CrowdsalePaymentInstruction) Compare(inst2 *CrowdsalePaymentInstruction) bool {
	return bytes.Equal(inst.PaymentAddress.Pk, inst2.PaymentAddress.Pk) &&
		inst.Amount == inst2.Amount &&
		inst.AssetID.IsEqual(&inst2.AssetID)
}
