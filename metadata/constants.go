package metadata

const (
	LoanKeyDigestLength = 32
	LoanKeyLength       = 32
)

const (
	InvalidMeta = iota
	LoanRequestMeta
	LoanResponseMeta
	LoanWithdrawMeta
	BuySellRequestMeta
)
