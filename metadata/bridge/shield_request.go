package bridge

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type AcceptedShieldRequest struct {
	Receiver   string                      `json:"Receiver"`
	IncTokenID common.Hash                 `json:"IncTokenID"`
	TxReqID    common.Hash                 `json:"TxReqID"`
	IsReward   bool                        `json:"IsReward"`
	ShardID    byte                        `json:"ShardID"`
	Data       []AcceptedShieldRequestData `json:"Data"`
}

type AcceptedShieldRequestData struct {
	IssuingAmount   uint64 `json:"IssuingAmount"`
	UniqTx          []byte `json:"UniqTx,omitempty"`
	ExternalTokenID []byte `json:"ExternalTokenID,omitempty"`
	NetworkID       uint   `json:"NetworkID"`
}

type ShieldRequestData struct {
	BlockHash string   `json:"BlockHash"`
	TxIndex   uint     `json:"TxIndex"`
	Proof     []string `json:"Proof"`
	NetworkID uint     `json:"NetworkID"`
}

type ShieldRequest struct {
	Data           []ShieldRequestData    `json:"Data"`
	IncTokenID     common.Hash            `json:"IncTokenID"`
	PaymentAddress privacy.PaymentAddress `json:"PaymentAddress,omitempty"`
	metadataCommon.MetadataBase
}

func NewShieldRequest() *ShieldRequest {
	return &ShieldRequest{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.ShieldUnifiedTokenRequestMeta,
		},
	}
}

func NewShieldRequestWithValue(
	data []ShieldRequestData, incTokenID common.Hash, paymentAddress privacy.PaymentAddress,
) *ShieldRequest {
	return &ShieldRequest{
		Data:           data,
		IncTokenID:     incTokenID,
		PaymentAddress: paymentAddress,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.ShieldUnifiedTokenRequestMeta,
		},
	}
}

func (request *ShieldRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (request *ShieldRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if len(request.Data) == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggShieldValidateSanityDataError, fmt.Errorf("Data can not be null"))
	}
	return true, true, nil
}

func (request *ShieldRequest) ValidateMetadataByItself() bool {
	if request.Type != metadataCommon.ShieldUnifiedTokenRequestMeta {
		return false
	}
	for _, data := range request.Data {
		switch data.NetworkID {
		case common.ETHNetworkID, common.BSCNetworkID, common.PLGNetworkID:
			evmShieldRequest, err := NewIssuingEVMRequestWithShieldRequest(data, request.IncTokenID)
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
		switch data.NetworkID {
		case common.ETHNetworkID, common.BSCNetworkID, common.PLGNetworkID:
			evmShieldRequest, err := NewIssuingEVMRequestWithShieldRequest(data, request.IncTokenID)
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
			content, err := json.Marshal(evmReceipt)
			if err != nil {
				return [][]string{}, errors.Errorf("The evm proof's receipt could not be null.")
			}
			extraData = append(extraData, content)
		case common.DefaultNetworkID:
			return [][]string{}, errors.New("NetworkID cannot be default")
		default:
			return [][]string{}, errors.New("Invalid networkID")
		}
	}
	content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), extraData).StringSlice(metadataCommon.ShieldUnifiedTokenRequestMeta)
	return [][]string{content}, err
}

func (request *ShieldRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(request)
}
