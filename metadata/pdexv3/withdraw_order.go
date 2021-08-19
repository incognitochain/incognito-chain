package pdexv3

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// WithdrawOrderRequest
type WithdrawOrderRequest struct {
	PoolPairID string              `json:"PoolPairID"`
	OrderID    string              `json:"OrderID"`
	TokenID    common.Hash         `json:"TokenID"`
	Amount     uint64              `json:"Amount"`
	Receiver   privacy.OTAReceiver `json:"Receiver"`
	NftID      common.Hash         `json:"NftID"`
	metadataCommon.MetadataBase
}

func NewWithdrawOrderRequest(
	pairID, orderID string,
	tokenID common.Hash,
	amount uint64,
	recv privacy.OTAReceiver,
	nftID common.Hash,
	metaType int,
) (*WithdrawOrderRequest, error) {
	r := &WithdrawOrderRequest{
		PoolPairID: pairID,
		OrderID:    orderID,
		TokenID:    tokenID,
		Amount:     amount,
		Receiver:   recv,
		NftID:      nftID,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
	return r, nil
}

func (req WithdrawOrderRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (req WithdrawOrderRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	// OTAReceiver check
	if !req.Receiver.IsValid() {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Invalid OTAReceiver %v", req.Receiver))
	}
	if tx.GetSenderAddrLastByte() != req.Receiver.GetShardID() {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("Invalid shard %d for Receiver - must equal sender shard",
				req.Receiver.GetShardID()))
	}

	// Burned coin check
	isBurn, burnedPRVCoin, burnedCoin, burnedTokenID, err := tx.GetTxFullBurnData()
	if err != nil || !isBurn {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("Burned coins not found in trade request - %v", err))
	}
	if *burnedTokenID != req.NftID {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("Burned nftID mismatch - %v vs %v on metadata", *burnedTokenID, req.NftID))
	}

	// Type vs burned token id + amount check
	switch tx.GetType() {
	case common.TxCustomTokenPrivacyType:
		if burnedCoin == nil || burnedPRVCoin != nil {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Burned token invalid - must be %v only", req.NftID))
		}

		// accept any positive governance-NFT amount
		if burnedCoin.GetValue() <= 0 {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Invalid burned NFT amount %d",
					burnedCoin.GetValue()))
		}
	default:
		return false, false, fmt.Errorf("Invalid transaction type %v for withdrawOrder request", tx.GetType())
	}
	return true, true, nil
}

func (req WithdrawOrderRequest) ValidateMetadataByItself() bool {
	return req.Type == metadataCommon.Pdexv3WithdrawOrderRequestMeta
}

func (req WithdrawOrderRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(req)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (req *WithdrawOrderRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(req)
}

func (req *WithdrawOrderRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	tokenID := req.TokenID
	if tokenID != common.PRVCoinID {
		tokenID = common.ConfidentialAssetID
	}
	result = append(result, metadataCommon.OTADeclaration{
		PublicKey: req.Receiver.PublicKey.ToBytes(), TokenID: tokenID,
	})
	return result
}
