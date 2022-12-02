package bridge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type ShieldRequest struct {
	Data           []ShieldRequestData `json:"Data"`
	UnifiedTokenID common.Hash         `json:"UnifiedTokenID"`
	metadataCommon.MetadataBase
}

type ShieldRequestData struct {
	Proof      []byte      `json:"Proof"`
	NetworkID  uint8       `json:"NetworkID"`
	IncTokenID common.Hash `json:"IncTokenID"`
}

type AcceptedInstShieldRequest struct {
	Receiver       privacy.PaymentAddress      `json:"Receiver"`
	UnifiedTokenID common.Hash                 `json:"UnifiedTokenID"`
	TxReqID        common.Hash                 `json:"TxReqID"`
	Data           []AcceptedShieldRequestData `json:"Data"`
}

type AcceptedShieldRequestData struct {
	ShieldAmount    uint64      `json:"ShieldAmount"`
	Reward          uint64      `json:"Reward"`
	UniqTx          []byte      `json:"UniqTx"`
	ExternalTokenID []byte      `json:"ExternalTokenID"`
	NetworkID       uint8       `json:"NetworkID"`
	IncTokenID      common.Hash `json:"IncTokenID"`
}

func NewShieldRequest() *ShieldRequest {
	return &ShieldRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.IssuingUnifiedTokenRequestMeta,
		},
	}
}

func NewShieldRequestWithValue(
	data []ShieldRequestData, unifiedTokenID common.Hash,
) *ShieldRequest {
	return &ShieldRequest{
		Data:           data,
		UnifiedTokenID: unifiedTokenID,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.IssuingUnifiedTokenRequestMeta,
		},
	}
}

func ParseShieldReqInstAcceptedContent(instAcceptedContentStr string) (*AcceptedInstShieldRequest, error) {
	contentBytes, err := base64.StdEncoding.DecodeString(instAcceptedContentStr)
	if err != nil {
		metadataCommon.Logger.Log.Errorf("BridgeAgg Can not decode accepted instruction for shielding %v\n", err)
		return nil, fmt.Errorf("BridgeAgg Can not decode accepted instruction for shielding %v\n", err)
	}
	var shieldReqAcceptedInst AcceptedInstShieldRequest
	err = json.Unmarshal(contentBytes, &shieldReqAcceptedInst)
	if err != nil {
		metadataCommon.Logger.Log.Errorf("BridgeAgg Can not unmarshal accepted instruction for shielding %v\n", err)
		return nil, fmt.Errorf("BridgeAgg Can not unmarshal accepted instruction for shielding %v\n", err)
	}
	return &shieldReqAcceptedInst, nil
}

func (request *ShieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (request *ShieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if request.UnifiedTokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, errors.New("UnifiedTokenID can not be empty"))
	}
	if len(request.Data) <= 0 || len(request.Data) > int(config.Param().BridgeAggParam.MaxLenOfPath) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, fmt.Errorf("Length of data %d need to be in [1..%d]", len(request.Data), config.Param().BridgeAggParam.MaxLenOfPath))
	}
	for _, data := range request.Data {
		if (data.NetworkID == common.AVAXNetworkID && shardViewRetriever.GetTriggeredFeature()["auroraavaxbridge"] == 0) ||
			(data.NetworkID == common.AURORANetworkID && shardViewRetriever.GetTriggeredFeature()["aurorahotfix"] == 0) {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.UnexpectedError, errors.New("Feature not enabled yet"))
		}

		if data.IncTokenID.IsZeroValue() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, errors.New("IncTokenID can not be empty"))
		}
		if data.IncTokenID.String() == request.UnifiedTokenID.String() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, fmt.Errorf("IncTokenID duplicate with unifiedTokenID %s", data.IncTokenID.String()))
		}
	}
	return true, true, nil
}

func (request *ShieldRequest) ValidateMetadataByItself() bool {
	if request.Type != metadataCommon.IssuingUnifiedTokenRequestMeta {
		return false
	}
	for _, data := range request.Data {
		switch data.NetworkID {
		case common.ETHNetworkID, common.BSCNetworkID, common.PLGNetworkID, common.FTMNetworkID, common.AVAXNetworkID:
			proofData := EVMProof{}
			err := json.Unmarshal(data.Proof, &proofData)
			if err != nil {
				metadataCommon.Logger.Log.Errorf("Can not unmarshal evm proof: %v\n", err)
				return false
			}
			evmShieldRequest, err := NewIssuingEVMRequestFromProofData(proofData, uint(data.NetworkID), request.UnifiedTokenID)
			if err != nil {
				return false
			}
			return evmShieldRequest.ValidateMetadataByItself()
		case common.AURORANetworkID:
			auroraTxId := common.BytesToHash(data.Proof)
			auroraShieldRequest, err := NewIssuingEVMAuroraRequest(
				auroraTxId,
				data.IncTokenID,
				common.AURORANetworkID,
				metadataCommon.IssuingUnifiedTokenRequestMeta,
			)
			if err != nil {
				return false
			}
			return auroraShieldRequest.ValidateMetadataByItself()
		case common.DefaultNetworkID:
			return false
		default:
			return false
		}
	}
	return true
}

func (request *ShieldRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&request)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (request *ShieldRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	extraData := [][]byte{}
	for _, data := range request.Data {
		networkType, err := GetNetworkTypeByNetworkID(data.NetworkID)
		if err != nil {
			return nil, err
		}
		switch networkType {
		case common.EVMNetworkType:
			proofData := EVMProof{}
			err := json.Unmarshal(data.Proof, &proofData)
			if err != nil {
				return [][]string{}, err
			}
			evmShieldRequest, err := NewIssuingEVMRequestFromProofData(proofData, uint(data.NetworkID), request.UnifiedTokenID)
			if err != nil {
				return [][]string{}, err
			}
			evmReceipt, err := evmShieldRequest.verifyProofAndParseReceipt()
			if err != nil {
				return [][]string{}, err
			}
			if evmReceipt == nil {
				return [][]string{}, errors.Errorf("The evm proof's receipt could not be null.")
			}
			content, err := MarshalActionDataForShieldEVMReq(evmReceipt)
			if err != nil {
				return [][]string{}, err
			}
			extraData = append(extraData, content)
		case common.AURORANetworkID:
			auroraTxId := common.BytesToHash(data.Proof)
			auroraShieldRequest, err := NewIssuingEVMAuroraRequest(
				auroraTxId,
				data.IncTokenID,
				common.AURORANetworkID,
				metadataCommon.IssuingUnifiedTokenRequestMeta,
			)
			if err != nil {
				return [][]string{}, err
			}
			evmReceipt, err := auroraShieldRequest.verifyProofAndParseReceipt()
			if err != nil {
				return [][]string{}, err
			}
			if evmReceipt == nil {
				return [][]string{}, errors.Errorf("The evm proof's receipt could not be null.")
			}
			content, err := MarshalActionDataForShieldEVMReq(evmReceipt)
			if err != nil {
				return [][]string{}, err
			}
			extraData = append(extraData, content)
		default:
			return [][]string{}, errors.New("Invalid networkID")
		}
	}
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), extraData).StringSlice(metadataCommon.IssuingUnifiedTokenRequestMeta)
	return [][]string{content}, err
}

func (request *ShieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}

func UnmarshalActionDataForShieldEVMReq(data []byte) (*types.Receipt, error) {
	txReceipt := types.Receipt{}
	err := json.Unmarshal(data, &txReceipt)
	return &txReceipt, err
}

func MarshalActionDataForShieldEVMReq(txReceipt *types.Receipt) ([]byte, error) {
	return json.Marshal(txReceipt)
}
