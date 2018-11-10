package lvdb

// TODO(@0xbunyip): implement these
func (db *db) StoreLoanRequest(loanID, txHash []byte) error {
	return nil
}

func (db *db) StoreLoanResponse(loanID, txHash []byte) error {
	return nil
}

func (db *db) GetLoanTxs(loanID []byte) ([][]byte, error) {
	return nil, nil
}
