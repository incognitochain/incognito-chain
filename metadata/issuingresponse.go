package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/pkg/errors"
)

type IssuingResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

type IssuingResAction struct {
	IncTokenID *common.Hash `json:"incTokenID"`
}

func NewIssuingResponse(requestedTxID common.Hash, metaType int) *IssuingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes *IssuingResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (iRes *IssuingResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes *IssuingResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes *IssuingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes *IssuingResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes *IssuingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	ac *AccumulatedValues,
) (bool, error) {
	db := bcr.GetDatabase()
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not IssuingETHRequest instruction
			continue
		}
		instMetaType := inst[0]
		issuingReqAction, err := ParseIssuingInstContent(inst[3])
		if err != nil {
			fmt.Println("WARNING: an error occured during parsing instruction content: ", err)
			continue
		}

		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(IssuingRequestMeta) ||
			!bytes.Equal(iRes.RequestedTxID[:], issuingReqAction.TxReqID[:]) {
			continue
		}

		issuingReq := issuingReqAction.Meta
		issuingTokenID := issuingReq.TokenID
		if !ac.CanProcessCIncToken(issuingTokenID) {
			fmt.Printf("WARNING: The issuing token (%s) was already used in the current block.", issuingTokenID.String())
			continue
		}

		ok, err := db.CanProcessCIncToken(issuingTokenID)
		if err != nil {
			fmt.Println("WARNING: an error occured during checking centralized inc token is valid or not: ", err)
			continue
		}
		if !ok {
			fmt.Printf("WARNING: The issuing token (%s) was already used in the previous blocks.", issuingTokenID.String())
			continue
		}

		_, pk, amount, assetID := tx.GetTransferData()
		if !bytes.Equal(issuingReq.ReceiverAddress.Pk[:], pk[:]) ||
			issuingReq.DepositedAmount != amount ||
			!bytes.Equal(issuingReq.TokenID[:], assetID[:]) {
			continue
		}
		ac.CBridgeTokens = append(ac.CBridgeTokens, &issuingTokenID)
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no IssuingRequest tx found for the IssuingResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}

func (iRes *IssuingResponse) BuildReqActions(
	tx Transaction,
	bcr BlockchainRetriever,
	shardID byte,
) ([][]string, error) {
	incTokenID := tx.GetTokenID()
	actionContent := map[string]interface{}{
		"incTokenID": incTokenID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(IssuingResponseMeta), actionContentBase64Str}
	return [][]string{action}, nil
}
