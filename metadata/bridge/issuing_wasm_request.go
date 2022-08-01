package bridge

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/pkg/errors"
)

type WasmShieldRequestData struct {
	TxHash uint `json:"TxHash"`
}

type IssuingWasmRequest struct {
	TxHash     string
	IncTokenID common.Hash
	NetworkID  uint `json:"NetworkID,omitempty"`
	metadataCommon.MetadataBase
}

type IssuingWasmReqAction struct {
	Meta          IssuingWasmRequest `json:"meta"`
	TxReqID       common.Hash        `json:"txReqId"`
	TokenId       string             `json:"tokenId"`
	IncognitoAddr string             `json:"incognitoAddr"`
	Amount        uint64             `json:"amount"`
	ContractId    string             `json:"contractId"`
}

type IssuingWasmAcceptedInst struct {
	ShardID         byte        `json:"shardId"`
	IssuingAmount   uint64      `json:"issuingAmount"`
	ReceiverAddrStr string      `json:"receiverAddrStr"`
	IncTokenID      common.Hash `json:"incTokenId"`
	TxReqID         common.Hash `json:"txReqId"`
	UniqTx          []byte      `json:"uniqWasmTx"` // don't update the jsontag to make it compatible with the old shielding Wasm tx
	ExternalTokenID []byte      `json:"externalTokenId"`
}

type GetWasmHeaderByHashRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetWasmHeaderByNumberRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetWasmBlockNumRes struct {
	rpccaller.RPCBaseRes
	Result string `json:"result"`
}

func ParseWasmIssuingInstContent(instContentStr string) (*IssuingWasmReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestDecodeInstructionError, err)
	}
	var issuingWasmReqAction IssuingWasmReqAction
	err = json.Unmarshal(contentBytes, &issuingWasmReqAction)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestUnmarshalJsonError, err)
	}
	return &issuingWasmReqAction, nil
}

func ParseWasmIssuingInstAcceptedContent(instAcceptedContentStr string) (*IssuingWasmAcceptedInst, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instAcceptedContentStr)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestDecodeInstructionError, err)
	}
	var issuingWasmAcceptedInst IssuingWasmAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingWasmAcceptedInst)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestUnmarshalJsonError, err)
	}
	return &issuingWasmAcceptedInst, nil
}

func NewIssuingWasmRequest(
	txHash string,
	incTokenID common.Hash,
	networkID uint,
	metaType int,
) (*IssuingWasmRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	issuingWasmReq := &IssuingWasmRequest{
		TxHash:     txHash,
		IncTokenID: incTokenID,
		NetworkID:  networkID,
	}
	issuingWasmReq.MetadataBase = metadataBase
	return issuingWasmReq, nil
}

func NewIssuingWasmRequestFromMap(
	data map[string]interface{},
	networkID uint,
	metatype int,
) (*IssuingWasmRequest, error) {
	txHash := data["TxHash"].(string)
	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestNewIssuingWasmRequestFromMapError, errors.Errorf("TokenID incorrect"))
	}

	req, _ := NewIssuingWasmRequest(
		txHash,
		*incTokenID,
		networkID,
		metatype,
	)
	return req, nil
}

func (iReq IssuingWasmRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq IssuingWasmRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if len(iReq.TxHash) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestValidateSanityDataError, errors.New("Wrong request info's tx hash"))
	}

	return true, true, nil
}

func (iReq IssuingWasmRequest) ValidateMetadataByItself() bool {
	if iReq.Type != metadataCommon.IssuingNearRequestMeta {
		return false
	}
	_, _, _, _, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		metadataCommon.Logger.Log.Error(metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestValidateTxWithBlockChainError, err))
		return false
	}

	return true
}

func (iReq IssuingWasmRequest) Hash() *common.Hash {
	record := iReq.TxHash
	record += iReq.MetadataBase.Hash().String()
	record += iReq.IncTokenID.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *IssuingWasmRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	incAddr, token, amount, contractId, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return [][]string{}, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestBuildReqActionsError, err)
	}
	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":          *iReq,
		"txReqId":       txReqID,
		"tokenId":       token,
		"incognitoAddr": incAddr,
		"amount":        amount,
		"contractId":    contractId,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(iReq.Type), actionContentBase64Str}

	return [][]string{action}, nil
}

func (iReq *IssuingWasmRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iReq)
}

func (iReq *IssuingWasmRequest) verifyProofAndParseReceipt() (string, string, uint64, string, error) {
	// get hosts, minWasmConfirmationBlocks, networkPrefix depend iReq.Type
	hosts, _, minWasmConfirmationBlocks, contractId, err := GetWasmInfoByMetadataType(iReq.Type)
	if err != nil {
		metadataCommon.Logger.Log.Errorf("Can not get Wasm info - Error: %+v", err)
		return "", "", 0, "", metadataCommon.NewMetadataTxError(metadataCommon.IssuingWasmRequestVerifyProofAndParseReceipt, err)
	}

	return VerifyWasmShieldTxId(iReq.TxHash, hosts, minWasmConfirmationBlocks, contractId)

}
