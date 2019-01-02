package metadata

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
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

func NewLoanRequest(data map[string]interface{}) (*LoanRequest, error) {
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
	fmt.Println("collat amount:", data["CollateralAmount"].(string))
	n, ok := n.SetString(data["CollateralAmount"].(string), 10)
	fmt.Printf("ok setstring: %v\n", ok)
	if !ok {
		return nil, errors.Errorf("Collateral amount incorrect")
	}
	result.CollateralAmount = n
	key, err := wallet.Base58CheckDeserialize(data["ReceiveAddress"].(string))
	fmt.Printf("err receiveaddress: %v\n", err)
	if err != nil {
		return nil, errors.Errorf("ReceiveAddress incorrect")
	}
	result.ReceiveAddress = &key.KeySet.PaymentAddress

	s, err := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s

	s, err = hex.DecodeString(data["KeyDigest"].(string))
	result.KeyDigest = s

	result.Type = LoanRequestMeta
	return &result, nil
}

func (lr *LoanRequest) Hash() *common.Hash {
	record := string(lr.LoanID)
	record += string(lr.Params.InterestRate)
	record += string(lr.Params.Maturity)
	record += string(lr.Params.LiquidationStart)
	record += lr.CollateralType
	record += lr.CollateralAmount.String()
	record += string(lr.LoanAmount)
	record += string(lr.ReceiveAddress.Bytes())
	record += string(lr.KeyDigest)

	// final hash
	record += string(lr.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lr *LoanRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
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

func (lr *LoanRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(lr.KeyDigest) != LoanKeyDigestLength {
		return false, false, errors.Errorf("KeyDigest is not 32 bytes")
	}
	return true, true, nil // continue to check for fee
}

func (lr *LoanRequest) ValidateMetadataByItself() bool {
	return true
}
