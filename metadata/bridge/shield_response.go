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
)

type ShieldResponseData struct {
	ExternalTokenID []byte      `json:"ExternalTokenID"`
	UniqTx          []byte      `json:"UniqTx"`
	IncTokenID      common.Hash `json:"IncTokenID"`
}

type ShieldResponse struct {
	metadataCommon.MetadataBase
	RequestedTxID common.Hash          `json:"RequestedTxID"`
	ShieldAmount  uint64               `json:"ShieldAmount"`
	Reward        uint64               `json:"Reward"`
	Data          []ShieldResponseData `json:"Data"`
	SharedRandom  []byte               `json:"SharedRandom,omitempty"`
}

func NewShieldResponse(metaType int) *ShieldResponse {
	return &ShieldResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
}

func NewShieldResponseWithValue(
	metaType int, shieldAmount, reward uint64, data []ShieldResponseData, requestedTxID common.Hash, sharedRandom []byte,
) *ShieldResponse {
	return &ShieldResponse{
		RequestedTxID: requestedTxID,
		ShieldAmount:  shieldAmount,
		Reward:        reward,
		Data:          data,
		SharedRandom:  sharedRandom,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
}

func (response *ShieldResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (response *ShieldResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {

	return true, true, nil
}

func (response *ShieldResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.IssuingUnifiedTokenResponseMeta
}

func (response *ShieldResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *ShieldResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response ShieldResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadataCommon.MintData, shardID byte, tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, ac *metadataCommon.AccumulatedValues, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not shieldRequest instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.IssuingUnifiedTokenRequestMeta)) {
			continue
		}

		tempInst := metadataCommon.NewInstruction()
		err := tempInst.FromStringSlice(inst)
		if err != nil {
			metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction: ", err)
			continue
		}

		if tempInst.Status != common.AcceptedStatusStr {
			continue
		}
		contentBytes, err := base64.StdEncoding.DecodeString(tempInst.Content)
		if err != nil {
			metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}
		acceptedContent := AcceptedInstShieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			continue
		}

		if !bytes.Equal(response.RequestedTxID[:], acceptedContent.TxReqID[:]) || shardID != tempInst.ShardID {
			continue
		}

		shieldAmt := uint64(0)
		reward := uint64(0)
		for i, data := range acceptedContent.Data {
			if !bytes.Equal(data.UniqTx, response.Data[i].UniqTx) {
				return false, fmt.Errorf("expect uniqTx %v but get %v", data.UniqTx, response.Data[i].UniqTx)
			}
			if !bytes.Equal(data.ExternalTokenID, response.Data[i].ExternalTokenID) {
				return false, fmt.Errorf("expect externalTokenID %v but get %v", data.ExternalTokenID, response.Data[i].ExternalTokenID)
			}
			if !data.IncTokenID.IsEqual(&response.Data[i].IncTokenID) {
				return false, fmt.Errorf("expect IncTokenID %v but get %v", data.IncTokenID, response.Data[i].IncTokenID)
			}
			shieldAmt += data.ShieldAmount
			reward += data.Reward
		}
		mintAmtFromInst := shieldAmt + reward

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != acceptedContent.UnifiedTokenID.String() {
			return false, fmt.Errorf("Invalid coinID")
		}
		if ok := mintCoin.CheckCoinValid(acceptedContent.Receiver, response.SharedRandom, mintAmtFromInst); !ok {
			return false, fmt.Errorf("Invalid coin")
		}

		idx = i
		break
	}
	if idx == -1 { // not found the issuance request inst for this response
		return false, errors.New(fmt.Sprintf("no shield request inst found for ShieldResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (response *ShieldResponse) SetSharedRandom(r []byte) {
	response.SharedRandom = r
}
