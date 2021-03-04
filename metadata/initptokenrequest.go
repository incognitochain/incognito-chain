package metadata

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strconv"
)

type InitPTokenRequest struct {
	OTAStr      string
	TxRandomStr string
	TokenID     common.Hash
	Amount      uint64
	TokenName   string
	TokenSymbol string
	MetadataBase
}

type InitPTokenReqAction struct {
	Meta    InitPTokenRequest `json:"meta"`
	TxReqID common.Hash       `json:"txReqId"`
}

type InitPTokenAcceptedInst struct {
	OTAStr        string      `json:"OTAStr"`
	TxRandomStr   string      `json:"TxRandomStr"`
	Amount        uint64      `json:"Amount"`
	TokenID       common.Hash `json:"TokenID"`
	ShardID       byte        `json:"ShardID"`
	RequestedTxID common.Hash `json:"txReqId"`
}

func ParseInitPTokenInstContent(instContentStr string) (*InitPTokenReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, NewMetadataTxError(InitPTokenRequestDecodeInstructionError, err)
	}
	var initPTokenReqAction InitPTokenReqAction
	err = json.Unmarshal(contentBytes, &initPTokenReqAction)
	if err != nil {
		return nil, NewMetadataTxError(InitPTokenRequestUnmarshalJsonError, err)
	}
	return &initPTokenReqAction, nil
}

func NewPTokenInitRequest(convertingAddress string, txRandomStr string, amount uint64, tokenID common.Hash, tokenName, tokenSymbol string, metaType int) (*InitPTokenRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	initPTokenMeta := &InitPTokenRequest{
		OTAStr:      convertingAddress,
		TxRandomStr: txRandomStr,
		TokenID:     tokenID,
		TokenName:   tokenName,
		TokenSymbol: tokenSymbol,
		Amount:      amount,
	}
	initPTokenMeta.MetadataBase = metadataBase
	return initPTokenMeta, nil
}

func (req InitPTokenRequest) ValidateMetadataByItself() bool {
	return req.Type == InitPTokenRequestMeta
}

func (req InitPTokenRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (req InitPTokenRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return true, true, nil
}

func (req InitPTokenRequest) Hash() *common.Hash {
	record := req.MetadataBase.Hash().String()
	record += req.TokenID.String()
	record += req.OTAStr
	record += req.TxRandomStr
	record += req.TokenName
	record += req.TokenSymbol
	record += strconv.FormatUint(req.Amount, 10)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *InitPTokenRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":    *iReq,
		"txReqId": txReqID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, NewMetadataTxError(InitPTokenRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(InitPTokenRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (req *InitPTokenRequest) CalculateSize() uint64 {
	return calculateSize(req)
}
