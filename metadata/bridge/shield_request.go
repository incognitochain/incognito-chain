package bridge

import (
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
	NetworkID  uint        `json:"NetworkID"`
	IncTokenID common.Hash `json:"IncTokenID"`
}

type AcceptedInstShieldRequest struct {
	Receiver       privacy.PaymentAddress      `json:"Receiver"`
	UnifiedTokenID common.Hash                 `json:"UnifiedTokenID"`
	TxReqID        common.Hash                 `json:"TxReqID"`
	IsReward       bool                        `json:"IsReward"`
	ShardID        byte                        `json:"ShardID"`
	Data           []AcceptedShieldRequestData `json:"Data"`
}

type AcceptedShieldRequestData struct {
	IssuingAmount   uint64      `json:"IssuingAmount"`
	UniqTx          []byte      `json:"UniqTx,omitempty"`          // empty for reward inst
	ExternalTokenID []byte      `json:"ExternalTokenID,omitempty"` // empty for reward inst
	NetworkID       uint        `json:"NetworkID"`
	IncTokenID      common.Hash `json:"IncTokenID"`
}

// type AcceptedShieldRewardData struct {
// 	RewardAmount uint64      `json:"RewardAmount"`
// 	NetworkID    uint        `json:"NetworkID"`
// 	IncTokenID   common.Hash `json:"IncTokenID"`
// }

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

func (request *ShieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (request *ShieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	usedTokenIDs := make(map[common.Hash]bool)
	if request.UnifiedTokenID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, errors.New("UnifiedTokenID can not be empty"))
	}
	if len(request.Data) <= 0 || len(request.Data) > config.Param().BridgeAggParam.MaxLenOfPath {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, fmt.Errorf("Length of data %d need to be in [1..%d]", len(request.Data), config.Param().BridgeAggParam.MaxLenOfPath))
	}
	usedTokenIDs[request.UnifiedTokenID] = true
	for _, data := range request.Data {
		if data.IncTokenID.IsZeroValue() {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, errors.New("IncTokenID can not be empty"))
		}
		if usedTokenIDs[data.IncTokenID] {
			return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, fmt.Errorf("Duplicate tokenID %s", data.IncTokenID.String()))
		}
		usedTokenIDs[data.IncTokenID] = true
	}
	return true, true, nil
}

func (request *ShieldRequest) ValidateMetadataByItself() bool {
	if request.Type != metadataCommon.IssuingUnifiedTokenRequestMeta {
		return false
	}
	for _, data := range request.Data {
		switch data.NetworkID {
		case common.ETHNetworkID, common.BSCNetworkID, common.PLGNetworkID, common.FTMNetworkID:
			proofData := EVMProof{}
			err := json.Unmarshal(data.Proof, &proofData)
			if err != nil {
				return false
			}
			evmShieldRequest, err := NewIssuingEVMRequestFromProofData(proofData, data.NetworkID, request.UnifiedTokenID)
			if err != nil {
				return false
			}
			return evmShieldRequest.ValidateMetadataByItself()
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
			evmShieldRequest, err := NewIssuingEVMRequestFromProofData(proofData, data.NetworkID, request.UnifiedTokenID)
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
