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
	Reject ValidLoanResponse = iota
	Accept
)

type LoanResponse struct {
	LoanID   []byte
	Response ValidLoanResponse

	MetadataBase
}

type ResponseData struct {
	PublicKey []byte
	Response  ValidLoanResponse
}

func NewLoanResponse(data map[string]interface{}) (Metadata, error) {
	result := LoanResponse{}
	s, _ := hex.DecodeString(data["LoanID"].(string))
	result.LoanID = s

	result.Response = ValidLoanResponse(int(data["Response"].(float64)))
	result.Type = LoanResponseMeta

	return &result, nil
}

func (lr *LoanResponse) Hash() *common.Hash {
	record := string(lr.LoanID)
	record += string(lr.Response)

	// final hash
	record += string(lr.MetadataBase.Hash()[:])
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func txCreatedByDCBBoardMember(txr Transaction, bcr BlockchainRetriever) bool {
	isBoard := false
	txPubKey := txr.GetJSPubKey()
	fmt.Printf("check if created by dcb board: %v\n", txPubKey)
	for _, member := range bcr.GetDCBBoardPubKeys() {
		fmt.Printf("member of board pubkey: %v\n", member)
		if bytes.Equal(member, txPubKey) {
			isBoard = true
		}
	}
	return isBoard
}

func (lr *LoanResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Validating LoanResponse with blockchain!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	// Check if only board members created this tx
	if !txCreatedByDCBBoardMember(txr, bcr) {
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

// GetLoanResponses returns list of members who responded to a loan; input the hashes of request and response txs of the loan
func GetLoanResponses(txHashes [][]byte, bcr BlockchainRetriever) []ResponseData {
	data := []ResponseData{}
	for _, txHash := range txHashes {
		hash, err := (&common.Hash{}).NewHash(txHash)
		if err != nil {
			continue
		}
		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			continue
		}
		if txOld.GetMetadataType() == LoanResponseMeta {
			meta := txOld.GetMetadata().(*LoanResponse)
			respData := ResponseData{
				PublicKey: txOld.GetJSPubKey(),
				Response:  meta.Response,
			}
			data = append(data, respData)
		}
	}
	return data
}
