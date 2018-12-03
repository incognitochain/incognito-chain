package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"strconv"
)

// TxCustomTokenPrivacy is class tx which is inherited from constant tx(supporting privacy) for fee
// and contain data(with supporting privacy format) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// TxCustomTokenPrivacy is an advance format of TxCustomToken
// so that user need to spend a lot fee to create this class tx
type TxCustomTokenPrivacy struct {
	Tx                                    // inherit from normal tx of constant(supporting privacy) with a high fee to ensure that tx could contain a big data of privacy for token
	TxTokenPrivacyData TxTokenPrivacyData // supporting privacy format
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomTokenPrivacy) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of tx custom token privacy
	record += tx.TxTokenPrivacyData.PropertyName
	record += tx.TxTokenPrivacyData.PropertySymbol
	record += strconv.Itoa(tx.TxTokenPrivacyData.Type)
	record += strconv.Itoa(int(tx.TxTokenPrivacyData.Amount))
	for _, desc := range tx.TxTokenPrivacyData.Descs {
		record += desc.toString()
	}

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
