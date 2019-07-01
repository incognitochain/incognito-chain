package metadata

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	rCommon "github.com/incognitochain/incognito-chain/ethrelaying/common"
	"github.com/pkg/errors"
)

type IssuingETHRequest struct {
	BlockHash rCommon.Hash
	TxIndex   uint
	// Proof     *light.NodeList
	ProofStrs []string
	MetadataBase
}

func NewIssuingETHRequest(
	blockHash rCommon.Hash,
	txIndex uint,
	proofStrs []string,
	metaType int,
) (*IssuingETHRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	issuingETHReq := &IssuingETHRequest{
		BlockHash: blockHash,
		TxIndex:   txIndex,
		ProofStrs: proofStrs,
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

	// nodeList := new(light.NodeList)
	// for _, item := range proofsRaw {
	// 	proofStr := item.(string)
	// 	proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	nodeList.Put([]byte{}, proofBytes)
	// }
	// proof := nodeList.NodeSet()
	// fmt.Println("proof str: ", proof.KeyCount())

	req, _ := NewIssuingETHRequest(
		blockHash,
		txIdx,
		proofStrs,
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
