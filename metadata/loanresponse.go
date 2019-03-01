package metadata

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
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
	record += lr.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func txCreatedByDCBBoardMember(txr Transaction, bcr BlockchainRetriever) bool {
	isBoard := false
	txPubKey := txr.GetSigPubKey()
	fmt.Printf("check if created by dcb board: %v\n", txPubKey)
	for _, member := range bcr.GetBoardPubKeys(DCBBoard) {
		fmt.Printf("member of board pubkey: %v\n", member)
		if bytes.Equal(member, txPubKey) {
			isBoard = true
		}
	}
	return isBoard
}

func (lr *LoanResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	fmt.Println("Validating LoanResponse with blockchain!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	// TODO(@0xbunyip): check that only one response for this loan request in this block
	// Check if only board members created this tx
	if !txCreatedByDCBBoardMember(txr, bcr) {
		return false, errors.New("Tx must be created by DCB Governor")
	}

	// Check if the loan request is accepted on beacon shard
	reqHash, err := bcr.GetLoanReq(lr.LoanID)
	if err != nil {
		return false, err
	}
	if len(reqHash) == 0 {
		// Request and response needn't be on the same shard
		return false, errors.New("Error finding corresponding loan request")
	}

	// Check if current board member responded
	senders, _, err := bcr.GetLoanResps(lr.LoanID)
	if err != nil {
		return false, errors.New("Error finding corresponding loan response")
	}
	for _, sender := range senders {
		if bytes.Equal(txr.GetSigPubKey(), sender) {
			return false, errors.New("Current board member already responded to loan request")
		}
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

type LoanResponseAction struct {
	LoanID   []byte
	Sender   []byte
	Response ValidLoanResponse
}

func (lr *LoanResponse) BuildReqActions(txr Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	lrActionValue := getLoanResponseActionValue(lr.LoanID, txr.GetSigPubKey(), lr.Response)
	lrAction := []string{strconv.Itoa(LoanResponseMeta), lrActionValue}
	return [][]string{lrAction}, nil
}

func getLoanResponseActionValue(loanID, sender []byte, response ValidLoanResponse) string {
	action := &LoanResponseAction{
		LoanID:   loanID,
		Sender:   sender,
		Response: response,
	}
	value, _ := json.Marshal(action)
	return string(value)
}

func ParseLoanResponseActionValue(value string) ([]byte, []byte, ValidLoanResponse, error) {
	action := &LoanResponseAction{}
	err := json.Unmarshal([]byte(value), action)
	if err != nil {
		return nil, nil, 0, err
	}
	return action.LoanID, action.Sender, action.Response, nil
}

func (lr *LoanResponse) CalculateSize() uint64 {
	return calculateSize(lr)
}
