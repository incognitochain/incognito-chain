package metadata

type TxRetriever interface {
	GetTxFee() uint64
}

type Metadata interface {
	Validate() error
	Process() error
	CheckTransactionFee(TxRetriever, uint64) bool
}
