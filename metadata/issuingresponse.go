package metadata

import (
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

type IssuingResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewIssuingResponse(requestedTxID common.Hash, metaType int) *IssuingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &IssuingResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes *IssuingResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}

func (iRes *IssuingResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via RequestedTxID) in current block
	return false, nil
}

func (iRes *IssuingResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes *IssuingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (iRes *IssuingResponse) Hash() *common.Hash {
	record := iRes.RequestedTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *IssuingResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes *IssuingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	fmt.Printf("[db] verifying issuing response tx\n")
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 ||
			inst[0] != strconv.Itoa(IssuingRequestMeta) ||
			inst[1] != strconv.Itoa(int(shardID)) ||
			inst[2] != "accepted" {
			continue
		}
		issuingInfo, err := component.ParseIssuingInfo(inst[3])
		if err != nil {
			continue
		}
		unique, pk, amount, assetID := tx.GetTransferData()
		txData := &component.IssuingInfo{
			ReceiverAddress: privacy.PaymentAddress{Pk: pk},
			Amount:          amount,
			TokenID:         *assetID,
		}

		if unique && txData.Compare(issuingInfo) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false, errors.Errorf("no instruction found for IssuingResponse tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return true, nil
}
