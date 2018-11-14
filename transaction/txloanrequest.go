package transaction

import (
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

type FeeArgs struct {
	SenderKey     *client.SpendingKey
	PaymentInfo   []*client.PaymentInfo
	Rts           map[byte]*common.Hash
	UsableTx      map[byte][]*Tx
	Commitments   map[byte]([][]byte)
	Fee           uint64
	SenderChainID byte
}

type LoanParams struct {
	InterestRate     uint64 // basis points, e.g. 125 represents 1.25%
	Maturity         uint64 // seconds
	LiquidationStart uint64 // ratio between collateral and debt to start auto-liquidation, stored in basis points
}

type LoanRequest struct {
	Params           LoanParams
	LoanID           []byte // 32 bytes
	CollateralType   string
	CollateralTx     []byte // Tx hash in case of ETH
	CollateralAmount *big.Int

	LoanAmount     uint64
	ReceiveAddress *client.PaymentAddress

	KeyDigest []byte // 32 bytes, from sha256
}

type TxWithFee struct {
	*Tx // for fee only
}

type TxLoanRequest struct {
	TxWithFee
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
		TxWithFee:   TxWithFee{Tx: tx},
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
	record += string(tx.CollateralTx)
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

func (tx *TxWithFee) GetType() string {
	return tx.Tx.Type
}

func (tx *TxWithFee) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *TxWithFee) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

func (tx *TxWithFee) GetTxFee() uint64 {
	return tx.Tx.Fee
}
