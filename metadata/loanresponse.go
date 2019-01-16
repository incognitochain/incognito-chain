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
	isBoard := common.FalseValue
	txPubKey := txr.GetSigPubKey()
	fmt.Printf("check if created by dcb board: %v\n", txPubKey)
	for _, member := range bcr.GetBoardPubKeys("dcb") {
		fmt.Printf("member of board pubkey: %v\n", member)
		if bytes.Equal(member, txPubKey) {
			isBoard = common.TrueValue
		}
	}
	return isBoard
}

func (lr *LoanResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Validating LoanResponse with blockchain!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	// Check if only board members created this tx
	if !txCreatedByDCBBoardMember(txr, bcr) {
		return common.FalseValue, fmt.Errorf("Tx must be created by DCB Governor")
	}

	// Check if a loan request with the same id exists on any chain
	txHashes, err := bcr.GetLoanTxs(lr.LoanID)
	if err != nil {
		return common.FalseValue, err
	}
	fmt.Printf("GetLoanTxs found:\n")
	found := common.FalseValue
	for _, txHash := range txHashes {
		fmt.Printf("%x\n", txHash)
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return common.FalseValue, fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetMetadataType() {
		case LoanResponseMeta:
			{
				_, ok := txOld.GetMetadata().(*LoanResponse)
				if !ok {
					continue
				}
				// Check if the same user responses twice
				if bytes.Equal(txOld.GetSigPubKey(), txr.GetSigPubKey()) {
					return common.FalseValue, fmt.Errorf("Current board member already responded to loan request")
				}
			}
		case LoanRequestMeta:
			{
				_, ok := txOld.GetMetadata().(*LoanRequest)
				if !ok {
					continue
				}
				found = common.TrueValue
			}
		}
	}

	if found == common.FalseValue {
		return common.FalseValue, fmt.Errorf("Corresponding loan request not found")
	}
	fmt.Printf("Validate returns common.TrueValue!!!\n")
	return common.TrueValue, nil
}

func (lr *LoanResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if lr.Response != Accept && lr.Response != Reject {
		return common.FalseValue, common.FalseValue, nil
	}
	return common.FalseValue, common.TrueValue, nil // No need to check for fee
}

func (lr *LoanResponse) ValidateMetadataByItself() bool {
	return common.TrueValue
}

// CheckTransactionFee returns common.TrueValue since loan response tx doesn't have fee
func (lr *LoanResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	return common.TrueValue
}

// GetLoanResponses returns list of members who responded to a loan; input the hashes of request and response txs of the loan
func GetLoanResponses(txHashes [][]byte, bcr BlockchainRetriever) []ResponseData {
	data := []ResponseData{}
	for _, txHash := range txHashes {
		hash, err := (&common.Hash{}).NewHash(txHash)
		if err != nil {
			fmt.Printf("NewHash err: %x\n", txHash)
			continue
		}
		_, _, _, txOld, err := bcr.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			fmt.Printf("GetTxByHash err: %x\n", hash)
			continue
		}
		fmt.Printf("Type: %d\n", txOld.GetMetadataType())
		if txOld.GetMetadataType() == LoanResponseMeta {
			meta := txOld.GetMetadata().(*LoanResponse)
			respData := ResponseData{
				PublicKey: txOld.GetSigPubKey(),
				Response:  meta.Response,
			}
			data = append(data, respData)
		}
	}
	return data
}
