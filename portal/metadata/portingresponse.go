package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"
	pCommon "github.com/incognitochain/incognito-chain/portal/common"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalFeeRefundResponse struct {
	basemeta.MetadataBase
	PortingRequestStatus string
	ReqTxID              common.Hash
}

func NewPortalFeeRefundResponse(
	portingRequestStatus string,
	reqTxID common.Hash,
	metaType int,
) *PortalFeeRefundResponse {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	return &PortalFeeRefundResponse{
		PortingRequestStatus: portingRequestStatus,
		ReqTxID:              reqTxID,
		MetadataBase:         metadataBase,
	}
}

func (iRes PortalFeeRefundResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalFeeRefundResponse) ValidateTxWithBlockChain(txr basemeta.Transaction,
	chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalFeeRefundResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, txr basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalFeeRefundResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == basemeta.PortalPortingResponseMeta
}

func (iRes PortalFeeRefundResponse) Hash() *common.Hash {
	record := iRes.PortingRequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalFeeRefundResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func parsePortingRequest(contentBytes []byte, shardID string) (string, common.Hash, string, uint64, error) {
	var portalPortingRequestContent PortalPortingRequestContent
	err := json.Unmarshal(contentBytes, &portalPortingRequestContent)
	if err != nil {
		Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal porting request content: ", err)
		return "", common.Hash{}, "", uint64(0), err
	}
	return shardID, portalPortingRequestContent.TxReqID, portalPortingRequestContent.IncogAddressStr, portalPortingRequestContent.PortingFee, nil
}

func parseValuesFromInst(inst []string) (string, common.Hash, string, uint64, error) {
	shardIDStr := inst[1]
	contentBytes := []byte(inst[3])
	return parsePortingRequest(contentBytes, shardIDStr)
}

func (iRes PortalFeeRefundResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
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
		if len(inst) < 4 { // this is not PortalFeeRefund response instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 || (instMetaType != strconv.Itoa(basemeta.PortalRequestPortingMeta) && instMetaType != strconv.Itoa(basemeta.PortalRequestPortingMetaV3)) {
			continue
		}
		status := inst[2]
		if status != iRes.PortingRequestStatus || status != pCommon.PortalRequestRejectedChainStatus {
			continue
		}

		shardIDFromInst, txReqIDFromInst, receiverAddrStrFromInst, portingFeeFromInst, err := parseValuesFromInst(inst)
		if err != nil {
			continue
		}

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			strconv.Itoa(int(shardID)) != shardIDFromInst {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(receiverAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing receiver address string: ", err)
			continue
		}

		// collateral must be PRV
		PRVIDStr := common.PRVCoinID.String()
		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			portingFeeFromInst != paidAmount ||
			PRVIDStr != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalFeeRefundRequest instruction found for PortalFeeRefundResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
