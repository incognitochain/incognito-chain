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

type AddLiquidityResponse struct {
	metadataCommon.MetadataBase
	status  string
	txReqID string
}

func NewAddLiquidityResponse() *AddLiquidityResponse {
	return &AddLiquidityResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3AddLiquidityResponseMeta,
		},
	}
}

func NewAddLiquidityResponseWithValue(
	status, txReqID string,
) *AddLiquidityResponse {
	return &AddLiquidityResponse{
		MetadataBase: metadataCommon.MetadataBase{
			Type: metadataCommon.Pdexv3AddLiquidityResponseMeta,
		},
		status:  status,
		txReqID: txReqID,
	}
}

func (response *AddLiquidityResponse) CheckTransactionFee(tx metadataCommon.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (response *AddLiquidityResponse) ValidateTxWithBlockChain(
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

func (response *AddLiquidityResponse) ValidateSanityData(
	chainRetriever metadataCommon.ChainRetriever,
	shardViewRetriever metadataCommon.ShardViewRetriever,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	beaconHeight uint64,
	tx metadataCommon.Transaction,
) (bool, bool, error) {
	if response.status == "" {
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

func (response *AddLiquidityResponse) ValidateMetadataByItself() bool {
	return response.Type == metadataCommon.Pdexv3AddLiquidityResponseMeta
}

func (response *AddLiquidityResponse) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(&response)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (response *AddLiquidityResponse) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(response)
}

func (response *AddLiquidityResponse) ToCompactBytes() ([]byte, error) {
	return metadataCommon.ToCompactBytes(response)
}

func (response *AddLiquidityResponse) MarshalJSON() ([]byte, error) {
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

func (response *AddLiquidityResponse) UnmarshalJSON(data []byte) error {
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

func (response *AddLiquidityResponse) TxReqID() string {
	return response.txReqID
}

func (response *AddLiquidityResponse) Status() string {
	return response.status
}

type RefundAddLiquidity struct {
	Contribution *statedb.Pdexv3ContributionState `json:"Contribution"`
}

type MatchAndReturnAddLiquidity struct {
	ShareAmount              uint64                           `json:"ShareAmount"`
	Contribution             *statedb.Pdexv3ContributionState `json:"Contribution"`
	ReturnAmount             uint64                           `json:"ReturnAmount"`
	ExistedTokenActualAmount uint64                           `json:"ExistedTokenActualAmount"`
	ExistedTokenReturnAmount uint64                           `json:"ExistedTokenReturnAmount"`
	ExistedTokenID           common.Hash                      `json:"ExistedTokenID"`
	NftID                    common.Hash                      `json:"NftID"`
}

func (response *AddLiquidityResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if len(inst) != 3 { // this is not PDEContribution instruction
			continue
		}
		metadataCommon.Logger.Log.Infof("BUGLOG currently processing inst: %v\n", inst)
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 || instMetaType != strconv.Itoa(metadataCommon.Pdexv3AddLiquidityRequestMeta) {
			continue
		}
		instContributionStatus := inst[1]
		if instContributionStatus != response.status ||
			(instContributionStatus != common.PDEContributionRefundChainStatus && instContributionStatus != common.PDEContributionMatchedNReturnedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var receiverAddrStrFromInst string
		var receivingAmtFromInst uint64
		var receivingTokenIDStr string

		switch instContributionStatus {
		case common.PDEContributionRefundChainStatus:
			contentBytes := []byte(inst[2])
			var refundAddLiquidity RefundAddLiquidity
			err := json.Unmarshal(contentBytes, &refundAddLiquidity)
			if err != nil {
				metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing refund contribution content: ", err)
				return false, err
			}
			contribution := refundAddLiquidity.Contribution
			value := contribution.Value()
			shardIDFromInst = value.ShardID()
			txReqIDFromInst = value.TxReqID()
			receiverAddrStrFromInst = value.OtaReceiver()
			receivingTokenIDStr = value.TokenID().String()
			receivingAmtFromInst = value.Amount()
		case common.PDEContributionMatchedNReturnedChainStatus:
			contentBytes := []byte(inst[2])
			var matchAndReturnAddLiquidity MatchAndReturnAddLiquidity
			err := json.Unmarshal(contentBytes, &matchAndReturnAddLiquidity)
			if err != nil {
				metadataCommon.Logger.Log.Error("WARNING - VALIDATION: an error occured while parsing matched and returned contribution content: ", err)
				return false, err
			}
			contribution := matchAndReturnAddLiquidity.Contribution
			value := contribution.Value()
			shardIDFromInst = value.ShardID()
			txReqIDFromInst = value.TxReqID()
			receiverAddrStrFromInst = value.OtaReceiver()
			receivingTokenIDStr = value.TokenID().String()
			receivingAmtFromInst = matchAndReturnAddLiquidity.ReturnAmount
		default:
			return false, errors.New("Not find status")
		}

		if response.TxReqID() != txReqIDFromInst.String() || shardID != shardIDFromInst {
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

		otaReceiver := coin.OTAReceiver{}
		err = otaReceiver.FromString(receiverAddrStrFromInst)
		if err != nil {
			return false, errors.New("Invalid ota receiver")
		}

		txR := mintCoin.(*coin.CoinV2).GetTxRandom()
		if !bytes.Equal(otaReceiver.PublicKey.ToBytesS(), pk[:]) ||
			receivingAmtFromInst != paidAmount ||
			!bytes.Equal(txR[:], otaReceiver.TxRandom[:]) ||
			receivingTokenIDStr != coinID.String() {
			return false, errors.New("Coin is invalid")
		}
		idx = i
		fmt.Println("BUGLOG Verify Metadata --- OK")
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		metadataCommon.Logger.Log.Debugf("no Pdexv3 addliquidity instruction tx %s", tx.Hash().String())
		return false, fmt.Errorf(fmt.Sprintf("no Pdexv3 addliquidity instruction tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}
