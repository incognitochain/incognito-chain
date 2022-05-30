package bridge

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

type UnshieldResponse struct {
	metadataCommon.MetadataBase
	Status        string      `json:"Status"`
	RequestedTxID common.Hash `json:"RequestedTxID"`
}

func NewUnshieldResponse() *UnshieldResponse {
	return &UnshieldResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BurningUnifiedTokenResponseMeta,
		},
	}
}

func NewUnshieldResponseWithValue(
	status string, requestedTxID common.Hash,
) *UnshieldResponse {
	return &UnshieldResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BurningUnifiedTokenResponseMeta,
		},
		Status:        status,
		RequestedTxID: requestedTxID,
	}
}

func (response *UnshieldResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (response *UnshieldResponse) ValidateTxWithBlockChain(
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

func (response *UnshieldResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if response.Status != common.RejectedStatusStr {
		return false, false, errors.New("Status is invalid")
	}
	return true, true, nil
}

func (response *UnshieldResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.BurningUnifiedTokenResponseMeta
}

func (response *UnshieldResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *UnshieldResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *UnshieldResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.BurningUnifiedTokenRequestMeta) {
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
		var rejectedData RejectedUnshieldRequest
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
			return false, errors.New("This is not tx mint")
		}
		pk := mintCoin.GetPublicKey().ToBytesS()
		paidAmount := mintCoin.GetValue()
		txR := mintCoin.(*coin.CoinV2).GetTxRandom()

		if !bytes.Equal(rejectedData.Receiver.PublicKey.ToBytesS(), pk[:]) {
			return false, errors.New("OTAReceiver public key is invalid")
		}

		if rejectedData.Amount != paidAmount {
			return false, fmt.Errorf("Amount is invalid receive %d paid %d", rejectedData.Amount, paidAmount)
		}

		if !bytes.Equal(txR[:], rejectedData.Receiver.TxRandom[:]) {
			return false, fmt.Errorf("otaReceiver tx random is invalid")
		}

		if rejectedData.TokenID.String() != coinID.String() {
			return false, fmt.Errorf("Coin is invalid receive %s expect %s", rejectedData.TokenID.String(), coinID.String())
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
