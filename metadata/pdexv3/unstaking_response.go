package pdexv3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

type UnstakingResponse struct {
	metadataCommon.MetadataBase
	status  string
	txReqID string
}

func NewUnstakingResponse() *UnstakingResponse {
	return &UnstakingResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UnstakingResponseMeta,
		},
	}
}

func NewUnstakingResponseWithValue(status, txReqID string) *UnstakingResponse {
	return &UnstakingResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UnstakingResponseMeta,
		},
		status:  status,
		txReqID: txReqID,
	}
}

func (response *UnstakingResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (response *UnstakingResponse) ValidateTxWithBlockChain(
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

func (response *UnstakingResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if response.status != common.Pdexv3AcceptUnstakingStatus {
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

func (response *UnstakingResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.Pdexv3UnstakingResponseMeta
}

func (response *UnstakingResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *UnstakingResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *UnstakingResponse) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(response)
}

func (response *UnstakingResponse) MarshalJSON() ([]byte, error) {
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

func (response *UnstakingResponse) UnmarshalJSON(data []byte) error {
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

func (response *UnstakingResponse) TxReqID() string {
	return response.txReqID
}

func (response *UnstakingResponse) Status() string {
	return response.status
}

type acceptUnstaking struct {
	StakingPoolID common.Hash `json:"StakingPoolID"`
	NftID         common.Hash `json:"NftID"`
	Amount        uint64      `json:"Amount"`
	OtaReceiver   string      `json:"OtaReceiver"`
	TxReqID       common.Hash `json:"TxReqID"`
	ShardID       byte        `json:"ShardID"`
}

func (response *UnstakingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if len(inst) != 3 {
			continue
		}
		metadataCommon.Logger.Log.Infof("BUGLOG currently processing inst: %v\n", inst)
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.Pdexv3UnstakingRequestMeta) {
			continue
		}
		instContributionStatus := inst[1]
		if instContributionStatus != response.status || instContributionStatus != common.Pdexv3AcceptUnstakingStatus {
			continue
		}

		contentBytes := []byte(inst[2])
		var instContent acceptUnstaking
		err := json.Unmarshal(contentBytes, &instContent)
		if err != nil {
			metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if response.TxReqID() != instContent.TxReqID.String() || shardID != instContent.ShardID {
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

		otaReceiver := coin.OTAReceiver{}
		metadataCommon.Logger.Log.Info("instContent.OtaReceive:", instContent.OtaReceiver)
		err = otaReceiver.FromString(instContent.OtaReceiver)
		if err != nil {
			return false, errors.New("Invalid ota receiver")
		}

		tempMintCoin, ok := mintCoin.(*coin.CoinV2)
		if !ok {
			return false, errors.New("mint coin is not coin v2")
		}
		txR := tempMintCoin.GetTxRandom()

		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) ||
			instContent.Amount != paidAmount ||
			!bytes.Equal(txR[:], otaReceiver.TxRandom[:]) ||
			instContent.StakingPoolID.String() != coinID.String() {
			metadataCommon.Logger.Log.Debugf("mintcoinPk %v otaReceiverPk %v", pk, otaReceiver.PublicKey.ToBytesS())
			metadataCommon.Logger.Log.Debugf("instContent.Amount %v paidAmount %v", instContent.Amount, paidAmount)
			metadataCommon.Logger.Log.Debugf("txR %v otaReceiver.TxRandom %v", txR, otaReceiver.TxRandom)
			metadataCommon.Logger.Log.Debugf("instContent.StakingPoolID %v coinID %v", instContent.StakingPoolID.String(), coinID.String())
			return false, errors.New("Coin is invalid")
		}
		idx = i
		metadataCommon.Logger.Log.Info("BUGLOG Verify Metadata --- OK")
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		metadataCommon.Logger.Log.Debugf("no Pdexv3 unstaking instruction tx %s", tx.Hash().String())
		return false, fmt.Errorf(fmt.Sprintf("no Pdexv3 unstaking instruction tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
