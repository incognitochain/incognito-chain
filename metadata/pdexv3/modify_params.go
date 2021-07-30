package pdexv3

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type Pdexv3Params struct {
	DefaultFeeRateBPS               uint            `json:"DefaultFeeRateBPS"`
	FeeRateBPS                      map[string]uint `json:"FeeRateBPS"`
	PRVDiscountPercent              uint            `json:"PRVDiscountPercent"`
	LimitProtocolFeePercent         uint            `json:"LimitProtocolFeePercent"`
	LimitStakingPoolRewardPercent   uint            `json:"LimitStakingPoolRewardPercent"`
	TradingProtocolFeePercent       uint            `json:"TradingProtocolFeePercent"`
	TradingStakingPoolRewardPercent uint            `json:"TradingStakingPoolRewardPercent"`
	DefaultStakingPoolsShare        uint            `json:"DefaultStakingPoolsShare"`
	StakingPoolsShare               map[string]uint `json:"StakingPoolsShare"`
}

type ParamsModifyingRequest struct {
	metadataCommon.MetadataBaseWithSignature
	Pdexv3Params `json:"Pdexv3Params"`
}

type ParamsModifyingContent struct {
	Content Pdexv3Params `json:"Content"`
	TxReqID common.Hash  `json:"TxReqID"`
	ShardID byte         `json:"ShardID"`
}

type ParamsModifyingRequestStatus struct {
	Status       int `json:"Status"`
	Pdexv3Params `json:"Pdexv3Params"`
}

func NewPdexv3ParamsModifyingRequestStatus(
	status int,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	limitProtocolFeePercent uint,
	limitStakingPoolRewardPercent uint,
	tradingProtocolFeePercent uint,
	tradingStakingPoolRewardPercent uint,
	stakingPoolsShare map[string]uint,
) *ParamsModifyingRequestStatus {
	return &ParamsModifyingRequestStatus{
		Pdexv3Params: Pdexv3Params{
			FeeRateBPS:                      feeRateBPS,
			PRVDiscountPercent:              prvDiscountPercent,
			LimitProtocolFeePercent:         limitProtocolFeePercent,
			LimitStakingPoolRewardPercent:   limitStakingPoolRewardPercent,
			TradingProtocolFeePercent:       tradingProtocolFeePercent,
			TradingStakingPoolRewardPercent: tradingStakingPoolRewardPercent,
			StakingPoolsShare:               stakingPoolsShare,
		},
		Status: status,
	}
}

func NewPdexv3ParamsModifyingRequest(
	metaType int,
	params Pdexv3Params,
) (*ParamsModifyingRequest, error) {
	metadataBase := metadataCommon.NewMetadataBaseWithSignature(metaType)
	paramsModifying := &ParamsModifyingRequest{}
	paramsModifying.MetadataBaseWithSignature = *metadataBase
	paramsModifying.Pdexv3Params = params

	return paramsModifying, nil
}

func (paramsModifying ParamsModifyingRequest) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (paramsModifying ParamsModifyingRequest) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	// validate IncAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(config.Param().PDexParams.AdminAddress)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3ModifyParamsValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}
	incAddr := keyWallet.KeySet.PaymentAddress
	if len(incAddr.Pk) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3ModifyParamsValidateSanityDataError, errors.New("Requester incognito address is invalid"))
	}

	if ok, err := paramsModifying.MetadataBaseWithSignature.VerifyMetadataSignature(incAddr.Pk, tx); err != nil || !ok {
		return false, false, errors.New("Sender is unauthorized")
	}

	// check tx type and version
	if tx.GetType() != common.TxNormalType {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.Pdexv3ModifyParamsValidateSanityDataError, errors.New("Tx pDex v3 modifying params must be TxNormalType"))
	}

	if tx.GetVersion() != 2 {
		return false, false, metadataCommon.NewMetadataTxError(0, errors.New("Tx pDex v3 modifying params must be version 2"))
	}

	return true, true, nil
}

func (paramsModifying ParamsModifyingRequest) ValidateMetadataByItself() bool {
	return paramsModifying.Type == metadataCommon.Pdexv3ModifyParamsMeta
}

func (paramsModifying ParamsModifyingRequest) Hash() *common.Hash {
	record := paramsModifying.MetadataBaseWithSignature.Hash().String()
	if paramsModifying.Sig != nil && len(paramsModifying.Sig) != 0 {
		record += string(paramsModifying.Sig)
	}
	contentBytes, _ := json.Marshal(paramsModifying.Pdexv3Params)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (paramsModifying ParamsModifyingRequest) HashWithoutSig() *common.Hash {
	record := paramsModifying.MetadataBaseWithSignature.Hash().String()
	contentBytes, _ := json.Marshal(paramsModifying.Pdexv3Params)
	hashParams := common.HashH(contentBytes)
	record += hashParams.String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (paramsModifying *ParamsModifyingRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(paramsModifying)
}
