package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

const (
	dataSep  = "-"
	valueSep = "_"
)

var (
	loanIDKeyPrefix         = "loanID-"
	loanRespKeyPrefix       = "loanResp-"
	saleDataPrefix          = "sale-"
	dividendPrefixDCB       = "divDCB"
	dividendPrefixGOV       = "divGOV"
	dividendSubmitPrefix    = "divSub"
	dividendAggregatePrefix = "divAgg"
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

func getSaleDataKeyBeacon(saleID []byte) string {
	return saleDataPrefix + string(saleID)
}

func getSaleDataValueBeacon(data *params.SaleData) string {
	value, _ := json.Marshal(data)
	return string(value)
}

func parseSaleDataValueBeacon(value string) (*params.SaleData, error) {
	data := &params.SaleData{}
	err := json.Unmarshal([]byte(value), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type CrowdsalePaymentInstruction struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	AssetID        common.Hash

	// Data for updating crowdsale on beacon params
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

func getDCBDividendKeyBeacon() string {
	return dividendPrefixDCB
}

func getGOVDividendKeyBeacon() string {
	return dividendPrefixGOV
}

func getDividendValueBeacon(amounts []uint64) string {
	value, _ := json.Marshal(amounts)
	return string(value)
}

func parseDividendValueBeacon(value string) ([]uint64, error) {
	data := []uint64{}
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getDividendSubmitKeyBeacon(shardID byte, dividendID uint64, tokenID *common.Hash) string {
	return strings.Join([]string{dividendSubmitPrefix, string(shardID), strconv.FormatUint(dividendID, 10), tokenID.String()}, "")
}

func getDividendSubmitValueBeacon(shardTokenAmount uint64) string {
	return strconv.FormatUint(shardTokenAmount, 10)
}

func parseDividendSubmitValueBeacon(value string) uint64 {
	shardTokenAmount, _ := strconv.ParseUint(value, 10, 64)
	return shardTokenAmount
}

func getDividendAggregatedKeyBeacon(dividendID uint64, tokenID *common.Hash) string {
	return strings.Join([]string{dividendAggregatePrefix, strconv.FormatUint(dividendID, 10), tokenID.String()}, "")
}

func getDividendAggregatedValueBeacon(totalTokenAmount, cstToPayout uint64) string {
	return strings.Join([]string{strconv.FormatUint(totalTokenAmount, 10), strconv.FormatUint(cstToPayout, 10)}, dataSep)
}

func parseDividendAggregatedValueBeacon(value string) (uint64, uint64) {
	splits := strings.Split(value, dataSep)
	totalTokenAmount, _ := strconv.ParseUint(splits[0], 10, 64)
	cstToPayout, _ := strconv.ParseUint(splits[1], 10, 64)
	return totalTokenAmount, cstToPayout
}
