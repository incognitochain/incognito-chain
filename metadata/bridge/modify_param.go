package bridge

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type ModifyBridgeAggParamReq struct {
	metadataCommon.MetadataBaseWithSignature
	PercentFeeWithDec uint64 `json:"PercentFeeWithDec"`
}

type ModifyBridgeAggParamContentInst struct {
	PercentFeeWithDec uint64      `json:"PercentFeeWithDec"`
	TxReqID           common.Hash `json:"TxReqID"`
}

func NewModifyBridgeAggParamReq() *ModifyBridgeAggParamReq {
	return &ModifyBridgeAggParamReq{}
}

func NewModifyBridgeAggParamReqWithValue(percentFeeWithDec uint64) *ModifyBridgeAggParamReq {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metadataCommon.BridgeAggModifyParamMeta)
	request := &ModifyBridgeAggParamReq{}
	request.MetadataBaseWithSignature = *metadataBase
	request.PercentFeeWithDec = percentFeeWithDec
	return request
}

func (request *ModifyBridgeAggParamReq) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (request *ModifyBridgeAggParamReq) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// validate requester
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().BridgeAggParam.AdminAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyParamValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyParamValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := request.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyParamValidateSanityDataError, errors.New("Sender is unauthorized"))
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyParamValidateSanityDataError, errors.New("Tx bridge agg modify param must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyParamValidateSanityDataError, errors.New("Tx bridge agg modify param must be version 2"))
	}

	// mustn't exceed 100%
	if request.PercentFeeWithDec >= config.Param().BridgeAggParam.PercentFeeDecimal {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggModifyParamValidateSanityDataError, errors.New("Tx bridge agg modify param invalid percent fee with dec"))
	}

	return true, true, nil
}

func (request *ModifyBridgeAggParamReq) ValidateMetadataByItself() bool {
	return request.Type == metadataCommon.BridgeAggModifyParamMeta
}

func (request *ModifyBridgeAggParamReq) Hash() *common.Hash {
	record := request.MetadataBaseWithSignature.Hash().String()
	if request.Sig != nil && len(request.Sig) != 0 {
		record += string(request.Sig)
	}
	contentBytes, _ := json.Marshal(request)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *ModifyBridgeAggParamReq) HashWithoutSig() *common.Hash {
	record := request.MetadataBaseWithSignature.Hash().String()
	contentBytes, _ := json.Marshal(request.PercentFeeWithDec)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (request *ModifyBridgeAggParamReq) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func (request *ModifyBridgeAggParamReq) BuildReqActions(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	shardHeight uint64,
) ([][]string, error) {
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.BridgeAggModifyParamMeta)
	return [][]string{content}, err
}
