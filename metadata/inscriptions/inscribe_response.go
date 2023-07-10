package ink

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

type InscribeResponse struct {
	metadataCommon.MetadataBase
	status  string
	txReqID string
}

func NewInscribeResponseWithValue(status, txReqID string) *InscribeResponse {
	return &InscribeResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.InscribeResponseMeta,
		},
		status:  status,
		txReqID: txReqID,
	}
}

func (response *InscribeResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (response *InscribeResponse) ValidateTxWithBlockChain(
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

func (response *InscribeResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if response.status != strconv.Itoa(metadataPdexv3.OrderAcceptedStatus) && response.status != strconv.Itoa(metadataPdexv3.OrderRefundedStatus) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("status can not be empty"))
	}
	txReqID, err := common.Hash{}.NewHashFromStr(response.txReqID)
	if err != nil {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
	}
	if txReqID.IsZeroValue() {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("TxReqID should not be empty"))
	}
	return true, true, nil
}

func (response *InscribeResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.InscribeResponseMeta
}

func (response *InscribeResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *InscribeResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *InscribeResponse) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Status  string `json:"Status"`
		TxReqID string `json:"TxReqID"`
		metadataCommon.MetadataBase
	}{
		Status:       response.status,
		TxReqID:      response.txReqID,
		MetadataBase: response.MetadataBase,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (response *InscribeResponse) UnmarshalJSON(data []byte) error {
	temp := struct {
		Status  string `json:"Status"`
		TxReqID string `json:"TxReqID"`
		metadataCommon.MetadataBase
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	response.txReqID = temp.TxReqID
	response.status = temp.Status
	response.MetadataBase = temp.MetadataBase
	return nil
}

func (response *InscribeResponse) TxReqID() string {
	return response.txReqID
}

func (response *InscribeResponse) Status() string {
	return response.status
}

type MintNftData struct {
	NftID       common.Hash `json:"NftID"`
	OtaReceiver string      `json:"OtaReceiver"`
	ShardID     byte        `json:"ShardID"`
}

func (response *InscribeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *metadataCommon.MintData,
	shardID byte,
	tx metadataCommon.Transaction,
	chainRetriever metadataCommon.ChainRetriever,
	ac *metadataCommon.AccumulatedValues,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
) (bool, error) {
	idx := -1
	metadataCommon.Logger.Log.Infof("Verifying ins: %v of %d", response, len(mintData.Insts))
	for i, inst := range mintData.Insts {
		metadataCommon.Logger.Log.Infof("currently processing inst: %v\n", inst)
		if len(inst) != 5 {
			continue
		}

		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.InscribeRequestMeta) {
			continue
		}
		instContributionStatus := inst[1]
		if instContributionStatus != response.status || (instContributionStatus != strconv.Itoa(metadataPdexv3.OrderAcceptedStatus) && instContributionStatus != strconv.Itoa(metadataPdexv3.OrderRefundedStatus)) {
			continue
		}

		contentBytes := []byte(inst[4])
		shardIDStr := inst[2]
		txReqIDStr := inst[3]

		var instShardID byte
		var tokenID common.Hash
		var txReqID string
		var amount uint64
		var otaReceiver privacy.OTAReceiver
		switch inst[1] {
		case strconv.Itoa(metadataPdexv3.OrderRefundedStatus):
			// var instContent refundUserMintNft
			// err := json.Unmarshal(contentBytes, &instContent)
			// if err != nil {
			// 	metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			// 	continue
			// }
			// instShardID = instContent.ShardID
			// tokenID = common.PRVCoinID
			// otaReceiverStr = instContent.OtaReceiver
			// amount = instContent.Amount
			// txReqID = instContent.TxReqID.String()
		case strconv.Itoa(metadataPdexv3.OrderAcceptedStatus):
			var instContent struct {
				Content InscribeAcceptedAction
			}
			err := json.Unmarshal(contentBytes, &instContent)
			if err != nil {
				metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				continue
			}
			tokenID = instContent.Content.TokenID
			otaReceiver = instContent.Content.Receiver
			amount = 1
			n, _ := strconv.Atoi(shardIDStr)
			instShardID = byte(n)
			h, _ := (common.Hash{}).NewHashFromStr(txReqIDStr)
			txReqID = h.String()
		default:
			continue
		}

		if response.TxReqID() != txReqID || shardID != instShardID {
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
		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) ||
			amount != paidAmount ||
			!bytes.Equal(txR[:], otaReceiver.TxRandom[:]) ||
			tokenID.String() != coinID.String() {
			return false, errors.New("Coin is invalid")
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		metadataCommon.Logger.Log.Debugf("no InscribeResponse instruction tx %s", tx.Hash().String())
		return false, fmt.Errorf(fmt.Sprintf("no InscribeResponse instruction tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
