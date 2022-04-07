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

type ShieldResponseData struct {
	ExternalTokenID []byte `json:"ExternalTokenID"`
	UniqTx          []byte `json:"UniqETHTx"`
	NetworkID       uint   `json:"NetworkID"`
}

type ShieldResponse struct {
	metadataCommon.MetadataBase
	RequestedTxID common.Hash          `json:"RequestedTxID"`
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
	metaType int, data []ShieldResponseData, requestedTxID common.Hash, shardRandom []byte,
) *ShieldResponse {
	return &ShieldResponse{
		Data: data,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
		SharedRandom:  shardRandom,
		RequestedTxID: requestedTxID,
	}
}

func (response *ShieldResponse) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (response *ShieldResponse) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {

	return true, true, nil
}

func (response *ShieldResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.IssuingUnifiedTokenResponseMeta || response.Type == metadataCommon.IssuingUnifiedRewardResponseMeta
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
		if mintData.InstsUsed[i] > 0 || (instMetaType != strconv.Itoa(metadataCommon.IssuingUnifiedTokenRequestMeta) && instMetaType != strconv.Itoa(metadataCommon.IssuingUnifiedRewardResponseMeta)) {
			continue
		}

		tempInst := metadataCommon.NewInstruction()
		err := tempInst.FromStringSlice(inst)
		if err != nil {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var address privacy.PaymentAddress
		var receivingAmtFromInst uint64
		var receivingTokenID common.Hash

		if tempInst.Status == common.AcceptedStatusStr {
			contentBytes, err := base64.StdEncoding.DecodeString(tempInst.Content)
			if err != nil {
				return false, err
			}
			acceptedContent := AcceptedShieldRequest{}
			err = json.Unmarshal(contentBytes, &acceptedContent)
			if err != nil {
				return false, err
			}
			shardIDFromInst = tempInst.ShardID
			txReqIDFromInst = acceptedContent.TxReqID
			receivingTokenID = acceptedContent.TokenID
			address = acceptedContent.Receiver
			for index, data := range acceptedContent.Data {
				if !bytes.Equal(data.UniqTx, response.Data[index].UniqTx) {
					continue
				}
				if !bytes.Equal(data.ExternalTokenID, response.Data[index].ExternalTokenID) {
					continue
				}
				if data.NetworkID != response.Data[index].NetworkID {
					continue
				}
				receivingAmtFromInst += data.IssuingAmount
			}
		} else {
			continue
		}
		if !bytes.Equal(response.RequestedTxID[:], txReqIDFromInst[:]) || shardID != shardIDFromInst {
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted || coinID.String() != receivingTokenID.String() {
			continue
		}
		if ok := mintCoin.CheckCoinValid(address, response.SharedRandom, receivingAmtFromInst); !ok {
			continue
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
