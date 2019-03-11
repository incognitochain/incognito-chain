package metadata

import (
	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/database"
	"github.com/big0t/constant-chain/privacy"
	"github.com/pkg/errors"
)

const CMBInitRefundPeriod = uint64(1000) // TODO(@0xbunyip): set appropriate value

const (
	CMBInvalid = uint8(iota)
	CMBRequested
	CMBApproved
	CMBRefunded
)

type CMBInitRefund struct {
	MainAccount privacy.PaymentAddress

	MetadataBase
}

func (cref *CMBInitRefund) Hash() *common.Hash {
	record := cref.MainAccount.String()

	// final hash
	record += cref.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (cref *CMBInitRefund) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if cmb init request existed
	_, _, _, txHash, state, _, err := bcr.GetCMB(cref.MainAccount.Bytes())
	if err != nil {
		return false, err
	}

	// Check if it's at least CMBInitRefundPeriod since request
	_, blockHash, _, _, err := bcr.GetTransactionByHash(txHash)
	if err != nil {
		return false, err
	}
	reqBlockHeight, _, _ := bcr.GetBlockHeightByBlockHash(blockHash)
	curBlockHeight, err := bcr.GetTxChainHeight(txr)
	if err != nil || curBlockHeight < CMBInitRefundPeriod+reqBlockHeight {
		return false, errors.Errorf("still waiting for repsponses, cannot refund cmb init request now")
	}
	return state == CMBRequested, nil
}

func (cref *CMBInitRefund) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, false, nil // DCB takes care of fee
}

func (cref *CMBInitRefund) ValidateMetadataByItself() bool {
	return true
}

func (cref *CMBInitRefund) CalculateSize() uint64 {
	return calculateSize(cref)
}
