package transaction

import (
	"strconv"

	"github.com/ninjadotorg/constant/common"
)

type LoanResponse struct {
	LoanID     []byte
	ValidUntil uint64
	KeyDigest  []byte // 32 bytes, from sha256
}

type TxLoanResponse struct {
	TxWithFee
	*LoanResponse // data for a loan response
}

func CreateTxLoanResponse(
	feeArgs FeeArgs,
	loanResponse *LoanResponse,
) (*TxLoanResponse, error) {
	// Create tx for fee
	tx, err := CreateTx(
		feeArgs.SenderKey,
		feeArgs.PaymentInfo,
		feeArgs.Rts,
		feeArgs.UsableTx,
		feeArgs.Nullifiers,
		feeArgs.Commitments,
		feeArgs.Fee,
		feeArgs.AssetType,
		feeArgs.SenderChainID,
	)
	if err != nil {
		return nil, err
	}

	txLoanResponse := &TxLoanResponse{
		TxWithFee:    TxWithFee{Tx: tx},
		LoanResponse: loanResponse,
	}

	return txLoanResponse, nil
}

func (tx *TxLoanResponse) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of loan response data
	record += string(tx.LoanID)
	record += strconv.Itoa(int(tx.ValidUntil))
	record += string(tx.KeyDigest)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxLoanResponse) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}

	// TODO(@0xbunyip): check if only board members created this tx
	if len(tx.KeyDigest) != LoanKeyDigestLen {
		return false
	}

	return true
}
