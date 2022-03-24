package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataBridgeAgg "github.com/incognitochain/incognito-chain/metadata/bridgeagg"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type AcceptedUnshieldRequest struct {
	BurnerAddress privacy.PaymentAddress `json:"BurnerAddress"`
	Amount        uint64                 `json:"BurningAmount"`
	TokenID       common.Hash            `json:"TokenID"`
	NetworkID     uint                   `json:"NetworkID,omitempty"`
	Fee           uint64                 `json:"Fee"`
	TxReqID       common.Hash            `json:"TxReqID"`
}

type BurningResponse struct {
	metadataCommon.MetadataBase
	Status       string      `json:"Status"`
	TxReqID      common.Hash `json:"TxReqID"`
	SharedRandom []byte      `json:"SharedRandom,omitempty"`
}

func NewBuringResponse() *BurningResponse {
	return &BurningResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BurningUnifiedTokenResponseMeta,
		},
	}
}

func NewBuringResponseWithValue(
	status string, txReqID common.Hash,
) *BurningResponse {
	return &BurningResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.BurningUnifiedTokenResponseMeta,
		},
		Status:  status,
		TxReqID: txReqID,
	}
}

func (response *BurningResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (response *BurningResponse) ValidateTxWithBlockChain(
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

func (response *BurningResponse) ValidateSanityData(
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

func (response *BurningResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.BurningUnifiedTokenResponseMeta
}

func (response *BurningResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *BurningResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *BurningResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.BurningUnifiedTokenResponseMeta) {
			continue
		}
		tempInst := metadataCommon.NewInstruction()
		if err := tempInst.FromStringSlice(inst); err != nil {
			return false, err
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var address privacy.PaymentAddress
		var receivingAmtFromInst uint64
		var receivingTokenID common.Hash

		switch tempInst.Status {
		case common.RejectedStatusStr:
			rejectContent := metadataCommon.NewRejectContent()
			if err := rejectContent.FromString(tempInst.Content); err != nil {
				return false, err
			}
			mdData, _ := rejectContent.Meta.(*metadataBridgeAgg.BurningRequest)
			shardIDFromInst = tempInst.ShardID
			txReqIDFromInst = rejectContent.TxReqID
			receivingTokenID = mdData.TokenID
			receivingAmtFromInst = mdData.BurningAmount
			address = mdData.BurnerAddress
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
			continue
		}
		if !isMinted {
			metadataCommon.Logger.Log.Info("WARNING - VALIDATION: this is not Tx Mint")
			continue
		}
		if coinID.String() != receivingTokenID.String() {
			metadataCommon.Logger.Log.Info("WARNING - VALIDATION: coinID is not similar to receivingTokenID")
			continue
		}

		if ok := mintCoin.CheckCoinValid(address, response.SharedRandom, receivingAmtFromInst); !ok {
			continue
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

func (response *BurningResponse) SetSharedRandom(r []byte) {
	response.SharedRandom = r
}
