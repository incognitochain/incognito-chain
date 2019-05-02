package component

import (
	"bytes"
	"encoding/json"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
)

type IssuingInfo struct {
	ReceiverAddress privacy.PaymentAddress
	Amount          uint64
	RequestedTxID   common.Hash
	TokenID         common.Hash
	CurrencyType    common.Hash
}

func ParseIssuingInfo(issuingInfoRaw string) (*IssuingInfo, error) {
	var issuingInfo IssuingInfo
	err := json.Unmarshal([]byte(issuingInfoRaw), &issuingInfo)
	if err != nil {
		return nil, err
	}
	return &issuingInfo, nil
}

func (info *IssuingInfo) Compare(info2 *IssuingInfo) bool {
	return bytes.Equal(info.ReceiverAddress.Pk, info2.ReceiverAddress.Pk) &&
		info.Amount == info2.Amount &&
		info.TokenID.IsEqual(&info2.TokenID)
}

type ContractingInfo struct {
	BurnerAddress     privacy.PaymentAddress
	BurnedConstAmount uint64
	RedeemAmount      uint64
	RequestedTxID     common.Hash
	CurrencyType      common.Hash
}

func ParseContractingInfo(contractingInfoRaw string) (*ContractingInfo, error) {
	var contractingInfo ContractingInfo
	err := json.Unmarshal([]byte(contractingInfoRaw), &contractingInfo)
	if err != nil {
		return nil, err
	}
	return &contractingInfo, nil
}

func (info *ContractingInfo) Compare(info2 *ContractingInfo) bool {
	return bytes.Equal(info.BurnerAddress.Pk, info2.BurnerAddress.Pk) &&
		info.BurnedConstAmount == info2.BurnedConstAmount
}
