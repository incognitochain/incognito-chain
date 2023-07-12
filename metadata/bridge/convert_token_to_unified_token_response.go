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
	ConvertAmount uint64      `json:"ConvertAmount"`
	Reward        uint64      `json:"Reward"`
	Status        string      `json:"Status"`
	TxReqID       common.Hash `json:"TxReqID"`
}

func NewConvertTokenToUnifiedTokenResponse() *ConvertTokenToUnifiedTokenResponse {
	return &ConvertTokenToUnifiedTokenResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta,
		},
	}
}

func NewBridgeAggConvertTokenToUnifiedTokenResponseWithValue(
	status string, txReqID common.Hash, convertAmount uint64, reward uint64,
) *ConvertTokenToUnifiedTokenResponse {
	return &ConvertTokenToUnifiedTokenResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BridgeAggConvertTokenToUnifiedTokenResponseMeta,
		},
		ConvertAmount: convertAmount,
		Reward:        reward,
		Status:        status,
		TxReqID:       txReqID,
	}
}

func (response *ConvertTokenToUnifiedTokenResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
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
	if response.ConvertAmount == 0 {
		return false, false, errors.New("Convert amount can not is zero")
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
			continue
		}

		var txReqIDFromInst common.Hash
		var otaReceiver privacy.OTAReceiver
		var mintingAmtFromInst uint64
		var convertAmtFromInst uint64
		var rewardFromInst uint64
		var mintedTokenID common.Hash
		shardIDFromInst := tempInst.ShardID

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

			txReqIDFromInst = rejectContent.TxReqID
			otaReceiver = rejectedData.Receiver
			mintedTokenID = rejectedData.TokenID
			convertAmtFromInst = rejectedData.Amount
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
			txReqIDFromInst = acceptedContent.TxReqID
			otaReceiver = acceptedContent.Receiver
			mintedTokenID = acceptedContent.UnifiedTokenID
			convertAmtFromInst = acceptedContent.ConvertPUnifiedAmount
			rewardFromInst = acceptedContent.Reward

		default:
			return false, errors.New("Not find status")
		}

		mintingAmtFromInst = convertAmtFromInst + rewardFromInst

		if response.TxReqID.String() != txReqIDFromInst.String() {
			metadataCommon.Logger.Log.Infof("BUGLOG txReqID: %v, %v\n", response.TxReqID.String(), txReqIDFromInst.String())
			continue
		}

		if shardID != shardIDFromInst {
			metadataCommon.Logger.Log.Infof("BUGLOG shardID: %v, %v\n", shardID, shardIDFromInst)
			continue
		}

		if response.Status != tempInst.Status {
			metadataCommon.Logger.Log.Error("ERROR - VALIDATION: an error occured while check response status: ")
			return false, errors.New("Invalid convert response status")
		}

		if response.ConvertAmount != convertAmtFromInst {
			metadataCommon.Logger.Log.Error("ERROR - VALIDATION: an error occured while check response convert amount: ")
			return false, errors.New("Invalid convert amount response")
		}

		if response.Reward != rewardFromInst {
			metadataCommon.Logger.Log.Error("ERROR - VALIDATION: an error occured while check response reward: ")
			return false, errors.New("Invalid convert reward response")
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

		if mintingAmtFromInst != paidAmount {
			return false, fmt.Errorf("Amount is invalid receive %d paid %d", mintingAmtFromInst, paidAmount)
		}

		if !bytes.Equal(txR[:], otaReceiver.TxRandom[:]) {
			return false, fmt.Errorf("otaReceiver tx random is invalid")
		}

		if mintedTokenID.String() != coinID.String() {
			return false, fmt.Errorf("Coin is invalid receive %s expect %s", mintedTokenID.String(), coinID.String())
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		metadataCommon.Logger.Log.Debugf("no bridgeagg convert instruction tx %s", tx.Hash().String())
		return false, fmt.Errorf(fmt.Sprintf("no bridgeagg convert instruction tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
