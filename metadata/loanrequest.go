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

func NewLoanRequest(data map[string]interface{}) (Metadata, error) {
	loanParams := data["Params"].(map[string]interface{})
	result := LoanRequest{
		Params: params.LoanParams{
			InterestRate:     uint64(loanParams["InterestRate"].(float64)),
			LiquidationStart: uint64(loanParams["LiquidationStart"].(float64)),
			Maturity:         uint64(loanParams["Maturity"].(float64)),
		},
		CollateralType: data["CollateralType"].(string),
		LoanAmount:     uint64(data["LoanAmount"].(float64)),
	}
	n := new(big.Int)
	n, ok := n.SetString(data["CollateralAmount"].(string), 10)
	if !ok {
		return nil, errors.Errorf("Collateral amount incorrect")
	}
	result.CollateralAmount = n
	keyWallet, err := wallet.Base58CheckDeserialize(data["ReceiveAddress"].(string))
	fmt.Printf("err receiveaddress: %v\n", err)
	if err != nil {
		return nil, errors.Errorf("ReceiveAddress incorrect")
	}
	result.ReceiveAddress = &keyWallet.KeySet.PaymentAddress

	s, err := hex.DecodeString(data["LoanID"].(string))
	if err != nil {
		return nil, errors.Errorf("LoanID incorrect")
	}
	result.LoanID = s

	s, err = hex.DecodeString(data["KeyDigest"].(string))
	if err != nil {
		return nil, errors.Errorf("KeyDigest incorrect")
	}
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
	record += lr.ReceiveAddress.String()
	record += string(lr.KeyDigest)

	// final hash
	record += string(lr.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lr *LoanRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Validating LoanRequest with blockchain!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	// Check if loan's params are correct
	dcbParams := bcr.GetDCBParams()
	validLoanParams := dcbParams.LoanParams
	ok := common.FalseValue
	for _, temp := range validLoanParams {
		if lr.Params == temp {
			ok = common.TrueValue
		}
	}
	if !ok {
		return common.FalseValue, fmt.Errorf("LoanRequest has incorrect params")
	}

	txs, err := bcr.GetLoanTxs(lr.LoanID)
	if err != nil {
		return common.FalseValue, err
	}

	if len(txs) > 0 {
		return common.FalseValue, fmt.Errorf("LoanID already existed")
	}
	return common.TrueValue, nil
}

func (lr *LoanRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(lr.KeyDigest) != LoanKeyDigestLength {
		return common.FalseValue, common.FalseValue, errors.Errorf("KeyDigest is not 32 bytes")
	}
	return common.TrueValue, common.TrueValue, nil // continue to check for fee
}

func (lr *LoanRequest) ValidateMetadataByItself() bool {
	return common.TrueValue
}
