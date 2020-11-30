package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"

	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PDETradeResponse struct {
	basemeta.MetadataBase
	TradeStatus   string
	RequestedTxID common.Hash
}

func NewPDETradeResponse(
	tradeStatus string,
	requestedTxID common.Hash,
	metaType int,
) *PDETradeResponse {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	return &PDETradeResponse{
		TradeStatus:   tradeStatus,
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes PDETradeResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PDETradeResponse) ValidateTxWithBlockChain(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PDETradeResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PDETradeResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == basemeta.PDETradeResponseMeta
}

func (iRes PDETradeResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.TradeStatus
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PDETradeResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func (iRes PDETradeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []basemeta.Transaction, txsUsed []int, insts [][]string, instUsed []int, shardID byte, tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, ac *basemeta.AccumulatedValues, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PDETradeRequest or PDECrossPoolTradeRequestMeta instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			(instMetaType != strconv.Itoa(basemeta.PDETradeRequestMeta) && instMetaType != strconv.Itoa(basemeta.PDECrossPoolTradeRequestMeta)) {
			continue
		}
		instTradeStatus := inst[2]
		if instTradeStatus != iRes.TradeStatus || (instTradeStatus != common.PDETradeRefundChainStatus && instTradeStatus != common.PDETradeAcceptedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var receiverAddrStrFromInst string
		var receivingAmtFromInst uint64
		var receivingTokenIDStr string
		if instTradeStatus == common.PDETradeRefundChainStatus {
			contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
			if err != nil {
				Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				continue
			}
			var pdeTradeRequestAction PDETradeRequestAction
			err = json.Unmarshal(contentBytes, &pdeTradeRequestAction)
			if err != nil {
				Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				continue
			}
			shardIDFromInst = pdeTradeRequestAction.ShardID
			txReqIDFromInst = pdeTradeRequestAction.TxReqID
			receiverAddrStrFromInst = pdeTradeRequestAction.Meta.TraderAddressStr
			receivingTokenIDStr = pdeTradeRequestAction.Meta.TokenIDToSellStr
			receivingAmtFromInst = pdeTradeRequestAction.Meta.SellAmount + pdeTradeRequestAction.Meta.TradingFee
		} else { // trade accepted
			contentBytes := []byte(inst[3])
			var pdeTradeAcceptedContent PDETradeAcceptedContent
			err := json.Unmarshal(contentBytes, &pdeTradeAcceptedContent)
			if err != nil {
				Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
				continue
			}
			shardIDFromInst = pdeTradeAcceptedContent.ShardID
			txReqIDFromInst = pdeTradeAcceptedContent.RequestedTxID
			receiverAddrStrFromInst = pdeTradeAcceptedContent.TraderAddressStr
			receivingTokenIDStr = pdeTradeAcceptedContent.TokenIDToBuyStr
			receivingAmtFromInst = pdeTradeAcceptedContent.ReceiveAmount
		}

		if !bytes.Equal(iRes.RequestedTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(receiverAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}
		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			receivingAmtFromInst != paidAmount ||
			receivingTokenIDStr != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PDETradeRequest or PDECrossPoolTradeRequestMeta tx found for PDETradeResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
