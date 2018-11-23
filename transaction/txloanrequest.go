package transaction

import (
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
	"encoding/hex"
)

type FeeArgs struct {
	SenderKey     *privacy.SpendingKey
	PaymentInfo   []*privacy.PaymentInfo
	Rts           map[byte]*common.Hash
	UsableTx      map[byte][]*Tx
	Commitments   map[byte]([][]byte)
	Fee           uint64
	SenderChainID byte
}

type LoanParams struct {
	InterestRate     uint64 `json:"InterestRate"`     // basis points, e.g. 125 represents 1.25%
	Maturity         uint64 `json:"Maturity"`         // seconds
	LiquidationStart uint64 `json:"LiquidationStart"` // ratio between collateral and debt to start auto-liquidation, stored in basis points
}

type LoanRequest struct {
	Params           LoanParams `json:"Params"`
	LoanID           []byte     `json:"LoanID"` // 32 bytes
	CollateralType   string     `json:"CollateralType"`
	CollateralAmount *big.Int   `json:"CollateralAmount"`

	LoanAmount     uint64                  `json:"LoanAmount"`
	ReceiveAddress *privacy.PaymentAddress `json:"ReceiveAddress"`

	KeyDigest []byte `json:"KeyDigest"` // 32 bytes, from sha256
}

func NewLoanRequest(data map[string]interface{}) *LoanRequest {
	loanParams := data["Params"].(map[string]interface{})
	result := LoanRequest{
		Params: LoanParams{
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

type TxLoanRequest struct {
	Tx
	*LoanRequest // data for a loan request
}

// CreateTxLoanRequest
// senderKey and paymentInfo is for paying fee
func CreateTxLoanRequest(
	feeArgs FeeArgs,
	loanRequest *LoanRequest,
) (*TxLoanRequest, error) {
	// Create tx for fee
	tx, err := CreateTx(
		feeArgs.SenderKey,
		feeArgs.PaymentInfo,
		feeArgs.Rts,
		feeArgs.UsableTx,
		feeArgs.Commitments,
		feeArgs.Fee,
		feeArgs.SenderChainID,
		false,
	)
	if err != nil {
		return nil, err
	}

	txLoanRequest := &TxLoanRequest{
		Tx:          *tx,
		LoanRequest: loanRequest,
	}

	return txLoanRequest, nil
}

func (tx *TxLoanRequest) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of collateral data
	record += string(tx.LoanID)
	record += tx.CollateralType
	record += tx.CollateralAmount.String()

	// add more hash of loan data
	record += string(tx.LoanID)
	record += string(tx.ReceiveAddress.ToBytes())

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxLoanRequest) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}

	// TODO: save and check type on-chain
	if tx.CollateralType != "ETH" {
		return false
	}

	if len(tx.KeyDigest) != LoanKeyDigestLen {
		return false
	}

	return true
}

func (tx *TxLoanRequest) GetType() string {
	return common.TxLoanRequest
}
