package pdexv3

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PDexV3Params struct {
	DefaultFeeRateBPS        uint            `json:"DefaultFeeRateBPS"`
	FeeRateBPS               map[string]uint `json:"FeeRateBPS"`
	PRVDiscountPercent       uint            `json:"PRVDiscountPercent"`
	ProtocolFeePercent       uint            `json:"ProtocolFeePercent"`
	StakingPoolRewardPercent uint            `json:"StakingPoolRewardPercent"`
	DefaultStakingPoolsShare uint            `json:"DefaultStakingPoolsShare"`
	StakingPoolsShare        map[string]uint `json:"StakingPoolsShare"`
}

type PDexV3ParamsModifyingRequest struct {
	metadataCommon.MetadataBaseWithSignature
	PDexV3Params `json:"PDexV3Params"`
}

type PDexV3ParamsModifyingRequestAction struct {
	Meta    PDexV3ParamsModifyingRequest `json:"Meta"`
	TxReqID common.Hash                  `json:"TxReqID"`
	ShardID byte                         `json:"ShardID"`
}

type PDexV3ParamsModifyingRequestContent struct {
	Content PDexV3Params `json:"Content"`
	TxReqID common.Hash  `json:"TxReqID"`
	ShardID byte         `json:"ShardID"`
}

type PDexV3ParamsModifyingRequestStatus struct {
	Status       int `json:"Status"`
	PDexV3Params `json:"PDexV3Params"`
}

func NewPDexV3ParamsModifyingRequestStatus(
	status int,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	protocolFeePercent uint,
	stakingPoolRewardPercent uint,
	stakingPoolsShare map[string]uint,
) *PDexV3ParamsModifyingRequestStatus {
	return &PDexV3ParamsModifyingRequestStatus{
		PDexV3Params: PDexV3Params{
			FeeRateBPS:               feeRateBPS,
			PRVDiscountPercent:       prvDiscountPercent,
			ProtocolFeePercent:       protocolFeePercent,
			StakingPoolRewardPercent: stakingPoolRewardPercent,
			StakingPoolsShare:        stakingPoolsShare,
		},
		Status: status,
	}
}

func NewPDexV3ParamsModifyingRequest(
	metaType int,
	params PDexV3Params,
) (*PDexV3ParamsModifyingRequest, error) {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metaType)
	paramsModifying := &PDexV3ParamsModifyingRequest{}
	paramsModifying.MetadataBaseWithSignature = *metadataBase
	paramsModifying.PDexV3Params = params

	return paramsModifying, nil
}

func (paramsModifying PDexV3ParamsModifyingRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (paramsModifying PDexV3ParamsModifyingRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().PDexParams.AdminAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3ModifyParamsValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3ModifyParamsValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := paramsModifying.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, errors.New("Sender is unauthorized")
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDexV3ModifyParamsValidateSanityDataError, errors.New("Tx pDex v3 modifying params must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 modifying params must be version 2"))
	}

	return true, true, nil
}

func (paramsModifying PDexV3ParamsModifyingRequest) ValidateMetadataByItself() bool {
	return paramsModifying.Type == metadataCommon.PDexV3ModifyParamsMeta
}

func (paramsModifying PDexV3ParamsModifyingRequest) Hash() *common.Hash {
	record := paramsModifying.MetadataBaseWithSignature.Hash().String()
	contentBytes, _ := json.Marshal(paramsModifying.PDexV3Params)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	if paramsModifying.Sig != nil && len(paramsModifying.Sig) != 0 {
		record += string(paramsModifying.Sig)
	}

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (paramsModifying PDexV3ParamsModifyingRequest) HashWithoutSig() *common.Hash {
	record := paramsModifying.MetadataBaseWithSignature.Hash().String()
	contentBytes, _ := json.Marshal(paramsModifying.PDexV3Params)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (paramsModifying *PDexV3ParamsModifyingRequest) BuildReqActions(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	shardHeight uint64,
) ([][]string, error) {
	actionContent := PDexV3ParamsModifyingRequestAction{
		Meta:    *paramsModifying,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}

	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(metadataCommon.PDexV3ModifyParamsMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (paramsModifying *PDexV3ParamsModifyingRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(paramsModifying)
}
