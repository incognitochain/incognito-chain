package bridge

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "bytes"
    "strconv"

    "github.com/incognitochain/incognito-chain/common"
    "github.com/incognitochain/incognito-chain/config"
    "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
    metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
    "github.com/incognitochain/incognito-chain/privacy"
)

type BurnForCallRequest struct {
    BurnTokenID common.Hash              `json:"BurnTokenID"`
    Data        []BurnForCallRequestData `json:"Data"`
    metadataCommon.MetadataBase
}

type BurnForCallRequestData struct {
    BurningAmount       uint64              `json:"BurningAmount"`
    ExternalNetworkID   uint8               `json:"ExternalNetworkID"`
    IncTokenID          common.Hash         `json:"IncTokenID"`
    ExternalCalldata    string              `json:"ExternalCalldata"`
    ExternalCallAddress string              `json:"ExternalCallAddress"`
    ReceiveToken        string              `json:"ReceiveToken"`
    RedepositReceiver   privacy.OTAReceiver `json:"RedepositReceiver"`
    WithdrawAddress     string              `json:"WithdrawAddress"`
}

type RejectedBurnForCallRequest struct {
    BurnTokenID common.Hash         `json:"BurnTokenID"`
    Amount      uint64              `json:"Amount"`
    Receiver    privacy.OTAReceiver `json:"Receiver"`
}

func (bReq BurnForCallRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
    return true, nil
}

func (bReq BurnForCallRequest) TotalBurningAmount() (uint64, error) {
    var totalBurningAmount uint64 = 0
    for _, d := range bReq.Data {
        totalBurningAmount += d.BurningAmount
        if totalBurningAmount < d.BurningAmount {
            return 0, fmt.Errorf("out of range uint64")
        }
    }
    return totalBurningAmount, nil
}

