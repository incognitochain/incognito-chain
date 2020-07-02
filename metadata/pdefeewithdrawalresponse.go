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

type PDEFeeWithdrawalResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewPDEFeeWithdrawalResponse(
	requestedTxID common.Hash,
	metaType int,
) *PDEFeeWithdrawalResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PDEFeeWithdrawalResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes PDEFeeWithdrawalResponse) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PDEFeeWithdrawalResponse) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PDEFeeWithdrawalResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PDEFeeWithdrawalResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PDEFeeWithdrawalResponseMeta
}

func (iRes PDEFeeWithdrawalResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PDEFeeWithdrawalResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PDEFeeWithdrawalResponse) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []Transaction, txsUsed []int, insts [][]string, instUsed []int, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1
	for i, inst := range insts {
		if len(inst) < 4 { // this is not PDEFeeWithdrawalRequest instruction
			continue
		}
		instMetaType := inst[0]
		if instUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PDEFeeWithdrawalRequestMeta) {
			continue
		}

		contentBytes := []byte(inst[3])
		var feeWithdrawalRequestAction PDEFeeWithdrawalRequestAction
		err := json.Unmarshal(contentBytes, &feeWithdrawalRequestAction)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing instruction content: ", err)
			continue
		}

		if !bytes.Equal(iRes.RequestedTxID[:], feeWithdrawalRequestAction.TxReqID[:]) ||
			shardID != feeWithdrawalRequestAction.ShardID {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(feeWithdrawalRequestAction.Meta.WithdrawerAddressStr)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occured while deserializing withdrawer address string: ", err)
			continue
		}

		_, pk, amount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			feeWithdrawalRequestAction.Meta.WithdrawalFeeAmt != amount ||
			common.PRVCoinID.String() != assetID.String() {
			continue
		}
		idx = i
		break
	}
	if idx == -1 { // not found the issuance request tx for this response
		return false, errors.Errorf("no PDEFeeWithdrawalRequest tx found for the PDEFeeWithdrawalResponse tx %s", tx.Hash().String())
	}
	instUsed[idx] = 1
	return true, nil
}
