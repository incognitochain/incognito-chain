package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalTopUpWaitingPortingResponse struct {
	basemeta.MetadataBase
	DepositStatus string
	ReqTxID       common.Hash
}

func NewPortalTopUpWaitingPortingResponse(
	depositStatus string,
	reqTxID common.Hash,
	metaType int,
) *PortalTopUpWaitingPortingResponse {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}

	return &PortalTopUpWaitingPortingResponse{
		DepositStatus: depositStatus,
		ReqTxID:       reqTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes PortalTopUpWaitingPortingResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalTopUpWaitingPortingResponse) ValidateTxWithBlockChain(txr basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalTopUpWaitingPortingResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalTopUpWaitingPortingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == basemeta.PortalTopUpWaitingPortingResponseMeta
}

func (iRes PortalTopUpWaitingPortingResponse) Hash() *common.Hash {
	record := iRes.DepositStatus
	record += iRes.ReqTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalTopUpWaitingPortingResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func (iRes PortalTopUpWaitingPortingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []basemeta.Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever,
	ac *basemeta.AccumulatedValues,
	shardViewRetriever basemeta.ShardViewRetriever,
	beaconViewRetriever basemeta.BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PortalTopUpWaitingPorting response instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(basemeta.PortalTopUpWaitingPortingRequestMeta) {
			continue
		}
		instDepositStatus := inst[2]
		if instDepositStatus != iRes.DepositStatus ||
			(instDepositStatus != common.PortalTopUpWaitingPortingRejectedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var depositorAddrStrFromInst string
		var depositedAmountFromInst uint64

		contentBytes := []byte(inst[3])
		var topUpWaitingPortingReqContent PortalTopUpWaitingPortingRequestContent
		err := json.Unmarshal(contentBytes, &topUpWaitingPortingReqContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal top up waiting porting request content: ", err)
			continue
		}
		shardIDFromInst = topUpWaitingPortingReqContent.ShardID
		txReqIDFromInst = topUpWaitingPortingReqContent.TxReqID
		depositorAddrStrFromInst = topUpWaitingPortingReqContent.IncogAddressStr
		depositedAmountFromInst = topUpWaitingPortingReqContent.DepositedAmount

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(depositorAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occurred while deserializing custodian address string: ", err)
			continue
		}

		// collateral must be PRV
		PRVIDStr := common.PRVCoinID.String()
		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			depositedAmountFromInst != paidAmount ||
			PRVIDStr != assetID.String() {
			continue
		}
		idx = i
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalTopUpWaitingPortingRequestMeta instruction found for PortalTopUpWaitingPortingResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
