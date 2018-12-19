package metadata

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type ValidLoanResponse int

const (
	Accept ValidLoanResponse = iota
	Reject
)

type LoanResponse struct {
	LoanID     []byte
	Response   ValidLoanResponse
	ValidUntil int32

	MetadataBase
}

func NewLoanResponse(data map[string]interface{}) *LoanResponse {
	result := LoanResponse{
		ValidUntil: int32(data["ValidUntil"].(float64)),
	}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s

	result.Response = ValidLoanResponse(int(data["Response"].(float64)))
	result.Type = LoanResponseMeta

	return &result
}

func (lr *LoanResponse) Hash() *common.Hash {
	record := string(lr.LoanID)
	record += string(lr.Response)
	record += string(lr.ValidUntil)

	// final hash
	record += string(lr.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (lr *LoanResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// Check if only board members created this tx
	isBoard := false
	for _, gov := range bcr.GetDCBBoardPubKeys() {
		// TODO(@0xbunyip): change gov to []byte or use Base58Decode for entire payment address of governors
		if bytes.Equal([]byte(gov), txr.GetJSPubKey()) {
			isBoard = true
		}
	}
	if !isBoard {
		return false, fmt.Errorf("Tx must be created by DCB Governor")
	}

	// Check if a loan request with the same id exists on any chain
	txHashes, err := bcr.GetLoanTxs(lr.LoanID)
	if err != nil {
		return false, err
	}
	found := false
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return false, fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetMetadataType() {
		case LoanResponseMeta:
			{
				// Check if the same user responses twice
				if bytes.Equal(txOld.GetJSPubKey(), txr.GetJSPubKey()) {
					return false, fmt.Errorf("Current board member already responded to loan request")
				}
				meta := txOld.GetMetadata()
				if meta == nil {
					continue
				}
				metaOld := meta.(*LoanResponse)
				if lr.ValidUntil != metaOld.ValidUntil {
					return false, fmt.Errorf("Valid deadline of all responses of a loan must be the same")
				}
			}
		case LoanRequestMeta:
			{
				meta := txOld.GetMetadata()
				if meta == nil {
					return false, fmt.Errorf("Error parsing loan request tx")
				}
				found = true
			}
		}
	}

	if found == false {
		return false, fmt.Errorf("Corresponding loan request not found")
	}
	return true, nil
}

func (lr *LoanResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if lr.Response != Accept && lr.Response != Reject {
		return false, false, nil
	}
	return false, true, nil // No need to check for fee
}

func (lr *LoanResponse) ValidateMetadataByItself() bool {
	return true
}

// CheckTransactionFee returns true since loan response tx doesn't have fee
func (lr *LoanResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	return true
}
