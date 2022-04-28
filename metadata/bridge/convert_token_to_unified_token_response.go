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
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

type ConvertTokenToUnifiedTokenResponse struct {
	metadataCommon.MetadataBase
	Status  string      `json:"Status"`
	TxReqID common.Hash `json:"TxReqID"`
}

func NewConvertTokenToUnifiedTokenResponse() *ConvertTokenToUnifiedTokenResponse {
	return &ConvertTokenToUnifiedTokenResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta,
		},
	}
}

func NewBridgeAggConvertTokenToUnifiedTokenResponseWithValue(
	status string, txReqID common.Hash,
) *ConvertTokenToUnifiedTokenResponse {
	return &ConvertTokenToUnifiedTokenResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta,
		},
		Status:  status,
		TxReqID: txReqID,
	}
}

func (response *ConvertTokenToUnifiedTokenResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (response *ConvertTokenToUnifiedTokenResponse) ValidateTxWithBlockChain(
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	// NOTE: verify supported tokens pair as needed
	return true, nil
}

func (response *ConvertTokenToUnifiedTokenResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if response.Status != common.AcceptedStatusStr && response.Status != common.RejectedStatusStr {
		return false, false, errors.New("Status is invalid")
	}
	return true, true, nil
}

func (response *ConvertTokenToUnifiedTokenResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta
}

func (response *ConvertTokenToUnifiedTokenResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *ConvertTokenToUnifiedTokenResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *ConvertTokenToUnifiedTokenResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
	metadataCommon.Logger.Log.Infof("BUGLOG There are %v inst\n", len(mintData.Insts))
	for i, inst := range mintData.Insts {
		if len(inst) != 4 { // this is not bridgeagg instruction
			continue
		}
		metadataCommon.Logger.Log.Infof("BUGLOG currently processing inst: %v\n", inst)
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta) {
			continue
		}
		tempInst := metadataCommon.NewInstruction()
		if err := tempInst.FromStringSlice(inst); err != nil {
			return false, err
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var otaReceiver privacy.OTAReceiver
		var receivingAmtFromInst uint64
		var receivingTokenID common.Hash

		switch tempInst.Status {
		case common.RejectedStatusStr:
			rejectContent := metadataCommon.NewRejectContent()
			if err := rejectContent.FromString(tempInst.Content); err != nil {
				return false, err
			}
			var rejectedData RejectedConvertTokenToUnifiedToken
			if err := json.Unmarshal(rejectContent.Data, &rejectedData); err != nil {
				return false, err
			}

			shardIDFromInst = tempInst.ShardID
			txReqIDFromInst = rejectContent.TxReqID
			otaReceiver = rejectedData.Receiver
			receivingTokenID = rejectedData.TokenID
			receivingAmtFromInst = rejectedData.Amount
		case common.AcceptedStatusStr:
			contentBytes, err := base64.StdEncoding.DecodeString(tempInst.Content)
			if err != nil {
				return false, err
			}
			acceptedContent := AcceptedConvertTokenToUnifiedToken{}
			err = json.Unmarshal(contentBytes, &acceptedContent)
			if err != nil {
				return false, err
			}
			shardIDFromInst = tempInst.ShardID
			txReqIDFromInst = acceptedContent.TxReqID
			otaReceiver = acceptedContent.Receiver
			receivingTokenID = acceptedContent.UnifiedTokenID
			receivingAmtFromInst = acceptedContent.MintAmount
		default:
			return false, errors.New("Not find status")
		}

		if response.TxReqID.String() != txReqIDFromInst.String() {
			metadataCommon.Logger.Log.Infof("BUGLOG txReqID: %v, %v\n", response.TxReqID.String(), txReqIDFromInst.String())
			continue
		}

		if shardID != shardIDFromInst {
			metadataCommon.Logger.Log.Infof("BUGLOG shardID: %v, %v\n", shardID, shardIDFromInst)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil {
			metadataCommon.Logger.Log.Error("ERROR - VALIDATION: an error occured while get tx mint data: ", err)
			return false, err
		}
		if !isMinted {
			metadataCommon.Logger.Log.Info("WARNING - VALIDATION: this is not Tx Mint: ")
			return false, errors.New("This is not tx mint")
		}
		pk := mintCoin.GetPublicKey().ToBytesS()
		paidAmount := mintCoin.GetValue()
		txR := mintCoin.(*coin.CoinV2).GetTxRandom()

		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) {
			return false, errors.New("OTAReceiver public key is invalid")
		}

		if receivingAmtFromInst != paidAmount {
			return false, fmt.Errorf("Amount is invalid receive %d paid %d", receivingAmtFromInst, paidAmount)
		}

		if !bytes.Equal(txR[:], otaReceiver.TxRandom[:]) {
			return false, fmt.Errorf("otaReceiver tx random is invalid")
		}

		if receivingTokenID.String() != coinID.String() {
			return false, fmt.Errorf("Coin is invalid receive %s expect %s", receivingTokenID.String(), coinID.String())
		}

		idx = i
		fmt.Println("BUGLOG Verify Metadata --- OK")
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		metadataCommon.Logger.Log.Debugf("no bridgeagg convert instruction tx %s", tx.Hash().String())
		return false, fmt.Errorf(fmt.Sprintf("no bridgeagg convert instruction tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
