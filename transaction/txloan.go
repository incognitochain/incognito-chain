package transaction

import (
	"math/big"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

type TxLoanRequest struct {
	Tx

	LoanID           []byte // 32 bytes
	CollateralType   string
	CollateralTx     []byte // Tx hash in case of ETH
	CollateralAmount *big.Int

	LoanAmount     uint64
	ReceiveAddress *client.PaymentAddress

	KeyDigest []byte // 32 bytes, from sha256
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
	record += strconv.Itoa(tx.LoanID)
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
