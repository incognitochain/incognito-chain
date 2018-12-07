package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

type DividendProposal struct {
	TotalAmount uint64 // total Constant to pay dividend
	PayoutID    uint64 // increasing ID for each type of token
	TokenID     *common.Hash
}

type DividendInfo struct {
	TokenHolder privacy.PaymentAddress
	Amount      uint64
}

type Dividend struct {
	PayoutID uint64
	TokenID  *common.Hash

	MetadataBase
}

func NewDividend(data map[string]interface{}) *Dividend {
	result := Dividend{
		PayoutID: uint64(data["PayoutID"].(float64)),
	}
	s, _ := hex.DecodeString(data["TokenID"].(string))
	copy(result.TokenID[:], s)
	return &result
}

func (div *Dividend) GetType() int {
	return DividendMeta
}

func (div *Dividend) Hash() *common.Hash {
	record := fmt.Sprintf("%d", div.PayoutID)
	record += string(div.TokenID[:])

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (div *Dividend) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte) (bool, error) {
	// Check if there's a proposal to pay dividend
	// TODO(@0xbunyip): get current proposal and check if it is dividend payout
	proposal := &DividendProposal{}
	_, tokenHolders, correctAmounts, err := bcr.GetAmountPerAccount(proposal)
	if err != nil {
		return false, err
	}

	// Check if user is not rewarded and amount is correct
	receivers, recAmounts := txr.GetReceivers()
	for j, rec := range receivers {
		// Check amount
		count := 0
		for i, holder := range tokenHolders {
			temp, _ := hex.DecodeString(holder)
			paymentAddress := (&privacy.PaymentAddress{}).FromBytes(temp)
			if bytes.Equal(paymentAddress.Pk[:], rec) {
				count += 1
				if correctAmounts[i] != recAmounts[j] {
					return false, fmt.Errorf("Payment amount for user %s incorrect, found %d instead of %d", holder, recAmounts[j], correctAmounts[i])
				}
			}
		}

		if count == 0 {
			return false, fmt.Errorf("User %s isn't eligible for receiving dividend", rec)
		} else if count > 1 {
			return false, fmt.Errorf("Multiple dividend payments found for user %s", rec)
		}
	}
	return false, nil
}

func (div *Dividend) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil // No need to check for fee
}

func (div *Dividend) ValidateMetadataByItself() bool {
	return true
}

// CheckTransactionFee returns true since loan response tx doesn't have fee
func (div *Dividend) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	return true
}
