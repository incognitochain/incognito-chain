package metadata

import (
	"bytes"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

type DividendPayment struct {
	// TODO(@0xbunyip): add reference back to instruction for validation
	DividendID uint64
	TokenID    *common.Hash

	MetadataBase
}

func (dp *DividendPayment) Hash() *common.Hash {
	record := strconv.FormatUint(dp.DividendID, 10)

	// final hash
	record += dp.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (dp *DividendPayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if there's a proposal to pay dividend
	forDCB := dp.TokenID.IsEqual(&common.DCBTokenID)
	id, _ := bcr.GetLatestDividendProposal(forDCB)
	if dp.DividendID != id {
		return false, errors.Errorf("Expect dividend id %d, got %d", id, dp.DividendID)
	}

	// Check if user is not rewarded and amount is correct
	unique, receiver, amount := txr.GetUniqueReceiver()
	if !unique {
		return false, errors.Errorf("Dividend payment tx must be sent to only one token holder")
	}

	receivers, amounts, hasValue, err := bcr.GetDividendReceiversForID(dp.DividendID, forDCB)
	if err != nil {
		return false, err
	}
	if !hasValue {
		return false, errors.Errorf("Error getting list of dividend receivers for dividend id %d", dp.DividendID)
	}

	found := false
	for j, rec := range receivers {
		if bytes.Equal(rec.Pk[:], receiver) {
			if amounts[j] != amount {
				return false, errors.Errorf("Incorrect amount of dividend payout to %x, expected %d, got %d", rec.Pk[:], amounts[j], amount)
			}
			found = true
		}
	}
	if !found {
		return false, errors.Errorf("User %x is not eligible for dividend paytout", receiver)
	}

	// TODO(@0xbunyip): check that only one payment to the same token holder in this block
	return false, nil
}

func (dp *DividendPayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil // No need to check for fee
}

func (dp *DividendPayment) ValidateMetadataByItself() bool {
	if !dp.TokenID.IsEqual(&common.DCBTokenID) && dp.TokenID.IsEqual(&common.GOVTokenID) {
		return false
	}
	return true
}

// CheckTransactionFee returns true since loan response tx doesn't have fee
func (dp *DividendPayment) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	return true
}
