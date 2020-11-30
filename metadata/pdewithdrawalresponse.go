package metadata

import (
	"bytes"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/basemeta"

	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type PDEWithdrawalResponse struct {
	basemeta.MetadataBase
	RequestedTxID common.Hash
	TokenIDStr    string
}

func NewPDEWithdrawalResponse(
	tokenIDStr string,
	requestedTxID common.Hash,
	metaType int,
) *PDEWithdrawalResponse {
	metadataBase := basemeta.MetadataBase{
		Type: metaType,
	}
	return &PDEWithdrawalResponse{
		RequestedTxID: requestedTxID,
		TokenIDStr:    tokenIDStr,
		MetadataBase:  metadataBase,
	}
}

func (iRes PDEWithdrawalResponse) CheckTransactionFee(tr basemeta.Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PDEWithdrawalResponse) ValidateTxWithBlockChain(tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PDEWithdrawalResponse) ValidateSanityData(chainRetriever basemeta.ChainRetriever, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever, beaconHeight uint64, tx basemeta.Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PDEWithdrawalResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == basemeta.PDEWithdrawalResponseMeta
}

func (iRes PDEWithdrawalResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.TokenIDStr
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PDEWithdrawalResponse) CalculateSize() uint64 {
	return basemeta.CalculateSize(iRes)
}

func (iRes PDEWithdrawalResponse) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []basemeta.Transaction, txsUsed []int, insts [][]string, instUsed []int, shardID byte, tx basemeta.Transaction, chainRetriever basemeta.ChainRetriever, ac *basemeta.AccumulatedValues, shardViewRetriever basemeta.ShardViewRetriever, beaconViewRetriever basemeta.BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PDEWithdrawalRequest instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(basemeta.PDEWithdrawalRequestMeta) {
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
		return false, errors.Errorf("no PDEWithdrawalRequest tx found for the PDEWithdrawalResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}
