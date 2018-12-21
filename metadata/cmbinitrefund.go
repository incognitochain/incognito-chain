package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/pkg/errors"
)

const CMBInitRefundPeriod = 1000 // TODO(@0xbunyip): set appropriate value

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

func (creq *CMBInitRefund) Hash() *common.Hash {
	record := string(creq.MainAccount.ToBytes())

	// final hash
	record += string(creq.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (creq *CMBInitRefund) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// TODO(@0xbunyip): only accept response if it's still earlier than height+CMBInitRefundPeriod

	// Check if cmb init request existed
	meta, ok := txr.GetMetadata().(*CMBInitRefund)
	if !ok {
		return false, errors.Errorf("error parsing cmb init refund metadata")
	}
	_, _, txHash, state, err := bcr.GetCMB(meta.MainAccount.Pk[:])
	if err != nil {
		return false, err
	}

	// Check if it's at least CMBInitRefundPeriod since request
	_, blockHash, _, _, err := bcr.GetTransactionByHash(txHash)
	if err != nil {
		return false, err
	}
	reqBlockHeight, _, err := bcr.GetBlockHeightByBlockHash(blockHash)
	curBlockHeight := bcr.GetHeight()
	if curBlockHeight-reqBlockHeight < CMBInitRefundPeriod {
		return false, errors.Errorf("still waiting for repsponses, cannot refund cmb init request now")
	}
	return state == CMBRequested, nil
}

func (creq *CMBInitRefund) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return true, false, nil // DCB takes care of fee
}

func (creq *CMBInitRefund) ValidateMetadataByItself() bool {
	return true
}
