package metadata

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type PDEWithdrawalWithPRVFeeResponse struct {
	MetadataBase
	RequestedTxID common.Hash
	TokenIDStr    string
	WithdrawalStatus string
}

func NewPDEWithdrawalWithPRVFeeResponse(
	withdrawalStatus string,
	tokenIDStr string,
	requestedTxID common.Hash,
	metaType int,
) *PDEWithdrawalWithPRVFeeResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PDEWithdrawalWithPRVFeeResponse{
		WithdrawalStatus: withdrawalStatus,
		RequestedTxID: requestedTxID,
		TokenIDStr:    tokenIDStr,
		MetadataBase:  metadataBase,
	}
}

func (iRes PDEWithdrawalWithPRVFeeResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PDEWithdrawalWithPRVFeeResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PDEWithdrawalWithPRVFeeResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PDEWithdrawalWithPRVFeeResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PDEWithdrawalWithPRVFeeResponseMeta
}

func (iRes PDEWithdrawalWithPRVFeeResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.TokenIDStr
	record += iRes.WithdrawalStatus
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PDEWithdrawalWithPRVFeeResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PDEWithdrawalWithPRVFeeResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
txsInBlock []Transaction,
	txsUsed []int,
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	chainRetriever ChainRetriever,
	ac *AccumulatedValues,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PDEWithdrawalRequest instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PDEWithdrawalWithPRVFeeRequestMeta) {
			continue
		}
		instWithdrawalStatus := inst[2]
		if instWithdrawalStatus != iRes.WithdrawalStatus ||
			(instWithdrawalStatus != common.PDEWithdrawalOnFeeAcceptedChainStatus &&
			instWithdrawalStatus != common.PDEWithdrawalOnPoolPairAcceptedChainStatus) {
			continue
		}

		contentBytes := []byte(inst[3])
		var withdrawalAcceptedContent PDEWithdrawalAcceptedContent
		err := json.Unmarshal(contentBytes, &withdrawalAcceptedContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], withdrawalAcceptedContent.TxReqID[:]) ||
			shardID != withdrawalAcceptedContent.ShardID {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(withdrawalAcceptedContent.WithdrawerAddressStr)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing withdrawer address string: ", err)
			continue
		}

		_, pk, amount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			withdrawalAcceptedContent.DeductingPoolValue != amount ||
			withdrawalAcceptedContent.WithdrawalTokenIDStr != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no PDEWithdrawalRequest tx found for the PDEWithdrawalWithPRVFeeResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}
