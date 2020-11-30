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

type PDEContributionResponse struct {
	basemeta.MetadataBase
	ContributionStatus string
	RequestedTxID      common.Hash
	TokenIDStr         string
}

func NewPDEContributionResponse(
	contributionStatus string,
	requestedTxID common.Hash,
	tokenIDStr string,
	metaType int,
) *PDEContributionResponse {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	return &PDEContributionResponse{
		ContributionStatus: contributionStatus,
		RequestedTxID:      requestedTxID,
		TokenIDStr:         tokenIDStr,
		MetadataBase:       metadataBase,
	}
}

func (iRes PDEContributionResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PDEContributionResponse) ValidateTxWithBlockChain(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PDEContributionResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PDEContributionResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == basemeta.PDEContributionResponseMeta
}

func (iRes PDEContributionResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.TokenIDStr
	record += iRes.ContributionStatus
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PDEContributionResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func (iRes PDEContributionResponse) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []basemeta.Transaction, txsUsed []int, insts [][]string, instUsed []int, shardID byte, tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, ac *basemeta.AccumulatedValues, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PDEContribution or PDEPRVRequiredContributionRequestMeta instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			(instMetaType != strconv.Itoa(basemeta.PDEContributionMeta) && instMetaType != strconv.Itoa(basemeta.PDEPRVRequiredContributionRequestMeta)) {
			continue
		}
		instContributionStatus := inst[2]
		if instContributionStatus != iRes.ContributionStatus || (instContributionStatus != common.PDEContributionRefundChainStatus && instContributionStatus != common.PDEContributionMatchedNReturnedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var receiverAddrStrFromInst string
		var receivingAmtFromInst uint64
		var receivingTokenIDStr string

		if instContributionStatus == common.PDEContributionRefundChainStatus {
			contentBytes := []byte(inst[3])
			var refundContribution PDERefundContribution
			err := json.Unmarshal(contentBytes, &refundContribution)
			if err != nil {
				Logger.log.Error("WARNING - VALIDATION: an error occured while parsing refund contribution content: ", err)
				continue
			}
			shardIDFromInst = refundContribution.ShardID
			txReqIDFromInst = refundContribution.TxReqID
			receiverAddrStrFromInst = refundContribution.ContributorAddressStr
			receivingTokenIDStr = refundContribution.TokenIDStr
			receivingAmtFromInst = refundContribution.ContributedAmount

		} else { // matched and returned
			contentBytes := []byte(inst[3])
			var matchedNReturnedContrib PDEMatchedNReturnedContribution
			err := json.Unmarshal(contentBytes, &matchedNReturnedContrib)
			if err != nil {
				Logger.log.Error("WARNING - VALIDATION: an error occured while parsing matched and returned contribution content: ", err)
				continue
			}
			shardIDFromInst = matchedNReturnedContrib.ShardID
			txReqIDFromInst = matchedNReturnedContrib.TxReqID
			receiverAddrStrFromInst = matchedNReturnedContrib.ContributorAddressStr
			receivingTokenIDStr = matchedNReturnedContrib.TokenIDStr
			receivingAmtFromInst = matchedNReturnedContrib.ReturnedContributedAmount
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
		return false, fmt.Errorf(fmt.Sprintf("no PDEContribution or PDEPRVRequiredContributionRequestMeta instruction found for PDEContributionResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
