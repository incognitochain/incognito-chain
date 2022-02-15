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

type UserMintNftResponse struct {
	metadataCommon.MetadataBase
	status  string
	txReqID string
}

func NewUserMintNftResponse() *UserMintNftResponse {
	return &UserMintNftResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UserMintNftResponseMeta,
		},
	}
}

func NewUserMintNftResponseWithValue(status, txReqID string) *UserMintNftResponse {
	return &UserMintNftResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3UserMintNftResponseMeta,
		},
		status:  status,
		txReqID: txReqID,
	}
}

func (response *UserMintNftResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (response *UserMintNftResponse) ValidateTxWithBlockChain(
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

func (response *UserMintNftResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if response.status != common.Pdexv3AcceptStringStatus && response.status != common.Pdexv3RejectStringStatus {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, errors.New("status cannot be empty"))
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

func (response *UserMintNftResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.Pdexv3UserMintNftResponseMeta
}

func (response *UserMintNftResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *UserMintNftResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *UserMintNftResponse) MarshalJSON() ([]byte, error) {
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

func (response *UserMintNftResponse) UnmarshalJSON(data []byte) error {
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

func (response *UserMintNftResponse) TxReqID() string {
	return response.txReqID
}

func (response *UserMintNftResponse) Status() string {
	return response.status
}

type acceptUserMintNft struct {
	NftID       common.Hash `json:"NftID"`
	BurntAmount uint64      `json:"BurntAmount"`
	OtaReceiver string      `json:"OtaReceiver"`
	ShardID     byte        `json:"ShardID"`
	TxReqID     common.Hash `json:"TxReqID"`
}

type refundUserMintNft struct {
	OtaReceiver string      `json:"OtaReceiver"`
	Amount      uint64      `json:"Amount"`
	ShardID     byte        `json:"ShardID"`
	TxReqID     common.Hash `json:"TxReqID"`
}

func (response *UserMintNftResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.Pdexv3UserMintNftRequestMeta) {
			continue
		}
		instContributionStatus := inst[1]
		if instContributionStatus != response.status || (instContributionStatus != common.Pdexv3AcceptStringStatus && instContributionStatus != common.Pdexv3RejectStringStatus) {
			continue
		}

		contentBytes := []byte(inst[2])

		var instShardID byte
		var tokenID common.Hash
		var otaReceiverStr, txReqID string
		var amount uint64
		switch inst[1] {
		case common.Pdexv3RejectStringStatus:
			var instContent refundUserMintNft
			err := json.Unmarshal(contentBytes, &instContent)
			if err != nil {
				metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				continue
			}
			instShardID = instContent.ShardID
			tokenID = common.PRVCoinID
			otaReceiverStr = instContent.OtaReceiver
			amount = instContent.Amount
			txReqID = instContent.TxReqID.String()
		case common.Pdexv3AcceptStringStatus:
			var instContent acceptUserMintNft
			err := json.Unmarshal(contentBytes, &instContent)
			if err != nil {
				metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				continue
			}
			instShardID = instContent.ShardID
			tokenID = instContent.NftID
			otaReceiverStr = instContent.OtaReceiver
			amount = 1
			txReqID = instContent.TxReqID.String()
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

		otaReceiver := coin.OTAReceiver{}
		err = otaReceiver.FromString(otaReceiverStr)
		if err != nil {
			return false, errors.New("Invalid ota receiver")
		}

		txR := mintCoin.(*coin.CoinV2).GetTxRandom()
		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) ||
			amount != paidAmount ||
			!bytes.Equal(txR[:], otaReceiver.TxRandom[:]) ||
			tokenID.String() != coinID.String() {
			return false, errors.New("Coin is invalid")
		}
		idx = i
		fmt.Println("BUGLOG Verify Metadata --- OK")
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		metadataCommon.Logger.Log.Debugf("no Pdexv3 user mint nft instruction tx %s", tx.Hash().String())
		return false, fmt.Errorf(fmt.Sprintf("no Pdexv3 user mint nft instruction tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
