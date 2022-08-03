package bridge

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "strconv"

    "github.com/incognitochain/incognito-chain/common"
    "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
    metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
    "github.com/incognitochain/incognito-chain/privacy"
)

type IssuingReshieldResponse struct {
    metadataCommon.MetadataBase
    RequestedTxID   common.Hash `json:"RequestedTxID"`
    UniqTx          []byte      `json:"UniqETHTx"`
    ExternalTokenID []byte      `json:"ExternalTokenID"`
}

type AcceptedReshieldRequest struct {
    UnifiedTokenID *common.Hash              `json:"UnifiedTokenID"`
    Receiver       privacy.OTAReceiver       `json:"Receiver"`
    TxReqID        common.Hash               `json:"TxReqID"`
    ReshieldData   AcceptedShieldRequestData `json:"ReshieldData"`
}

func NewIssuingReshieldResponse(
    requestedTxID common.Hash,
    uniqTx []byte,
    externalTokenID []byte,
    metaType int,
) *IssuingReshieldResponse {
    metadataBase := metadataCommon.MetadataBase{
        Type: metaType,
    }
    return &IssuingReshieldResponse{
        RequestedTxID:   requestedTxID,
        UniqTx:          uniqTx,
        ExternalTokenID: externalTokenID,
        MetadataBase:    metadataBase,
    }
}

func (iRes IssuingReshieldResponse) CheckTransactionFee(tr metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
    // no need to have fee for this tx
    return true
}

func (iRes IssuingReshieldResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
    // no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
    return false, nil
}

func (iRes IssuingReshieldResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
    return false, true, nil
}

func (iRes IssuingReshieldResponse) ValidateMetadataByItself() bool {
    return iRes.Type == metadataCommon.IssuingReshieldResponseMeta
}

func (iRes IssuingReshieldResponse) Hash() *common.Hash {
    record := iRes.RequestedTxID.String()
    record += string(iRes.UniqTx)
    record += string(iRes.ExternalTokenID)
    record += iRes.MetadataBase.Hash().String()

    // final hash
    hash := common.HashH([]byte(record))
    return &hash
}

func (iRes *IssuingReshieldResponse) CalculateSize() uint64 {
    return metadataCommon.CalculateSize(iRes)
}

func (iRes IssuingReshieldResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
    idx := -1
    for i, inst := range mintData.Insts {
        if len(inst) < 4 { // this is not IssuingEVMRequest instruction
            continue
        }
        instMetaType := inst[0]
        if mintData.InstsUsed[i] > 0 ||
            instMetaType != strconv.Itoa(metadataCommon.IssuingReshieldResponseMeta) {
            continue
        }

        contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
        if err != nil {
            metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
            continue
        }
        var acceptedInst AcceptedReshieldRequest
        err = json.Unmarshal(contentBytes, &acceptedInst)
        if err != nil {
            metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
            continue
        }

        if !bytes.Equal(iRes.RequestedTxID[:], acceptedInst.TxReqID[:]) ||
            !bytes.Equal(iRes.UniqTx, acceptedInst.ReshieldData.UniqTx) ||
            !bytes.Equal(iRes.ExternalTokenID, acceptedInst.ReshieldData.ExternalTokenID) ||
            strconv.Itoa(int(shardID)) != inst[1] {
            b, _ := json.Marshal(iRes)
            metadataCommon.Logger.Log.Warnf("WARNING - VALIDATION: response / instruction content mismatch: %v vs %s - skipping", inst, string(b))
            continue
        }
        expectedMintTokenID := acceptedInst.ReshieldData.IncTokenID
        if acceptedInst.UnifiedTokenID != nil {
            expectedMintTokenID = *acceptedInst.UnifiedTokenID
        }

        isMinted, mintCoin, coinID, err := tx.GetTxMintData()
        if err != nil || !isMinted || coinID.String() != expectedMintTokenID.String() {
            continue
        }
        otaReceiver := acceptedInst.Receiver
        cv2, ok := mintCoin.(*privacy.CoinV2)
        if !ok {
            metadataCommon.Logger.Log.Info("WARNING - VALIDATION: unrecognized mint coin version")
            continue
        }
        pk := cv2.GetPublicKey().ToBytesS()
        txR := cv2.GetTxRandom()
        if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) || !bytes.Equal(txR[:], otaReceiver.TxRandom[:]) {
            metadataCommon.Logger.Log.Warnf("WARNING - VALIDATION: reshield PublicKey or TxRandom mismatch")
            continue
        }
        if cv2.GetValue() != acceptedInst.ReshieldData.ShieldAmount {
            metadataCommon.Logger.Log.Warnf("WARNING - VALIDATION: reshield amount mismatch - %d vs %d", cv2.GetValue(), acceptedInst.ReshieldData.ShieldAmount)
            continue
        }
        idx = i
        break
    }
    if idx == -1 { // not found the issuance request tx for this response
        return false, errors.New(fmt.Sprintf("reshield: no IssuingRequest tx found for IssuingReshieldResponse tx %s", tx.Hash().String()))
    }
    mintData.InstsUsed[idx] = 1
    return true, nil
}
