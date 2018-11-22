package transaction

const (
	// TxVersion is the current latest supported transaction version.
	TxVersion = 1

	// NumDescInputs max number of input notes in a JSDesc
	NumDescInputs = 2

	// NumDescOutputs max number of output notes in a JSDesc
	NumDescOutputs = 2 // b

	LoanKeyDigestLen = 32 // number of bytes of a loan's secret key digest

	LoanKeyLen = 32
)

const (
	NoSort = iota
	SortByAmount
	SortByCreatedTime
)

const (
	CustomTokenInit = iota
	CustomTokenTransfer
	VoteDCBBoard
	VoteGOVBoard
)
