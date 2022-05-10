package bridge

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/metadata/rpccaller"
	"github.com/pkg/errors"
)

type IssuingEVMRequest struct {
	BlockHash  rCommon.Hash
	TxIndex    uint
	ProofStrs  []string
	IncTokenID common.Hash
	NetworkID  uint `json:"NetworkID,omitempty"`
	metadataCommon.MetadataBase
}

type IssuingEVMReqAction struct {
	Meta       IssuingEVMRequest `json:"meta"`
	TxReqID    common.Hash       `json:"txReqId"`
	EVMReceipt *types.Receipt    `json:"ethReceipt"` // don't update the jsontag to make it compatible with the old shielding eth tx
}

type IssuingEVMAcceptedInst struct {
	ShardID         byte        `json:"shardId"`
	IssuingAmount   uint64      `json:"issuingAmount"`
	ReceiverAddrStr string      `json:"receiverAddrStr"`
	IncTokenID      common.Hash `json:"incTokenId"`
	TxReqID         common.Hash `json:"txReqId"`
	UniqTx          []byte      `json:"uniqETHTx"` // don't update the jsontag to make it compatible with the old shielding eth tx
	ExternalTokenID []byte      `json:"externalTokenId"`
}

type GetEVMHeaderByHashRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetEVMHeaderByNumberRes struct {
	rpccaller.RPCBaseRes
	Result *types.Header `json:"result"`
}

type GetEVMBlockNumRes struct {
	rpccaller.RPCBaseRes
	Result string `json:"result"`
}

const (
	LegacyTxType = iota
	AccessListTxType
	DynamicFeeTxType
)

func ParseEVMIssuingInstContent(instContentStr string) (*IssuingEVMReqAction, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instContentStr)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestDecodeInstructionError, err)
	}
	var issuingEVMReqAction IssuingEVMReqAction
	err = json.Unmarshal(contentBytes, &issuingEVMReqAction)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestUnmarshalJsonError, err)
	}
	return &issuingEVMReqAction, nil
}

func ParseEVMIssuingInstAcceptedContent(instAcceptedContentStr string) (*IssuingEVMAcceptedInst, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instAcceptedContentStr)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestDecodeInstructionError, err)
	}
	var issuingETHAcceptedInst IssuingEVMAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestUnmarshalJsonError, err)
	}
	return &issuingETHAcceptedInst, nil
}

func NewIssuingEVMRequest(
	blockHash rCommon.Hash,
	txIndex uint,
	proofStrs []string,
	incTokenID common.Hash,
	networkID uint,
	metaType int,
) (*IssuingEVMRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	issuingEVMReq := &IssuingEVMRequest{
		BlockHash:  blockHash,
		TxIndex:    txIndex,
		ProofStrs:  proofStrs,
		IncTokenID: incTokenID,
		NetworkID:  networkID,
	}
	issuingEVMReq.MetadataBase = metadataBase
	return issuingEVMReq, nil
}

func NewIssuingEVMRequestFromMap(
	data map[string]interface{},
	networkID uint,
	metatype int,
) (*IssuingEVMRequest, error) {
	blockHash := rCommon.HexToHash(data["BlockHash"].(string))
	txIdx := uint(data["TxIndex"].(float64))
	proofsRaw := data["ProofStrs"].([]interface{})
	proofStrs := []string{}
	for _, item := range proofsRaw {
		proofStrs = append(proofStrs, item.(string))
	}

	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestNewIssuingEVMRequestFromMapError, errors.Errorf("TokenID incorrect"))
	}

	req, _ := NewIssuingEVMRequest(
		blockHash,
		txIdx,
		proofStrs,
		*incTokenID,
		networkID,
		metatype,
	)
	return req, nil
}

func NewIssuingEVMRequestWithShieldRequest(data ShieldRequestData, incTokenID common.Hash) (*IssuingEVMRequest, error) {
	blockHash := rCommon.Hash{}
	err := blockHash.UnmarshalText([]byte(data.BlockHash))
	if err != nil {
		return nil, err
	}

	evmShieldRequest, _ := NewIssuingEVMRequest(
		blockHash, data.TxIndex, data.Proof, incTokenID, data.NetworkID,
		metadataCommon.IssuingUnifiedTokenRequestMeta,
	) // error always null
	return evmShieldRequest, nil
}

func (iReq IssuingEVMRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq IssuingEVMRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if len(iReq.ProofStrs) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateSanityDataError, errors.New("Wrong request info's proof"))
	}

	if (iReq.Type == metadataCommon.IssuingPRVBEP20RequestMeta || iReq.Type == metadataCommon.IssuingPRVERC20RequestMeta) && iReq.IncTokenID.String() != common.PRVIDStr {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateSanityDataError, errors.New("Invalid token id"))
	}

	return true, true, nil
}

func (iReq IssuingEVMRequest) ValidateMetadataByItself() bool {
	if iReq.Type != metadataCommon.IssuingETHRequestMeta && iReq.Type != metadataCommon.IssuingBSCRequestMeta &&
		iReq.Type != metadataCommon.IssuingPRVERC20RequestMeta && iReq.Type != metadataCommon.IssuingPRVBEP20RequestMeta &&
		iReq.Type != metadataCommon.IssuingPLGRequestMeta && !(iReq.Type == metadataCommon.IssuingUnifiedTokenRequestMeta &&
		iReq.NetworkID != common.DefaultNetworkID) && iReq.Type != metadataCommon.IssuingFantomRequestMeta {
		return false
	}

	evmReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		metadataCommon.Logger.Log.Error(metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateTxWithBlockChainError, err))
		return false
	}
	if evmReceipt == nil {
		metadataCommon.Logger.Log.Error(errors.Errorf("The evm proof's receipt could not be null."))
		return false
	}
	return true
}

func (iReq IssuingEVMRequest) Hash() *common.Hash {
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

// thachtb todo: update for terra here
func (iReq *IssuingEVMRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	evmReceipt, err := iReq.verifyProofAndParseReceipt()
	if err != nil {
		return [][]string{}, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestBuildReqActionsError, err)
	}
	if evmReceipt == nil {
		return [][]string{}, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestBuildReqActionsError, errors.Errorf("The evm proof's receipt could not be null."))
	}
	txReqID := *(tx.Hash())
	actionContent := map[string]interface{}{
		"meta":       *iReq,
		"txReqId":    txReqID,
		"ethReceipt": *evmReceipt,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestBuildReqActionsError, err)
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(iReq.Type), actionContentBase64Str}

	return [][]string{action}, nil
}

func (iReq *IssuingEVMRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iReq)
}

func (iReq *IssuingEVMRequest) verifyProofAndParseReceipt() (*types.Receipt, error) {
	// get hosts, minEVMConfirmationBlocks, networkPrefix depend iReq.Type
	hosts, networkPrefix, minEVMConfirmationBlocks, checkEVMHardFork, err := GetEVMInfoByMetadataType(iReq.Type, iReq.NetworkID)
	if err != nil {
		metadataCommon.Logger.Log.Errorf("Can not get EVM info - Error: %+v", err)
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}

	return VerifyProofAndParseEVMReceipt(iReq.BlockHash, iReq.TxIndex, iReq.ProofStrs, hosts, minEVMConfirmationBlocks, networkPrefix, checkEVMHardFork)
}