func (bReq BurnForCallRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
    if len(bReq.Data) <= 0 || len(bReq.Data) > int(config.Param().BridgeAggParam.MaxLenOfPath) {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("Length of data %d need to be in [1..%d]", len(bReq.Data), config.Param().BridgeAggParam.MaxLenOfPath))
    }
    isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
    if err != nil || !isBurned {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burn missing from tx %s - %v", tx.Hash(), err))
    }
    if *burnedTokenID != bReq.BurnTokenID {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("burned tokenID mismatch - %s, %s", burnedTokenID.String(), bReq.BurnTokenID.String()))
    }

    tokenIDs := []common.Hash{bReq.BurnTokenID}
    for _, d := range bReq.Data {
        hexStrings := []string{d.ExternalCalldata}
        for _, s := range hexStrings {
            if _, err := hex.DecodeString(s); err != nil {
                return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("invalid data %s, expect hex string", s))
            }
        }
        extAddressStrings := []string{d.ExternalCallAddress, d.ReceiveToken, d.WithdrawAddress}
        for _, s := range extAddressStrings {
            if decoded, err := hex.DecodeString(s); err != nil || len(decoded) != metadataCommon.ExternalAddressLen {
                return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("invalid data %s, expect hex string", s))
            }
        }

        if d.RedepositReceiver.GetShardID() != byte(tx.GetValidationEnv().ShardID()) {
            return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("otaReceiver shardID is different from txShardID"))
        }
        if d.BurningAmount == 0 {
            return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("wrong request info's burned amount"))
        }

        tokenIDs = append(tokenIDs, d.IncTokenID)
    }

    burnAmount := burnCoin.GetValue()
    totalBurningAmount, err := bReq.TotalBurningAmount()
    if err != nil {
        return false, false, err
    }
    if burnAmount != totalBurningAmount {
        return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("invalid burn amount - %v, %v", burnAmount, totalBurningAmount))
    }

    // check validity of token IDs
    for _, t := range tokenIDs {
        if t == common.PDEXCoinID {
            return false, false, metadataCommon.NewMetadataTxError(metadataCommon.BridgeAggUnshieldValidateSanityDataError, fmt.Errorf("tokenID must not be special token"))
        }
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

func (bReq *BurnForCallRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
    var result []metadataCommon.OTADeclaration
    for _, d := range bReq.Data {
        result = append(result, metadataCommon.OTADeclaration{
            PublicKey: d.RedepositReceiver.PublicKey.ToBytes(), TokenID: common.ConfidentialAssetID,
        })
    }
    return result
}

type BurnForCallResponse struct {
    UnshieldResponse
}

func NewBurnForCallResponseWithValue(
    status string, requestedTxID common.Hash,
) *BurnForCallResponse {
    return &BurnForCallResponse{
        UnshieldResponse{
            MetadataBase: metadataCommon.MetadataBase{
                Type: metadataCommon.BurnForCallResponseMeta,
            },
            Status:        status,
            RequestedTxID: requestedTxID,
        }}
}

func (response *BurnForCallResponse) ValidateMetadataByItself() bool {
    return response.Type == metadataCommon.BurnForCallResponseMeta
}

func (response *BurnForCallResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
    mintData *metadataCommon.MintData,
    shardID byte,
    tx metadataCommon.Transaction,
    chainRetriever metadataCommon.ChainRetriever,
    ac *metadataCommon.AccumulatedValues,
    shardViewRetriever metadataCommon.ShardViewRetriever,
    beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
    idx := -1
    metadataCommon.Logger.Log.Infof("Currently verifying ins: %v\n", response)
    for i, inst := range mintData.Insts {
        if len(inst) != 4 { // this is not bridgeagg instruction
            continue
        }
        instMetaType := inst[0]
        if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.BurnForCallResponseMeta) {
            continue
        }
        tempInst := metadataCommon.NewInstruction()
        if err := tempInst.FromStringSlice(inst); err != nil {
            return false, err
        }
        if tempInst.Status != common.RejectedStatusStr {
            continue
        }
        rejectContent := metadataCommon.NewRejectContent()
        if err := rejectContent.FromString(tempInst.Content); err != nil {
            return false, err
        }
        var rejectedData RejectedBurnForCallRequest
        if err := json.Unmarshal(rejectContent.Data, &rejectedData); err != nil {
            return false, err
        }
        if shardID != tempInst.ShardID {
            metadataCommon.Logger.Log.Infof("BUGLOG shardID: %v, %v\n", shardID, tempInst.ShardID)
            continue
        }
        if response.RequestedTxID.String() != rejectContent.TxReqID.String() {
            metadataCommon.Logger.Log.Infof("BUGLOG txReqID: %v, %v\n", response.RequestedTxID.String(), rejectContent.TxReqID.String())
            continue
        }
        isMinted, mintCoin, coinID, err := tx.GetTxMintData()
        if err != nil {
            metadataCommon.Logger.Log.Error("ERROR - VALIDATION: an error occured while get tx mint data: ", err)
            return false, err
        }
        if !isMinted {
            metadataCommon.Logger.Log.Info("WARNING - VALIDATION: this is not Tx Mint: ")
            return false, fmt.Errorf("This is not tx mint")
        }
        paidAmount := mintCoin.GetValue()
        cv2, ok := mintCoin.(*privacy.CoinV2)
        if !ok {
            metadataCommon.Logger.Log.Info("WARNING - VALIDATION: unrecognized mint coin version")
            continue
        }
        pk := cv2.GetPublicKey().ToBytesS()
        txR := cv2.GetTxRandom()

        if !bytes.Equal(rejectedData.Receiver.PublicKey.ToBytesS(), pk[:]) {
            return false, fmt.Errorf("OTAReceiver public key is invalid")
        }

        if rejectedData.Amount != paidAmount {
            return false, fmt.Errorf("Amount is invalid receive %d paid %d", rejectedData.Amount, paidAmount)
        }

        if !bytes.Equal(txR[:], rejectedData.Receiver.TxRandom[:]) {
            return false, fmt.Errorf("otaReceiver tx random is invalid")
        }

        if rejectedData.BurnTokenID.String() != coinID.String() {
            return false, fmt.Errorf("Coin is invalid receive %s expect %s", rejectedData.BurnTokenID.String(), coinID.String())
        }
        idx = i
        break
    }
    if idx == -1 { // not found the issuance request tx for this response
        metadataCommon.Logger.Log.Debugf("no bridgeagg unshield instruction tx %s", tx.Hash().String())
        return false, fmt.Errorf(fmt.Sprintf("no bridgeagg unshield instruction tx %s", tx.Hash().String()))
    }
    mintData.InstsUsed[idx] = 1
    return true, nil
}
