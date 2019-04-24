package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/pkg/errors"
)

const (
	dataSep  = "-"
	valueSep = "_"
)

var (
	loanIDKeyPrefix   = "loanID-"
	loanRespKeyPrefix = "loanResp-"
	saleDataPrefix    = "sale-"
)

func getLoanRequestKeyBeacon(loanID []byte) string {
	return loanIDKeyPrefix + string(loanID)
}

type LoanRespData struct {
	SenderPubkey []byte
	Response     metadata.ValidLoanResponse
}

func (lrd *LoanRespData) String() string {
	return strings.Join([]string{base64.StdEncoding.EncodeToString(lrd.SenderPubkey), string(lrd.Response)}, dataSep)
}

func parseLoanRespData(data string) (*LoanRespData, error) {
	s := strings.Split(data, dataSep)
	if len(s) != 2 {
		return nil, errors.Errorf("Error parsing loan response data")
	}
	errSaver := &metadata.ErrorSaver{}
	sender, errSender := base64.StdEncoding.DecodeString(s[0])
	response, errResp := strconv.Atoi(s[1])
	if errSaver.Save(errSender, errResp) != nil {
		return nil, errSaver.Get()
	}
	lrd := &LoanRespData{
		SenderPubkey: sender,
		Response:     metadata.ValidLoanResponse(response),
	}
	return lrd, nil
}

func getLoanResponseKeyBeacon(loanID []byte) string {
	return loanRespKeyPrefix + string(loanID)
}

func getLoanResponseValueBeacon(data []*LoanRespData) string {
	value, _ := json.Marshal(data)
	return string(value)
}

func parseLoanResponseValueBeacon(data string) ([]*LoanRespData, error) {
	lrds := []*LoanRespData{}
	err := json.Unmarshal([]byte(data), &lrds)
	return lrds, err
}

//// Crowdsale bond
func getSaleDataKeyBeacon(saleID []byte) string {
	return saleDataPrefix + string(saleID)
}

func getSaleDataValueBeacon(data *component.SaleData) string {
	value, _ := json.Marshal(data)
	return string(value)
}

func parseSaleDataValueBeacon(value string) (*component.SaleData, error) {
	data := &component.SaleData{}
	err := json.Unmarshal([]byte(value), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// type CrowdsalePaymentInstruction struct {
// 	PaymentAddress privacy.PaymentAddress
// 	Amount         uint64
// 	AssetID        common.Hash

// 	// Data for updating crowdsale on beacon component
// 	SaleID     []byte
// 	SentAmount uint64
// 	UpdateSale bool
// }

// func (inst *CrowdsalePaymentInstruction) String() (string, error) {
// 	data, err := json.Marshal(inst)
// 	return string(data), err
// }

// func ParseCrowdsalePaymentInstruction(data string) (*CrowdsalePaymentInstruction, error) {
// 	inst := &CrowdsalePaymentInstruction{}
// 	err := json.Unmarshal([]byte(data), inst)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return inst, nil
// }

// func (inst *CrowdsalePaymentInstruction) Compare(inst2 *CrowdsalePaymentInstruction) bool {
// 	return bytes.Equal(inst.PaymentAddress.Pk, inst2.PaymentAddress.Pk) &&
// 		inst.Amount == inst2.Amount &&
// 		inst.AssetID.IsEqual(&inst2.AssetID)
// }
