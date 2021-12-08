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
	PoolPairID string                              `json:"PoolPairID"`
	OrderID    string                              `json:"OrderID"`
	Amount     uint64                              `json:"Amount"`
	Receiver   map[common.Hash]privacy.OTAReceiver `json:"Receiver"`
	NftID      common.Hash                         `json:"NftID"`
	NextOTA    *AccessOTA                          `json:"NextOTA,omitempty"`
	metadataCommon.MetadataBase
}

func NewWithdrawOrderRequest(
	pairID, orderID string,
	amount uint64,
	recv map[common.Hash]privacy.OTAReceiver,
	nftID common.Hash,
	nextOTA *AccessOTA,
	metaType int,
) (*WithdrawOrderRequest, error) {
	r := &WithdrawOrderRequest{
		PoolPairID: pairID,
		OrderID:    orderID,
		Amount:     amount,
		Receiver:   recv,
		NftID:      nftID,
		NextOTA: nextOTA,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
	return r, nil
}

func (req WithdrawOrderRequest) BurntAccessOTA() *AccessOTA {
	result := AccessOTA{}
	err := result.FromBytes([32]byte(req.NftID))
	if err != nil {
		return nil
	}
	return &result
}

func (req WithdrawOrderRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	err := beaconViewRetriever.IsValidPoolPairID(req.PoolPairID)
	if err != nil {
		return false, err
	}

	_, isReceivingNFT := req.Receiver[req.NftID]
	err = beaconViewRetriever.IsValidNftID(req.NftID.String())
	if err != nil {
		if isReceivingNFT || req.NextOTA == nil || req.BurntAccessOTA() == nil {
			metadataCommon.Logger.Log.Errorf("TX %s: invalid access with OTA. NFT: %v, NextOTA: %v, BurnOTA: %v", tx.Hash().String(), isReceivingNFT, req.NextOTA, req.BurntAccessOTA())
			return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		valid, err1 := ValidPdexv3Access(req.BurntAccessOTA(), *req.NextOTA, tx, common.ConfidentialAssetID, transactionStateDB)
		if !valid {
			return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v - %v", err, err1))
		}
	} else {
		if !isReceivingNFT || req.NextOTA != nil {
			metadataCommon.Logger.Log.Errorf("TX %s: invalid access with NFT. NFT: %v, NextOTA: %v", tx.Hash().String(), isReceivingNFT, req.NextOTA)
			return false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Invalid access for NftID"))
		}
		// Burned coin check
		isBurn, burnedPRVCoin, burnedCoin, burnedTokenID, err := tx.GetTxFullBurnData()
		if err != nil || !isBurn {
			return false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Burned coins not found in trade request - %v", err))
		}
		if *burnedTokenID != req.NftID {
			return false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Burned nftID mismatch - %v vs %v on metadata", *burnedTokenID, req.NftID))
		}
		// Type vs burned token id + amount check
		switch tx.GetType() {
		case common.TxCustomTokenPrivacyType:
			if burnedCoin == nil || burnedPRVCoin != nil {
				return false, metadataCommon.NewMetadataTxError(
					metadataCommon.PDEInvalidMetadataValueError,
					fmt.Errorf("Burned token invalid - must be %v only", req.NftID))
			}

			// accept any positive governance-NFT amount
			if burnedCoin.GetValue() <= 0 {
				return false, metadataCommon.NewMetadataTxError(
					metadataCommon.PDEInvalidMetadataValueError,
					fmt.Errorf("Invalid burned NFT amount %d",
						burnedCoin.GetValue()))
			}
		default:
			return false, fmt.Errorf("Invalid transaction type %v for withdrawOrder request", tx.GetType())
		}
	}
	return true, nil
}

func (req WithdrawOrderRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Feature pdexv3 has not been activated yet"))
	}

	// OTAReceiver check
	for _, item := range req.Receiver {
		if !item.IsValid() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Invalid OTAReceiver %v", item))
		}
		if tx.GetSenderAddrLastByte() != item.GetShardID() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Invalid shard %d for Receiver - must equal sender shard",
					item.GetShardID()))
		}
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
	for currentTokenID, val := range req.Receiver {
		if currentTokenID != common.PRVCoinID {
			currentTokenID = common.ConfidentialAssetID
		}
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: val.PublicKey.ToBytes(), TokenID: currentTokenID,
		})
	}
	return result
}
