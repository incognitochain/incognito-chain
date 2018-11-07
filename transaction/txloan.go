package transaction

import (
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

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

type TxLoanRequest struct {
	*Tx          // for fee
	*LoanRequest // data for a loan request
}

// CreateTxLoanRequest
// senderKey and paymentInfo is for paying fee
func CreateTxLoanRequest(
	senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	nullifiers map[byte]([][]byte),
	commitments map[byte]([][]byte),
	fee uint64,
	assetType string,
	senderChainID byte,
	loanRequest *LoanRequest,
) (*TxLoanRequest, error) {
	// Create tx for fee
	tx, err := CreateTx(
		senderKey,
		paymentInfo,
		rts,
		usableTx,
		nullifiers,
		commitments,
		fee,
		assetType,
		senderChainID,
	)
	if err != nil {
		return nil, err
	}

	txLoanRequest := &TxLoanRequest{
		Tx:          tx,
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

	// TODO: LoanID unique
	// TODO: save and check type on-chain
	if tx.CollateralType != "ETH" {
		return false
	}

	return true
}

func (tx *TxLoanRequest) GetType() string {
	return tx.Type
}

func (tx *TxLoanRequest) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *TxLoanRequest) GetSenderAddrLastByte() byte {
	return tx.AddressLastByte
}

func (tx *TxLoanRequest) GetTxFee() uint64 {
	return tx.Fee
}
