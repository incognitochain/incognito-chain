package transaction

// TxCustomTokenPrivacy is class tx which is inherited from constant tx(supporting privacy) for fee
// and contain data(with supporting privacy format) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// TxCustomTokenPrivacy is an advance format of TxCustomToken
type TxCustomTokenPrivacy struct {
	Tx                                    // inherit from normal tx of constant(supporting privacy)
	TxTokenPrivacyData TxTokenPrivacyData // v
}
