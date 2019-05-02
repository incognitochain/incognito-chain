package metadata

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

type LoanRequest struct {
	Params           component.LoanParams `json:"Params"`
	LoanID           []byte               `json:"LoanID"` // 32 bytes
	CollateralType   string               `json:"CollateralType"`
	CollateralAmount *big.Int             `json:"CollateralAmount"`

	LoanAmount     uint64                  `json:"LoanAmount"`
	ReceiveAddress *privacy.PaymentAddress `json:"ReceiveAddress"`

	KeyDigest []byte `json:"KeyDigest"` // 32 bytes, from sha256

	MetadataBase
}

func NewLoanRequest(data map[string]interface{}) (Metadata, error) {
	loanParams := data["Params"].(map[string]interface{})
	result := LoanRequest{
		Params: component.LoanParams{
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
	hash := common.HashH([]byte(record))
	return &hash
}

func (lr *LoanRequest) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Validating LoanRequest with blockchain!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	// Check if loan's component are correct
	dcbParams := bcr.GetDCBParams()
	validLoanParams := dcbParams.ListLoanParams
	ok := false
	for _, temp := range validLoanParams {
		if lr.Params == temp {
			ok = true
		}
	}
	if !ok {
		return false, errors.New("LoanRequest has incorrect component")
	}

	txHash, _ := bcr.GetLoanReq(lr.LoanID)
	if txHash != nil && len(txHash) > 0 {
		return false, errors.New("LoanID already existed")
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

type LoanRequestAction struct {
	LoanID []byte
	TxID   *common.Hash
}

func (lr *LoanRequest) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	lrActionValue := getLoanRequestActionValue(lr.LoanID, txr.Hash())
	lrAction := []string{strconv.Itoa(LoanRequestMeta), lrActionValue}
	return [][]string{lrAction}, nil
}

func getLoanRequestActionValue(loanID []byte, txHash *common.Hash) string {
	value, _ := json.Marshal(LoanRequestAction{
		LoanID: loanID,
		TxID:   txHash,
	})
	return string(value)
}

func ParseLoanRequestActionValue(value string) ([]byte, *common.Hash, error) {
	data := &LoanRequestAction{}
	err := json.Unmarshal([]byte(value), data)
	if err != nil {
		return nil, nil, err
	}
	return data.LoanID, data.TxID, nil
}

func (lr *LoanRequest) CalculateSize() uint64 {
	return calculateSize(lr)
}
