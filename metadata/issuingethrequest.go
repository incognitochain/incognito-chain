package metadata

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type IssuingETHRequest struct {
	BlockHash  rCommon.Hash
	TxIndex    uint
	ProofStrs  []string
	IncTokenID common.Hash
	MetadataBase
}

func NewIssuingETHRequest(
	blockHash rCommon.Hash,
	txIndex uint,
	proofStrs []string,
	incTokenID common.Hash,
	metaType int,
) (*IssuingETHRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingETHReq := &IssuingETHRequest{
		BlockHash:  blockHash,
		TxIndex:    txIndex,
		ProofStrs:  proofStrs,
		IncTokenID: incTokenID,
	}
	issuingETHReq.MetadataBase = metadataBase
	return issuingETHReq, nil
}

func NewIssuingETHRequestFromMap(
	data map[string]interface{},
) (*IssuingETHRequest, error) {
	blockHash := rCommon.HexToHash(data["BlockHash"].(string))
	txIdx := uint(data["TxIndex"].(float64))
	proofsRaw := data["ProofStrs"].([]interface{})
	proofStrs := []string{}
	for _, item := range proofsRaw {
		proofStrs = append(proofStrs, item.(string))
	}

	incTokenID, err := common.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, errors.Errorf("TokenID incorrect")
	}

	req, _ := NewIssuingETHRequest(
		blockHash,
		txIdx,
		proofStrs,
		*incTokenID,
		IssuingETHRequestMeta,
	)
	return req, nil
}

func (iReq *IssuingETHRequest) ValidateTxWithBlockChain(
	txr Transaction,
	bcr BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) (bool, error) {
	return true, nil
}

func (iReq *IssuingETHRequest) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(iReq.ProofStrs) == 0 {
		return false, false, errors.New("Wrong request info's proof")
	}
	return true, true, nil
}

func (iReq *IssuingETHRequest) ValidateMetadataByItself() bool {
	if iReq.Type != IssuingETHRequestMeta {
		return false
	}
	return true
}

func (iReq *IssuingETHRequest) Hash() *common.Hash {
	record := iReq.BlockHash.String()
	record += string(iReq.TxIndex)
	proofStrs := iReq.ProofStrs
	for _, proofStr := range proofStrs {
		record += proofStr
	}
	record += iReq.MetadataBase.Hash().String()
	record += iReq.IncTokenID.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *IssuingETHRequest) BuildReqActions(tx Transaction, bcr BlockchainRetriever, shardID byte) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":    *iReq,
		"txReqId": *(tx.Hash()),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(IssuingETHRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (iReq *IssuingETHRequest) CalculateSize() uint64 {
	return calculateSize(iReq)
}
