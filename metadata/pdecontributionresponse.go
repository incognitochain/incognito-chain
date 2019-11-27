package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PDEContributionResponse struct {
	MetadataBase
	ContributionStatus string
	RequestedTxID      common.Hash
}

func NewPDEContributionResponse(
	contributionStatus string,
	requestedTxID common.Hash,
	metaType int,
) *PDEContributionResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PDEContributionResponse{
		ContributionStatus: contributionStatus,
		RequestedTxID:      requestedTxID,
		MetadataBase:       metadataBase,
	}
}

func (iRes PDEContributionResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db incdb.Database) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PDEContributionResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db incdb.Database) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PDEContributionResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PDEContributionResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PDEContributionResponseMeta
}

func (iRes PDEContributionResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.ContributionStatus
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PDEContributionResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PDEContributionResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	ac *AccumulatedValues,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PDEContribution instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PDEContributionMeta) {
			continue
		}
		instContributionStatus := inst[2]
		if instContributionStatus != iRes.ContributionStatus || instContributionStatus != "refund" {
			continue
		}

		contentBytes := []byte(inst[3])
		var refundContribution PDERefundContribution
		err := json.Unmarshal(contentBytes, &refundContribution)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing refund contribution content: ", err)
			continue
		}
		shardIDFromInst := refundContribution.ShardID
		txReqIDFromInst := refundContribution.TxReqID
		receiverAddrStrFromInst := refundContribution.ContributorAddressStr
		receivingTokenIDStr := refundContribution.TokenIDStr
		receivingAmtFromInst := refundContribution.ContributedAmount

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
		return false, fmt.Errorf(fmt.Sprintf("no PDEContribution instruction found for PDEContributionResponse tx %s", tx.Hash().String()))
	}
	instUsed[idx] = 1
	return true, nil
}
