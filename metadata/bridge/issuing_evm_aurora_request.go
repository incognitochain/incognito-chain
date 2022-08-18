package bridge

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/config"
	"strconv"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/pkg/errors"
)

type IssuingEVMAuroraRequest struct {
	TxHash     common.Hash
	IncTokenID common.Hash
	NetworkID  uint `json:"NetworkID,omitempty"`
	metadataCommon.MetadataBase
}

func NewIssuingEVMAuroraRequest(
	txHash common.Hash,
	incTokendId common.Hash,
	networkID uint,
	metaType int,
) (*IssuingEVMAuroraRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	issuingEVMReq := &IssuingEVMAuroraRequest{
		TxHash:     txHash,
		IncTokenID: incTokendId,
		NetworkID:  networkID,
	}
	issuingEVMReq.MetadataBase = metadataBase
	return issuingEVMReq, nil
}

func NewIssuingEVMAuroraRequestFromMap(
	data map[string]interface{},
	networkID uint,
	metatype int,
) (*IssuingEVMAuroraRequest, error) {
	incTokenID, err := common.Hash{}.NewHashFromStr(data["IncTokenID"].(string))
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestNewIssuingEVMRequestFromMapError, errors.Errorf("TokenID incorrect"))
	}

	txHash, err := common.Hash{}.NewHashFromStr(data["TxHash"].(string))
	if err != nil {
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestNewIssuingEVMRequestFromMapError, errors.Errorf("TxHash incorrect"))
	}

	req, _ := NewIssuingEVMAuroraRequest(
		*txHash,
		*incTokenID,
		networkID,
		metatype,
	)
	return req, nil
}

func (iReq IssuingEVMAuroraRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (iReq IssuingEVMAuroraRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if iReq.TxHash.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestValidateSanityDataError, errors.New("Wrong request info's txhash"))
	}

	return true, true, nil
}

func (iReq IssuingEVMAuroraRequest) ValidateMetadataByItself() bool {
	if iReq.Type != metadataCommon.IssuingAuroraRequestMeta {
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

func (iReq IssuingEVMAuroraRequest) Hash() *common.Hash {
	record := iReq.MetadataBase.Hash().String()
	record += iReq.TxHash.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iReq *IssuingEVMAuroraRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
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

func (iReq *IssuingEVMAuroraRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iReq)
}

func (iReq *IssuingEVMAuroraRequest) verifyProofAndParseReceipt() (*types.Receipt, error) {
	// get hosts, minEVMConfirmationBlocks, networkPrefix depend on iReq.Type
	hosts, networkPrefix, minEVMConfirmationBlocks, _, err := GetEVMInfoByMetadataType(iReq.Type, iReq.NetworkID)
	if err != nil {
		metadataCommon.Logger.Log.Errorf("Can not get EVM info - Error: %+v", err)
		return nil, metadataCommon.NewMetadataTxError(metadataCommon.IssuingEvmRequestVerifyProofAndParseReceipt, err)
	}
	nearParam := config.Param().NEARParam

	return VerifyProofAndParseAuroraReceipt(iReq.TxHash, hosts, nearParam.Host, minEVMConfirmationBlocks, networkPrefix)
}
