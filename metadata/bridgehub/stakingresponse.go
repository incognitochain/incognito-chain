package bridgehub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

type BridgeHubStakingResponse struct {
	metadataCommon.MetadataBase
	RequestedTxID common.Hash         `json:"RequestedTxID"`
	StakeAmount   uint64              `json:"StakeAmount"`
	IncTokenID    common.Hash         `json:"IncTokenID"`
	Staker        privacy.OTAReceiver `json:"Staker"`
}

type BridgeHubStakeResAction struct {
	Meta       *BridgeHubStakingResponse `json:"meta"`
	IncTokenID *common.Hash              `json:"incTokenID"`
}

func NewBridgeHubStakingResponse(
	requestedTxID common.Hash,
	stakeAmount uint64,
	incTokenID common.Hash,
	staker privacy.OTAReceiver,
	metaType int,
) *BridgeHubStakingResponse {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	return &BridgeHubStakingResponse{
		RequestedTxID: requestedTxID,
		StakeAmount:   stakeAmount,
		IncTokenID:    incTokenID,
		Staker:        staker,
		MetadataBase:  metadataBase,
	}
}

func (iRes BridgeHubStakingResponse) CheckTransactionFee(tr metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes BridgeHubStakingResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes BridgeHubStakingResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes BridgeHubStakingResponse) ValidateMetadataByItself() bool {
	return iRes.Type == metadataCommon.BridgeHubStakeResponse
}

func (iRes BridgeHubStakingResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&iRes)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (iRes *BridgeHubStakingResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(iRes)
}

func (iRes BridgeHubStakingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not IssuingEVMRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.StakePRVRequestMeta) {
			continue
		}
		tempInst := metadataCommon.NewInstruction()
		err := tempInst.FromStringSlice(inst)
		if err != nil {
			metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction: ", err)
			continue
		}

		if tempInst.Status != common.RejectedStatusStr {
			continue
		}
		contentBytes, err := base64.StdEncoding.DecodeString(tempInst.Content)
		if err != nil {
			metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		failedContent := StakePRVRequestContentInst{}
		err = json.Unmarshal(contentBytes, &failedContent)
		if err != nil {
			continue
		}

		txHash := &common.Hash{}
		txHash, _ = txHash.NewHashFromStr(failedContent.TxReqID)
		if !bytes.Equal(iRes.RequestedTxID[:], txHash[:]) || shardID != tempInst.ShardID {
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != failedContent.TokenID.String() {
			continue
		}
		otaReceiver := failedContent.Staker
		cv2, ok := mintCoin.(*privacy.CoinV2)
		if !ok {
			metadataCommon.Logger.Log.Info("WARNING - VALIDATION: unrecognized mint coin version")
			continue
		}
		pk := cv2.GetPublicKey().ToBytesS()
		txR := cv2.GetTxRandom()
		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) || !bytes.Equal(txR[:], otaReceiver.TxRandom[:]) {
			metadataCommon.Logger.Log.Warnf("WARNING - VALIDATION: bridgehub stake failed PublicKey or TxRandom mismatch")
			continue
		}
		if cv2.GetValue() != failedContent.StakeAmount { // range check was done by producer
			metadataCommon.Logger.Log.Warnf("WARNING - VALIDATION: bridgehub stake failed amount mismatch - %d vs %d", cv2.GetValue(), failedContent.StakeAmount)
			continue
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.New(fmt.Sprintf("no IssuingETHRequest tx found for BridgeHubStakingResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
