package metadata

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
)

type TxRetriever interface {
	GetTxFee() uint64
}

type MempoolRetriever interface {
	GetPoolNullifiers() map[common.Hash][][]byte
}

type BlockchainRetriever interface {
	GetDCBParams() params.DCBParams
	GetLoanTxs([]byte) ([][]byte, error)
}

type Metadata interface {
	ValidateWithBlockChain(BlockchainRetriever) error
	Validate() error
	Process() error
	CheckTransactionFee(TxRetriever, uint64) bool
}

type MetadataBase struct {
}

func (mb *MetadataBase) Validate() error {
	return nil
}

func (mb *MetadataBase) Process() error {
	return nil
}

func (mb *MetadataBase) CheckTransactionFee(tr TxRetriever, minFee uint64) bool {
	txFee := tr.GetTxFee()
	if txFee < minFee {
		return false
	}
	return true
}
