package bridge

import (
    "encoding/hex"
    "encoding/json"
    "fmt"

    "github.com/incognitochain/incognito-chain/common"
    "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
    metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
    "github.com/incognitochain/incognito-chain/privacy"
)

type BurnForCallRequest struct {
    BurnTokenID common.Hash `json:"BurnTokenID"`
    BurnForCallRequestData
    metadataCommon.MetadataBase
}

type BurnForCallRequestData struct {
    BurningAmount       uint64              `json:"BurningAmount"`
    MinExpectedAmount   uint64              `json:"MinExpectedAmount"`
    IncTokenID          common.Hash         `json:"IncTokenID"`
    ExternalCalldata    string              `json:"ExternalCalldata"`
    ExternalCallAddress string              `json:"ExternalCallAddress"`
    ReceiveToken        string              `json:"ReceiveToken"`
    RedepositReceiver   privacy.OTAReceiver `json:"RedepositReceiver"`
    WithdrawAddress     string              `json:"WithdrawAddress"`
}

func (bReq BurnForCallRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
    return true, nil
}

func (bReq BurnForCallRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
    hexStrings := []string{bReq.ExternalCalldata}
    for _, s := range hexStrings {
        if _, err := hex.DecodeString(s); err != nil {
            return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("invalid data %s, expect hex string", s))
        }
    }

    extAddressStrings := []string{bReq.ExternalCallAddress, bReq.ReceiveToken, bReq.WithdrawAddress}
    for _, s := range extAddressStrings {
        if decoded, err := hex.DecodeString(s); err != nil || len(decoded) != metadataCommon.ExternalAddressLen {
            return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("invalid data %s, expect hex string", s))
        }
    }

    isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
    if err != nil || !isBurned {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burn missing from tx %s - %v", tx.Hash(), err))
    }
    if *burnedTokenID != bReq.BurnTokenID {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burned tokenID mismatch - %s, %s", burnedTokenID.String(), bReq.BurnTokenID.String()))
    }

    // check validity of token IDs
    for _, t := range []common.Hash{bReq.IncTokenID, bReq.BurnTokenID} {
        if t == common.PDEXCoinID {
            return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("tokenID must not be special token"))
        }
    }
    burnAmount := burnCoin.GetValue()
    if bReq.BurningAmount == 0 || burnAmount == 0 || burnAmount != bReq.BurningAmount {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("invalid burn amount - %v, %v", burnAmount, bReq.BurningAmount))
    }
    if bReq.BurningAmount < bReq.MinExpectedAmount {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burningAmount %v < expectedAmount %v", bReq.BurningAmount, bReq.MinExpectedAmount))
    }

    return true, true, nil
}

func (bReq BurnForCallRequest) ValidateMetadataByItself() bool {
    return bReq.Type == metadataCommon.BurnForCallRequestMeta
}

func (bReq BurnForCallRequest) Hash() *common.Hash {
    rawBytes, _ := json.Marshal(bReq)
    hash := common.HashH([]byte(rawBytes))
    return &hash
}

func (request *BurnForCallRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
    content, err := metadataCommon.NewActionWithValue(request, *tx.Hash(), nil).StringSlice(metadataCommon.BurnForCallRequestMeta)
    return [][]string{content}, err
}

func (bReq *BurnForCallRequest) CalculateSize() uint64 {
    return metadataCommon.CalculateSize(bReq)
}

func (request *BurnForCallRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
    var result []metadataCommon.OTADeclaration
    result = append(result, metadataCommon.OTADeclaration{
        PublicKey: request.RedepositReceiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
    })
    return result
}
