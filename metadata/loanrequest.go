package metadata

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
)

type LoanRequest struct {
	Params           params.LoanParams `json:"Params"`
	LoanID           []byte            `json:"LoanID"` // 32 bytes
	CollateralType   string            `json:"CollateralType"`
	CollateralAmount *big.Int          `json:"CollateralAmount"`

	LoanAmount     uint64                  `json:"LoanAmount"`
	ReceiveAddress *privacy.PaymentAddress `json:"ReceiveAddress"`

	KeyDigest []byte `json:"KeyDigest"` // 32 bytes, from sha256

	MetadataBase
}

func NewLoanRequest(data map[string]interface{}) *LoanRequest {
	loanParams := data["Params"].(map[string]interface{})
	result := LoanRequest{
		Params: params.LoanParams{
			InterestRate:     uint64(loanParams["InterestRate"].(float64)),
			LiquidationStart: uint64(loanParams["LiquidationStart"].(float64)),
			Maturity:         uint32(loanParams["Maturity"].(float64)),
		},
		CollateralType: data["CollateralType"].(string),
		LoanAmount:     uint64(data["LoanAmount"].(float64)),
	}
	n := new(big.Int)
	n, ok := n.SetString(data["CollateralAmount"].(string), 10)
	if !ok {
		return nil
	}
	result.CollateralAmount = n
	key, err := wallet.Base58CheckDeserialize(data["ReceiveAddress"].(string))
	if err != nil {
		return nil
	}
	result.ReceiveAddress = &key.KeySet.PaymentAddress

	s, err := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s

	s, err = hex.DecodeString(data["KeyDigest"].(string))
	result.KeyDigest = s

	return &result
}

func (lr *LoanRequest) GetType() int {
	return LoanRequestMeta
}

func (lr *LoanRequest) Hash() *common.Hash {
	record := string(lr.LoanID)
	record += string(lr.Params.InterestRate)
	record += string(lr.Params.Maturity)
	record += string(lr.Params.LiquidationStart)
	record += lr.CollateralType
	record += lr.CollateralAmount.String()
	record += string(lr.LoanAmount)
	record += string(lr.ReceiveAddress.ToBytes())
	record += string(lr.KeyDigest)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lr *LoanRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	// Check if loan's params are correct
	dcbParams := bcr.GetDCBParams()
	validLoanParams := dcbParams.LoanParams
	ok := false
	for _, temp := range validLoanParams {
		if lr.Params == temp {
			ok = true
		}
	}
	if !ok {
		return false, fmt.Errorf("LoanRequest has incorrect params")
	}

	txs, err := bcr.GetLoanTxs(lr.LoanID)
	if err != nil {
		return false, err
	}

	if len(txs) > 0 {
		return false, fmt.Errorf("LoanID already existed")
	}
	return true, nil
}

func (lr *LoanRequest) ValidateSanityData() (bool, bool, error) {
	return true, true, nil
}

func (lr *LoanRequest) ValidateMetadataByItself() bool {
	return true
}
